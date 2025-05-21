package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/openshift/cluster-openshift-controller-manager-operator/bindata"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	operatorapiv1 "github.com/openshift/api/operator/v1"

	configv1 "github.com/openshift/api/config/v1"
	controlplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	configlisterv1 "github.com/openshift/client-go/config/listers/config/v1"
	proxyvclient1 "github.com/openshift/client-go/config/listers/config/v1"

	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcehash"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

const rcmConditionTypePrefix = "RouteControllerManager"

// ControllerCapabilities maps controllers to capabilities, so we can enable/disable controllers
// based on capabilities.
var controllerCapabilities = map[controlplanev1.OpenShiftControllerName]configv1.ClusterVersionCapability{
	controlplanev1.OpenShiftBuildController:                     configv1.ClusterVersionCapabilityBuild,
	controlplanev1.OpenShiftBuildConfigChangeController:         configv1.ClusterVersionCapabilityBuild,
	controlplanev1.OpenShiftBuilderServiceAccountController:     configv1.ClusterVersionCapabilityBuild,
	controlplanev1.OpenShiftBuilderRoleBindingsController:       configv1.ClusterVersionCapabilityBuild,
	controlplanev1.OpenShiftDeploymentConfigController:          configv1.ClusterVersionCapabilityDeploymentConfig,
	controlplanev1.OpenShiftDeployerServiceAccountController:    configv1.ClusterVersionCapabilityDeploymentConfig,
	controlplanev1.OpenShiftDeployerController:                  configv1.ClusterVersionCapabilityDeploymentConfig,
	controlplanev1.OpenShiftDeployerRoleBindingsController:      configv1.ClusterVersionCapabilityDeploymentConfig,
	controlplanev1.OpenShiftServiceAccountPullSecretsController: configv1.ClusterVersionCapabilityImageRegistry,
	controlplanev1.OpenShiftImagePullerRoleBindingsController:   configv1.ClusterVersionCapabilityImageRegistry,
}

