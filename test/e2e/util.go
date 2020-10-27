package e2e

import (
	"context"
	"testing"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/test/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getConfig helper to execute get against operator's config.
func getConfig(t *testing.T, client *framework.Clientset) *operatorv1.OpenShiftControllerManager {
	opts := metav1.GetOptions{}
	cfg, err := client.OpenShiftControllerManagers().Get(context.TODO(), "cluster", opts)
	if err != nil {
		t.Fatalf("error getting openshift controller manager: '%v'", err)
	}
	return cfg
}

// updateConfig helper to execute update against operator's config.
func updateConfig(
	t *testing.T,
	client *framework.Clientset,
	cfg *operatorv1.OpenShiftControllerManager,
) *operatorv1.OpenShiftControllerManager {
	opts := metav1.UpdateOptions{}
	updated, err := client.OpenShiftControllerManagers().Update(context.TODO(), cfg, opts)
	if err != nil {
		t.Fatalf("error updating openshift controller manager: '%v'", err)
	}
	return updated
}

// assertOperatorConditions compare two slices of operator conditions ignoring timestamps.
func assertOperatorConditions(t *testing.T, expected, actual []operatorv1.OperatorCondition) bool {
	for _, e := range expected {
		matches := false
		for _, a := range actual {
			if matches {
				continue
			}
			if e.Type == a.Type && e.Status == a.Status && e.Reason == a.Reason {
				matches = true
			}
		}
		if !matches {
			t.Logf("Expected condition '%#v' is not found", e)
			return false
		}
	}
	return true
}
