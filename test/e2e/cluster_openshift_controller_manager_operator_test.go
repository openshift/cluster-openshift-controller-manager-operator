package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	configv1 "github.com/openshift/api/config/v1"

	"github.com/openshift/cluster-openshift-controller-manager-operator/test/framework"
)

func TestClusterOpenshiftControllerManagerOperator(t *testing.T) {
	client := framework.MustNewClientset(t, nil)
	// make sure the operator is fully up
	framework.MustEnsureClusterOperatorStatusIsSet(t, client)
}

func TestClusterBuildConfigObservation(t *testing.T) {
	ctx := context.Background()
	client := framework.MustNewClientset(t, nil)
	// make sure the operator is fully up
	framework.MustEnsureClusterOperatorStatusIsSet(t, client)

	// The CVO should be creating these cluster configuration objects on cluster install
	var buildConfig *configv1.Build

	err := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		buildConfig, err = client.Builds().Get(ctx, "cluster", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("error getting cluster build config: %v", err)
	}

	buildConfig.Spec.BuildDefaults = configv1.BuildDefaults{
		Env: []corev1.EnvVar{
			{
				Name:  "FOO",
				Value: "BAR",
			},
		},
	}

	if _, err := client.Builds().Update(ctx, buildConfig, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("could not update cluster build configuration: %v", err)
	}

	defer func() {
		// Other things may update the cluster build config - get a fresh copy
		buildConfig, err = client.Builds().Get(ctx, "cluster", metav1.GetOptions{})
		if err != nil {
			t.Logf("failed to get cluster build configuration: %v", err)
			return
		}
		buildConfig.Spec.BuildDefaults = configv1.BuildDefaults{}
		if _, err := client.Builds().Update(ctx, buildConfig, metav1.UpdateOptions{}); err != nil {
			t.Logf("failed to clean up cluster build config: %v", err)
		}
	}()

	err = wait.Poll(5*time.Second, 1*time.Minute, func() (bool, error) {
		cfg, err := client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
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
	ctx := context.Background()
	client := framework.MustNewClientset(t, nil)
	// make sure the operator is fully up
	framework.MustEnsureClusterOperatorStatusIsSet(t, client)

	err := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		cfg, err := client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
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
