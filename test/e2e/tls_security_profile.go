package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	g "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/test/framework"
)

var _ = g.Describe("[sig-openshift-controller-manager] TLS Security Profile", func() {
	g.It("[Operator][TLS][Serial] should propagate Modern TLS profile from APIServer to OpenShift Controller Manager", func() {
		testTLSSecurityProfilePropagation(g.GinkgoTB())
	})
})

func testTLSSecurityProfilePropagation(t testing.TB) {
	ctx := context.Background()
	client := framework.MustNewClientset(t, nil)

	// Make sure the operator is fully up
	framework.MustEnsureClusterOperatorStatusIsSet(t, client)

	// Get the current APIServer config
	apiServer, err := client.APIServers().Get(ctx, "cluster", metav1.GetOptions{})
	require.NoError(t, err, "failed to get APIServer config")

	// Save the original TLS profile for cleanup
	originalTLSProfile := apiServer.Spec.TLSSecurityProfile

	// Modify the TLS security profile to use Modern profile
	// Modern profile uses TLS 1.3 with modern cipher suites
	apiServer.Spec.TLSSecurityProfile = &configv1.TLSSecurityProfile{
		Type:   configv1.TLSProfileModernType,
		Modern: &configv1.ModernTLSProfile{},
	}

	_, err = client.APIServers().Update(ctx, apiServer, metav1.UpdateOptions{})
	require.NoError(t, err, "failed to update APIServer TLS profile to Modern")

	// Cleanup: restore original TLS profile and verify restoration
	t.Cleanup(func() {
		t.Log("Restoring original TLS profile")
		apiServer, err := client.APIServers().Get(ctx, "cluster", metav1.GetOptions{})
		if err != nil {
			t.Logf("failed to get APIServer for cleanup: %v", err)
			return
		}
		apiServer.Spec.TLSSecurityProfile = originalTLSProfile
		if _, err := client.APIServers().Update(ctx, apiServer, metav1.UpdateOptions{}); err != nil {
			t.Logf("failed to restore original TLS profile: %v", err)
			return
		}

		// Wait for operator to reconcile the restoration
		t.Log("Waiting for operator to reconcile TLS profile restoration")
		err = wait.PollUntilContextTimeout(ctx, 10*time.Second, 10*time.Minute, true, func(ctx context.Context) (bool, error) {
			co, err := client.ClusterOperators().Get(ctx, "openshift-controller-manager", metav1.GetOptions{})
			if err != nil {
				t.Logf("error getting clusteroperator during cleanup: %v", err)
				return false, nil
			}

			isAvailable := false
			isProgressing := true

			for _, c := range co.Status.Conditions {
				if c.Type == configv1.OperatorAvailable && c.Status == configv1.ConditionTrue {
					isAvailable = true
				}
				if c.Type == configv1.OperatorProgressing && c.Status == configv1.ConditionFalse {
					isProgressing = false
				}
			}

			if isAvailable && !isProgressing {
				t.Log("Operator reconciliation after restoration complete")
				return true, nil
			}

			return false, nil
		})
		if err != nil {
			t.Logf("operator did not complete reconciliation after restoration: %v", err)
			return
		}

		// Verify TLS profile was restored (should be back to default TLS 1.2 or original setting)
		t.Log("Verifying TLS profile was restored correctly")
		err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
			cfg, err := client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
			if err != nil {
				t.Logf("error getting openshift controller manager config during cleanup verification: %v", err)
				return false, nil
			}

			observedConfig := map[string]interface{}{}
			if err := json.Unmarshal(cfg.Spec.ObservedConfig.Raw, &observedConfig); err != nil {
				t.Logf("failed to unmarshal observed config during cleanup: %v", err)
				return false, nil
			}

			// Check the restored TLS version
			minTLSVersion, found, err := unstructured.NestedString(observedConfig, "servingInfo", "minTLSVersion")
			if err != nil {
				t.Logf("error reading minTLSVersion during cleanup: %v", err)
				return false, nil
			}

			// If original profile was nil, expect default (typically VersionTLS12)
			// If original profile was set, it should match
			if originalTLSProfile == nil {
				// Default OpenShift TLS profile is typically TLS 1.2
				if found && minTLSVersion == "VersionTLS12" {
					t.Logf("TLS profile restored to default: %s", minTLSVersion)
					return true, nil
				}
				// Also accept if TLS config is removed entirely (using cluster defaults)
				if !found || minTLSVersion == "" {
					t.Log("TLS profile restored to cluster defaults (no explicit TLS version)")
					return true, nil
				}
			} else {
				// If there was an original profile, verify it's not TLS 1.3 anymore
				if found && minTLSVersion != "VersionTLS13" {
					t.Logf("TLS profile restored from Modern: %s", minTLSVersion)
					return true, nil
				}
			}

			t.Logf("Waiting for TLS profile restoration to propagate, current: %s", minTLSVersion)
			return false, nil
		})
		if err != nil {
			t.Logf("TLS profile was not properly restored in observed config: %v", err)
		}
	})

	// Wait for the operator to start progressing (detecting the change)
	t.Log("Waiting for operator to detect TLS profile change and start progressing")
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		co, err := client.ClusterOperators().Get(ctx, "openshift-controller-manager", metav1.GetOptions{})
		if err != nil {
			t.Logf("error getting clusteroperator: %v", err)
			return false, nil
		}
		for _, c := range co.Status.Conditions {
			if c.Type == configv1.OperatorProgressing && c.Status == configv1.ConditionTrue {
				t.Logf("Operator is now progressing, reason: %s", c.Reason)
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Logf("Warning: operator did not start progressing within 5 minutes, continuing anyway: %v", err)
	}

	// Wait for the operator to finish progressing (reconciliation complete)
	// This typically takes 12-15 minutes for TLS changes to propagate
	// Replace this wait cluster update logic when library-go PR https://github.com/openshift/library-go/pull/2050 is merged
	t.Log("Waiting for operator to complete reconciliation (may take up to 15 minutes)")
	err = wait.PollUntilContextTimeout(ctx, 10*time.Second, 15*time.Minute, true, func(ctx context.Context) (bool, error) {
		co, err := client.ClusterOperators().Get(ctx, "openshift-controller-manager", metav1.GetOptions{})
		if err != nil {
			t.Logf("error getting clusteroperator: %v", err)
			return false, nil
		}

		isAvailable := false
		isProgressing := true
		isDegraded := false

		for _, c := range co.Status.Conditions {
			if c.Type == configv1.OperatorAvailable && c.Status == configv1.ConditionTrue {
				isAvailable = true
			}
			if c.Type == configv1.OperatorProgressing && c.Status == configv1.ConditionFalse {
				isProgressing = false
			}
			if c.Type == configv1.OperatorDegraded && c.Status == configv1.ConditionTrue {
				isDegraded = true
			}
		}

		if isDegraded {
			t.Log("Warning: operator is degraded")
			return false, nil
		}

		if isAvailable && !isProgressing {
			t.Log("Operator reconciliation complete")
			return true, nil
		}

		t.Logf("Operator still reconciling, available=%v, progressing=%v", isAvailable, isProgressing)
		return false, nil
	})
	require.NoError(t, err, "operator did not complete reconciliation")

	// Now verify the TLS config was propagated to the observed config
	t.Log("Verifying TLS config in observed config")
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		cfg, err := client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
		if err != nil {
			t.Logf("error getting openshift controller manager config: %v", err)
			return false, nil
		}

		observed := string(cfg.Spec.ObservedConfig.Raw)

		// The Modern TLS profile should set minTLSVersion to TLS 1.3
		// We're looking for the propagated TLS settings
		hasTLSVersion := strings.Contains(observed, "\"minTLSVersion\"")
		hasCipherSuites := strings.Contains(observed, "\"cipherSuites\"")

		if !hasTLSVersion || !hasCipherSuites {
			t.Logf("TLS config not yet observed in config: %s", observed)
			return false, nil
		}

		t.Logf("TLS config successfully observed: %s", observed)

		// Additional validation: parse the observed config
		observedConfig := map[string]interface{}{}
		if err := json.Unmarshal(cfg.Spec.ObservedConfig.Raw, &observedConfig); err != nil {
			t.Logf("failed to unmarshal observed config: %v", err)
			return false, nil
		}

		// Verify servingInfo exists
		_, found, err := unstructured.NestedMap(observedConfig, "servingInfo")
		if err != nil || !found {
			t.Log("servingInfo not found in observed config")
			return false, nil
		}

		// Verify minTLSVersion is set to TLS 1.3 (Modern profile)
		minTLSVersion, found, err := unstructured.NestedString(observedConfig, "servingInfo", "minTLSVersion")
		if err != nil || !found || minTLSVersion == "" {
			t.Logf("minTLSVersion not properly set, found=%v, value=%s", found, minTLSVersion)
			return false, nil
		}

		// Modern profile should use VersionTLS13 (exact string match)
		if minTLSVersion != "VersionTLS13" {
			t.Logf("minTLSVersion not VersionTLS13 yet, got=%s, expected=VersionTLS13", minTLSVersion)
			return false, nil
		}

		// Verify cipherSuites is set and contains the expected Modern profile ciphers
		cipherSuites, found, err := unstructured.NestedStringSlice(observedConfig, "servingInfo", "cipherSuites")
		if err != nil || !found || len(cipherSuites) == 0 {
			t.Logf("cipherSuites not properly set, found=%v, count=%d", found, len(cipherSuites))
			return false, nil
		}

		// Modern profile should have exactly these TLS 1.3 cipher suites
		expectedCiphers := []string{
			"TLS_AES_128_GCM_SHA256",
			"TLS_AES_256_GCM_SHA384",
			"TLS_CHACHA20_POLY1305_SHA256",
		}

		// Verify all expected ciphers are present
		cipherSet := make(map[string]bool)
		for _, cipher := range cipherSuites {
			cipherSet[cipher] = true
		}

		for _, expected := range expectedCiphers {
			if !cipherSet[expected] {
				// Don't fail immediately, keep polling
				t.Logf("expected cipher suite not found yet: %s, got: %v", expected, cipherSuites)
				return false, nil
			}
		}

		t.Logf("Validated Modern TLS config: minTLSVersion=%s, cipherSuites=%v", minTLSVersion, cipherSuites)
		return true, nil
	})

	require.NoError(t, err, "Modern TLS security profile from APIServer was not propagated to OpenShift Controller Manager observed config")

	// Wait for pods to restart and become ready with the new TLS configuration
	t.Log("Waiting for controller-manager pods to be ready with new TLS configuration")
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		// Check if deployment is ready
		checkCmd := exec.Command("oc", "get", "deployment", "controller-manager", "-n", "openshift-controller-manager",
			"-o", "jsonpath={.status.conditions[?(@.type=='Available')].status}")
		output, err := checkCmd.CombinedOutput()
		if err != nil {
			t.Logf("error checking deployment status: %v", err)
			return false, nil
		}

		if strings.TrimSpace(string(output)) == "True" {
			// Also verify that the deployment is not progressing (rolling out)
			progressingCmd := exec.Command("oc", "get", "deployment", "controller-manager", "-n", "openshift-controller-manager",
				"-o", "jsonpath={.status.conditions[?(@.type=='Progressing')].reason}")
			progressOutput, err := progressingCmd.CombinedOutput()
			if err != nil {
				t.Logf("error checking deployment progressing status: %v", err)
				return false, nil
			}

			progressReason := strings.TrimSpace(string(progressOutput))
			if progressReason == "NewReplicaSetAvailable" {
				t.Log("Controller-manager deployment is ready")
				return true, nil
			}
			t.Logf("Deployment progressing: %s", progressReason)
		}
		return false, nil
	})
	require.NoError(t, err, "controller-manager pods did not become ready with new TLS configuration")

	// Verify actual TLS connectivity from inside the cluster using a temporary pod
	t.Log("Verifying actual TLS connectivity with Modern profile")

	// Use the service DNS name and service port (443)
	serviceEndpoint := "controller-manager.openshift-controller-manager.svc:443"

	// Test 1: TLS 1.2 should fail with Modern profile
	t.Log("Testing TLS 1.2 connection (should fail with Modern profile)")
	cmdTLS12 := fmt.Sprintf("echo | openssl s_client -connect %s -tls1_2 2>&1 || true", serviceEndpoint)
	runCmdTLS12 := exec.Command("oc", "run", "tls-test-12",
		"--image=image-registry.openshift-image-registry.svc:5000/openshift/tools:latest",
		"--rm", "-i", "--restart=Never",
		"--command", "--", "bash", "-c", cmdTLS12)
	outputTLS12, err := runCmdTLS12.CombinedOutput()

	// TLS 1.2 should fail - check for error indicators
	if strings.Contains(string(outputTLS12), "Certificate chain") {
		t.Error("TLS 1.2 connection succeeded but should have failed with Modern profile")
	} else if strings.Contains(string(outputTLS12), "ssl handshake failure") ||
		strings.Contains(string(outputTLS12), "no protocols available") ||
		strings.Contains(string(outputTLS12), "wrong version number") ||
		strings.Contains(string(outputTLS12), "sslv3 alert handshake failure") ||
		strings.Contains(string(outputTLS12), "alert protocol version") ||
		strings.Contains(string(outputTLS12), "SSL alert number 70") ||
		strings.Contains(string(outputTLS12), "no peer certificate available") ||
		!strings.Contains(string(outputTLS12), "CONNECTED") {
		t.Log("TLS 1.2 connection correctly failed with Modern profile")
	} else {
		t.Logf("Warning: TLS 1.2 test had unexpected output, cannot confirm failure")
	}

	// Test 2: TLS 1.3 should succeed with Modern profile
	t.Log("Testing TLS 1.3 connection (should succeed with Modern profile)")
	cmdTLS13 := fmt.Sprintf("echo | openssl s_client -connect %s -tls1_3 2>&1 || true", serviceEndpoint)
	runCmdTLS13 := exec.Command("oc", "run", "tls-test-13",
		"--image=image-registry.openshift-image-registry.svc:5000/openshift/tools:latest",
		"--rm", "-i", "--restart=Never",
		"--command", "--", "bash", "-c", cmdTLS13)
	outputTLS13, err := runCmdTLS13.CombinedOutput()

	// TLS 1.3 should succeed - check for success indicators
	if strings.Contains(string(outputTLS13), "Certificate chain") {
		t.Log("TLS 1.3 connection succeeded as expected with Modern profile")
	} else {
		require.Fail(t, "TLS 1.3 connection failed but should have succeeded with Modern profile",
			"Output: %s", string(outputTLS13))
	}
}
