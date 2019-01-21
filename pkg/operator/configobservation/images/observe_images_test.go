package images

import (
	"testing"

	"github.com/openshift/library-go/pkg/operator/events"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

func TestObserveRegistryConfig(t *testing.T) {
	const (
		expectedInternalRegistryHostname = "docker-registry.openshift-image-registry.svc.cluster.local:5000"
	)

	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	imageConfig := &configv1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: configv1.ImageStatus{
			InternalRegistryHostname: expectedInternalRegistryHostname,
		},
	}
	indexer.Add(imageConfig)

	listers := configobservation.Listers{
		ImageConfigLister: configlistersv1.NewImageLister(indexer),
		ImageConfigSynced: func() bool { return true },
	}

	result, errs := ObserveInternalRegistryHostname(listers, events.NewInMemoryRecorder(""), map[string]interface{}{})
	if len(errs) != 0 {
		t.Errorf("expected no errors: %v", errs)
	}
	internalRegistryHostname, _, err := unstructured.NestedString(result, "dockerPullSecret", "internalRegistryHostname")
	if err != nil {
		t.Fatal(err)
	}
	if internalRegistryHostname != expectedInternalRegistryHostname {
		t.Errorf("expected internal registry hostname: %s, got %s", expectedInternalRegistryHostname, internalRegistryHostname)
	}
}
