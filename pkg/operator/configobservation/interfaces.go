package configobservation

import (
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
)

type Listers struct {
	ImageConfigLister  configlistersv1.ImageLister
	BuildConfigLister  configlistersv1.BuildLister
	ConfigMapLister    corelistersv1.ConfigMapLister
	PreRunCachesSynced []cache.InformerSynced
}

func (l Listers) ResourceSyncer() resourcesynccontroller.ResourceSyncer {
	return nil
}

func (l Listers) PreRunHasSynced() []cache.InformerSynced {
	return l.PreRunCachesSynced
}
