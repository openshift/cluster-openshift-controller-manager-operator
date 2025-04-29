package internalimageregistry

import (
	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

// ControllerManagerIsProgressing returns true if the operand for this operator
// is progressing. This function observes the "progressing", the "available", and
// the "degraded" conditions of the "openshift-controller-manager".
func ControllerManagerIsProgressing(clusterOperatorLister configlistersv1.ClusterOperatorLister) (bool, error) {
	co, err := clusterOperatorLister.Get("openshift-controller-manager")
	if err != nil {
		return false, err
	}

	if len(co.Status.Conditions) == 0 {
		klog.V(4).InfoS("clusteroperators.config.openshift.io/openshift-controller-manager conditions do not yet exist")
		return true, nil
	}

	missingProgressing, missingAvailable, missingDegraded := true, true, true
	for _, cond := range co.Status.Conditions {
		if cond.Type == configv1.OperatorProgressing {
			missingProgressing = false
			if cond.Status != configv1.ConditionFalse {
				klog.V(4).InfoS("clusteroperators.config.openshift.io/openshift-controller-manager is progressing")
				return true, nil
			}
		}
		if cond.Type == configv1.OperatorAvailable {
			missingAvailable = false
			if cond.Status != configv1.ConditionTrue {
				klog.V(4).InfoS("clusteroperators.config.openshift-controller-manager is not available")
				return true, nil
			}
		}
		if cond.Type == configv1.OperatorDegraded {
			missingDegraded = false
			if cond.Status != configv1.ConditionFalse {
				klog.V(4).InfoS("clusteroperators.config.openshift.io/openshift-controller-manager is degraded")
				return true, nil
			}
		}
	}

	if missingProgressing || missingAvailable || missingDegraded {
		klog.V(4).InfoS("clusteroperators.config.openshift.io/openshift-controller-manager conditions available/progressing/degraded not found")
		return true, nil
	}

	return false, nil
}

// ImageRegistryIsEnabled returns true if the ImageRegistry capability
// is enabled and the internal image registry has not been disabled.
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

	// Given that the capability is enabled, assume the internal image registry is
	// enabled unless we can explicitly determine otherwise in via the management status.

	co, err := clusterOperatorLister.Get("image-registry")
	if err != nil && !errors.IsNotFound(err) {
		klog.V(4).ErrorS(err, "unable to retrieve clusteroperators.config.openshift.io/image-registry")
		return true, nil
	}
	if errors.IsNotFound(err) {
		klog.V(4).InfoS("clusteroperators.config.openshift.io/image-registry does not exist yet.")
		return true, nil
	}

	// Check if internal image registry is "Removed". Any condition should do.
	if len(co.Status.Conditions) == 0 {
		klog.V(4).InfoS("clusteroperators.config.openshift.io/image-registry conditions do not yet exist")
		return true, nil
	}
	if co.Status.Conditions[0].Reason == "Removed" {
		return false, nil
	}

	return true, nil
}
