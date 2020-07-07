package usercaobservation

import (
	"context"
	"time"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/management"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

// Controller watches the cluster proxy config resource to see if a custom trusted CA has been
// added or removed. In the event a change is detected, this Controller makes appropriate calls to
// the provided ResourceSyncer instance.
type Controller struct {
	name                 string
	operatorConfigClient v1helpers.OperatorClient
	proxyLister          configlistersv1.ProxyLister
	resourceSyncer       resourcesynccontroller.ResourceSyncer
	runFn                func(ctx context.Context, workers int)
	syncCtxt             factory.SyncContext
}

// NewController creates a new usercaobservation.Controller instance.
func NewController(operatorConfigClient v1helpers.OperatorClient,
	configInformers configinformers.SharedInformerFactory,
	resourceSyncer resourcesynccontroller.ResourceSyncer,
	eventRecorder events.Recorder) *Controller {
	c := &Controller{
		name:                 "UserCAObservationController",
		operatorConfigClient: operatorConfigClient,
		proxyLister:          configInformers.Config().V1().Proxies().Lister(),
		resourceSyncer:       resourceSyncer,
	}
	informers := []factory.Informer{
		operatorConfigClient.Informer(),
		configInformers.Config().V1().Proxies().Informer(),
	}
	f := factory.New().
		WithSync(c.Sync).
		WithSyncContext(c.syncCtxt).
		WithInformers(informers...).
		ResyncEvery(10*time.Minute).
		ToController(c.name, eventRecorder.WithComponentSuffix("user-ca-observation-controller"))
	c.runFn = f.Run
	return c
}

// Run starts the controller with the provided context and number of workers.
func (c *Controller) Run(ctx context.Context, workers int) {
	c.runFn(ctx, workers)
}

// Sync runs the main synchronization logic for the controller.
func (c *Controller) Sync(ctx context.Context, syncCtx factory.SyncContext) error {
	operatorSpec, _, _, err := c.operatorConfigClient.GetOperatorState()
	if err != nil {
		return nil
	}

	if !management.IsOperatorManaged(operatorSpec.ManagementState) {
		return nil
	}

	// Bug 1826183: copy the proxy CA trust bundle to the openshift-controller-manager namespace.
	// If this ConfigMap exists, the build controller will copy the contents into a ConfigMap for
	// the build pod and mount it into each build pod container. The build pod containers will then
	// run `update-ca-trust extract` on startup, merging the proxy CA with the default trust bundle
	// provided by the openshift/builder image.
	//
	// If `source` returns an empty instance, the resourceSyncer will delete the destination
	// ConfigMap. The build controller will expect this and mount an empty file in this situation.
	source, err := c.findProxyCASource()
	if err != nil {
		return err
	}
	destination := resourcesynccontroller.ResourceLocation{
		Namespace: util.TargetNamespace,
		Name:      "openshift-user-ca",
	}
	err = c.resourceSyncer.SyncConfigMap(destination, source)
	return err
}

func (c *Controller) findProxyCASource() (resourcesynccontroller.ResourceLocation, error) {
	source := resourcesynccontroller.ResourceLocation{}
	proxy, err := c.proxyLister.Get("cluster")
	if errors.IsNotFound(err) {
		return source, nil
	}
	if err != nil {
		return source, err
	}
	if len(proxy.Spec.TrustedCA.Name) > 0 {
		source = resourcesynccontroller.ResourceLocation{
			Namespace: util.UserSpecifiedGlobalConfigNamespace,
			Name:      proxy.Spec.TrustedCA.Name,
		}
	}
	return source, nil
}