// syncOpenShiftControllerManager_v311_00_to_latest takes care of synchronizing (not upgrading) the thing we're managing.
// most of the time the sync method will be good for a large span of minor versions
func syncOpenShiftControllerManager_v311_00_to_latest(
	c OpenShiftControllerManagerOperator,
	originalOperatorConfig *operatorapiv1.OpenShiftControllerManager,
	countNodes nodeCountFunc,
	ensureAtMostOnePodPerNodeFn ensureAtMostOnePodPerNodeFunc,
) (bool, error) {
	var (
		ocmErrors, rcmErrors []error
		err                  error
	)
	operatorConfig := originalOperatorConfig.DeepCopy()

	operandName := "openshift-controller-manager"
	rcOperandName := "route-controller-manager"

	specAnnotations := map[string]string{
		"openshiftcontrollermanagers.operator.openshift.io/cluster": strconv.FormatInt(operatorConfig.ObjectMeta.Generation, 10),
	}

	rcSpecAnnotations := map[string]string{
		"openshiftcontrollermanagers.operator.openshift.io/cluster": strconv.FormatInt(operatorConfig.ObjectMeta.Generation, 10),
	}

	// OpenShift Controller Manager
	configMap, _, err := manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(c.clusterVersionLister, c.kubeClient, c.configMapsGetter, c.recorder, operatorConfig)
	if err != nil {
		ocmErrors = append(ocmErrors, fmt.Errorf("%q %q: %w", operandName, "config", err))
	} else {
		specAnnotations["configmaps/config"] = configMap.ObjectMeta.ResourceVersion
	}
	// the kube-apiserver is the source of truth for client CA bundles
	clientCAConfigMap, _, err := manageOpenShiftControllerManagerClientCA_v311_00_to_latest(c.kubeClient.CoreV1(), c.recorder)
	if err != nil {
		ocmErrors = append(ocmErrors, fmt.Errorf("%q %q: %v", operandName, "client-ca", err))
	} else {
		resourceVersion := "0"
		if clientCAConfigMap != nil { // SyncConfigMap can return nil
			resourceVersion = clientCAConfigMap.ObjectMeta.ResourceVersion
		}
		specAnnotations["configmaps/client-ca"] = resourceVersion
	}

	serviceCAConfigMap, _, err := manageOpenShiftServiceCAConfigMap_v311_00_to_latest(c.kubeClient, c.configMapsGetter, c.recorder)
	if err != nil {
		ocmErrors = append(ocmErrors, fmt.Errorf("%q %q: %v", operandName, "openshift-service-ca", err))
	} else {
		specAnnotations["configmaps/openshift-service-ca"] = serviceCAConfigMap.ObjectMeta.ResourceVersion
	}

	globalCAConfigMap, _, err := manageOpenShiftGlobalCAConfigMap_v311_00_to_latest(c.kubeClient, c.configMapsGetter, c.recorder)
	if err != nil {
		ocmErrors = append(ocmErrors, fmt.Errorf("%q %q: %v", operandName, "openshift-global-ca", err))
	} else {
		specAnnotations["configmaps/openshift-global-ca"] = globalCAConfigMap.ObjectMeta.ResourceVersion
	}

	// Route Controller Manager
	rcConfigMap, _, err := manageRouteControllerManagerConfigMap_v311_00_to_latest(c.kubeClient, c.configMapsGetter, c.recorder, operatorConfig)
	if err != nil {
		rcmErrors = append(rcmErrors, fmt.Errorf("%q %q: %v", rcOperandName, "configmap", err))
	} else {
		rcSpecAnnotations["configmaps/config"] = rcConfigMap.ObjectMeta.ResourceVersion
	}

	// the kube-apiserver is the source of truth for client CA bundles
	rcClientCAConfigMap, _, err := manageRouteControllerManagerClientCA_v311_00_to_latest(c.kubeClient.CoreV1(), c.recorder)
	if err != nil {
		rcmErrors = append(rcmErrors, fmt.Errorf("%q %q: %v", rcOperandName, "client-ca", err))
	} else {
		resourceVersion := "0"
		if rcClientCAConfigMap != nil { // SyncConfigMap can return nil
			resourceVersion = rcClientCAConfigMap.ObjectMeta.ResourceVersion
		}
		rcSpecAnnotations["configmaps/client-ca"] = resourceVersion
	}

	// our configmaps and secrets are in order, now it is time to create the Deployment
	actualDeployment, _, openshiftControllerManagerError := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(
		bindata.MustAsset,
		c.kubeClient.AppsV1(),
		countNodes,
		ensureAtMostOnePodPerNodeFn,
		c.recorder,
		operatorConfig,
		c.targetImagePullSpec,
		operatorConfig.Status.Generations,
		c.proxyLister,
		specAnnotations,
	)
	if openshiftControllerManagerError != nil {
		ocmErrors = append(ocmErrors, fmt.Errorf("%q %q: %v", operandName, "deployment", openshiftControllerManagerError))
	}

	actualRCDeployment, _, routerControllerManagerError := manageRouteControllerManagerDeployment_v311_00_to_latest(
		c.kubeClient.AppsV1(),
		countNodes,
		ensureAtMostOnePodPerNodeFn,
		c.recorder,
		operatorConfig,
		c.routeControllerManagerTargetImagePullSpec,
		operatorConfig.Status.Generations,
		rcSpecAnnotations,
	)
	if routerControllerManagerError != nil {
		rcmErrors = append(rcmErrors, fmt.Errorf("%q %q: %v", rcOperandName, "deployment", routerControllerManagerError))
	}

	// library-go func called by manageOpenShiftControllerManagerDeployment_v311_00_to_latest can return nil with errors
	if openshiftControllerManagerError != nil || routerControllerManagerError != nil {
		return syncReturn(c, ocmErrors, rcmErrors, originalOperatorConfig, operatorConfig)
	}

	// manage status
	// first we need to update operator config status version based on the OCM deployment
	if d := actualDeployment; d.Status.AvailableReplicas > 0 && d.Status.UpdatedReplicas == d.Status.Replicas && len(d.Annotations[util.VersionAnnotation]) > 0 {
		operatorConfig.Status.Version = d.Annotations[util.VersionAnnotation]
	}

	setControllerManagerStatusConditions(operatorConfig, actualDeployment, "openshift controller manager", "")
	setControllerManagerStatusConditions(operatorConfig, actualRCDeployment, "route controller manager", rcmConditionTypePrefix)

	v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
		Type:   operatorapiv1.OperatorStatusTypeUpgradeable,
		Status: operatorapiv1.ConditionTrue,
	})

	operatorConfig.Status.ObservedGeneration = operatorConfig.ObjectMeta.Generation
	resourcemerge.SetDeploymentGeneration(&operatorConfig.Status.Generations, actualDeployment)
	resourcemerge.SetDeploymentGeneration(&operatorConfig.Status.Generations, actualRCDeployment)

	return syncReturn(c, ocmErrors, rcmErrors, originalOperatorConfig, operatorConfig)
}

