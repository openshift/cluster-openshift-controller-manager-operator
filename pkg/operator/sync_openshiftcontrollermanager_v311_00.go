package operator

import (
	"fmt"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/apis/openshiftcontrollermanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	operatorsv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

// syncOpenShiftControllerManager_v311_00_to_latest takes care of synchronizing (not upgrading) the thing we're managing.
// most of the time the sync method will be good for a large span of minor versions
func syncOpenShiftControllerManager_v311_00_to_latest(c OpenShiftControllerManagerOperator, operatorConfig *v1alpha1.OpenShiftControllerManagerOperatorConfig, previousAvailability *operatorsv1alpha1.VersionAvailablity) (operatorsv1alpha1.VersionAvailablity, []error) {
	versionAvailability := operatorsv1alpha1.VersionAvailablity{
		Version: operatorConfig.Spec.Version,
	}

	errors := []error{}
	var err error

	directResourceResults := resourceapply.ApplyDirectly(c.kubeClient, v311_00_assets.Asset,
		"v3.11.0/openshift-controller-manager/ns.yaml",
		"v3.11.0/openshift-controller-manager/informer-clusterrole.yaml",
		"v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml",
		"v3.11.0/openshift-controller-manager/public-info-role.yaml",
		"v3.11.0/openshift-controller-manager/public-info-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/leader-role.yaml",
		"v3.11.0/openshift-controller-manager/leader-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/separate-sa-role.yaml",
		"v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/sa.yaml",
		"v3.11.0/openshift-controller-manager/svc.yaml",
	)
	resourcesThatForceRedeployment := sets.NewString("v3.11.0/openshift-controller-manager/sa.yaml")
	forceDeployment := false

	for _, currResult := range directResourceResults {
		if currResult.Error != nil {
			errors = append(errors, fmt.Errorf("%q (%T): %v", currResult.File, currResult.Type, currResult.Error))
			continue
		}

		if currResult.Changed && resourcesThatForceRedeployment.Has(currResult.File) {
			forceDeployment = true
		}
	}

	controllerManagerConfig, configMapModified, err := manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(c.kubeClient.CoreV1(), operatorConfig)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "configmap", err))
	}
	// the kube-apiserver is the source of truth for client CA bundles
	clientCAModified, err := manageOpenShiftAPIServerClientCA_v311_00_to_latest(c.kubeClient.CoreV1())
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "client-ca", err))
	}

	forceDeployment = forceDeployment || operatorConfig.ObjectMeta.Generation != operatorConfig.Status.ObservedGeneration
	forceDeployment = forceDeployment || configMapModified || clientCAModified

	// our configmaps and secrets are in order, now it is time to create the DS
	// TODO check basic preconditions here
	actualDeployment, _, err := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(c.kubeClient.AppsV1(), operatorConfig, previousAvailability, forceDeployment)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "deployment", err))
	}

	configData := ""
	if controllerManagerConfig != nil {
		configData = controllerManagerConfig.Data["config.yaml"]
	}
	_, _, err = manageOpenShiftControllerManagerPublicConfigMap_v311_00_to_latest(c.kubeClient.CoreV1(), configData, operatorConfig)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "configmap/public-info", err))
	}

	return resourcemerge.ApplyGenerationAvailability(versionAvailability, actualDeployment, errors...), errors
}

func manageOpenShiftAPIServerClientCA_v311_00_to_latest(client coreclientv1.CoreV1Interface) (bool, error) {
	const apiserverClientCA = "client-ca"
	_, caChanged, err := resourceapply.SyncConfigMap(client, kubeAPIServerNamespaceName, apiserverClientCA, targetNamespaceName, apiserverClientCA)
	if err != nil {
		return false, err
	}
	return caChanged, nil
}

func manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(client coreclientv1.ConfigMapsGetter, operatorConfig *v1alpha1.OpenShiftControllerManagerOperatorConfig) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/cm.yaml"))
	defaultConfig := v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/defaultconfig.yaml")
	requiredConfigMap, _, err := resourcemerge.MergeConfigMap(configMap, "config.yaml", nil, defaultConfig, operatorConfig.Spec.OpenShiftControllerManagerConfig.Raw)
	if err != nil {
		return nil, false, err
	}
	return resourceapply.ApplyConfigMap(client, requiredConfigMap)
}

func manageOpenShiftControllerManagerDeployment_v311_00_to_latest(client appsclientv1.DeploymentsGetter, options *v1alpha1.OpenShiftControllerManagerOperatorConfig, previousAvailability *operatorsv1alpha1.VersionAvailablity, forceDeployment bool) (*appsv1.Deployment, bool, error) {
	required := resourceread.ReadDeploymentV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/deployment.yaml"))
	required.Spec.Template.Spec.Containers[0].Image = options.Spec.ImagePullSpec
	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", options.Spec.Logging.Level))

	return resourceapply.ApplyDeployment(client, required, resourcemerge.ExpectedDeploymentGeneration(required, previousAvailability), forceDeployment)
}

func manageOpenShiftControllerManagerPublicConfigMap_v311_00_to_latest(client coreclientv1.ConfigMapsGetter, apiserverConfigString string, operatorConfig *v1alpha1.OpenShiftControllerManagerOperatorConfig) (*corev1.ConfigMap, bool, error) {
	uncastUnstructured, err := runtime.Decode(unstructured.UnstructuredJSONScheme, []byte(apiserverConfigString))
	if err != nil {
		return nil, false, err
	}
	apiserverConfig := uncastUnstructured.(runtime.Unstructured)

	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/public-info.yaml"))
	if operatorConfig.Status.CurrentAvailability != nil {
		configMap.Data["version"] = operatorConfig.Status.CurrentAvailability.Version
	} else {
		configMap.Data["version"] = ""
	}
	configMap.Data["imagePolicyConfig.internalRegistryHostname"], _, err = unstructured.NestedString(apiserverConfig.UnstructuredContent(), "imagePolicyConfig", "internalRegistryHostname")
	if err != nil {
		return nil, false, err
	}
	configMap.Data["imagePolicyConfig.externalRegistryHostname"], _, err = unstructured.NestedString(apiserverConfig.UnstructuredContent(), "imagePolicyConfig", "externalRegistryHostname")
	if err != nil {
		return nil, false, err
	}
	configMap.Data["projectConfig.defaultNodeSelector"], _, err = unstructured.NestedString(apiserverConfig.UnstructuredContent(), "projectConfig", "defaultNodeSelector")
	if err != nil {
		return nil, false, err
	}

	return resourceapply.ApplyConfigMap(client, configMap)
}
