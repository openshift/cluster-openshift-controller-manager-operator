package operator

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	operatorapiv1 "github.com/openshift/api/operator/v1"

	proxyvclient1 "github.com/openshift/client-go/config/listers/config/v1"

	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcehash"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

// syncOpenShiftControllerManager_v311_00_to_latest takes care of synchronizing (not upgrading) the thing we're managing.
// most of the time the sync method will be good for a large span of minor versions
func syncOpenShiftControllerManager_v311_00_to_latest(
	c OpenShiftControllerManagerOperator,
	originalOperatorConfig *operatorapiv1.OpenShiftControllerManager,
	countNodes nodeCountFunc,
	ensureAtMostOnePodPerNodeFn ensureAtMostOnePodPerNodeFunc,
) (bool, error) {
	errors := []error{}
	var err error
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
	configMap, _, err := manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(c.kubeClient, c.configMapsGetter, c.recorder, operatorConfig)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q %q: %v", operandName, "config", err))
	} else {
		specAnnotations["configmaps/config"] = configMap.ObjectMeta.ResourceVersion
	}
	// the kube-apiserver is the source of truth for client CA bundles
	clientCAConfigMap, _, err := manageOpenShiftControllerManagerClientCA_v311_00_to_latest(c.kubeClient.CoreV1(), c.recorder)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q %q: %v", operandName, "client-ca", err))
	} else {
		resourceVersion := "0"
		if clientCAConfigMap != nil { // SyncConfigMap can return nil
			resourceVersion = clientCAConfigMap.ObjectMeta.ResourceVersion
		}
		specAnnotations["configmaps/client-ca"] = resourceVersion
	}

	serviceCAConfigMap, _, err := manageOpenShiftServiceCAConfigMap_v311_00_to_latest(c.kubeClient, c.configMapsGetter, c.recorder)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q %q: %v", operandName, "openshift-service-ca", err))
	} else {
		specAnnotations["configmaps/openshift-service-ca"] = serviceCAConfigMap.ObjectMeta.ResourceVersion
	}

	globalCAConfigMap, _, err := manageOpenShiftGlobalCAConfigMap_v311_00_to_latest(c.kubeClient, c.configMapsGetter, c.recorder)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q %q: %v", operandName, "openshift-global-ca", err))
	} else {
		specAnnotations["configmaps/openshift-global-ca"] = globalCAConfigMap.ObjectMeta.ResourceVersion
	}

	// Route Controller Manager
	rcConfigMap, _, err := manageRouteControllerManagerConfigMap_v311_00_to_latest(c.kubeClient, c.configMapsGetter, c.recorder, operatorConfig)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q %q: %v", rcOperandName, "configmap", err))
	} else {
		rcSpecAnnotations["configmaps/config"] = rcConfigMap.ObjectMeta.ResourceVersion
	}

	// the kube-apiserver is the source of truth for client CA bundles
	rcClientCAConfigMap, _, err := manageRouteControllerManagerClientCA_v311_00_to_latest(c.kubeClient.CoreV1(), c.recorder)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q %q: %v", rcOperandName, "client-ca", err))
	} else {
		resourceVersion := "0"
		if rcClientCAConfigMap != nil { // SyncConfigMap can return nil
			resourceVersion = rcClientCAConfigMap.ObjectMeta.ResourceVersion
		}
		rcSpecAnnotations["configmaps/client-ca"] = resourceVersion
	}

	// our configmaps and secrets are in order, now it is time to create the Deployment
	var progressingMessages []string
	actualDeployment, _, openshiftControllerManagerError := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(
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
		msg := fmt.Sprintf("%q %q: %v", operandName, "deployment", openshiftControllerManagerError)
		progressingMessages = append(progressingMessages, msg)
		errors = append(errors, fmt.Errorf(msg))
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
		msg := fmt.Sprintf("%q %q: %v", rcOperandName, "deployment", routerControllerManagerError)
		progressingMessages = append(progressingMessages, msg)
		errors = append(errors, fmt.Errorf(msg))
	}

	// library-go func called by manageOpenShiftControllerManagerDeployment_v311_00_to_latest can return nil with errors
	if openshiftControllerManagerError != nil || routerControllerManagerError != nil {
		return syncReturn(c, errors, originalOperatorConfig, operatorConfig)
	}

	// at this point we know that the actualDeployment and actualRCDeployment are both non-nil and non-empty
	available := actualDeployment.Status.AvailableReplicas > 0
	rcAvailable := actualRCDeployment.Status.AvailableReplicas > 0

	// manage status
	if available && rcAvailable {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:   operatorapiv1.OperatorStatusTypeAvailable,
			Status: operatorapiv1.ConditionTrue,
		})
	} else {
		msg := "no pods available on any node."
		if !available && rcAvailable {
			msg = fmt.Sprintf("no openshift controller manager daemon pods available on any node.")
		}
		if available && !rcAvailable {
			msg = fmt.Sprintf("no route controller manager deployment pods available on any node.")
		}

		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:    operatorapiv1.OperatorStatusTypeAvailable,
			Status:  operatorapiv1.ConditionFalse,
			Reason:  "NoPodsAvailable",
			Message: msg,
		})
	}

	if available && actualDeployment.Status.UpdatedReplicas == actualDeployment.Status.Replicas {
		if len(actualDeployment.Annotations[util.VersionAnnotation]) > 0 {
			operatorConfig.Status.Version = actualDeployment.Annotations[util.VersionAnnotation]
		} else {
			progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/controller-manager: version annotation %s missing.", util.VersionAnnotation))
		}
	}

	if rcAvailable && actualRCDeployment.Status.UpdatedReplicas == actualRCDeployment.Status.Replicas {
		if len(actualRCDeployment.Annotations[util.VersionAnnotation]) > 0 {
			// version should be the same as the controller-manager, just do a check the route-controller-manager has the same
			if len(operatorConfig.Status.Version) != 0 && operatorConfig.Status.Version != actualRCDeployment.Annotations[util.VersionAnnotation] {
				progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/route-controller-manager: has invalid version annotation %s, desired version %s.", util.VersionAnnotation, operatorConfig.Status.Version))
			}
		} else {
			progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/route-controller-manager: version annotation %s missing.", util.VersionAnnotation))
		}
	}

	if actualDeployment != nil && actualDeployment.ObjectMeta.Generation != actualDeployment.Status.ObservedGeneration {
		progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/controller-manager: observed generation is %d, desired generation is %d.", actualDeployment.Status.ObservedGeneration, actualDeployment.ObjectMeta.Generation))
	}
	if actualDeployment.Status.AvailableReplicas == 0 {
		progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/controller-manager: available replicas is %d, desired available replicas > %d", actualDeployment.Status.AvailableReplicas, 1))
	}
	if actualDeployment.Status.UpdatedReplicas != *actualDeployment.Spec.Replicas {
		progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/controller-manager: updated replicas is %d, desired replicas is %d", actualDeployment.Status.UpdatedReplicas, *actualDeployment.Spec.Replicas))
	}
	if actualRCDeployment != nil && actualRCDeployment.ObjectMeta.Generation != actualRCDeployment.Status.ObservedGeneration {
		progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/route-controller-manager: observed generation is %d, desired generation is %d.", actualRCDeployment.Status.ObservedGeneration, actualRCDeployment.ObjectMeta.Generation))
	}
	if actualRCDeployment.Status.AvailableReplicas == 0 {
		progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/route-controller-manager: available replicas is %d, desired available replicas > %d", actualRCDeployment.Status.AvailableReplicas, 1))
	}
	if actualRCDeployment.Status.UpdatedReplicas != *actualRCDeployment.Spec.Replicas {
		progressingMessages = append(progressingMessages, fmt.Sprintf("deployment/route-controller-manager: updated replicas is %d, desired replicas is %d", actualRCDeployment.Status.UpdatedReplicas, *actualRCDeployment.Spec.Replicas))
	}
	if operatorConfig.ObjectMeta.Generation != operatorConfig.Status.ObservedGeneration {
		progressingMessages = append(progressingMessages, fmt.Sprintf("openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is %d, desired generation is %d.", operatorConfig.Status.ObservedGeneration, operatorConfig.ObjectMeta.Generation))
	}
	if len(progressingMessages) == 0 {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:   operatorapiv1.OperatorStatusTypeProgressing,
			Status: operatorapiv1.ConditionFalse,
		})
	} else {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:    operatorapiv1.OperatorStatusTypeProgressing,
			Status:  operatorapiv1.ConditionTrue,
			Reason:  "DesiredStateNotYetAchieved",
			Message: strings.Join(progressingMessages, "\n"),
		})
	}

	v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
		Type:   operatorapiv1.OperatorStatusTypeUpgradeable,
		Status: operatorapiv1.ConditionTrue,
	})

	operatorConfig.Status.ObservedGeneration = operatorConfig.ObjectMeta.Generation
	resourcemerge.SetDeploymentGeneration(&operatorConfig.Status.Generations, actualDeployment)
	resourcemerge.SetDeploymentGeneration(&operatorConfig.Status.Generations, actualRCDeployment)

	return syncReturn(c, errors, originalOperatorConfig, operatorConfig)
}

