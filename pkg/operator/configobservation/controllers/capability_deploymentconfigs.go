package controllers

import (
	configv1 "github.com/openshift/api/config/v1"
	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

func disabledDeploymentConfigControllers(listers configobservation.Listers) ([]openshiftcontrolplanev1.OpenShiftControllerName, error) {
	cv, err := listers.ClusterVersionLister.Get("version")
	if err != nil {
		return nil, err
	}
	var capabilityEnabled bool
	for _, capability := range cv.Status.Capabilities.EnabledCapabilities {
		if capability == configv1.ClusterVersionCapabilityDeploymentConfig {
			capabilityEnabled = true
			break
		}
	}
	if capabilityEnabled {
		return nil, nil
	}
	return []openshiftcontrolplanev1.OpenShiftControllerName{
		openshiftcontrolplanev1.OpenShiftDeploymentConfigController,
		openshiftcontrolplanev1.OpenShiftDeployerServiceAccountController,
		openshiftcontrolplanev1.OpenShiftDeployerController,
	}, nil

}
