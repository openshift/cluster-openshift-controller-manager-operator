package controllers

import (
	"fmt"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var allControllers = []string{
	string(openshiftcontrolplanev1.OpenShiftServiceAccountController),
	string(openshiftcontrolplanev1.OpenShiftDefaultRoleBindingsController),
	string(openshiftcontrolplanev1.OpenShiftServiceAccountPullSecretsController),
	string(openshiftcontrolplanev1.OpenShiftOriginNamespaceController),
	string(openshiftcontrolplanev1.OpenShiftBuildController),
	string(openshiftcontrolplanev1.OpenShiftBuildConfigChangeController),
	string(openshiftcontrolplanev1.OpenShiftBuilderServiceAccountController),
	string(openshiftcontrolplanev1.OpenShiftDeployerController),
	string(openshiftcontrolplanev1.OpenShiftDeployerServiceAccountController),
	string(openshiftcontrolplanev1.OpenShiftDeploymentConfigController),
	string(openshiftcontrolplanev1.OpenShiftImageTriggerController),
	string(openshiftcontrolplanev1.OpenShiftImageImportController),
	string(openshiftcontrolplanev1.OpenShiftImageSignatureImportController),
	string(openshiftcontrolplanev1.OpenShiftTemplateInstanceController),
	string(openshiftcontrolplanev1.OpenShiftTemplateInstanceFinalizerController),
	string(openshiftcontrolplanev1.OpenShiftUnidlingController),
	// the following two controllers are now part of route-controller-manager, which split
	// some crontollers off from  openshift-controller-manager, but still uses the same config.
	string(openshiftcontrolplanev1.OpenShiftIngressIPController),
	string(openshiftcontrolplanev1.OpenShiftIngressToRouteController),
}

type disabledControllersFunc func(listers configobservation.Listers) ([]openshiftcontrolplanev1.OpenShiftControllerName, error)

var disabledControllerFuncs = []disabledControllersFunc{
	disabledImageRegistryControllers,
	disabledBuildControllers,
	disabledDeploymentConfigControllers,
}

func ObserveControllers(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	observedConfig := map[string]interface{}{}
	var errs []error

	previousValue, _, err := unstructured.NestedStringSlice(existingConfig, "controllers")
	if err != nil {
		return observedConfig, append(errs, fmt.Errorf("unable to parse existing controllers value: %w", err))
	}
	previousConfig := map[string]interface{}{}
	unstructured.SetNestedStringSlice(previousConfig, previousValue, "controllers")

	controllers := append([]string{}, allControllers...)
	unstructured.SetNestedStringSlice(observedConfig, controllers, "controllers")

	// compile list of controllers to disable
	var disabledControllers []openshiftcontrolplanev1.OpenShiftControllerName
	for _, getDisabledControllers := range disabledControllerFuncs {
		disabled, err := getDisabledControllers(listers)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		disabledControllers = append(disabledControllers, disabled...)
	}
	if len(errs) > 0 {
		return previousConfig, errs
	}
	// mark controllers as disabled
	for _, name := range disabledControllers {
		controllers = disableController(controllers, string(name))
	}
	controllersSort(controllers).Sort()
	err = unstructured.SetNestedStringSlice(observedConfig, controllers, "controllers")
	if err != nil {
		return previousConfig, append(errs, fmt.Errorf("error setting controllers value: %w", err))
	}
	return observedConfig, nil
}

func disableController(controllers []string, controller string) []string {
	for i, c := range controllers {
		switch c {
		case controller:
			controllers[i] = "-" + controller
			return controllers
		case "-" + controller:
			return controllers
		}
	}
	return append(controllers, "-"+controller)
}
