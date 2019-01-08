package operator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/workqueue"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	operatorconfigclientv1alpha1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned/typed/openshiftcontrollermanager/v1alpha1"
	operatorconfiginformerv1alpha1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/informers/externalversions/openshiftcontrollermanager/v1alpha1"
)

type Listers struct {
	buildConfigLister configlistersv1.BuildLister
	imageConfigLister configlistersv1.ImageLister
	configmapLister   corelistersv1.ConfigMapLister
}

type observeConfigFunc func(Listers, map[string]interface{}) (map[string]interface{}, error)

type ConfigObserver struct {
	operatorConfigClient operatorconfigclientv1alpha1.OpenShiftControllerManagerOperatorConfigInterface

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface

	rateLimiter flowcontrol.RateLimiter
	observers   []observeConfigFunc

	// listers are used by config observers to retrieve necessary resources
	listers Listers

	operatorConfigSynced cache.InformerSynced
	configmapSynced      cache.InformerSynced
	imageConfigSynced    cache.InformerSynced
	buildConfigSynced    cache.InformerSynced
}

func NewConfigObserver(
	operatorConfigInformer operatorconfiginformerv1alpha1.OpenShiftControllerManagerOperatorConfigInformer,
	operatorConfigClient operatorconfigclientv1alpha1.OpenshiftcontrollermanagerV1alpha1Interface,
	kubeInformersForOperator informers.SharedInformerFactory,
	configInformer configinformers.SharedInformerFactory,
) *ConfigObserver {
	c := &ConfigObserver{
		operatorConfigClient: operatorConfigClient.OpenShiftControllerManagerOperatorConfigs(),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ConfigObserver"),

		rateLimiter: flowcontrol.NewTokenBucketRateLimiter(0.05 /*3 per minute*/, 4),
		observers: []observeConfigFunc{
			observeControllerManagerImagesConfig,
			observeInternalRegistryHostname,
			observeBuildControllerConfig,
		},
		listers: Listers{
			imageConfigLister: configInformer.Config().V1().Images().Lister(),
			configmapLister:   kubeInformersForOperator.Core().V1().ConfigMaps().Lister(),
			buildConfigLister: configInformer.Config().V1().Builds().Lister(),
		},
	}

	c.operatorConfigSynced = operatorConfigInformer.Informer().HasSynced
	c.configmapSynced = kubeInformersForOperator.Core().V1().ConfigMaps().Informer().HasSynced
	c.imageConfigSynced = configInformer.Config().V1().Images().Informer().HasSynced
	c.buildConfigSynced = configInformer.Config().V1().Builds().Informer().HasSynced

	operatorConfigInformer.Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOperator.Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	configInformer.Config().V1().Images().Informer().AddEventHandler(c.eventHandler())

	return c
}

// sync reacts to a change in controller manager images.
func (c ConfigObserver) sync() error {
	var err error
	observedConfig := map[string]interface{}{}

	for _, observer := range c.observers {
		observedConfig, err = observer(c.listers, observedConfig)
		if err != nil {
			return err
		}
	}

	operatorConfig, err := c.operatorConfigClient.Get("instance", metav1.GetOptions{})
	if err != nil {
		return err
	}

	// don't worry about errors
	currentConfig := map[string]interface{}{}
	json.NewDecoder(bytes.NewBuffer(operatorConfig.Spec.ObservedConfig.Raw)).Decode(&currentConfig)
	if reflect.DeepEqual(currentConfig, observedConfig) {
		return nil
	}

	glog.Infof("writing updated observedConfig: %v", diff.ObjectDiff(operatorConfig.Spec.ObservedConfig.Object, observedConfig))
	operatorConfig.Spec.ObservedConfig = runtime.RawExtension{Object: &unstructured.Unstructured{Object: observedConfig}}

	if _, err := c.operatorConfigClient.Update(operatorConfig); err != nil {
		return err
	}

	return nil
}

// observeControllerManagerImagesConfig observes image paths from openshift-controller-manager-images in order to determine which deployer and builder images to use
func observeControllerManagerImagesConfig(listers Listers, observedConfig map[string]interface{}) (map[string]interface{}, error) {
	controllerManagerImages, err := listers.configmapLister.ConfigMaps(operatorNamespaceName).Get("openshift-controller-manager-images")
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if controllerManagerImages != nil {
		// TODO(juanvallejo): reflect any issues in operator status
		if err = observeField(observedConfig, controllerManagerImages.Data["builderImage"], "build.imageTemplateFormat.format", true); err != nil {
			return nil, err
		}
		if err = observeField(observedConfig, controllerManagerImages.Data["deployerImage"], "deployer.imageTemplateFormat.format", true); err != nil {
			return nil, err
		}
	}

	return observedConfig, nil
}

// observeInternalRegistryHostname reads the internal registry hostname from the cluster configuration as provided by
// the registry operator.
func observeInternalRegistryHostname(listers Listers, observedConfig map[string]interface{}) (map[string]interface{}, error) {
	configImage, err := listers.imageConfigLister.Get("cluster")
	if errors.IsNotFound(err) {
		return observedConfig, nil
	}
	if err != nil {
		return nil, err
	}
	if err = observeField(observedConfig, configImage.Status.InternalRegistryHostname, "dockerPullSecret.internalRegistryHostname", true); err != nil {
		return nil, err
	}
	return observedConfig, nil
}

