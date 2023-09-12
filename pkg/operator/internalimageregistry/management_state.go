package internalimageregistry

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

func ImageRegistryIsEnabled(clusterVersionLister configlistersv1.ClusterVersionLister, clusterOperatorLister configlistersv1.ClusterOperatorLister) (bool, error) {
	cv, err := clusterVersionLister.Get("version")
	if err != nil {
		return false, err
	}
	var imageRegistryCapabilityEnabled bool
	for _, capability := range cv.Status.Capabilities.EnabledCapabilities {
		if capability == configv1.ClusterVersionCapabilityImageRegistry {
			imageRegistryCapabilityEnabled = true
			break
		}
	}
	if !imageRegistryCapabilityEnabled {
		return false, nil
	}

	co, err := clusterOperatorLister.Get("image-registry")
	if err != nil && !errors.IsNotFound(err) {
		return false, fmt.Errorf("unable to retrieve clusteroperators.config.openshift.io/image-registry: %w", err)
	}
	if errors.IsNotFound(err) {
		klog.V(4).Infof("clusteroperators.config.openshift.io/image-registry does not exist yet.")
		return false, nil
	}

	// Check if internal image registry is "Removed". Any condition should do.
	if len(co.Status.Conditions) == 0 {
		return false, fmt.Errorf("clusteroperators.config.openshift.io/image-registry conditions do not yet exist")
	}
	if co.Status.Conditions[0].Reason == "Removed" {
		return false, nil
	}
	// ImageRegistry capability is enabled, and internal image registry is enabled, nothing to disable.
	return true, nil
}
