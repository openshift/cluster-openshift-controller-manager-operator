package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/workload"
	"github.com/openshift/cluster-openshift-controller-manager-operator/test/framework"
	"github.com/openshift/library-go/pkg/operator/condition"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// clusterOpenshiftControllerManagerOperatorClient instantiate a client and return it, making sure
// the operator is fully up first.
func clusterOpenshiftControllerManagerOperatorClient(t *testing.T) framework.Clientset {
	client := framework.MustNewClientset(t, nil)
	framework.MustEnsureClusterOperatorStatusIsSet(t, client)
	return *client
}

// TestClusterOpenshiftControllerManagerOperator confirm operator is up before running other tests.
func TestClusterOpenshiftControllerManagerOperator(t *testing.T) {
	_ = clusterOpenshiftControllerManagerOperatorClient(t)
}

func TestClusterBuildConfigObservation(t *testing.T) {
	client := clusterOpenshiftControllerManagerOperatorClient(t)

	buildConfig, err := client.Builds().Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		t.Logf("error getting openshift controller manager config: %v", err)
	}

	buildDefaults := configv1.BuildDefaults{
		Env: []corev1.EnvVar{
			{
				Name:  "FOO",
				Value: "BAR",
			},
		},
	}

	if buildConfig == nil {
		buildConfig = &configv1.Build{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster",
			},
			Spec: configv1.BuildSpec{
				BuildDefaults: buildDefaults,
			},
		}

		if _, err := client.Builds().Create(context.TODO(), buildConfig, metav1.CreateOptions{}); err != nil {
			t.Fatalf("could not create cluster build configuration: %v", err)
		}
	} else {
		buildConfig.Spec.BuildDefaults = buildDefaults

		if _, err := client.Builds().Update(context.TODO(), buildConfig, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("could not create cluster build configuration: %v", err)
		}
	}

	defer func() {
		buildConfig.Spec.BuildDefaults.Env = []corev1.EnvVar{}

		if _, err := client.Builds().Update(context.TODO(), buildConfig, metav1.UpdateOptions{}); err != nil {
			t.Logf("failed to clean up cluster build config: %v", err)
		}
	}()

	err = wait.Poll(5*time.Second, 1*time.Minute, func() (bool, error) {
		cfg, err := client.OpenShiftControllerManagers().Get(context.TODO(), "cluster", metav1.GetOptions{})
		if cfg == nil || err != nil {
			t.Logf("error getting openshift controller manager config: %v", err)
			return false, nil
		}
		observed := string(cfg.Spec.ObservedConfig.Raw)
		if strings.Contains(observed, "FOO") {
			return true, nil
		}
		t.Logf("observed config missing env config: %s", observed)
		return false, nil
	})
	if err != nil {
		t.Fatalf("did not see cluster build env config propagated to openshift controller config: %v", err)
	}
}

func TestClusterImageConfigObservation(t *testing.T) {
	client := clusterOpenshiftControllerManagerOperatorClient(t)

	err := wait.Poll(5*time.Second, 1*time.Minute, func() (bool, error) {
		cfg, err := client.OpenShiftControllerManagers().Get(context.TODO(), "cluster", metav1.GetOptions{})
		if cfg == nil || err != nil {
			t.Logf("error getting openshift controller manager config: %v", err)
			return false, nil
		}
		observed := string(cfg.Spec.ObservedConfig.Raw)

		// on a healthy cluster this should always be set because the registry operator should
		// have created an images config object and the openshift controller operator should have
		// observed it at this point.
		if strings.Contains(observed, "\"internalRegistryHostname\"") {
			return true, nil
		}
		t.Logf("observed config missing internalregistryhostname config: %s", observed)
		return false, nil
	})
	if err != nil {
		t.Fatalf("did not see cluster image internalregistryhostname config propagated to openshift controller config: %v", err)
	}
}

// TestClusterConfigStatusPerManagementState change management state and assert status conditions.
func TestClusterConfigStatusPerManagementState(t *testing.T) {
	type testCase struct {
		managementState    operatorv1.ManagementState
		expectedConditions []operatorv1.OperatorCondition
	}

	managed := testCase{
		managementState: operatorv1.Managed,
		expectedConditions: []operatorv1.OperatorCondition{{
			Type:   operatorv1.OperatorStatusTypeAvailable,
			Status: operatorv1.ConditionTrue,
		}, {
			Type:   condition.ConfigObservationDegradedConditionType,
			Status: operatorv1.ConditionFalse,
		}, {
			Type:   operatorv1.OperatorStatusTypeProgressing,
			Status: operatorv1.ConditionFalse,
		}, {
			Type:   condition.ResourceSyncControllerDegradedConditionType,
			Status: operatorv1.ConditionFalse,
		}, {
			Type:   workload.WorkloadDegradedCondition,
			Status: operatorv1.ConditionFalse,
		}},
	}
	unmanaged := testCase{
		managementState: operatorv1.Unmanaged,
		expectedConditions: []operatorv1.OperatorCondition{{
			Type:   operatorv1.OperatorStatusTypeAvailable,
			Reason: "UnmanagedUnsupported",
			Status: operatorv1.ConditionTrue,
		}, {
			Type:   condition.ConfigObservationDegradedConditionType,
			Status: operatorv1.ConditionFalse,
		}, {
			Type:   operatorv1.OperatorStatusTypeProgressing,
			Reason: "UnmanagedUnsupported",
			Status: operatorv1.ConditionFalse,
		}, {
			Type:   condition.ResourceSyncControllerDegradedConditionType,
			Status: operatorv1.ConditionFalse,
		}, {
			Type:   workload.WorkloadDegradedCondition,
			Reason: "UnmanagedUnsupported",
			Status: operatorv1.ConditionFalse,
		}},
	}

	// going from managed to unmanaged, and again to original state
	testCases := []testCase{managed, unmanaged, managed}

	client := clusterOpenshiftControllerManagerOperatorClient(t)
	for _, tc := range testCases {
		t.Logf("Testing conditions for '%s' management state", tc.managementState)
		cfg := getConfig(t, &client)
		cfg.Spec.ManagementState = tc.managementState
		_ = updateConfig(t, &client, cfg)

		err := wait.Poll(5*time.Second, 1*time.Minute, func() (bool, error) {
			cfg, err := client.OpenShiftControllerManagers().Get(context.TODO(), "cluster", metav1.GetOptions{})
			if cfg == nil || err != nil {
				t.Logf("error getting openshift controller manager config: %v", err)
				return false, nil
			}
			matches := assertOperatorConditions(t, tc.expectedConditions, cfg.Status.Conditions)
			return matches, nil
		})
		if err != nil {
			t.Fatalf("reported status did not match expected: %v", err)
		}
	}
}
