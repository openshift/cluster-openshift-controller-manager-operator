package configobservercontroller

import (
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/builds"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/deployimages"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/images"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/network"
)

type ConfigObserver struct {
	*configobserver.ConfigObserver
}

// NewConfigObserver initializes a new configuration observer.
func NewConfigObserver(
	operatorClient v1helpers.OperatorClient,
	configInformers configinformers.SharedInformerFactory,
	kubeInformersForOperatorNamespace kubeinformers.SharedInformerFactory,
	eventRecorder events.Recorder,
) factory.Controller {
	imageInformer := configInformers.Config().V1().Images().Informer()
	buildInformer := configInformers.Config().V1().Builds().Informer()
	networkInformer := configInformers.Config().V1().Networks().Informer()
	configMapInformer := kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Informer()
	configObserverController := configobserver.NewConfigObserver(
		operatorClient,
		eventRecorder,
		configobservation.Listers{
			ImageConfigLister: configInformers.Config().V1().Images().Lister(),
			BuildConfigLister: configInformers.Config().V1().Builds().Lister(),
			NetworkLister:     configInformers.Config().V1().Networks().Lister(),
			ConfigMapLister:   kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Lister(),
			PreRunCachesSynced: []cache.InformerSynced{
				imageInformer.HasSynced,
				buildInformer.HasSynced,
				networkInformer.HasSynced,
				configMapInformer.HasSynced,
				//TODO why is images, builds, configmaps listed twice, but not networks?
				imageInformer.HasSynced,
				buildInformer.HasSynced,
				configMapInformer.HasSynced,
			},
		},
		[]factory.Informer{imageInformer},
		images.ObserveInternalRegistryHostname,
		builds.ObserveBuildControllerConfig,
		network.ObserveExternalIPAutoAssignCIDRs,
		deployimages.ObserveControllerManagerImagesConfig,
	)

	return configObserverController
}