func syncReturn(c OpenShiftControllerManagerOperator, errors []error, originalOperatorConfig, operatorConfig *operatorapiv1.OpenShiftControllerManager) (bool, error) {
	if len(errors) > 0 {
		message := ""
		for _, err := range errors {
			message = message + err.Error() + "\n"
		}
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:    workloadDegradedCondition,
			Status:  operatorapiv1.ConditionTrue,
			Message: message,
			Reason:  "SyncError",
		})
	} else {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:   workloadDegradedCondition,
			Status: operatorapiv1.ConditionFalse,
		})
	}

	if !equality.Semantic.DeepEqual(operatorConfig.Status, originalOperatorConfig.Status) {
		if _, err := c.operatorConfigClient.OpenShiftControllerManagers().UpdateStatus(context.TODO(), operatorConfig, metav1.UpdateOptions{}); err != nil {
			return false, err
		}
	}

	if len(errors) > 0 {
		return true, nil
	}
	return false, nil
}

func manageOpenShiftControllerManagerClientCA_v311_00_to_latest(client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	const apiserverClientCA = "client-ca"
	return resourceapply.SyncConfigMap(context.Background(), client, recorder, util.KubeAPIServerNamespace, apiserverClientCA, util.TargetNamespace, apiserverClientCA, []metav1.OwnerReference{})
}

