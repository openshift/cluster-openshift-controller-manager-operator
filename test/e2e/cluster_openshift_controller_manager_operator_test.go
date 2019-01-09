package e2e

import (
	"testing"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/testframework"
)

func TestClusterOpenshiftControllerManagerOperator(t *testing.T) {
	client := testframework.MustNewClientset(t, nil)
	testframework.MustEnsureClusterOperatorStatusIsSet(t, client)
}
