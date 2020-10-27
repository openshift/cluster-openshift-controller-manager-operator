package workload

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	operatorapiv1 "github.com/openshift/api/operator/v1"
	proxyclientv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcehash"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type manageConfigMapFn func(
	kubernetes.Interface,
	coreclientv1.ConfigMapsGetter,
	events.Recorder,
	*operatorapiv1.OpenShiftControllerManager,
) (*corev1.ConfigMap, bool, error)

type manageClientCAFn func(
	client coreclientv1.CoreV1Interface,
	recorder events.Recorder,
) (bool, error)

func manageOpenShiftControllerManagerClientCA_v311_00_to_latest(
	client coreclientv1.CoreV1Interface,
	recorder events.Recorder,
) (bool, error) {
	const apiserverClientCA = "client-ca"
	_, caChanged, err := resourceapply.SyncConfigMap(client, recorder, util.KubeAPIServerNamespace, apiserverClientCA, util.TargetNamespace, apiserverClientCA, []metav1.OwnerReference{})
	if err != nil {
		return false, err
	}
	return caChanged, nil
}

func manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(
	kubeClient kubernetes.Interface,
	client coreclientv1.ConfigMapsGetter,
	recorder events.Recorder,
	operatorConfig *operatorapiv1.OpenShiftControllerManager,
) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/cm.yaml"))
	defaultConfig := v311_00_assets.MustAsset("v3.11.0/config/defaultconfig.yaml")
	requiredConfigMap, _, err := resourcemerge.MergeConfigMap(configMap, "config.yaml", nil, defaultConfig, operatorConfig.Spec.UnsupportedConfigOverrides.Raw, operatorConfig.Spec.ObservedConfig.Raw)
	if err != nil {
		return nil, false, err
	}

	// we can embed input hashes on our main configmap to drive rollouts when they change.
	inputHashes, err := resourcehash.MultipleObjectHashStringMapForObjectReferences(
		kubeClient,
		resourcehash.NewObjectRef().ForConfigMap().InNamespace(util.TargetNamespace).Named("client-ca"),
		resourcehash.NewObjectRef().ForSecret().InNamespace(util.TargetNamespace).Named("serving-cert"),
		resourcehash.NewObjectRef().ForConfigMap().InNamespace(util.TargetNamespace).Named("openshift-global-ca"),
		resourcehash.NewObjectRef().ForConfigMap().InNamespace(util.TargetNamespace).Named("openshift-user-ca"),
	)
	if err != nil {
		return nil, false, err
	}
	for k, v := range inputHashes {
		requiredConfigMap.Data[k] = v
	}

	return resourceapply.ApplyConfigMap(client, recorder, requiredConfigMap)
}

func manageOpenShiftServiceCAConfigMap_v311_00_to_latest(
	kubeClient kubernetes.Interface,
	client coreclientv1.ConfigMapsGetter,
	recorder events.Recorder,
	operatorConfig *operatorapiv1.OpenShiftControllerManager,
) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/openshift-service-ca-cm.yaml"))
	existing, err := client.ConfigMaps(util.TargetNamespace).Get(context.TODO(), "openshift-service-ca", metav1.GetOptions{})
	// Ensure we create the ConfigMap for the registry CA, and that it has the right annotations
	// Lifted from library-go for the most part
	if apierrors.IsNotFound(err) {
		new, err := client.ConfigMaps(util.TargetNamespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
		if err != nil {
			recorder.Eventf("ConfigMapCreateFailed", "Failed to create %s%s/%s%s: %v", "configmap", "", "openshift-service-ca", "-n openshift-controller-manager", err)
			return nil, true, err
		}
		recorder.Eventf("ConfigMapCreated", "Created %s%s/%s%s because it was missing", "configmap", "", "openshift-service-ca", "-n openshift-controller-manager")
		return new, true, nil
	}

	// Ensure the openshift-service-ca ConfigMap has the service.beta.openshift.io/inject-cabundle annotation
	// Otherwise ignore the contents of the ConfigMap
	modified := resourcemerge.BoolPtr(false)
	existingCopy := existing.DeepCopy()
	resourcemerge.EnsureObjectMeta(modified, &existingCopy.ObjectMeta, configMap.ObjectMeta)
	if !*modified {
		return existing, false, nil
	}
	updated, err := client.ConfigMaps(util.TargetNamespace).Update(context.TODO(), existingCopy, metav1.UpdateOptions{})
	if err != nil {
		recorder.Eventf("ConfigMapUpdateFailed", "Failed to update %s%s/%s%s: %v", "configmap", "", "openshift-service-ca", "-n openshift-controller-manager", err)
		return nil, true, err
	}
	recorder.Eventf("ConfigMapUpdated", "Updated %s%s/%s%s", "configmap", "", "openshift-service-ca", "-n openshift-controller-manager")
	return updated, true, nil
}

