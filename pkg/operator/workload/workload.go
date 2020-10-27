package workload

import (
	"errors"
	"fmt"
	"strings"

	operatorapiv1 "github.com/openshift/api/operator/v1"
	proxyclientv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

// Workload reprents the operator components and the approach taken for resources deployment. The
// rollout sequence keeps all errors in a single slice, where some errors require different handling.
type Workload struct {
	targetImagePullSpec string                    // daemon-set image pull-spec
	proxyLister         proxyclientv1.ProxyLister // proxy client lister
	kubeClient          kubernetes.Interface      // kubernetes client
	recorder            events.Recorder           // event recorder
}

// WorkloadDegradedCondition status condition to denote manage workload is degraded.
const WorkloadDegradedCondition = "WorkloadDegraded"

var (
	// ErrForceRollout certain resource changes should force the daemon-set rollout.
	ErrForceRollout = errors.New("force rollout")

	// ErrDaemonSetNotFound when target daemon-set can't be loaded from k8s api, all subsequent
	// actions against daemon-set can be skipped.
	ErrDaemonSetNotFound = errors.New("daemonset not found")
)

var (
	// directResources all resources managed by this operator.
	directResources = []string{
		"v3.11.0/openshift-controller-manager/informer-clusterrole.yaml",
		"v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml",
		"v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml",
		"v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml",
		"v3.11.0/openshift-controller-manager/leader-role.yaml",
		"v3.11.0/openshift-controller-manager/leader-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/ns.yaml",
		"v3.11.0/openshift-controller-manager/old-leader-role.yaml",
		"v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/separate-sa-role.yaml",
		"v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml",
		"v3.11.0/openshift-controller-manager/sa.yaml",
		"v3.11.0/openshift-controller-manager/svc.yaml",
		"v3.11.0/openshift-controller-manager/servicemonitor-role.yaml",
		"v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml",
	}
	// forceRolloutResources resources that will force daemon-set redeployment when change.
	forceRolloutResources = []string{
		"v3.11.0/openshift-controller-manager/sa.yaml",
	}
)

// manageConfigMap execute the informed interface function, managing modified state and errors.
func (w *Workload) manageConfigMap(
	errSlice []error,
	name string, fn manageConfigMapFn,
	operatorConfig *operatorapiv1.OpenShiftControllerManager,
) []error {
	_, modified, err := fn(w.kubeClient, w.kubeClient.CoreV1(), w.recorder, operatorConfig)
	return manageError(errSlice, name, err, modified)
}

// manageClientCA execute the informed interface function, managing modified state and errors.
func (w *Workload) manageClientCA(errSlice []error, name string, fn manageClientCAFn) []error {
	modified, err := fn(w.kubeClient.CoreV1(), w.recorder)
	return manageError(errSlice, name, err, modified)
}

// setupDirectResources apply the resources managed by this operator.
func (w *Workload) setupDirectResources() []error {
	ch := resourceapply.NewKubeClientHolder(w.kubeClient)
	results := resourceapply.ApplyDirectly(ch, w.recorder, v311_00_assets.Asset, directResources...)
	forceRedeploymentResource := sets.NewString(forceRolloutResources...)
	errSlice := []error{}
	for _, r := range results {
		if r.Error != nil {
			errSlice = append(errSlice, fmt.Errorf("%q (%T): %v", r.File, r.Type, r.Error))
			continue
		}
		// checking if the modified resource triggers a forceful re-deployment
		if r.Changed && forceRedeploymentResource.Has(r.File) {
			errSlice = append(errSlice, fmt.Errorf("%w: %q (%T)", ErrForceRollout, r.File, r.Type))
		}
	}
	return errSlice
}

// manageDaemonSet rollout daemon-set resource, and inspecting installed resource to create rollout
// progress messages. It also updates the operator configuration with observed generation.
func (w *Workload) manageDaemonSet(
	errSlice []error,
	operatorConfig *operatorapiv1.OpenShiftControllerManager,
) []error {
	progress := []string{}
	forceRollout := hasError(errSlice, ErrForceRollout)

	ds, _, err := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(
		w.kubeClient.AppsV1(),
		w.recorder,
		operatorConfig,
		w.targetImagePullSpec,
		operatorConfig.Status.Generations,
		forceRollout,
		w.proxyLister,
	)
	if err != nil {
		manageError(errSlice, "deployment", err, false)
		progress = append(progress, fmt.Sprintf("deployment: %v", err))
	}

	if ds == nil {
		return append(errSlice, fmt.Errorf("%w: unable to find daemon-set", ErrDaemonSetNotFound))
	}

	if ds.Status.NumberAvailable > 0 {
		v1helpers.SetOperatorCondition(
			&operatorConfig.Status.Conditions,
			operatorapiv1.OperatorCondition{
				Type:   operatorapiv1.OperatorStatusTypeAvailable,
				Status: operatorapiv1.ConditionTrue,
			},
		)
	} else {
		v1helpers.SetOperatorCondition(
			&operatorConfig.Status.Conditions,
			operatorapiv1.OperatorCondition{
				Type:    operatorapiv1.OperatorStatusTypeAvailable,
				Status:  operatorapiv1.ConditionFalse,
				Reason:  "NoPodsAvailable",
				Message: "no daemon pods available on any node.",
			},
		)
	}

	if ds.Status.UpdatedNumberScheduled == ds.Status.DesiredNumberScheduled {
		if len(ds.Annotations[util.VersionAnnotation]) > 0 {
			operatorConfig.Status.Version = ds.Annotations[util.VersionAnnotation]
		} else {
			msg := fmt.Sprintf("daemonset/controller-manager: version annotation %s missing.",
				util.VersionAnnotation)
			progress = append(progress, msg)
		}
	}

	if ds.ObjectMeta.Generation != ds.Status.ObservedGeneration {
		msg := fmt.Sprintf(
			"daemonset/controller-manager: observed generation is %d, desired generation is %d.",
			ds.Status.ObservedGeneration,
			ds.ObjectMeta.Generation,
		)
		progress = append(progress, msg)
	}

	if ds.Status.NumberAvailable == 0 {
		msg := fmt.Sprintf(
			"daemonset/controller-manager: number available is %d, desired number available > 1",
			ds.Status.NumberAvailable,
		)
		progress = append(progress, msg)
	}

	if ds.Status.UpdatedNumberScheduled != ds.Status.DesiredNumberScheduled {
		msg := fmt.Sprintf(
			"daemonset/controller-manager: updated number scheduled is %d, desired number scheduled is %d",
			ds.Status.UpdatedNumberScheduled,
			ds.Status.DesiredNumberScheduled,
		)
		progress = append(progress, msg)
	}

	if operatorConfig.ObjectMeta.Generation != operatorConfig.Status.ObservedGeneration {
		msg := fmt.Sprintf(
			"openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is %d, desired generation is %d.",
			operatorConfig.Status.ObservedGeneration,
			operatorConfig.ObjectMeta.Generation,
		)
		progress = append(progress, msg)
	}

	if len(progress) == 0 {
		v1helpers.SetOperatorCondition(
			&operatorConfig.Status.Conditions,
			operatorapiv1.OperatorCondition{
				Type:   operatorapiv1.OperatorStatusTypeProgressing,
				Status: operatorapiv1.ConditionFalse,
			},
		)
	} else {
		v1helpers.SetOperatorCondition(
			&operatorConfig.Status.Conditions,
			operatorapiv1.OperatorCondition{
				Type:    operatorapiv1.OperatorStatusTypeProgressing,
				Status:  operatorapiv1.ConditionTrue,
				Reason:  "DesiredStateNotYetAchieved",
				Message: strings.Join(progress, "\n"),
			},
		)
	}

	if len(errSlice) > 0 {
		message := ""
		for _, err := range errSlice {
			message = message + err.Error() + "\n"
		}
		v1helpers.SetOperatorCondition(
			&operatorConfig.Status.Conditions,
			operatorapiv1.OperatorCondition{
				Type:    WorkloadDegradedCondition,
				Status:  operatorapiv1.ConditionTrue,
				Message: message,
				Reason:  "SyncError",
			},
		)
	} else {
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, operatorapiv1.OperatorCondition{
			Type:   WorkloadDegradedCondition,
			Status: operatorapiv1.ConditionFalse,
		})
	}

	operatorConfig.Status.ObservedGeneration = operatorConfig.ObjectMeta.Generation
	resourcemerge.SetDaemonSetGeneration(&operatorConfig.Status.Generations, ds)

	return errSlice
}

