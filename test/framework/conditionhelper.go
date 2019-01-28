package framework

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	configv1 "github.com/openshift/api/config/v1"
)

func hasExpectedClusterOperatorConditions(status *configv1.ClusterOperator) bool {
	gotAvailable := false
	gotProgressing := false
	gotFailing := false
	for _, c := range status.Status.Conditions {
		if c.Type == configv1.OperatorAvailable && c.Status == configv1.ConditionTrue {
			gotAvailable = true
		}
		if c.Type == configv1.OperatorProgressing && c.Status == configv1.ConditionFalse {
			gotProgressing = true
		}
		if c.Type == configv1.OperatorFailing && c.Status == configv1.ConditionFalse {
			gotFailing = true
		}
	}
	return gotAvailable && gotProgressing && gotFailing
}

func ensureClusterOperatorStatusIsSet(logger Logger, client *Clientset) error {
	var status *configv1.ClusterOperator
	err := wait.Poll(1*time.Second, 2*time.Minute, func() (stop bool, err error) {
		status, err = client.ClusterOperators().Get("openshift-controller-manager-operator", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			logger.Logf("waiting for the cluster operator resource: the resource does not exist")
			return false, nil
		} else if err != nil {
			return false, err
		}
		if hasExpectedClusterOperatorConditions(status) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		logger.Logf("clusteroperator status resource was not updated with the expected status: %v", err)
		if status != nil {
			logger.Logf("clusteroperator conditions are: %#v", status.Status.Conditions)
		}
	}
	return err
}

func MustEnsureClusterOperatorStatusIsSet(t *testing.T, client *Clientset) {
	if err := ensureClusterOperatorStatusIsSet(t, client); err != nil {
		t.Fatal(err)
	}
}
