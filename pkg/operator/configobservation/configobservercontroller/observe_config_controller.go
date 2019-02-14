package configobservercontroller

import (
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/builds"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/deployimages"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/images"
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
) *ConfigObserver {
	c := &ConfigObserver{
		ConfigObserver: configobserver.NewConfigObserver(
			operatorClient,
			eventRecorder,
			configobservation.Listers{
				ImageConfigLister: configInformers.Config().V1().Images().Lister(),
				BuildConfigLister: configInformers.Config().V1().Builds().Lister(),
				ConfigMapLister:   kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Lister(),
				PreRunCachesSynced: []cache.InformerSynced{
					configInformers.Config().V1().Images().Informer().HasSynced,
					configInformers.Config().V1().Builds().Informer().HasSynced,
					kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Informer().HasSynced,
					configInformers.Config().V1().Images().Informer().HasSynced,
					configInformers.Config().V1().Builds().Informer().HasSynced,
					kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Informer().HasSynced,
				},
			},
			images.ObserveInternalRegistryHostname,
			builds.ObserveBuildControllerConfig,
			deployimages.ObserveControllerManagerImagesConfig,
		),
	}

	kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Informer().AddEventHandler(c.EventHandler())
	configInformers.Config().V1().Images().Informer().AddEventHandler(c.EventHandler())
	configInformers.Config().V1().Builds().Informer().AddEventHandler(c.EventHandler())

	return c
}