// setControllerManagerStatusConditions sets operator status conditions for an operand.
// Available and Progressing are being handled here, Degraded is handled in syncReturn.
//
// Make sure operatorConfig.Status.Version is set properly before calling this helper function.
func setControllerManagerStatusConditions(
	operatorConfig *operatorapiv1.OpenShiftControllerManager,
	d *appsv1.Deployment,
	operandReadableName string,
	conditionTypePrefix string,
) {
	available := d.Status.AvailableReplicas > 0

	// Available
	if available {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:   conditionTypePrefix + operatorapiv1.OperatorStatusTypeAvailable,
			Status: operatorapiv1.ConditionTrue,
		})
	} else {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:    conditionTypePrefix + operatorapiv1.OperatorStatusTypeAvailable,
			Status:  operatorapiv1.ConditionFalse,
			Reason:  "NoPodsAvailable",
			Message: fmt.Sprintf("no %s deployment pods available on any node", operandReadableName),
		})
	}

	// Progressing
	var progressingMessages []string
	if available && d.Status.UpdatedReplicas == d.Status.Replicas {
		if len(d.Annotations[util.VersionAnnotation]) > 0 {
			if len(operatorConfig.Status.Version) != 0 && operatorConfig.Status.Version != d.Annotations[util.VersionAnnotation] {
				progressingMessages = append(progressingMessages, fmt.Sprintf(
					"deployment/%s: has invalid version annotation %s, desired version %s",
					d.Name, util.VersionAnnotation, operatorConfig.Status.Version))
			}
		} else {
			progressingMessages = append(progressingMessages, fmt.Sprintf(
				"deployment/%s: version annotation %s missing", d.Name, util.VersionAnnotation))
		}
	}
	if d.ObjectMeta.Generation != d.Status.ObservedGeneration {
		progressingMessages = append(progressingMessages, fmt.Sprintf(
			"deployment/%s: observed generation is %d, desired generation is %d",
			d.Name, d.Status.ObservedGeneration, d.ObjectMeta.Generation))
	}
	if d.Status.AvailableReplicas == 0 {
		progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/%s: no available replica found", d.Name))
	}
	if d.Status.UpdatedReplicas != *d.Spec.Replicas {
		progressingMessages = append(progressingMessages, fmt.Sprintf(
			"deployment/%s: updated replicas is %d, desired replicas is %d",
			d.Name, d.Status.UpdatedReplicas, *d.Spec.Replicas))
	}

	// This is actually not depending on the deployment object,
	// but generation mismatch sets all operands to Progressing.
	if operatorConfig.ObjectMeta.Generation != operatorConfig.Status.ObservedGeneration {
		progressingMessages = append(progressingMessages, fmt.Sprintf(
			"openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is %d, desired generation is %d",
			operatorConfig.Status.ObservedGeneration, operatorConfig.ObjectMeta.Generation))
	}

	if len(progressingMessages) == 0 {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:   conditionTypePrefix + operatorapiv1.OperatorStatusTypeProgressing,
			Status: operatorapiv1.ConditionFalse,
		})
	} else {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:    conditionTypePrefix + operatorapiv1.OperatorStatusTypeProgressing,
			Status:  operatorapiv1.ConditionTrue,
			Reason:  "DesiredStateNotYetAchieved",
			Message: strings.Join(progressingMessages, "\n"),
		})
	}
}

