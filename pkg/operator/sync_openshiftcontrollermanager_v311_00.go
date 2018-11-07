package operator

import (
	"bytes"
	"crypto/sha1"
	"encoding/pem"
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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	configv1 "github.com/openshift/api/config/v1"
	configclientv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
)

// syncOpenShiftControllerManager_v311_00_to_latest takes care of synchronizing (not upgrading) the thing we're managing.
// most of the time the sync method will be good for a large span of minor versions
func syncOpenShiftControllerManager_v311_00_to_latest(c OpenShiftControllerManagerOperator, operatorConfig *v1alpha1.OpenShiftControllerManagerOperatorConfig, previousAvailability *operatorsv1alpha1.VersionAvailability) (operatorsv1alpha1.VersionAvailability, []error) {
	versionAvailability := operatorsv1alpha1.VersionAvailability{
		Version: operatorConfig.Spec.Version,
	}

	syncErrors := []error{}
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
	forceRollout := false

	for _, currResult := range directResourceResults {
		if currResult.Error != nil {
			syncErrors = append(syncErrors, fmt.Errorf("%q (%T): %v", currResult.File, currResult.Type, currResult.Error))
			continue
		}

		if currResult.Changed && resourcesThatForceRedeployment.Has(currResult.File) {
			forceRollout = true
		}
	}

	controllerManagerConfig, configMapModified, err := manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(c.kubeClient.CoreV1(), operatorConfig)
	if err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("%q: %v", "configmap", err))
	}
	// the kube-apiserver is the source of truth for client CA bundles
	clientCAModified, err := manageOpenShiftAPIServerClientCA_v311_00_to_latest(c.kubeClient.CoreV1())
	if err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("%q: %v", "client-ca", err))
	}
	caHash, additionalCAModified, err := manageBuildAdditionalCAConfigMap(c.kubeClient.CoreV1(), c.configClient.ConfigV1(), operatorConfig)
	if err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("%q: %v", "build-additional-ca", err))
	}

	forceRollout = forceRollout || operatorConfig.ObjectMeta.Generation != operatorConfig.Status.ObservedGeneration
	forceRollout = forceRollout || configMapModified || clientCAModified || additionalCAModified

	// our configmaps and secrets are in order, now it is time to create the DS
	// TODO check basic preconditions here
	actualDaemonSet, _, err := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(c.kubeClient.AppsV1(), operatorConfig, previousAvailability, caHash, forceRollout)
	if err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("%q: %v", "deployment", err))
	}

	configData := ""
	if controllerManagerConfig != nil {
		configData = controllerManagerConfig.Data["config.yaml"]
	}
	_, _, err = manageOpenShiftControllerManagerPublicConfigMap_v311_00_to_latest(c.kubeClient.CoreV1(), configData, operatorConfig)
	if err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("%q: %v", "configmap/public-info", err))
	}

	return resourcemerge.ApplyDaemonSetGenerationAvailability(versionAvailability, actualDaemonSet, syncErrors...), syncErrors
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
	requiredConfigMap, _, err := resourcemerge.MergeConfigMap(configMap, "config.yaml", nil, defaultConfig, operatorConfig.Spec.UserConfig.Raw, operatorConfig.Spec.ObservedConfig.Raw)
	if err != nil {
		return nil, false, err
	}
	return resourceapply.ApplyConfigMap(client, requiredConfigMap)
}

func manageOpenShiftControllerManagerDeployment_v311_00_to_latest(client appsclientv1.DaemonSetsGetter, options *v1alpha1.OpenShiftControllerManagerOperatorConfig, previousAvailability *operatorsv1alpha1.VersionAvailability, additionalCAHash string, forceRollout bool) (*appsv1.DaemonSet, bool, error) {
	required := resourceread.ReadDaemonSetV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/ds.yaml"))
	required.Spec.Template.Spec.Containers[0].Image = options.Spec.ImagePullSpec
	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, fmt.Sprintf("-v=%d", options.Spec.Logging.Level))

	if required.Annotations == nil {
		required.Annotations = make(map[string]string)
	}
	required.Annotations["config.openshift.io/build.additional-ca-hash"] = additionalCAHash

	return resourceapply.ApplyDaemonSet(client, required, resourcemerge.ExpectedDaemonSetGeneration(required, previousAvailability), forceRollout)
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

func manageBuildAdditionalCAConfigMap(client coreclientv1.ConfigMapsGetter, configClient configclientv1.BuildsGetter, operatorConfig *v1alpha1.OpenShiftControllerManagerOperatorConfig) (string, bool, error) {
	caMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/build-additional-ca-cm.yaml"))
	buildConfig, err := configClient.Builds().Get("cluster", metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return "", false, err
	}
	if buildConfig == nil {
		caMap.Data = nil
		_, modified, err := resourceapply.ApplyConfigMap(client, caMap)
		operatorConfig.Status.AdditionalTrustedCA = nil
		return "", modified, err
	}

	caData, err := mergeCAConfigMap(client, buildConfig.Spec.AdditionalTrustedCA)
	if err != nil {
		return "", false, err
	}
	if len(caData) == 0 && operatorConfig.Status.AdditionalTrustedCA != nil {
		caMap.Data = nil
		_, modified, err := resourceapply.ApplyConfigMap(client, caMap)
		operatorConfig.Status.AdditionalTrustedCA = nil
		return "", modified, err
	}
	h := sha1.New()
	h.Write(caData)
	caHash := fmt.Sprintf("%x", h.Sum(nil))
	if operatorConfig.Status.AdditionalTrustedCA != nil &&
		operatorConfig.Status.AdditionalTrustedCA.SHA1Hash == caHash {
		return caHash, false, nil
	}
	operatorConfig.Status.AdditionalTrustedCA = &v1alpha1.AdditionalTrustedCA{
		SHA1Hash:      caHash,
		ConfigMapName: buildConfig.Spec.AdditionalTrustedCA.Name,
	}
	caMap.Data["additional-ca.crt"] = string(caData)
	_, modified, err := resourceapply.ApplyConfigMap(client, caMap)
	return caHash, modified, err
}

// mergeCAConfigMap merges the CA content within the provided ConfigMap. Returns the merged CA bundle bytes.
//
// If the ConfigMap contains invalid PEM-encoded data, an error is thrown.
func mergeCAConfigMap(client coreclientv1.ConfigMapsGetter, cmRef configv1.ConfigMapReference) ([]byte, error) {
	if len(cmRef.Name) == 0 || len(cmRef.Namespace) == 0 {
		return nil, nil
	}
	configMap, err := client.ConfigMaps(cmRef.Namespace).Get(cmRef.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if configMap == nil {
		return nil, nil
	}
	return mergePEMData(configMap.Data)
}

func mergePEMData(data map[string]string) ([]byte, error) {
	pemData := &bytes.Buffer{}
	for key, certData := range data {
		data := []byte(certData)
		for len(data) > 0 {
			block, rest := pem.Decode(data)
			if block == nil && len(rest) > 0 {
				return nil, fmt.Errorf("map key %s contains invalid PEM data: %s", key, string(rest))
			}
			err := pem.Encode(pemData, block)
			if err != nil {
				return nil, err
			}
			data = rest
		}
	}
	return pemData.Bytes(), nil
}
