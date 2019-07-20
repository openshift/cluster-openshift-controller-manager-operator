package network

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

func TestObserveExternalIPAutoAssignCIDRs(t *testing.T) {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	recorder := events.NewInMemoryRecorder("")

	listers := configobservation.Listers{
		NetworkLister: configlistersv1.NewNetworkLister(indexer),
	}

	configPath := []string{"ingress", "ingressIPNetworkCIDR"}

	expectValue := func(v string, result map[string]interface{}) {
		t.Helper()
		val, _, err := unstructured.NestedString(result, configPath...)
		if err != nil {
			t.Fatal(err)
		}
		if val != v {
			t.Errorf("expected %q, got %q", v, val)
		}
	}

	result, errs := ObserveExternalIPAutoAssignCIDRs(listers, recorder, map[string]interface{}{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors: %v", errs)
	}
	expectValue("", result)

	err := indexer.Add(&configv1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.NetworkSpec{},
	})
	if err != nil {
		t.Fatal(err)
	}

	result, errs = ObserveExternalIPAutoAssignCIDRs(listers, recorder, map[string]interface{}{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors: %v", errs)
	}
	expectValue("", result)

	err = indexer.Update(&configv1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.NetworkSpec{
			ExternalIP: &configv1.ExternalIPConfig{
				AutoAssignCIDRs: []string{"1.2.3.0/24"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	result, errs = ObserveExternalIPAutoAssignCIDRs(listers, recorder, map[string]interface{}{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors: %v", errs)
	}
	expectValue("1.2.3.0/24", result)

	// save this result, make invalid, ensure old result preserved
	oldResult := result
	err = indexer.Update(&configv1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.NetworkSpec{
			ExternalIP: &configv1.ExternalIPConfig{
				AutoAssignCIDRs: []string{"invalid"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	result, errs = ObserveExternalIPAutoAssignCIDRs(listers, recorder, oldResult)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error: %v", errs)
	}
	expectValue("1.2.3.0/24", result)
}
