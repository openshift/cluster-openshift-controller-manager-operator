package operator

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	operatorapiv1 "github.com/openshift/api/operator/v1"
	configinformerv1 "github.com/openshift/client-go/config/informers/externalversions/config/v1"
	proxyvclient1 "github.com/openshift/client-go/config/listers/config/v1"
	operatorclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorinformersv1 "github.com/openshift/client-go/operator/informers/externalversions/operator/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/management"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

const (
	workQueueKey              = "key"
	workloadDegradedCondition = "WorkloadDegraded"
)

type OpenShiftControllerManagerOperator struct {
	targetImagePullSpec  string
	operatorConfigClient operatorclientv1.OperatorV1Interface
	proxyLister          proxyvclient1.ProxyLister

	kubeClient kubernetes.Interface

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface

	rateLimiter flowcontrol.RateLimiter
	recorder    events.Recorder
}

func NewOpenShiftControllerManagerOperator(
	targetImagePullSpec string,
	operatorConfigInformer operatorinformersv1.OpenShiftControllerManagerInformer,
	proxyInformer configinformerv1.ProxyInformer,
	kubeInformersForOpenshiftControllerManager informers.SharedInformerFactory,
	operatorConfigClient operatorclientv1.OperatorV1Interface,
	kubeClient kubernetes.Interface,
	recorder events.Recorder,
) *OpenShiftControllerManagerOperator {
	c := &OpenShiftControllerManagerOperator{
		targetImagePullSpec:  targetImagePullSpec,
		operatorConfigClient: operatorConfigClient,
		proxyLister:          proxyInformer.Lister(),
		kubeClient:           kubeClient,
		queue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "KubeApiserverOperator"),
		rateLimiter:          flowcontrol.NewTokenBucketRateLimiter(0.05 /*3 per minute*/, 4),
		recorder:             recorder,
	}

	operatorConfigInformer.Informer().AddEventHandler(c.eventHandler())
	proxyInformer.Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Core().V1().ServiceAccounts().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Core().V1().Services().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftControllerManager.Apps().V1().Deployments().Informer().AddEventHandler(c.eventHandler())

	// we only watch some namespaces
	kubeInformersForOpenshiftControllerManager.Core().V1().Namespaces().Informer().AddEventHandler(c.namespaceEventHandler())

	// set this bit so the library-go code knows we opt-out from supporting the "unmanaged" state.
	management.SetOperatorAlwaysManaged()
	// set this bit so the library-go code know we do not support removing of the operand
	management.SetOperatorNotRemovable()

	return c
}

func (c OpenShiftControllerManagerOperator) sync() error {
	operatorConfig, err := c.operatorConfigClient.OpenShiftControllerManagers().Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		return err
	}
	// manage status
	originalOperatorConfig := operatorConfig.DeepCopy()
	reasonString := ""
	messageString := ""
	switch operatorConfig.Spec.ManagementState {
	case operatorapiv1.Removed:
		fallthrough
		// we equally do not allow removed/unmanaged
	case operatorapiv1.Unmanaged:
		reasonString = fmt.Sprintf("%sUnsupported", string(operatorConfig.Spec.ManagementState))
		messageString = fmt.Sprintf("the controller manager spec was set to %s state, but that is unsupported, and has no effect on this condition", string(operatorConfig.Spec.ManagementState))

		// as we are ignoring / not supporting unmanaged/removed, we still process any other inputs to the sync/reconciliation
		// unlike what we used to do
	case operatorapiv1.Managed:
		// we want to empty out the reason/message string if transitioning from the other phases so the default setting above is good
	}

	for _, condition := range operatorConfig.Status.Conditions {
		// do not change the current status as part of noting that unmanaged/removed are not supported
		condition.Reason = reasonString
		condition.Message = messageString
		v1helpers.SetOperatorCondition(&operatorConfig.Status.Conditions, condition)
	}
	if !equality.Semantic.DeepEqual(operatorConfig.Status, originalOperatorConfig.Status) {
		if _, err := c.operatorConfigClient.OpenShiftControllerManagers().UpdateStatus(context.TODO(), operatorConfig, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	forceRequeue, err := syncOpenShiftControllerManager_v311_00_to_latest(c, operatorConfig)
	if forceRequeue && err != nil {
		c.queue.AddRateLimited(workQueueKey)
	}

	return err
}

// Run starts the openshift-controller-manager and blocks until stopCh is closed.
func (c *OpenShiftControllerManagerOperator) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting OpenShiftControllerManagerOperator")
	defer klog.Infof("Shutting down OpenShiftControllerManagerOperator")

	// doesn't matter what workers say, only start one.
	go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	<-ctx.Done()
}

func (c *OpenShiftControllerManagerOperator) runWorker(_ context.Context) {
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

func (c *OpenShiftControllerManagerOperator) namespaceEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				c.queue.Add(workQueueKey)
			}
			if ns.Name == util.TargetNamespace {
				c.queue.Add(workQueueKey)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			ns, ok := old.(*corev1.Namespace)
			if !ok {
				c.queue.Add(workQueueKey)
			}
			if ns.Name == util.TargetNamespace {
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
			if ns.Name == util.TargetNamespace {
				c.queue.Add(workQueueKey)
			}
		},
	}
}
