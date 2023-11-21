package configobservercontroller

import (
	"k8s.io/apimachinery/pkg/util/sets"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	configv1 "github.com/openshift/api/config/v1"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorv1informers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/builds"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/controllers"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/deployimages"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/images"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/network"
)

// NewConfigObserver initializes a new configuration observer.
func NewConfigObserver(
	operatorClient v1helpers.OperatorClient,
	operatorConfigInformers operatorv1informers.SharedInformerFactory,
	configInformers configinformers.SharedInformerFactory,
	kubeInformersForOperatorNamespace kubeinformers.SharedInformerFactory,
	featureGateAccessor featuregates.FeatureGateAccess,
	eventRecorder events.Recorder,
	buildEnabled bool,
) factory.Controller {
	informersSynced := []cache.InformerSynced{
		configInformers.Config().V1().Images().Informer().HasSynced,
		configInformers.Config().V1().Networks().Informer().HasSynced,
		configInformers.Config().V1().ClusterVersions().Informer().HasSynced,
		configInformers.Config().V1().ClusterOperators().Informer().HasSynced,
		kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Informer().HasSynced,
		operatorConfigInformers.Operator().V1().OpenShiftControllerManagers().Informer().HasSynced,
	}

	if buildEnabled {
		informersSynced = append(informersSynced, configInformers.Config().V1().Builds().Informer().HasSynced)
	}

	configObservationListers := configobservation.Listers{
		ImageConfigLister:     configInformers.Config().V1().Images().Lister(),
		NetworkLister:         configInformers.Config().V1().Networks().Lister(),
		FeatureGateLister_:    configInformers.Config().V1().FeatureGates().Lister(),
		ClusterVersionLister:  configInformers.Config().V1().ClusterVersions().Lister(),
		ClusterOperatorLister: configInformers.Config().V1().ClusterOperators().Lister(),
		ConfigMapLister:       kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Lister(),
		PreRunCachesSynced:    informersSynced,
	}

	observerFuncs := []configobserver.ObserveConfigFunc{
		images.ObserveInternalRegistryHostname,
		images.ObserveExternalRegistryHostnames,
		network.ObserveExternalIPAutoAssignCIDRs,
		deployimages.ObserveControllerManagerImagesConfig,
		controllers.ObserveControllers,
		featuregates.NewObserveFeatureFlagsFunc(
			sets.New[configv1.FeatureGateName]("BuildCSIVolumes"),
			nil,
			[]string{"featureGates"},
			featureGateAccessor,
		),
	}

	if buildEnabled {
		configObservationListers.BuildConfigLister = configInformers.Config().V1().Builds().Lister()
		observerFuncs = append(observerFuncs, builds.ObserveBuildControllerConfig)
	}

	c := configobserver.NewConfigObserver(
		operatorClient,
		eventRecorder,
		configObservationListers,
		[]factory.Informer{operatorConfigInformers.Operator().V1().OpenShiftControllerManagers().Informer()},
		observerFuncs...,
	)

	return c
}
