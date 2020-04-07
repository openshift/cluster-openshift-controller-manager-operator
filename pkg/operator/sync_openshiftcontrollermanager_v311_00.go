package operator

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

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
func syncOpenShiftControllerManager_v311_00_to_latest(c OpenShiftControllerManagerOperator, originalOperatorConfig *operatorapiv1.OpenShiftControllerManager) (bool, error) {
	errors := []error{}
	var err error
	operatorConfig := originalOperatorConfig.DeepCopy()
	clientHolder := resourceapply.NewKubeClientHolder(c.kubeClient)
	directResourceResults := resourceapply.ApplyDirectly(clientHolder, c.recorder, v311_00_assets.Asset,
		"v3.11.0/openshift-controller-manager/informer-clusterrole.yaml",
		"v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml",
		"v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml",
		"v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml",
		"v3.11.0/openshift-controller-manager/leader-role.yaml",
		"v3.11.0/openshift-controller-manager/leader-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/old-leader-role.yaml",
		"v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/separate-sa-role.yaml",
		"v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/sa.yaml",
		"v3.11.0/openshift-controller-manager/svc.yaml",
		"v3.11.0/openshift-controller-manager/servicemonitor-role.yaml",
		"v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml",
	)
	resourcesThatForceRedeployment := sets.NewString("v3.11.0/openshift-controller-manager/sa.yaml")
	forceRollout := false

	for _, currResult := range directResourceResults {
		if currResult.Error != nil {
			errors = append(errors, fmt.Errorf("%q (%T): %v", currResult.File, currResult.Type, currResult.Error))
			continue
		}

		if currResult.Changed && resourcesThatForceRedeployment.Has(currResult.File) {
			forceRollout = true
		}
	}

	_, configMapModified, err := manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(c.kubeClient, c.kubeClient.CoreV1(), c.recorder, operatorConfig)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "configmap", err))
	}
	// the kube-apiserver is the source of truth for client CA bundles
	clientCAModified, err := manageOpenShiftControllerManagerClientCA_v311_00_to_latest(c.kubeClient.CoreV1(), c.recorder)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "client-ca", err))
	}

	_, serviceCAModified, err := manageOpenShiftServiceCAConfigMap_v311_00_to_latest(c.kubeClient, c.kubeClient.CoreV1(), c.recorder)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "openshift-service-ca", err))
	}

	_, globalCAModified, err := manageOpenShiftGlobalCAConfigMap_v311_00_to_latest(c.kubeClient, c.kubeClient.CoreV1(), c.recorder)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "openshift-global-ca", err))
	}

	forceRollout = forceRollout || operatorConfig.ObjectMeta.Generation != operatorConfig.Status.ObservedGeneration
	forceRollout = forceRollout || configMapModified || clientCAModified || serviceCAModified || globalCAModified

	// our configmaps and secrets are in order, now it is time to create the DS
	// TODO check basic preconditions here
	actualDaemonSet, _, err := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(c.kubeClient.AppsV1(), c.recorder, operatorConfig, c.targetImagePullSpec, operatorConfig.Status.Generations, forceRollout, c.proxyLister)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "deployment", err))
	}

	// manage status
	if actualDaemonSet.Status.NumberAvailable > 0 {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:   operatorapiv1.OperatorStatusTypeAvailable,
			Status: operatorapiv1.ConditionTrue,
		})
	} else {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:    operatorapiv1.OperatorStatusTypeAvailable,
			Status:  operatorapiv1.ConditionFalse,
			Reason:  "NoPodsAvailable",
			Message: "no daemon pods available on any node.",
		})
	}
	if actualDaemonSet.Status.NumberAvailable > 0 && actualDaemonSet.Status.UpdatedNumberScheduled == actualDaemonSet.Status.DesiredNumberScheduled {
		if len(actualDaemonSet.Annotations[util.VersionAnnotation]) > 0 {
			operatorConfig.Status.Version = actualDaemonSet.Annotations[util.VersionAnnotation]
		}
	}

	var progressingMessages []string
	if actualDaemonSet != nil && actualDaemonSet.ObjectMeta.Generation != actualDaemonSet.Status.ObservedGeneration {
		progressingMessages = append(progressingMessages, fmt.Sprintf("daemonset/controller-manager: observed generation is %d, desired generation is %d.", actualDaemonSet.Status.ObservedGeneration, actualDaemonSet.ObjectMeta.Generation))
	}
	if actualDaemonSet.Status.NumberAvailable == 0 {
		progressingMessages = append(progressingMessages, fmt.Sprintf("daemonset/controller-manager: number available is %d, desired number available > %d", actualDaemonSet.Status.NumberAvailable, 1))
	}
	if actualDaemonSet.Status.UpdatedNumberScheduled != actualDaemonSet.Status.DesiredNumberScheduled {
		progressingMessages = append(progressingMessages, fmt.Sprintf("daemonset/controller-manager: updated number scheduled is %d, desired number scheduled is %d", actualDaemonSet.Status.UpdatedNumberScheduled, actualDaemonSet.Status.DesiredNumberScheduled))
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

	operatorConfig.Status.ObservedGeneration = operatorConfig.ObjectMeta.Generation
	resourcemerge.SetDaemonSetGeneration(&operatorConfig.Status.Generations, actualDaemonSet)

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

func manageOpenShiftControllerManagerClientCA_v311_00_to_latest(client coreclientv1.CoreV1Interface, recorder events.Recorder) (bool, error) {
	const apiserverClientCA = "client-ca"
	_, caChanged, err := resourceapply.SyncConfigMap(client, recorder, util.KubeAPIServerNamespace, apiserverClientCA, util.TargetNamespace, apiserverClientCA, []metav1.OwnerReference{})
	if err != nil {
		return false, err
	}
	return caChanged, nil
}

func manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder, operatorConfig *operatorapiv1.OpenShiftControllerManager) (*corev1.ConfigMap, bool, error) {
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
	)
	if err != nil {
		return nil, false, err
	}
	for k, v := range inputHashes {
		requiredConfigMap.Data[k] = v
	}

	return resourceapply.ApplyConfigMap(client, recorder, requiredConfigMap)
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

func manageOpenShiftGlobalCAConfigMap_v311_00_to_latest(kubeClient kubernetes.Interface, client coreclientv1.ConfigMapsGetter, recorder events.Recorder) (*corev1.ConfigMap, bool, error) {
	configMap := resourceread.ReadConfigMapV1OrDie(v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/openshift-global-ca-cm.yaml"))
	existing, err := client.ConfigMaps(util.TargetNamespace).Get(context.TODO(), "openshift-global-ca", metav1.GetOptions{})
	// Ensure we create the ConfigMap for the global CA, and that it has the right labels
	// Lifted from library-go for the most part
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

func manageOpenShiftControllerManagerDeployment_v311_00_to_latest(client appsclientv1.DaemonSetsGetter, recorder events.Recorder, options *operatorapiv1.OpenShiftControllerManager, imagePullSpec string, generationStatus []operatorapiv1.GenerationStatus, forceRollout bool, proxyLister proxyvclient1.ProxyLister) (*appsv1.DaemonSet, bool, error) {
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

	return resourceapply.ApplyDaemonSet(client, recorder, required, resourcemerge.ExpectedDaemonSetGeneration(required, generationStatus), forceRollout)
}
