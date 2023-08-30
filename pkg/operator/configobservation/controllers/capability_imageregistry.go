package controllers

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

func disabledImageRegistryControllers(listers configobservation.Listers) ([]openshiftcontrolplanev1.OpenShiftControllerName, error) {
	cv, err := listers.ClusterVersionLister.Get("version")
	if err != nil {
		return nil, err
	}
	var imageRegistryCapabilityEnabled bool
	for _, capability := range cv.Status.Capabilities.EnabledCapabilities {
		if capability == configv1.ClusterVersionCapabilityImageRegistry {
			imageRegistryCapabilityEnabled = true
			break
		}
	}
	controllers := []openshiftcontrolplanev1.OpenShiftControllerName{
		openshiftcontrolplanev1.OpenShiftServiceAccountPullSecretsController,
	}
	if !imageRegistryCapabilityEnabled {
		return controllers, nil
	}

	co, err := listers.ClusterOperatorLister.Get("image-registry")
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("unable to retrieve clusteroperators.config.openshift.io/image-registry: %w", err)
	}
	if errors.IsNotFound(err) {
		klog.V(4).Infof("clusteroperators.config.openshift.io/image-registry does not exist yet.")
		return controllers, nil
	}

	// Check if internal image registry is "Removed". Any condition should do.
	if len(co.Status.Conditions) == 0 {
		return nil, fmt.Errorf("clusteroperators.config.openshift.io/image-registry conditions do not yet exist")
	}
	if co.Status.Conditions[0].Reason == "Removed" {
		return controllers, nil
	}
	// ImageRegistry capability is enabled, and internal image registry is enabled, nothing to disable.
	return nil, nil
}
