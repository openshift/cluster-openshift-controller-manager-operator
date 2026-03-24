package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

const (
	defaultDenyAllPolicyName         = "default-deny"
	controllerManagerPolicyName      = "allow-controller-manager"
	routeControllerManagerPolicyName = "allow-route-controller-manager"
	operatorPolicyName               = "allow-operator"
)

func TestControllerManagerNetworkPolicies(t *testing.T) {
	ctx := context.Background()
	t.Log("Creating Kubernetes clients")
	kubeConfig, err := getKubeConfig()
	if err != nil {
		t.Fatalf("failed to get kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Fatalf("failed to create kubernetes client: %v", err)
	}

	t.Log("Validating NetworkPolicies in openshift-controller-manager")
	controllerManagerDefaultDeny := getNetworkPolicyT(t, ctx, kubeClient, util.TargetNamespace, defaultDenyAllPolicyName)
	logNetworkPolicySummary(t, "controller-manager/default-deny-all", controllerManagerDefaultDeny)
	logNetworkPolicyDetails(t, "controller-manager/default-deny-all", controllerManagerDefaultDeny)
	requireDefaultDenyAll(t, controllerManagerDefaultDeny)

	controllerManagerPolicy := getNetworkPolicyT(t, ctx, kubeClient, util.TargetNamespace, controllerManagerPolicyName)
	logNetworkPolicySummary(t, "controller-manager/allow-controller-manager", controllerManagerPolicy)
	logNetworkPolicyDetails(t, "controller-manager/allow-controller-manager", controllerManagerPolicy)
	requirePodSelectorHasLabel(t, controllerManagerPolicy, "controller-manager")
	requireIngressPort(t, controllerManagerPolicy, corev1.ProtocolTCP, 8443)
	logEgressAllowAll(t, controllerManagerPolicy)

	t.Log("Validating NetworkPolicies in openshift-route-controller-manager")
	routeControllerManagerDefaultDeny := getNetworkPolicyT(t, ctx, kubeClient, util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	logNetworkPolicySummary(t, "route-controller-manager/default-deny-all", routeControllerManagerDefaultDeny)
	logNetworkPolicyDetails(t, "route-controller-manager/default-deny-all", routeControllerManagerDefaultDeny)
	requireDefaultDenyAll(t, routeControllerManagerDefaultDeny)

	routeControllerManagerPolicy := getNetworkPolicyT(t, ctx, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	logNetworkPolicySummary(t, "route-controller-manager/allow-route-controller-manager", routeControllerManagerPolicy)
	logNetworkPolicyDetails(t, "route-controller-manager/allow-route-controller-manager", routeControllerManagerPolicy)
	requirePodSelectorHasLabel(t, routeControllerManagerPolicy, "route-controller-manager")
	requireIngressPort(t, routeControllerManagerPolicy, corev1.ProtocolTCP, 8443)
	logEgressAllowAll(t, routeControllerManagerPolicy)

	t.Log("Validating NetworkPolicies in openshift-controller-manager-operator")
	operatorDefaultDeny := getNetworkPolicyT(t, ctx, kubeClient, util.OperatorNamespace, defaultDenyAllPolicyName)
	logNetworkPolicySummary(t, "operator/default-deny-all", operatorDefaultDeny)
	logNetworkPolicyDetails(t, "operator/default-deny-all", operatorDefaultDeny)
	requireDefaultDenyAll(t, operatorDefaultDeny)

	operatorPolicy := getNetworkPolicyT(t, ctx, kubeClient, util.OperatorNamespace, operatorPolicyName)
	logNetworkPolicySummary(t, "operator/allow-operator", operatorPolicy)
	logNetworkPolicyDetails(t, "operator/allow-operator", operatorPolicy)
	requirePodSelectorLabel(t, operatorPolicy, "app", "openshift-controller-manager-operator")
	requireIngressPort(t, operatorPolicy, corev1.ProtocolTCP, 8443)
	logEgressAllowAll(t, operatorPolicy)

	t.Log("Verifying pods are ready in controller manager namespaces")
	waitForPodsReadyByLabel(t, ctx, kubeClient, util.TargetNamespace, "controller-manager=true")
	waitForPodsReadyByLabel(t, ctx, kubeClient, util.RouteControllerTargetNamespace, "route-controller-manager=true")
	waitForPodsReadyByLabel(t, ctx, kubeClient, util.OperatorNamespace, "app=openshift-controller-manager-operator")
}

func TestControllerManagerNetworkPolicyReconcile(t *testing.T) {
	ctx := context.Background()
	t.Log("Creating Kubernetes clients")
	kubeConfig, err := getKubeConfig()
	if err != nil {
		t.Fatalf("failed to get kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Fatalf("failed to create kubernetes client: %v", err)
	}

	t.Log("Capturing expected NetworkPolicy specs")
	expectedControllerManagerDefaultDeny := getNetworkPolicyT(t, ctx, kubeClient, util.TargetNamespace, defaultDenyAllPolicyName)
	expectedControllerManagerPolicy := getNetworkPolicyT(t, ctx, kubeClient, util.TargetNamespace, controllerManagerPolicyName)
	expectedRouteControllerManagerDefaultDeny := getNetworkPolicyT(t, ctx, kubeClient, util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	expectedRouteControllerManagerPolicy := getNetworkPolicyT(t, ctx, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	expectedOperatorDefaultDeny := getNetworkPolicyT(t, ctx, kubeClient, util.OperatorNamespace, defaultDenyAllPolicyName)
	expectedOperatorPolicy := getNetworkPolicyT(t, ctx, kubeClient, util.OperatorNamespace, operatorPolicyName)

	t.Log("Deleting main policies and waiting for restoration")
	t.Logf("deleting NetworkPolicy %s/%s", util.TargetNamespace, controllerManagerPolicyName)
	restoreNetworkPolicy(t, ctx, kubeClient, expectedControllerManagerPolicy)
	t.Logf("deleting NetworkPolicy %s/%s", util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	restoreNetworkPolicy(t, ctx, kubeClient, expectedRouteControllerManagerPolicy)
	t.Logf("deleting NetworkPolicy %s/%s (operator namespace may need longer to reconcile)", util.OperatorNamespace, operatorPolicyName)
	restoreNetworkPolicyWithTimeout(t, ctx, kubeClient, expectedOperatorPolicy, 15*time.Minute)

	t.Log("Deleting default-deny-all policies and waiting for restoration")
	t.Logf("deleting NetworkPolicy %s/%s", util.TargetNamespace, defaultDenyAllPolicyName)
	restoreNetworkPolicy(t, ctx, kubeClient, expectedControllerManagerDefaultDeny)
	t.Logf("deleting NetworkPolicy %s/%s", util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	restoreNetworkPolicy(t, ctx, kubeClient, expectedRouteControllerManagerDefaultDeny)
	t.Logf("deleting NetworkPolicy %s/%s (operator namespace may need longer to reconcile)", util.OperatorNamespace, defaultDenyAllPolicyName)
	restoreNetworkPolicyWithTimeout(t, ctx, kubeClient, expectedOperatorDefaultDeny, 15*time.Minute)

	t.Log("Mutating main policies and waiting for reconciliation")
	t.Logf("mutating NetworkPolicy %s/%s", util.TargetNamespace, controllerManagerPolicyName)
	mutateAndRestoreNetworkPolicy(t, ctx, kubeClient, util.TargetNamespace, controllerManagerPolicyName)
	t.Logf("mutating NetworkPolicy %s/%s", util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	mutateAndRestoreNetworkPolicy(t, ctx, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	t.Logf("mutating NetworkPolicy %s/%s (operator namespace may need longer to reconcile)", util.OperatorNamespace, operatorPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(t, ctx, kubeClient, util.OperatorNamespace, operatorPolicyName, 15*time.Minute)

	t.Log("Mutating default-deny-all policies and waiting for reconciliation")
	t.Logf("mutating NetworkPolicy %s/%s", util.TargetNamespace, defaultDenyAllPolicyName)
	mutateAndRestoreNetworkPolicy(t, ctx, kubeClient, util.TargetNamespace, defaultDenyAllPolicyName)
	t.Logf("mutating NetworkPolicy %s/%s", util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	mutateAndRestoreNetworkPolicy(t, ctx, kubeClient, util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	t.Logf("mutating NetworkPolicy %s/%s (operator namespace may need longer to reconcile)", util.OperatorNamespace, defaultDenyAllPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(t, ctx, kubeClient, util.OperatorNamespace, defaultDenyAllPolicyName, 15*time.Minute)

	t.Log("Checking NetworkPolicy-related events (best-effort)")
	logNetworkPolicyEvents(t, ctx, kubeClient, []string{util.OperatorNamespace, util.TargetNamespace, util.RouteControllerTargetNamespace}, controllerManagerPolicyName)
}

func getKubeConfig() (*restclient.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}

func getNetworkPolicyT(t *testing.T, ctx context.Context, client kubernetes.Interface, namespace, name string) *networkingv1.NetworkPolicy {
	t.Helper()
	policy, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get NetworkPolicy %s/%s: %v", namespace, name, err)
	}
	return policy
}

func requireDefaultDenyAll(t *testing.T, policy *networkingv1.NetworkPolicy) {
	t.Helper()
	if len(policy.Spec.PodSelector.MatchLabels) != 0 || len(policy.Spec.PodSelector.MatchExpressions) != 0 {
		t.Errorf("%s/%s: expected empty podSelector", policy.Namespace, policy.Name)
	}
	if len(policy.Spec.Ingress) != 0 || len(policy.Spec.Egress) != 0 {
		t.Errorf("%s/%s: expected no ingress/egress rules, got ingress=%d egress=%d", policy.Namespace, policy.Name, len(policy.Spec.Ingress), len(policy.Spec.Egress))
	}

	policyTypes := sets.NewString()
	for _, policyType := range policy.Spec.PolicyTypes {
		policyTypes.Insert(string(policyType))
	}
	if !policyTypes.Has(string(networkingv1.PolicyTypeIngress)) || !policyTypes.Has(string(networkingv1.PolicyTypeEgress)) {
		t.Errorf("%s/%s: expected both Ingress and Egress policyTypes, got %v", policy.Namespace, policy.Name, policy.Spec.PolicyTypes)
	}
}

func requirePodSelectorLabel(t *testing.T, policy *networkingv1.NetworkPolicy, key, value string) {
	t.Helper()
	actual, ok := policy.Spec.PodSelector.MatchLabels[key]
	if !ok || actual != value {
		t.Errorf("%s/%s: expected podSelector %s=%s, got %v", policy.Namespace, policy.Name, key, value, policy.Spec.PodSelector.MatchLabels)
	}
}

func requirePodSelectorHasLabel(t *testing.T, policy *networkingv1.NetworkPolicy, key string) {
	t.Helper()
	if _, ok := policy.Spec.PodSelector.MatchLabels[key]; !ok {
		t.Errorf("%s/%s: expected podSelector to have label %s, got %v", policy.Namespace, policy.Name, key, policy.Spec.PodSelector.MatchLabels)
	}
}

func requireIngressPort(t *testing.T, policy *networkingv1.NetworkPolicy, protocol corev1.Protocol, port int32) {
	t.Helper()
	if !hasPortInIngress(policy.Spec.Ingress, protocol, port) {
		t.Errorf("%s/%s: expected ingress port %s/%d", policy.Namespace, policy.Name, protocol, port)
	}
}

func requireEgressPort(t *testing.T, policy *networkingv1.NetworkPolicy, protocol corev1.Protocol, port int32) {
	t.Helper()
	if !hasPortInEgress(policy.Spec.Egress, protocol, port) {
		t.Errorf("%s/%s: expected egress port %s/%d", policy.Namespace, policy.Name, protocol, port)
	}
}

func hasPortInIngress(rules []networkingv1.NetworkPolicyIngressRule, protocol corev1.Protocol, port int32) bool {
	for _, rule := range rules {
		if hasPort(rule.Ports, protocol, port) {
			return true
		}
	}
	return false
}

func hasPortInEgress(rules []networkingv1.NetworkPolicyEgressRule, protocol corev1.Protocol, port int32) bool {
	for _, rule := range rules {
		if hasPort(rule.Ports, protocol, port) {
			return true
		}
	}
	return false
}

func hasPort(ports []networkingv1.NetworkPolicyPort, protocol corev1.Protocol, port int32) bool {
	for _, p := range ports {
		if p.Port == nil || p.Port.IntValue() != int(port) {
			continue
		}
		if p.Protocol == nil || *p.Protocol == protocol {
			return true
		}
	}
	return false
}

func logEgressAllowAll(t *testing.T, policy *networkingv1.NetworkPolicy) {
	t.Helper()
	if hasEgressAllowAll(policy.Spec.Egress) {
		t.Logf("networkpolicy %s/%s: egress allow-all rule present", policy.Namespace, policy.Name)
		return
	}
	t.Logf("networkpolicy %s/%s: no egress allow-all rule", policy.Namespace, policy.Name)
}

func hasEgressAllowAll(rules []networkingv1.NetworkPolicyEgressRule) bool {
	for _, rule := range rules {
		if len(rule.To) == 0 && len(rule.Ports) == 0 {
			return true
		}
	}
	return false
}

func restoreNetworkPolicy(t *testing.T, ctx context.Context, client kubernetes.Interface, expected *networkingv1.NetworkPolicy) {
	restoreNetworkPolicyWithTimeout(t, ctx, client, expected, 10*time.Minute)
}

func restoreNetworkPolicyWithTimeout(t *testing.T, ctx context.Context, client kubernetes.Interface, expected *networkingv1.NetworkPolicy, timeout time.Duration) {
	t.Helper()
	namespace := expected.Namespace
	name := expected.Name
	t.Logf("deleting NetworkPolicy %s/%s", namespace, name)
	if err := client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("failed to delete NetworkPolicy %s/%s: %v", namespace, name, err)
	}
	err := wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		current, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return equality.Semantic.DeepEqual(expected.Spec, current.Spec), nil
	})
	if err != nil {
		t.Fatalf("timed out waiting for NetworkPolicy %s/%s spec to be restored after %v: %v", namespace, name, timeout, err)
	}
	t.Logf("NetworkPolicy %s/%s spec restored after delete", namespace, name)
}

func mutateAndRestoreNetworkPolicy(t *testing.T, ctx context.Context, client kubernetes.Interface, namespace, name string) {
	mutateAndRestoreNetworkPolicyWithTimeout(t, ctx, client, namespace, name, 10*time.Minute)
}

func mutateAndRestoreNetworkPolicyWithTimeout(t *testing.T, ctx context.Context, client kubernetes.Interface, namespace, name string, timeout time.Duration) {
	t.Helper()
	original := getNetworkPolicyT(t, ctx, client, namespace, name)
	t.Logf("mutating NetworkPolicy %s/%s (podSelector override)", namespace, name)
	patch := []byte(`{"spec":{"podSelector":{"matchLabels":{"np-reconcile":"mutated"}}}}`)
	_, err := client.NetworkingV1().NetworkPolicies(namespace).Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		t.Fatalf("failed to patch NetworkPolicy %s/%s: %v", namespace, name, err)
	}

	err = wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		current := getNetworkPolicyT(t, ctx, client, namespace, name)
		return equality.Semantic.DeepEqual(original.Spec, current.Spec), nil
	})
	if err != nil {
		t.Fatalf("timed out waiting for NetworkPolicy %s/%s spec to be restored after %v: %v", namespace, name, timeout, err)
	}
	t.Logf("NetworkPolicy %s/%s spec restored", namespace, name)
}

func waitForPodsReadyByLabel(t *testing.T, ctx context.Context, client kubernetes.Interface, namespace, labelSelector string) {
	t.Helper()
	t.Logf("waiting for pods ready in %s with selector %s", namespace, labelSelector)
	err := wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return false, err
		}
		if len(pods.Items) == 0 {
			return false, nil
		}
		for _, pod := range pods.Items {
			if !isPodReady(&pod) {
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("timed out waiting for pods in %s with selector %s to be ready: %v", namespace, labelSelector, err)
	}
}

func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func logNetworkPolicyEvents(t *testing.T, ctx context.Context, client kubernetes.Interface, namespaces []string, policyName string) {
	t.Helper()
	found := false
	_ = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		for _, namespace := range namespaces {
			events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				t.Logf("unable to list events in %s: %v", namespace, err)
				continue
			}
			for _, event := range events.Items {
				if event.InvolvedObject.Kind == "NetworkPolicy" && event.InvolvedObject.Name == policyName {
					t.Logf("event in %s: %s %s %s", namespace, event.Type, event.Reason, event.Message)
					found = true
				}
				if event.Message != "" && (event.InvolvedObject.Name == policyName || event.InvolvedObject.Kind == "NetworkPolicy") {
					t.Logf("event in %s: %s %s %s", namespace, event.Type, event.Reason, event.Message)
					found = true
				}
			}
		}
		if found {
			return true, nil
		}
		t.Logf("no NetworkPolicy events yet for %s (namespaces: %v)", policyName, namespaces)
		return false, nil
	})
	if !found {
		t.Logf("no NetworkPolicy events observed for %s (best-effort)", policyName)
	}
}

func logNetworkPolicySummary(t *testing.T, label string, policy *networkingv1.NetworkPolicy) {
	t.Logf("networkpolicy %s namespace=%s name=%s podSelector=%v policyTypes=%v ingress=%d egress=%d",
		label,
		policy.Namespace,
		policy.Name,
		policy.Spec.PodSelector.MatchLabels,
		policy.Spec.PolicyTypes,
		len(policy.Spec.Ingress),
		len(policy.Spec.Egress),
	)
}

func logNetworkPolicyDetails(t *testing.T, label string, policy *networkingv1.NetworkPolicy) {
	t.Helper()
	t.Logf("networkpolicy %s details:", label)
	t.Logf("  podSelector=%v policyTypes=%v", policy.Spec.PodSelector.MatchLabels, policy.Spec.PolicyTypes)
	for i, rule := range policy.Spec.Ingress {
		t.Logf("  ingress[%d]: ports=%s from=%s", i, formatPorts(rule.Ports), formatPeers(rule.From))
	}
	for i, rule := range policy.Spec.Egress {
		t.Logf("  egress[%d]: ports=%s to=%s", i, formatPorts(rule.Ports), formatPeers(rule.To))
	}
}

func formatPorts(ports []networkingv1.NetworkPolicyPort) string {
	if len(ports) == 0 {
		return "[]"
	}
	out := make([]string, 0, len(ports))
	for _, p := range ports {
		proto := "TCP"
		if p.Protocol != nil {
			proto = string(*p.Protocol)
		}
		if p.Port == nil {
			out = append(out, fmt.Sprintf("%s:any", proto))
			continue
		}
		out = append(out, fmt.Sprintf("%s:%s", proto, p.Port.String()))
	}
	return fmt.Sprintf("[%s]", joinStrings(out))
}

func formatPeers(peers []networkingv1.NetworkPolicyPeer) string {
	if len(peers) == 0 {
		return "[]"
	}
	out := make([]string, 0, len(peers))
	for _, peer := range peers {
		ns := formatSelector(peer.NamespaceSelector)
		pod := formatSelector(peer.PodSelector)
		if ns == "" && pod == "" {
			out = append(out, "{}")
			continue
		}
		out = append(out, fmt.Sprintf("ns=%s pod=%s", ns, pod))
	}
	return fmt.Sprintf("[%s]", joinStrings(out))
}

func formatSelector(sel *metav1.LabelSelector) string {
	if sel == nil {
		return ""
	}
	if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
		return "{}"
	}
	return fmt.Sprintf("labels=%v exprs=%v", sel.MatchLabels, sel.MatchExpressions)
}

func joinStrings(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for i := 1; i < len(items); i++ {
		out += ", " + items[i]
	}
	return out
}
