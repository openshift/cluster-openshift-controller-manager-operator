package operator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corelistersv1 "k8s.io/client-go/listers/core/v1"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	operatorconfigclientv1alpha1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned/typed/openshiftcontrollermanager/v1alpha1"
	operatorconfiginformerv1alpha1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/informers/externalversions/openshiftcontrollermanager/v1alpha1"
	"k8s.io/apimachinery/pkg/util/diff"
)

type Listers struct {
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
	configSynced         cache.InformerSynced
}

func NewConfigObserver(
	operatorConfigInformer operatorconfiginformerv1alpha1.OpenShiftControllerManagerOperatorConfigInformer,
	operatorConfigClient operatorconfigclientv1alpha1.OpenshiftcontrollermanagerV1alpha1Interface,
	kubeInformersForOpenshiftCoreOperators informers.SharedInformerFactory,
	configInformer configinformers.SharedInformerFactory,
) *ConfigObserver {
	c := &ConfigObserver{
		operatorConfigClient: operatorConfigClient.OpenShiftControllerManagerOperatorConfigs(),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ConfigObserver"),

		rateLimiter: flowcontrol.NewTokenBucketRateLimiter(0.05 /*3 per minute*/, 4),
		observers: []observeConfigFunc{
			observeControllerManagerImagesConfig,
			observeInternalRegistryHostname,
		},
		listers: Listers{
			imageConfigLister: configInformer.Config().V1().Images().Lister(),
			configmapLister:   kubeInformersForOpenshiftCoreOperators.Core().V1().ConfigMaps().Lister(),
		},
	}

	c.operatorConfigSynced = operatorConfigInformer.Informer().HasSynced
	c.configmapSynced = kubeInformersForOpenshiftCoreOperators.Core().V1().ConfigMaps().Informer().HasSynced
	c.configSynced = configInformer.Config().V1().Images().Informer().HasSynced

	operatorConfigInformer.Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftCoreOperators.Core().V1().Namespaces().Informer().AddEventHandler(c.eventHandler())
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
		if val := controllerManagerImages.Data["builderImage"]; len(val) > 0 {
			unstructured.SetNestedField(observedConfig, val, "build", "imageTemplateFormat", "format")
		}
		if val := controllerManagerImages.Data["deployerImage"]; len(val) > 0 {
			unstructured.SetNestedField(observedConfig, val, "deployer", "imageTemplateFormat", "format")
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
	internalRegistryHostName := configImage.Status.InternalRegistryHostname
	if len(internalRegistryHostName) > 0 {
		unstructured.SetNestedField(observedConfig, internalRegistryHostName, "dockerPullSecret", "internalRegistryHostname")
	}
	return observedConfig, nil
}

func (c *ConfigObserver) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Infof("Starting ConfigObserver")
	defer glog.Infof("Shutting down ConfigObserver")

	cache.WaitForCacheSync(stopCh,
		c.operatorConfigSynced,
		c.configmapSynced,
		c.configSynced,
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