// syncReturn checks the error slices and sets Degraded conditions in the operator config before returning.
func syncReturn(
	c OpenShiftControllerManagerOperator,
	ocmErrors, rcmErrors []error,
	originalOperatorConfig, operatorConfig *operatorapiv1.OpenShiftControllerManager,
) (bool, error) {
	setDegraded := func(errs []error, conditionType string) {
		if len(errs) == 0 {
			v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
				Type:   conditionType,
				Status: operatorapiv1.ConditionFalse,
			})
			return
		}

		var msg strings.Builder
		for _, err := range errs {
			msg.WriteString(err.Error())
			msg.WriteString("\n")
		}
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:    conditionType,
			Status:  operatorapiv1.ConditionTrue,
			Message: msg.String(),
			Reason:  "SyncError",
		})
	}

	setDegraded(ocmErrors, workloadDegradedCondition)
	setDegraded(rcmErrors, rcmConditionTypePrefix+"Degraded")

	if !equality.Semantic.DeepEqual(operatorConfig.Status, originalOperatorConfig.Status) {
		if _, err := c.operatorConfigClient.OpenShiftControllerManagers().UpdateStatus(context.TODO(), operatorConfig, metav1.UpdateOptions{}); err != nil {
			return false, err
		}
	}
	return len(ocmErrors) > 0 || len(rcmErrors) > 0, nil
}

func manageOpenShiftControllerManagerClientCA_v311_00_to_latest(client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	const apiserverClientCA = "client-ca"
	return resourceapply.SyncConfigMap(context.Background(), client, recorder, util.KubeAPIServerNamespace, apiserverClientCA, util.TargetNamespace, apiserverClientCA, []metav1.OwnerReference{})
}

func manageRouteControllerManagerClientCA_v311_00_to_latest(client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	const apiserverClientCA = "client-ca"
	return resourceapply.SyncConfigMap(context.Background(), client, recorder, util.KubeAPIServerNamespace, apiserverClientCA, util.RouteControllerTargetNamespace, apiserverClientCA, []metav1.OwnerReference{})
}

// disableControllers disables controllers based on capabilities.
// By default, all controllers are enabled. To disable specific controllers,
// '-foo' disables the controller named 'foo'. For example, to disable the
// OpenShiftServiceAccountController, one would use "-openshift.io/serviceaccount".
func disableControllers(clusterVersion *configv1.ClusterVersion) []string {
	controllers := []string{"*"}
	knownCaps := sets.New[configv1.ClusterVersionCapability](clusterVersion.Status.Capabilities.KnownCapabilities...)
	capsEnabled := sets.New[configv1.ClusterVersionCapability](clusterVersion.Status.Capabilities.EnabledCapabilities...)

	// OCPBUILD-8: the default pull secrets controller was refactored so each role binding can
	// be matched with a cluster capability. The original default pull secrets controller was
	// preserved to facilitate incremental upgrades. Starting in 4.16, the default pull secrets
	// controller should be always be disabled to prevent race conditions and conflicts.

	// TODO - the default rolebindings controller code should be removed in 4.17+
	controllers = append(controllers, "-openshift.io/default-rolebindings")

	for cont, cap := range controllerCapabilities {
		if knownCaps.Has(cap) && !capsEnabled.Has(cap) {
			controllers = append(controllers, fmt.Sprintf("-%s", cont))
		}
	}
	// we need to sort this slice to have always same list
	// otherwise, change in the order would modify configmap and causes roll out
	sort.Strings(controllers)
	return controllers
}

