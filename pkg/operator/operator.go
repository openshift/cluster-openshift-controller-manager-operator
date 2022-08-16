package operator

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

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

// nodeCountFunction a function to return count of nodes
type nodeCountFunc func(nodeSelector map[string]string) (*int32, error)

// ensureAtMostOnePodPerNode a function that updates the deployment spec to prevent more than
// one pod of a given replicaset from landing on a node.
type ensureAtMostOnePodPerNodeFunc func(spec *appsv1.DeploymentSpec, component string) error

type OpenShiftControllerManagerOperator struct {
	targetImagePullSpec  string
	operatorConfigClient operatorclientv1.OperatorV1Interface
	proxyLister          proxyvclient1.ProxyLister

	kubeClient       kubernetes.Interface
	configMapsGetter coreclientv1.ConfigMapsGetter

	// countNodes a function to return count of nodes on which the workload will be installed
	countNodes nodeCountFunc

	// ensureAtMostOnePodPerNode a function that updates the deployment spec to prevent more than
	// one pod of a given replicaset from landing on a node.
	ensureAtMostOnePodPerNode ensureAtMostOnePodPerNodeFunc

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface

	rateLimiter flowcontrol.RateLimiter
	recorder    events.Recorder
}

func NewOpenShiftControllerManagerOperator(
	targetImagePullSpec string,
	operatorConfigInformer operatorinformersv1.OpenShiftControllerManagerInformer,
	proxyInformer configinformerv1.ProxyInformer,
	kubeInformers v1helpers.KubeInformersForNamespaces,
	operatorConfigClient operatorclientv1.OperatorV1Interface,
	countNodes nodeCountFunc,
	ensureAtMostOnePodPerNode ensureAtMostOnePodPerNodeFunc,
	kubeClient kubernetes.Interface,
	recorder events.Recorder,
) *OpenShiftControllerManagerOperator {
	c := &OpenShiftControllerManagerOperator{
		targetImagePullSpec:       targetImagePullSpec,
		operatorConfigClient:      operatorConfigClient,
		countNodes:                countNodes,
		ensureAtMostOnePodPerNode: ensureAtMostOnePodPerNode,
		proxyLister:               proxyInformer.Lister(),
		kubeClient:                kubeClient,
		configMapsGetter:          v1helpers.CachedConfigMapGetter(kubeClient.CoreV1(), kubeInformers),
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "KubeApiserverOperator"),
		rateLimiter:               flowcontrol.NewTokenBucketRateLimiter(0.05 /*3 per minute*/, 4),
		recorder:                  recorder,
	}

	operatorConfigInformer.Informer().AddEventHandler(c.eventHandler())
	proxyInformer.Informer().AddEventHandler(c.eventHandler())

	targetInformers := kubeInformers.InformersFor(util.TargetNamespace)

	targetInformers.Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	targetInformers.Core().V1().ServiceAccounts().Informer().AddEventHandler(c.eventHandler())
	targetInformers.Core().V1().Services().Informer().AddEventHandler(c.eventHandler())
	targetInformers.Apps().V1().Deployments().Informer().AddEventHandler(c.eventHandler())

	// we only watch some namespaces
	targetInformers.Core().V1().Namespaces().Informer().AddEventHandler(c.namespaceEventHandler(util.TargetNamespace))

	rcTargetInformers := kubeInformers.InformersFor(util.RouteControllerTargetNamespace)

	rcTargetInformers.Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	rcTargetInformers.Core().V1().ServiceAccounts().Informer().AddEventHandler(c.eventHandler())
	rcTargetInformers.Core().V1().Services().Informer().AddEventHandler(c.eventHandler())
	rcTargetInformers.Apps().V1().Deployments().Informer().AddEventHandler(c.eventHandler())

	// we only watch some namespaces
	rcTargetInformers.Core().V1().Namespaces().Informer().AddEventHandler(c.namespaceEventHandler(util.RouteControllerTargetNamespace))

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

	forceRequeue, err := syncOpenShiftControllerManager_v311_00_to_latest(c, operatorConfig, c.countNodes, c.ensureAtMostOnePodPerNode)
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

func (c *OpenShiftControllerManagerOperator) namespaceEventHandler(targetNamespace string) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				c.queue.Add(workQueueKey)
			}
			if ns.Name == targetNamespace {
				c.queue.Add(workQueueKey)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			ns, ok := old.(*corev1.Namespace)
			if !ok {
				c.queue.Add(workQueueKey)
			}
			if ns.Name == targetNamespace {
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
			if ns.Name == targetNamespace {
				c.queue.Add(workQueueKey)
			}
		},
	}
}
