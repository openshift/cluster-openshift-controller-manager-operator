package operator

import (
	"fmt"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/workqueue"

	operatorsv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	operatorconfigclientv1alpha1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned/typed/openshiftcontrollermanager/v1alpha1"
	operatorconfiginformerv1alpha1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/informers/externalversions/openshiftcontrollermanager/v1alpha1"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1alpha1helpers"
	"github.com/openshift/library-go/pkg/operator/versioning"
)

const (
	kubeAPIServerNamespaceName = "openshift-kube-apiserver"
	targetNamespaceName        = "openshift-controller-manager"
	operatorNamespaceName      = "openshift-core-operators"
	workQueueKey               = "key"
)

type OpenShiftControllerManagerOperator struct {
	operatorConfigClient operatorconfigclientv1alpha1.OpenshiftcontrollermanagerV1alpha1Interface

	kubeClient kubernetes.Interface

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface

	rateLimiter flowcontrol.RateLimiter
	recorder    events.Recorder
}

func NewOpenShiftControllerManagerOperator(
	operatorConfigInformer operatorconfiginformerv1alpha1.OpenShiftControllerManagerOperatorConfigInformer,
	kubeInformersForOpenshiftControllerManager informers.SharedInformerFactory,
	operatorConfigClient operatorconfigclientv1alpha1.OpenshiftcontrollermanagerV1alpha1Interface,
	kubeClient kubernetes.Interface,
	recorder events.Recorder,
) *OpenShiftControllerManagerOperator {
	c := &OpenShiftControllerManagerOperator{
		operatorConfigClient: operatorConfigClient,
		kubeClient:           kubeClient,
		queue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "KubeApiserverOperator"),
		rateLimiter:          flowcontrol.NewTokenBucketRateLimiter(0.05 /*3 per minute*/, 4),
		recorder:             recorder,
	}

	operatorConfigInformer.Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Core().V1().ServiceAccounts().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Core().V1().Services().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Apps().V1().Deployments().Informer().AddEventHandler(c.eventHandler())

	// we only watch some namespaces
	kubeInformersForOpenshiftControllerManager.Core().V1().Namespaces().Informer().AddEventHandler(c.namespaceEventHandler())

	return c
}

func (c OpenShiftControllerManagerOperator) sync() error {
	operatorConfig, err := c.operatorConfigClient.OpenShiftControllerManagerOperatorConfigs().Get("instance", metav1.GetOptions{})
	if err != nil {
		return err
	}
	switch operatorConfig.Spec.ManagementState {
	case operatorsv1alpha1.Unmanaged:
		return nil

	case operatorsv1alpha1.Removed:
		// TODO probably need to watch until the NS is really gone
		if err := c.kubeClient.CoreV1().Namespaces().Delete(targetNamespaceName, nil); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		operatorConfig.Status.TaskSummary = "Remove"
		operatorConfig.Status.TargetAvailability = nil
		operatorConfig.Status.CurrentAvailability = nil
		operatorConfig.Status.Conditions = []operatorsv1alpha1.OperatorCondition{
			{
				Type:   operatorsv1alpha1.OperatorStatusTypeAvailable,
				Status: operatorsv1alpha1.ConditionFalse,
			},
		}
		if _, err := c.operatorConfigClient.OpenShiftControllerManagerOperatorConfigs().Update(operatorConfig); err != nil {
			return err
		}
		return nil
	}

	// block until config is obvserved
	if len(operatorConfig.Spec.ObservedConfig.Raw) == 0 {
		glog.Info("Waiting for observed configuration to be available")
		return nil
	}

	var currentActualVerion *semver.Version

	if operatorConfig.Status.CurrentAvailability != nil {
		ver, err := semver.Parse(operatorConfig.Status.CurrentAvailability.Version)
		if err != nil {
			utilruntime.HandleError(err)
		} else {
			currentActualVerion = &ver
		}
	}
	desiredVersion, err := semver.Parse(operatorConfig.Spec.Version)
	if err != nil {
		// TODO report failing status, we may actually attempt to do this in the "normal" error handling
		return err
	}

	v311_00_to_unknown := versioning.NewRangeOrDie("3.11.0", "3.12.0")

	var versionAvailability operatorsv1alpha1.VersionAvailability
	errors := []error{}
	switch {
	case v311_00_to_unknown.BetweenOrEmpty(currentActualVerion) && v311_00_to_unknown.Between(&desiredVersion):
		operatorConfig.Status.TaskSummary = "sync-[3.11.0,3.12.0)"
		operatorConfig.Status.TargetAvailability = nil
		versionAvailability, errors = syncOpenShiftControllerManager_v311_00_to_latest(c, operatorConfig, operatorConfig.Status.CurrentAvailability)

	default:
		operatorConfig.Status.TaskSummary = "unrecognized"
		if _, err := c.operatorConfigClient.OpenShiftControllerManagerOperatorConfigs().UpdateStatus(operatorConfig); err != nil {
			utilruntime.HandleError(err)
		}

		return fmt.Errorf("unrecognized state")
	}

	v1alpha1helpers.SetStatusFromAvailability(&operatorConfig.Status.OperatorStatus, operatorConfig.ObjectMeta.Generation, &versionAvailability)
	if _, err := c.operatorConfigClient.OpenShiftControllerManagerOperatorConfigs().UpdateStatus(operatorConfig); err != nil {
		errors = append(errors, err)
	}

	return utilerrors.NewAggregate(errors)
}

// Run starts the openshift-controller-manager and blocks until stopCh is closed.
func (c *OpenShiftControllerManagerOperator) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Infof("Starting OpenShiftControllerManagerOperator")
	defer glog.Infof("Shutting down OpenShiftControllerManagerOperator")

	// doesn't matter what workers say, only start one.
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *OpenShiftControllerManagerOperator) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *OpenShiftControllerManagerOperator) processNextWorkItem() bool {
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
func (c *OpenShiftControllerManagerOperator) eventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.queue.Add(workQueueKey) },
	}
}

// this set of namespaces will include things like logging and metrics which are used to drive
var interestingNamespaces = sets.NewString(targetNamespaceName)

func (c *OpenShiftControllerManagerOperator) namespaceEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				c.queue.Add(workQueueKey)
			}
			if ns.Name == targetNamespaceName {
				c.queue.Add(workQueueKey)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			ns, ok := old.(*corev1.Namespace)
			if !ok {
				c.queue.Add(workQueueKey)
			}
			if ns.Name == targetNamespaceName {
				c.queue.Add(workQueueKey)
			}
		},
		DeleteFunc: func(obj interface{}) {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
					return
				}
				ns, ok = tombstone.Obj.(*corev1.Namespace)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a Namespace %#v", obj))
					return
				}
			}
			if ns.Name == targetNamespaceName {
				c.queue.Add(workQueueKey)
			}
		},
	}
}