func manageOpenShiftGlobalCAConfigMap_v311_00_to_latest(
	kubeClient kubernetes.Interface,
	client coreclientv1.ConfigMapsGetter,
	recorder events.Recorder,
	operatorConfig *operatorapiv1.OpenShiftControllerManager,
) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml"))
	existing, err := client.ConfigMaps(util.TargetNamespace).Get(context.TODO(), "openshift-global-ca", metav1.GetOptions{})
	// Ensure we create the ConfigMap for the global CA, and that it has the right labels
	// Lifted from library-go for the most part

	// Bug 1826183: Build pod containers now run `update-ca-trust extract` on startup if a custom
	// PKI is added to the cluster. Prior to 4.6, builds used the global CA trust bundle that was
	// injected into this global-ca configmap. However, the global CA bundle is not intended to be
	// used with workloads which run `update-ca-trust extract` on their own. Instead, this operator
	// will directly copy the admin-provided custom PKI via the UserCAObservationController.
	//
	// TODO: In 4.6 we need to continue creating this ConfigMap to ensure smooth upgrades.
	// In 4.7 or beyond this ConfigMap should be deleted.
	if apierrors.IsNotFound(err) {
		new, err := client.ConfigMaps(util.TargetNamespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
		if err != nil {
			recorder.Eventf("ConfigMapCreateFailed", "Failed to create %s%s/%s%s: %v", "configmap", "", "openshift-global-ca", "-n openshift-controller-manager", err)
			return nil, true, err
		}
		recorder.Eventf("ConfigMapCreated", "Created %s%s/%s%s because it was missing", "configmap", "", "openshift-global-ca", "-n openshift-controller-manager")
		return new, true, nil
	}

	// Ensure the openshift-global-ca ConfigMap has the config.openshift.io/inject-trusted-cabundle Label
	// Otherwise ignore the contents of the ConfigMap
	modified := resourcemerge.BoolPtr(false)
	existingCopy := existing.DeepCopy()
	resourcemerge.EnsureObjectMeta(modified, &existingCopy.ObjectMeta, configMap.ObjectMeta)
	if !*modified {
		return existing, false, nil
	}
	updated, err := client.ConfigMaps(util.TargetNamespace).Update(context.TODO(), existingCopy, metav1.UpdateOptions{})
	if err != nil {
		recorder.Eventf("ConfigMapUpdateFailed", "Failed to update %s%s/%s%s: %v", "configmap", "", "openshift-global-ca", "-n openshift-controller-manager", err)
		return nil, true, err
	}
	recorder.Eventf("ConfigMapUpdated", "Updated %s%s/%s%s", "configmap", "", "openshift-global-ca", "-n openshift-controller-manager")
	return updated, true, nil
}

func manageOpenShiftControllerManagerDeployment_v311_00_to_latest(
	client appsclientv1.DaemonSetsGetter,
	recorder events.Recorder,
	options *operatorapiv1.OpenShiftControllerManager,
	imagePullSpec string,
	generationStatus []operatorapiv1.GenerationStatus,
	forceRollout bool,
	proxyLister proxyclientv1.ProxyLister,
) (*appsv1.DaemonSet, bool, error) {
	required := resourceread.ReadDaemonSetV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/ds.yaml"))

	if len(imagePullSpec) > 0 {
		required.Spec.Template.Spec.Containers[0].Image = imagePullSpec
	}

	level := 2
	switch options.Spec.LogLevel {
	case operatorapiv1.TraceAll:
		level = 8
	case operatorapiv1.Trace:
		level = 6
	case operatorapiv1.Debug:
		level = 4
	case operatorapiv1.Normal:
		level = 2
	}
	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", level))
	if required.Annotations == nil {
		required.Annotations = map[string]string{}
	}
	required.Annotations[util.VersionAnnotation] = os.Getenv("RELEASE_VERSION")

	proxyCfg, err := proxyLister.Get("cluster")
	if err != nil {
		recorder.Eventf("ProxyConfigGetFailed", "Error retrieving global proxy config: %s", err.Error())
		if !apierrors.IsNotFound(err) {
			// return daemonset since it is still referenced by caller even with errors
			return required, false, err
		}
	} else {
		for i, c := range required.Spec.Template.Spec.Containers {
			newEnvs := []corev1.EnvVar{}

			if len(c.Env) == 0 {
				if len(proxyCfg.Status.NoProxy) > 0 {
					newEnvs = append(newEnvs, corev1.EnvVar{Name: "NO_PROXY", Value: proxyCfg.Status.NoProxy})
				}
				if len(proxyCfg.Status.HTTPProxy) > 0 {
					newEnvs = append(newEnvs, corev1.EnvVar{Name: "HTTP_PROXY", Value: proxyCfg.Status.HTTPProxy})
				}
				if len(proxyCfg.Status.HTTPSProxy) > 0 {
					newEnvs = append(newEnvs, corev1.EnvVar{Name: "HTTPS_PROXY", Value: proxyCfg.Status.HTTPSProxy})
				}
			}

			for _, env := range c.Env {
				name := strings.TrimSpace(env.Name)
				switch name {
				case "HTTPS_PROXY":
					if len(proxyCfg.Status.HTTPSProxy) == 0 {
						continue
					}
					env.Value = proxyCfg.Status.HTTPSProxy

				case "HTTP_PROXY":
					if len(proxyCfg.Status.HTTPProxy) == 0 {
						continue
					}
					env.Value = proxyCfg.Status.HTTPProxy

				case "NO_PROXY":
					if len(proxyCfg.Status.NoProxy) == 0 {
						continue
					}
					env.Value = proxyCfg.Status.NoProxy

				}
				newEnvs = append(newEnvs, env)
			}
			// reflect.DeepEqual does not consider this case equal
			envsEqual := c.Env == nil && len(newEnvs) == 0
			envsEqual = envsEqual || !reflect.DeepEqual(newEnvs, c.Env)
			forceRollout = forceRollout || !envsEqual
			required.Spec.Template.Spec.Containers[i].Env = newEnvs
		}
	}

	return resourceapply.ApplyDaemonSetWithForce(client, recorder, required, resourcemerge.ExpectedDaemonSetGeneration(required, generationStatus), forceRollout)
}
