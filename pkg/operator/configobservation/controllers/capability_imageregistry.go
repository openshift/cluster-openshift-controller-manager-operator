package controllers

import (
	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/internalimageregistry"
)

func disabledImageRegistryControllers(listers configobservation.Listers) ([]openshiftcontrolplanev1.OpenShiftControllerName, error) {
	enabled, err := internalimageregistry.ImageRegistryIsEnabled(listers.ClusterVersionLister, listers.ClusterOperatorLister)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return []openshiftcontrolplanev1.OpenShiftControllerName{
			openshiftcontrolplanev1.OpenShiftServiceAccountPullSecretsController,
			openshiftcontrolplanev1.OpenShiftImagePullerRoleBindingsController,
		}, nil
	}
	// ImageRegistry capability is enabled, and internal image registry is enabled, nothing to disable.
	return nil, nil
}