func manageRouteControllerManagerClientCA_v311_00_to_latest(client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	const apiserverClientCA = "client-ca"
	return resourceapply.SyncConfigMap(context.Background(), client, recorder, util.KubeAPIServerNamespace, apiserverClientCA, util.RouteControllerTargetNamespace, apiserverClientCA, []metav1.OwnerReference{})
}

// similar logic for route-controller-manager in manageRouteControllerManagerConfigMap_v311_00_to_latest
func manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder, operatorConfig *operatorapiv1.OpenShiftControllerManager) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/cm.yaml"))
	defaultConfig := v311_00_assets.MustAsset("v3.11.0/config/defaultconfig.yaml")
	requiredConfigMap, _, err := resourcemerge.MergeConfigMap(configMap, "config.yaml", nil, defaultConfig, operatorConfig.Spec.ObservedConfig.Raw, operatorConfig.Spec.UnsupportedConfigOverrides.Raw)
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
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/route-controller-cm.yaml"))
	rcmDefaultConfig := v311_00_assets.MustAsset("v3.11.0/config/route-controller-defaultconfig.yaml")
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

// manageOpenShiftGlobalCAConfigMap_v311_00_to_latest syncs a ConfigMap that has the cluster's
// global trust bundle injected. This CA is used by ocm to communicate with external services,
// such as container registries that the image signature import controller downloads data from.
// The global trust bundle is needed in the event the service uses a custom PKI certificate, or
// OpenShift is run behind a proxy that uses a custom PKI certificate.
func manageOpenShiftGlobalCAConfigMap_v311_00_to_latest(kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml"))
	// ApplyConfigMap now handles the injection of CA certificates.
	return resourceapply.ApplyConfigMap(context.TODO(), client, recorder, configMap)
}

func manageOpenShiftControllerManagerDeployment_v311_00_to_latest(
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
	required := resourceread.ReadDeploymentV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/deploy.yaml"))

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
	required := resourceread.ReadDeploymentV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/route-controller-deploy.yaml"))

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
