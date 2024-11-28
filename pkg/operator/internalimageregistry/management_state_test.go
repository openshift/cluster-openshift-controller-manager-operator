package internalimageregistry

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestImageRegistryIsEnabled(t *testing.T) {
	tests := []struct {
		name            string
		capabilities    configv1.ClusterVersionCapabilitiesStatus
		clusterOperator *configv1.ClusterOperator
		expectEnabled   bool
	}{
		{
			name: "with capability and populated CO status conditions",
			capabilities: configv1.ClusterVersionCapabilitiesStatus{
				EnabledCapabilities: []configv1.ClusterVersionCapability{
					configv1.ClusterVersionCapabilityImageRegistry,
				},
				KnownCapabilities: []configv1.ClusterVersionCapability{
					configv1.ClusterVersionCapabilityImageRegistry,
				},
			},
			clusterOperator: &configv1.ClusterOperator{
				ObjectMeta: metav1.ObjectMeta{
					Name: "image-registry",
				},
				Status: configv1.ClusterOperatorStatus{
					Conditions: []configv1.ClusterOperatorStatusCondition{
						{Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue},
					},
				},
			},
			expectEnabled: true,
		},
		{
			name: "with capability and missing CO status conditions",
			capabilities: configv1.ClusterVersionCapabilitiesStatus{
				EnabledCapabilities: []configv1.ClusterVersionCapability{
					configv1.ClusterVersionCapabilityImageRegistry,
				},
				KnownCapabilities: []configv1.ClusterVersionCapability{
					configv1.ClusterVersionCapabilityImageRegistry,
				},
			},
			clusterOperator: &configv1.ClusterOperator{
				ObjectMeta: metav1.ObjectMeta{
					Name: "image-registry",
				},
				Status: configv1.ClusterOperatorStatus{
					Conditions: []configv1.ClusterOperatorStatusCondition{},
				},
			},
			expectEnabled: true,
		},
		{
			name: "without capability and without CO",
			capabilities: configv1.ClusterVersionCapabilitiesStatus{
				EnabledCapabilities: []configv1.ClusterVersionCapability{},
				KnownCapabilities: []configv1.ClusterVersionCapability{
					configv1.ClusterVersionCapabilityImageRegistry,
				},
			},
			expectEnabled: false,
		},
		{
			// this is an invalid cluster state - when the capability is disabled
			// the operator is not expected to be present in the cluster.
			// see OCPBUGS-35228 and OCPBUGS-43043 for background.
			name: "without capability and populated CO status conditions",
			capabilities: configv1.ClusterVersionCapabilitiesStatus{
				EnabledCapabilities: []configv1.ClusterVersionCapability{},
				KnownCapabilities: []configv1.ClusterVersionCapability{
					configv1.ClusterVersionCapabilityImageRegistry,
				},
			},
			clusterOperator: &configv1.ClusterOperator{
				ObjectMeta: metav1.ObjectMeta{
					Name: "image-registry",
				},
				Status: configv1.ClusterOperatorStatus{
					Conditions: []configv1.ClusterOperatorStatusCondition{
						{Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue},
					},
				},
			},
			expectEnabled: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cv := &configv1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{Name: "version"},
				Status: configv1.ClusterVersionStatus{
					Capabilities: test.capabilities,
				},
			}
			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			indexer.Add(cv)
			if test.clusterOperator != nil {
				indexer.Add(test.clusterOperator)
			}
			clusterVersionLister := configlistersv1.NewClusterVersionLister(indexer)
			clusterOperatorLister := configlistersv1.NewClusterOperatorLister(indexer)

			enabled, err := ImageRegistryIsEnabled(clusterVersionLister, clusterOperatorLister)
			if err != nil {
				t.Fatalf("ImageRegistryIsEnabled errored: %s", err)
			}
			if test.expectEnabled && !enabled {
				t.Fatal("expected image registry to be enabled but ImageRegistryIsEnabled returned disabled")
			}
			if !test.expectEnabled && enabled {
				t.Fatal("expected image registry to be disabled but ImageRegistryIsEnabled returned enabled")
			}
		})
	}
}