// Sync reconcile all resources managed by this operator.
func (w *Workload) Sync(operatorConfig *operatorapiv1.OpenShiftControllerManager) bool {
	errSlice := w.setupDirectResources()

	errSlice = w.manageConfigMap(
		errSlice,
		"configmap",
		manageOpenShiftControllerManagerConfigMap_v311_00_to_latest,
		operatorConfig,
	)
	errSlice = w.manageClientCA(
		errSlice,
		"client-ca",
		manageOpenShiftControllerManagerClientCA_v311_00_to_latest,
	)
	errSlice = w.manageConfigMap(
		errSlice,
		"openshift-service-ca",
		manageOpenShiftServiceCAConfigMap_v311_00_to_latest,
		operatorConfig,
	)
	errSlice = w.manageConfigMap(
		errSlice,
		"openshift-global-ca",
		manageOpenShiftGlobalCAConfigMap_v311_00_to_latest,
		operatorConfig,
	)

	errSlice = w.manageDaemonSet(errSlice, operatorConfig)

	return len(errSlice) > 0
}

// NewWorkload intantiate Workload.
func NewWorkload(
	targetImagePullSpec string,
	proxyLister proxyclientv1.ProxyLister,
	kubeClient kubernetes.Interface,
	recorder events.Recorder,
) *Workload {
	return &Workload{
		targetImagePullSpec: targetImagePullSpec,
		proxyLister:         proxyLister,
		kubeClient:          kubeClient,
		recorder:            recorder,
	}
}
