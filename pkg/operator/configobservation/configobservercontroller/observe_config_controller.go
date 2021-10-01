package configobservercontroller

import (
	"k8s.io/apimachinery/pkg/util/sets"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorv1informers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/builds"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/deployimages"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/images"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/network"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
)

// NewConfigObserver initializes a new configuration observer.
func NewConfigObserver(
	operatorClient v1helpers.OperatorClient,
	operatorConfigInformers operatorv1informers.SharedInformerFactory,
	configInformers configinformers.SharedInformerFactory,
	kubeInformersForOperatorNamespace kubeinformers.SharedInformerFactory,
	eventRecorder events.Recorder,
) factory.Controller {
	c := configobserver.NewConfigObserver(
		operatorClient,
		eventRecorder,
		configobservation.Listers{
			ImageConfigLister:  configInformers.Config().V1().Images().Lister(),
			BuildConfigLister:  configInformers.Config().V1().Builds().Lister(),
			NetworkLister:      configInformers.Config().V1().Networks().Lister(),
			FeatureGateLister_: configInformers.Config().V1().FeatureGates().Lister(),
			ConfigMapLister:    kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Lister(),
			PreRunCachesSynced: []cache.InformerSynced{
				configInformers.Config().V1().Builds().Informer().HasSynced,
				configInformers.Config().V1().Images().Informer().HasSynced,
				configInformers.Config().V1().Networks().Informer().HasSynced,
				kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Informer().HasSynced,
				operatorConfigInformers.Operator().V1().OpenShiftControllerManagers().Informer().HasSynced,
			},
		},
		[]factory.Informer{operatorConfigInformers.Operator().V1().OpenShiftControllerManagers().Informer()},
		images.ObserveInternalRegistryHostname,
		images.ObserveExternalRegistryHostnames,
		builds.ObserveBuildControllerConfig,
		network.ObserveExternalIPAutoAssignCIDRs,
		deployimages.ObserveControllerManagerImagesConfig,
		featuregates.NewObserveFeatureFlagsFunc(sets.NewString("BuildCSIVolumes"), sets.String{}, []string{"featureGates"}),
	)

	return c
}