// similar logic for route-controller-manager in manageRouteControllerManagerConfigMap_v311_00_to_latest
func manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(clusterVersionLister configlisterv1.ClusterVersionLister, kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder, operatorConfig *operatorapiv1.OpenShiftControllerManager) (*corev1.ConfigMap, bool, error) {
	clusterVersion, err := clusterVersionLister.Get("version")
	if err != nil {
		return nil, false, err
	}
	controllers := disableControllers(clusterVersion)
	configMap := resourceread.ReadConfigMapV1OrDie(bindata.MustAsset("assets/openshift-controller-manager/cm.yaml"))
	ocmDefaultConfig := bindata.MustAsset("assets/config/openshift-controller-manager-defaultconfig.yaml")
	controllersBytes, err := json.Marshal(controllers)
	bytes := []byte(fmt.Sprintf("{\"apiVersion\": \"openshiftcontrolplane.config.openshift.io/v1\", \"kind\": \"OpenShiftControllerManagerConfig\", \"controllers\": %s}", controllersBytes))
	if err != nil {
		return nil, false, err
	}
	requiredConfigMap, _, err := resourcemerge.MergeConfigMap(configMap, "config.yaml", nil, ocmDefaultConfig, bytes, operatorConfig.Spec.ObservedConfig.Raw, operatorConfig.Spec.UnsupportedConfigOverrides.Raw)
	if err != nil {
		return nil, false, err
	}

	// we can embed input hashes on our main configmap to drive rollouts when they change.
	inputHashes, err := resourcehash.MultipleObjectHashStringMapForObjectReferences(
		context.TODO(),
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

	return resourceapply.ApplyConfigMap(context.Background(), client, recorder, requiredConfigMap)
}

// similar logic for route-controller-manager in manageOpenShiftControllerManagerConfigMap_v311_00_to_latest
func manageRouteControllerManagerConfigMap_v311_00_to_latest(kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder, operatorConfig *operatorapiv1.OpenShiftControllerManager) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(bindata.MustAsset("assets/openshift-controller-manager/route-controller-manager-cm.yaml"))
	rcmDefaultConfig := bindata.MustAsset("assets/config/route-controller-manager-defaultconfig.yaml")
	requiredConfigMap, _, err := resourcemerge.MergeConfigMap(configMap, "config.yaml", nil, rcmDefaultConfig, operatorConfig.Spec.ObservedConfig.Raw, operatorConfig.Spec.UnsupportedConfigOverrides.Raw)
	if err != nil {
		return nil, false, err
	}

	// we can embed input hashes on our main configmap to drive rollouts when they change.
	inputHashes, err := resourcehash.MultipleObjectHashStringMapForObjectReferences(
		context.TODO(),
		kubeClient,
		resourcehash.NewObjectRef().ForConfigMap().InNamespace(util.RouteControllerTargetNamespace).Named("client-ca"),
		resourcehash.NewObjectRef().ForSecret().InNamespace(util.RouteControllerTargetNamespace).Named("serving-cert"),
	)
	if err != nil {
		return nil, false, err
	}
	for k, v := range inputHashes {
		requiredConfigMap.Data[k] = v
	}

	return resourceapply.ApplyConfigMap(context.Background(), client, recorder, requiredConfigMap)
}

