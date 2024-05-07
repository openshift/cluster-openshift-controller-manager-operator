package controllers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	configv1 "github.com/openshift/api/config/v1"
	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	configlisterv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/library-go/pkg/operator/events"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

func TestObserveControllers(t *testing.T) {

	withDisabled := func(op openshiftcontrolplanev1.OpenShiftControllerName) func([]string) []string {
		return func(arr []string) []string {
			return disableController(arr, string(op))
		}
	}

	defaultConfig := func(opts ...func([]string) []string) []string {
		result := append([]string{}, allControllers...)
		for _, fn := range opts {
			result = fn(result)
		}
		controllersSort(result).Sort()
		return result
	}

	clusterOperator := func(s string) *configv1.ClusterOperator {
		return &configv1.ClusterOperator{
			ObjectMeta: metav1.ObjectMeta{Name: "image-registry"},
			Status: configv1.ClusterOperatorStatus{
				Conditions: []configv1.ClusterOperatorStatusCondition{
					{
						Type:   configv1.OperatorAvailable,
						Status: configv1.ConditionTrue,
						Reason: s,
					},
				},
			},
		}
	}

	clusterVersion := func(opts ...func(*configv1.ClusterVersion)) *configv1.ClusterVersion {
		cv := &configv1.ClusterVersion{
			ObjectMeta: metav1.ObjectMeta{Name: "version"},
			Status: configv1.ClusterVersionStatus{
				Capabilities: configv1.ClusterVersionCapabilitiesStatus{
					EnabledCapabilities: []configv1.ClusterVersionCapability{configv1.ClusterVersionCapabilityBaremetal, configv1.ClusterVersionCapabilityConsole}},
			},
		}
		for _, f := range opts {
			f(cv)
		}
		return cv
	}

	withImageRegistryCapability := func(cv *configv1.ClusterVersion) {
		cv.Status.Capabilities.EnabledCapabilities = append(cv.Status.Capabilities.EnabledCapabilities, configv1.ClusterVersionCapabilityImageRegistry)
	}

	withBuildCapability := func(cv *configv1.ClusterVersion) {

		cv.Status.Capabilities.EnabledCapabilities = append(cv.Status.Capabilities.EnabledCapabilities, configv1.ClusterVersionCapabilityBuild)
	}
	withDeploymentConfigCapability := func(cv *configv1.ClusterVersion) {
		cv.Status.Capabilities.EnabledCapabilities = append(cv.Status.Capabilities.EnabledCapabilities, configv1.ClusterVersionCapabilityDeploymentConfig)
	}

	testCases := []struct {
		name            string
		clusterVersion  *configv1.ClusterVersion
		clusterOperator *configv1.ClusterOperator
		existingConfig  []string
		expectedConfig  []string
		expectErr       bool
	}{
		{
			name:           "NoClusterOperator",
			clusterVersion: clusterVersion(withImageRegistryCapability, withBuildCapability, withDeploymentConfigCapability),
			expectedConfig: defaultConfig(),
		},
		{
			name:            "RegistryRemoved",
			clusterVersion:  clusterVersion(withImageRegistryCapability, withBuildCapability, withDeploymentConfigCapability),
			clusterOperator: clusterOperator("Removed"),
			expectedConfig: defaultConfig(
				withDisabled(openshiftcontrolplanev1.OpenShiftServiceAccountPullSecretsController),
				withDisabled(openshiftcontrolplanev1.OpenShiftImagePullerRoleBindingsController),
			),
		},
		{
			name:            "RegistryNotRemoved",
			clusterVersion:  clusterVersion(withImageRegistryCapability, withBuildCapability, withDeploymentConfigCapability),
			clusterOperator: clusterOperator("Managed"),
			expectedConfig:  defaultConfig(),
		},
		{
			name:           "NoImageRegistryCapabilityNoRegistryConfig",
			clusterVersion: clusterVersion(withBuildCapability, withDeploymentConfigCapability),
			expectedConfig: defaultConfig(
				withDisabled(openshiftcontrolplanev1.OpenShiftServiceAccountPullSecretsController),
				withDisabled(openshiftcontrolplanev1.OpenShiftImagePullerRoleBindingsController),
			),
		},
		{
			name:            "NoImageRegistryCapabilityRegistryRemoved",
			clusterVersion:  clusterVersion(withBuildCapability, withDeploymentConfigCapability),
			clusterOperator: clusterOperator("Removed"),
			expectedConfig: defaultConfig(
				withDisabled(openshiftcontrolplanev1.OpenShiftServiceAccountPullSecretsController),
				withDisabled(openshiftcontrolplanev1.OpenShiftImagePullerRoleBindingsController),
			),
		},
		{
			name:            "NoImageRegistryCapabilityRegistryNotRemoved",
			clusterVersion:  clusterVersion(withBuildCapability, withDeploymentConfigCapability),
			clusterOperator: clusterOperator("Managed"),
			expectedConfig: defaultConfig(
				withDisabled(openshiftcontrolplanev1.OpenShiftServiceAccountPullSecretsController),
				withDisabled(openshiftcontrolplanev1.OpenShiftImagePullerRoleBindingsController),
			),
		},
		{
			name:            "NoBuildCapability",
			clusterVersion:  clusterVersion(withImageRegistryCapability, withDeploymentConfigCapability),
			clusterOperator: clusterOperator("Managed"),
			expectedConfig: defaultConfig(
				withDisabled(openshiftcontrolplanev1.OpenShiftBuildController),
				withDisabled(openshiftcontrolplanev1.OpenShiftBuildConfigChangeController),
				withDisabled(openshiftcontrolplanev1.OpenShiftBuilderServiceAccountController),
				withDisabled(openshiftcontrolplanev1.OpenShiftBuilderRoleBindingsController),
			),
		},
		{
			name:            "NoDeployerConfigCapability",
			clusterVersion:  clusterVersion(withImageRegistryCapability, withBuildCapability),
			clusterOperator: clusterOperator("Managed"),
			expectedConfig: defaultConfig(
				withDisabled(openshiftcontrolplanev1.OpenShiftDeploymentConfigController),
				withDisabled(openshiftcontrolplanev1.OpenShiftDeployerServiceAccountController),
				withDisabled(openshiftcontrolplanev1.OpenShiftDeployerController),
				withDisabled(openshiftcontrolplanev1.OpenShiftDeployerRoleBindingsController),
			),
		},
		{
			name:            "NoBuildOrDeployerConfigCapability",
			clusterVersion:  clusterVersion(withImageRegistryCapability),
			clusterOperator: clusterOperator("Managed"),
			expectedConfig: defaultConfig(
				withDisabled(openshiftcontrolplanev1.OpenShiftDeploymentConfigController),
				withDisabled(openshiftcontrolplanev1.OpenShiftBuildController),
				withDisabled(openshiftcontrolplanev1.OpenShiftBuildConfigChangeController),
				withDisabled(openshiftcontrolplanev1.OpenShiftBuilderServiceAccountController),
				withDisabled(openshiftcontrolplanev1.OpenShiftBuilderRoleBindingsController),
				withDisabled(openshiftcontrolplanev1.OpenShiftDeployerServiceAccountController),
				withDisabled(openshiftcontrolplanev1.OpenShiftDeployerController),
				withDisabled(openshiftcontrolplanev1.OpenShiftDeployerRoleBindingsController),
			),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.clusterVersion == nil {
				tc.clusterVersion = clusterVersion()
			}
			clusterVersionIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			clusterVersionIndexer.Add(tc.clusterVersion)
			clusterOperatorIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			if tc.clusterOperator != nil {
				clusterOperatorIndexer.Add(tc.clusterOperator)
			}
			listers := configobservation.Listers{
				ClusterOperatorLister: configlisterv1.NewClusterOperatorLister(clusterOperatorIndexer),
				ClusterVersionLister:  configlisterv1.NewClusterVersionLister(clusterVersionIndexer),
			}
			existingConfig := map[string]any{}
			if tc.existingConfig != nil {
				unstructured.SetNestedStringSlice(existingConfig, tc.existingConfig, "controllers")
			}
			actualConfig, actualErr := ObserveControllers(listers, events.NewInMemoryRecorder(t.Name()), existingConfig)

			expectedConfig := map[string]any{}
			if tc.expectedConfig != nil {
				unstructured.SetNestedStringSlice(expectedConfig, tc.expectedConfig, "controllers")
			}
			if !cmp.Equal(actualConfig, expectedConfig) {
				t.Errorf(cmp.Diff(actualConfig, expectedConfig))
			}
			if tc.expectErr == (actualErr == nil) {
				t.Errorf("expected an error: %v, got an error: %v", tc.expectErr, actualErr)
			}
		})
	}
}