// observeBuildControllerConfig reads the cluster-wide build controller configuration as provided by the cluster admin.
func observeBuildControllerConfig(listers Listers, observedConfig map[string]interface{}) (map[string]interface{}, error) {
	build, err := listers.buildConfigLister.Get("cluster")
	if errors.IsNotFound(err) {
		return observedConfig, nil
	}
	if err != nil {
		return nil, err
	}
	// set build defaults

	if build.Spec.BuildDefaults.GitProxy != nil {
		if err = observeField(observedConfig, build.Spec.BuildDefaults.GitProxy.HTTPProxy, "build.buildDefaults.gitHTTPProxy", false); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildDefaults.gitHTTPProxy", err)
		}
		if err = observeField(observedConfig, build.Spec.BuildDefaults.GitProxy.HTTPSProxy, "build.buildDefaults.gitHTTPSProxy", false); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildDefaults.gitHTTPSProxy", err)
		}
		if err = observeField(observedConfig, build.Spec.BuildDefaults.GitProxy.NoProxy, "build.buildDefaults.gitNoProxy", false); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildDefaults.gitNoProxy", err)
		}
	}

	buildEnv := build.Spec.BuildDefaults.Env
	if len(buildEnv) > 0 {
		if err = observeField(observedConfig, buildEnv, "build.buildDefaults.env", true); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildDefaults.env", err)
		}
	}
	if len(build.Spec.BuildDefaults.ImageLabels) > 0 {
		if err = observeField(observedConfig, build.Spec.BuildDefaults.ImageLabels, "build.buildDefaults.imageLabels", true); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildDefaults.imageLabels", err)
		}
	}

	// set build overrides
	if len(build.Spec.BuildOverrides.ImageLabels) > 0 {
		if err = observeField(observedConfig, build.Spec.BuildOverrides.ImageLabels, "build.buildOverrides.imageLabels", true); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildOverrides.imageLabels", err)
		}
	}
	nodeSelector := build.Spec.BuildOverrides.NodeSelector
	if len(build.Spec.BuildOverrides.NodeSelector.MatchLabels) > 0 {
		if err = observeField(observedConfig, nodeSelector.MatchLabels, "build.buildOverrides.nodeSelector", true); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildOverrides.nodeSelector", err)
		}
	}
	// Control plane config does not support MatchExpressions yet
	if len(nodeSelector.MatchExpressions) > 0 {
		glog.Warningf("config.Build: %s is not supported", "buildOverrides.nodeSelector.matchExpressions")
	}
	if len(build.Spec.BuildOverrides.Tolerations) > 0 {
		if err = observeField(observedConfig, build.Spec.BuildOverrides.Tolerations, "build.buildOverrides.tolerations", true); err != nil {
			return nil, fmt.Errorf("failed to observe %s: %v", "build.buildOverrides.tolerations", err)
		}
	}

	// TODO: 1) generate CA bundle from ConfigMapRef
	//       2) write CA bundle to a ConfigMap in the controller-manager's namespace
	//       3) wire in logic to force a rollout if CA bundle changes
	// additionalCA := build.Spec.AdditionalTrustedCA
	// if len(additionalCA.Name) > 0 {
	// 	unstructured.SetNestedField(observedConfig, "/var/run/openshift.io/config/certs/additional-ca.crt", "build", "additionalTrustedCA")
	// }
	return observedConfig, nil
}

// observeField sets the nested fieldName's value in the provided observedConfig.
// If the provided value is nil, no value is set.
// If skipIfEmpty is true, the value
func observeField(observedConfig map[string]interface{}, val interface{}, fieldName string, skipIfEmpty bool) error {
	nestedFields := strings.Split(fieldName, ".")
	if val == nil {
		return nil
	}
	var err error
	switch v := val.(type) {
	case int64, bool:
		err = unstructured.SetNestedField(observedConfig, v, nestedFields...)
	case string:
		if skipIfEmpty && len(v) == 0 {
			return nil
		}
		err = unstructured.SetNestedField(observedConfig, v, nestedFields...)
	case []interface{}:
		if skipIfEmpty && len(v) == 0 {
			return nil
		}
		err = unstructured.SetNestedSlice(observedConfig, v, nestedFields...)
	case map[string]string:
		if skipIfEmpty && len(v) == 0 {
			return nil
		}
		err = unstructured.SetNestedStringMap(observedConfig, v, nestedFields...)
	case map[string]interface{}:
		if skipIfEmpty && len(v) == 0 {
			return nil
		}
		err = unstructured.SetNestedMap(observedConfig, v, nestedFields...)
	default:
		data, err := ConvertJSON(v)
		if err != nil {
			return err
		}
		if skipIfEmpty && data == nil {
			return nil
		}
		err = unstructured.SetNestedField(observedConfig, data, nestedFields...)
	}
	return err
}

// ConvertJSON returns the provided object's JSON-encoded representation. The object
// must support JSON serialization and deserialization.
func ConvertJSON(o interface{}) (interface{}, error) {
	if o == nil {
		return nil, nil
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(o); err != nil {
		return nil, err
	}
	ret := []interface{}{}
	if err := json.NewDecoder(buf).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *ConfigObserver) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Infof("Starting ConfigObserver")
	defer glog.Infof("Shutting down ConfigObserver")

	cache.WaitForCacheSync(stopCh,
		c.operatorConfigSynced,
		c.configmapSynced,
		c.imageConfigSynced,
		c.buildConfigSynced,
	)

	// doesn't matter what workers say, only start one.
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *ConfigObserver) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *ConfigObserver) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	// before we call sync, we want to wait for token.  We do this to avoid hot looping.
	c.rateLimiter.Accept()

	err := c.sync()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

// eventHandler queues the operator to check spec and status
func (c *ConfigObserver) eventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.queue.Add(workQueueKey) },
	}
}