func manageOpenShiftServiceCAConfigMap_v311_00_to_latest(kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(bindata.MustAsset("assets/openshift-controller-manager/openshift-service-ca-cm.yaml"))
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

// manageOpenShiftGlobalCAConfigMap_v311_00_to_latest syncs a ConfigMap that has the cluster's
// global trust bundle injected. This CA is used by ocm to communicate with external services,
// such as container registries that the image signature import controller downloads data from.
// The global trust bundle is needed in the event the service uses a custom PKI certificate, or
// OpenShift is run behind a proxy that uses a custom PKI certificate.
func manageOpenShiftGlobalCAConfigMap_v311_00_to_latest(kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(bindata.MustAsset("assets/openshift-controller-manager/openshift-global-ca-cm.yaml"))
	// ApplyConfigMap now handles the injection of CA certificates.
	return resourceapply.ApplyConfigMap(context.TODO(), client, recorder, configMap)
}

func manageOpenShiftControllerManagerDeployment_v311_00_to_latest(
	mustLoadAsset func(string) []byte,
	client appsclientv1.DeploymentsGetter,
	countNodes nodeCountFunc,
	ensureAtMostOnePodPerNodeFn ensureAtMostOnePodPerNodeFunc,
	recorder events.Recorder,
	options *operatorapiv1.OpenShiftControllerManager,
	imagePullSpec string,
	generationStatus []operatorapiv1.GenerationStatus,
	proxyLister proxyvclient1.ProxyLister,
	specAnnotations map[string]string,
) (*appsv1.Deployment, bool, error) {
	required := resourceread.ReadDeploymentV1OrDie(mustLoadAsset("assets/openshift-controller-manager/deploy.yaml"))

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
	resourcemerge.MergeMap(resourcemerge.BoolPtr(false), &required.Spec.Template.Annotations, specAnnotations)

	// Set the replica count to the number of master nodes.
	masterNodeCount, err := countNodes(required.Spec.Template.Spec.NodeSelector)
	if err != nil {
		return nil, false, fmt.Errorf("failed to determine number of master nodes: %v", err)
	}
	required.Spec.Replicas = masterNodeCount

	err = ensureAtMostOnePodPerNodeFn(&required.Spec, util.RouteControllerTargetNamespace)
	if err != nil {
		return nil, false, fmt.Errorf("unable to ensure at most one pod per node: %v", err)
	}

	proxyCfg, err := proxyLister.Get("cluster")
	if err != nil {
		recorder.Eventf("ProxyConfigGetFailed", "Error retrieving global proxy config: %s", err.Error())
		if !apierrors.IsNotFound(err) {
			// return deployment since it is still referenced by caller even with errors
			return required, false, err
		}
	} else {
		for i, c := range required.Spec.Template.Spec.Containers {
			// Remove the relevant variables from the spec.
			newEnvs := make([]corev1.EnvVar, 0, len(c.Env))
			for _, env := range c.Env {
				switch strings.TrimSpace(env.Name) {
				case "HTTPS_PROXY", "HTTP_PROXY", "NO_PROXY":
				default:
					newEnvs = append(newEnvs, env)
				}
			}

			// Add back the values that are available in the proxy config.
			if len(proxyCfg.Status.HTTPSProxy) > 0 {
				newEnvs = append(newEnvs, corev1.EnvVar{Name: "HTTPS_PROXY", Value: proxyCfg.Status.HTTPSProxy})
			}
			if len(proxyCfg.Status.HTTPProxy) > 0 {
				newEnvs = append(newEnvs, corev1.EnvVar{Name: "HTTP_PROXY", Value: proxyCfg.Status.HTTPProxy})
			}
			if len(proxyCfg.Status.NoProxy) > 0 {
				newEnvs = append(newEnvs, corev1.EnvVar{Name: "NO_PROXY", Value: proxyCfg.Status.NoProxy})
			}

			required.Spec.Template.Spec.Containers[i].Env = newEnvs
		}
	}

	return resourceapply.ApplyDeployment(context.Background(), client, recorder, required, resourcemerge.ExpectedDeploymentGeneration(required, generationStatus))
}

func manageRouteControllerManagerDeployment_v311_00_to_latest(
	client appsclientv1.DeploymentsGetter,
	countNodes nodeCountFunc,
	ensureAtMostOnePodPerNodeFn ensureAtMostOnePodPerNodeFunc,
	recorder events.Recorder,
	options *operatorapiv1.OpenShiftControllerManager,
	imagePullSpec string,
	generationStatus []operatorapiv1.GenerationStatus,
	specAnnotations map[string]string,
) (*appsv1.Deployment, bool, error) {
	required := resourceread.ReadDeploymentV1OrDie(bindata.MustAsset("assets/openshift-controller-manager/route-controller-manager-deploy.yaml"))

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
	resourcemerge.MergeMap(resourcemerge.BoolPtr(false), &required.Spec.Template.Annotations, specAnnotations)

	// Set the replica count to the number of master nodes.
	masterNodeCount, err := countNodes(required.Spec.Template.Spec.NodeSelector)
	if err != nil {
		return nil, false, fmt.Errorf("failed to determine number of master nodes: %v", err)
	}
	required.Spec.Replicas = masterNodeCount

	err = ensureAtMostOnePodPerNodeFn(&required.Spec, util.RouteControllerTargetNamespace)
	if err != nil {
		return nil, false, fmt.Errorf("unable to ensure at most one pod per node: %v", err)
	}

	return resourceapply.ApplyDeployment(context.Background(), client, recorder, required, resourcemerge.ExpectedDeploymentGeneration(required, generationStatus))
}
