package e2e

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	agnhostImage = "registry.k8s.io/e2e-test-images/agnhost:2.45"

	// Poll intervals
	defaultPollInterval = 5 * time.Second
	fastPollInterval    = 2 * time.Second

	// Timeouts
	shortTimeout             = 30 * time.Second
	connectivityTimeout      = 2 * time.Minute
	podReadyTimeout          = 5 * time.Minute
	reconcileTimeout         = 10 * time.Minute
	operatorReconcileTimeout = 15 * time.Minute
)

// getKubeConfig creates a *rest.Config for talking to a Kubernetes apiserver.
func getKubeConfig() (*restclient.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}

// getNetworkPolicy retrieves a NetworkPolicy from the cluster or fails the test.
func getNetworkPolicy(ctx context.Context, client kubernetes.Interface, namespace, name string) *networkingv1.NetworkPolicy {
	g.GinkgoHelper()
	policy, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get NetworkPolicy %s/%s", namespace, name)
	return policy
}

// requireDefaultDenyAll validates that a policy is a default deny-all policy.
func requireDefaultDenyAll(policy *networkingv1.NetworkPolicy) {
	g.GinkgoHelper()
	o.Expect(len(policy.Spec.PodSelector.MatchLabels) == 0 && len(policy.Spec.PodSelector.MatchExpressions) == 0).To(o.BeTrue(),
		"%s/%s: expected empty podSelector", policy.Namespace, policy.Name)
	o.Expect(len(policy.Spec.Ingress) == 0 && len(policy.Spec.Egress) == 0).To(o.BeTrue(),
		"%s/%s: expected no ingress/egress rules, got ingress=%d egress=%d", policy.Namespace, policy.Name, len(policy.Spec.Ingress), len(policy.Spec.Egress))

	policyTypes := sets.New[string]()
	for _, policyType := range policy.Spec.PolicyTypes {
		policyTypes.Insert(string(policyType))
	}
	o.Expect(policyTypes.Has(string(networkingv1.PolicyTypeIngress)) && policyTypes.Has(string(networkingv1.PolicyTypeEgress))).To(o.BeTrue(),
		"%s/%s: expected both Ingress and Egress policyTypes, got %v", policy.Namespace, policy.Name, policy.Spec.PolicyTypes)
}

// requirePodSelectorLabel validates that a policy has a specific podSelector label with value.
func requirePodSelectorLabel(policy *networkingv1.NetworkPolicy, key, value string) {
	g.GinkgoHelper()
	actual, ok := policy.Spec.PodSelector.MatchLabels[key]
	o.Expect(ok && actual == value).To(o.BeTrue(),
		"%s/%s: expected podSelector %s=%s, got %v", policy.Namespace, policy.Name, key, value, policy.Spec.PodSelector.MatchLabels)
}

// requireIngressPort validates that a policy allows a specific ingress port.
func requireIngressPort(policy *networkingv1.NetworkPolicy, protocol corev1.Protocol, port int32) {
	g.GinkgoHelper()
	o.Expect(hasPortInIngress(policy.Spec.Ingress, protocol, port)).To(o.BeTrue(),
		"%s/%s: expected ingress port %s/%d", policy.Namespace, policy.Name, protocol, port)
}

// requireIngressFromNamespace validates that a policy allows ingress from a specific namespace.
func requireIngressFromNamespace(policy *networkingv1.NetworkPolicy, port int32, namespace string) {
	g.GinkgoHelper()
	o.Expect(hasIngressFromNamespace(policy.Spec.Ingress, port, namespace)).To(o.BeTrue(),
		"%s/%s: expected ingress from namespace %s on port %d", policy.Namespace, policy.Name, namespace, port)
}

// logIngressFromNamespaceOptional logs whether ingress from a namespace is present.
func logIngressFromNamespaceOptional(policy *networkingv1.NetworkPolicy, port int32, namespace string) {
	g.GinkgoHelper()
	if hasIngressFromNamespace(policy.Spec.Ingress, port, namespace) {
		g.GinkgoWriter.Printf("networkpolicy %s/%s: ingress from namespace %s present on port %d\n", policy.Namespace, policy.Name, namespace, port)
		return
	}
	g.GinkgoWriter.Printf("networkpolicy %s/%s: no ingress from namespace %s on port %d\n", policy.Namespace, policy.Name, namespace, port)
}

// requireIngressFromNamespaceOrPolicyGroup validates ingress from either a namespace or a policy group.
func requireIngressFromNamespaceOrPolicyGroup(policy *networkingv1.NetworkPolicy, port int32, namespace, policyGroupLabelKey string) {
	g.GinkgoHelper()
	if hasIngressFromNamespace(policy.Spec.Ingress, port, namespace) {
		return
	}
	if hasIngressFromPolicyGroup(policy.Spec.Ingress, port, policyGroupLabelKey) {
		return
	}
	g.Fail(fmt.Sprintf("%s/%s: expected ingress from namespace %s or policy-group %s on port %d", policy.Namespace, policy.Name, namespace, policyGroupLabelKey, port))
}

// logIngressHostNetworkOrAllowAll logs whether a policy has host-network or allow-all ingress.
func logIngressHostNetworkOrAllowAll(policy *networkingv1.NetworkPolicy, port int32) {
	g.GinkgoHelper()
	if hasIngressAllowAll(policy.Spec.Ingress, port) {
		g.GinkgoWriter.Printf("networkpolicy %s/%s: ingress allow-all present on port %d\n", policy.Namespace, policy.Name, port)
		return
	}
	if hasIngressFromPolicyGroup(policy.Spec.Ingress, port, "policy-group.network.openshift.io/host-network") {
		g.GinkgoWriter.Printf("networkpolicy %s/%s: ingress host-network policy-group present on port %d\n", policy.Namespace, policy.Name, port)
		return
	}
	g.GinkgoWriter.Printf("networkpolicy %s/%s: no ingress allow-all/host-network rule on port %d\n", policy.Namespace, policy.Name, port)
}

// requireEgressPort validates that a policy allows a specific egress port.
func requireEgressPort(policy *networkingv1.NetworkPolicy, protocol corev1.Protocol, port int32) {
	g.GinkgoHelper()
	o.Expect(hasPortInEgress(policy.Spec.Egress, protocol, port)).To(o.BeTrue(),
		"%s/%s: expected egress port %s/%d", policy.Namespace, policy.Name, protocol, port)
}

// requireEgressAllowAllTCP validates that a policy has an egress allow-all TCP rule.
func requireEgressAllowAllTCP(policy *networkingv1.NetworkPolicy) {
	g.GinkgoHelper()
	o.Expect(hasEgressAllowAllTCP(policy.Spec.Egress)).To(o.BeTrue(),
		"%s/%s: expected egress allow-all TCP rule", policy.Namespace, policy.Name)
}

// hasPortInIngress checks if any ingress rule contains the specified port.
func hasPortInIngress(rules []networkingv1.NetworkPolicyIngressRule, protocol corev1.Protocol, port int32) bool {
	for _, rule := range rules {
		if hasPort(rule.Ports, protocol, port) {
			return true
		}
	}
	return false
}

// hasPort checks if a port specification contains the specified port and protocol.
func hasPort(ports []networkingv1.NetworkPolicyPort, protocol corev1.Protocol, port int32) bool {
	for _, p := range ports {
		if p.Port == nil || p.Port.IntValue() != int(port) {
			continue
		}
		pProto := corev1.ProtocolTCP
		if p.Protocol != nil {
			pProto = *p.Protocol
		}
		if pProto == protocol {
			return true
		}
	}
	return false
}

// hasPortInEgress checks if any egress rule contains the specified port.
func hasPortInEgress(rules []networkingv1.NetworkPolicyEgressRule, protocol corev1.Protocol, port int32) bool {
	for _, rule := range rules {
		if hasPort(rule.Ports, protocol, port) {
			return true
		}
	}
	return false
}

// hasIngressFromNamespace checks if any ingress rule allows traffic from a specific namespace.
func hasIngressFromNamespace(rules []networkingv1.NetworkPolicyIngressRule, port int32, namespace string) bool {
	for _, rule := range rules {
		if !hasPort(rule.Ports, corev1.ProtocolTCP, port) {
			continue
		}
		for _, peer := range rule.From {
			if peer.NamespaceSelector != nil && nsMatch(peer.NamespaceSelector, namespace) {
				return true
			}
		}
	}
	return false
}

// hasIngressAllowAll checks if any ingress rule allows all traffic on a specific port.
func hasIngressAllowAll(rules []networkingv1.NetworkPolicyIngressRule, port int32) bool {
	for _, rule := range rules {
		if !hasPort(rule.Ports, corev1.ProtocolTCP, port) {
			continue
		}
		if len(rule.From) == 0 {
			return true
		}
	}
	return false
}

// hasIngressFromPolicyGroup checks if any ingress rule allows traffic from a policy group.
func hasIngressFromPolicyGroup(rules []networkingv1.NetworkPolicyIngressRule, port int32, policyGroupLabelKey string) bool {
	for _, rule := range rules {
		if !hasPort(rule.Ports, corev1.ProtocolTCP, port) {
			continue
		}
		for _, peer := range rule.From {
			if peer.NamespaceSelector == nil || peer.NamespaceSelector.MatchLabels == nil {
				continue
			}
			if _, ok := peer.NamespaceSelector.MatchLabels[policyGroupLabelKey]; ok {
				return true
			}
		}
	}
	return false
}

// logEgressAllowAll logs whether a policy has an egress allow-all rule.
func logEgressAllowAll(policy *networkingv1.NetworkPolicy) {
	g.GinkgoHelper()
	if hasEgressAllowAll(policy.Spec.Egress) {
		g.GinkgoWriter.Printf("networkpolicy %s/%s: egress allow-all rule present\n", policy.Namespace, policy.Name)
		return
	}
	g.GinkgoWriter.Printf("networkpolicy %s/%s: no egress allow-all rule\n", policy.Namespace, policy.Name)
}

// logEgressAllowAllTCP logs whether a policy has an egress allow-all TCP rule.
func logEgressAllowAllTCP(policy *networkingv1.NetworkPolicy) {
	g.GinkgoHelper()
	if hasEgressAllowAllTCP(policy.Spec.Egress) {
		g.GinkgoWriter.Printf("networkpolicy %s/%s: egress allow-all TCP rule present\n", policy.Namespace, policy.Name)
		return
	}
	g.GinkgoWriter.Printf("networkpolicy %s/%s: no egress allow-all TCP rule\n", policy.Namespace, policy.Name)
}

// hasEgressAllowAll checks if any egress rule is an allow-all rule.
func hasEgressAllowAll(rules []networkingv1.NetworkPolicyEgressRule) bool {
	for _, rule := range rules {
		if len(rule.To) == 0 && len(rule.Ports) == 0 {
			return true
		}
	}
	return false
}

// hasEgressAllowAllTCP checks if any egress rule allows all TCP traffic.
func hasEgressAllowAllTCP(rules []networkingv1.NetworkPolicyEgressRule) bool {
	for _, rule := range rules {
		if len(rule.To) != 0 {
			continue
		}
		if hasAnyTCPPort(rule.Ports) {
			return true
		}
	}
	return false
}

// hasAnyTCPPort checks if a ports list contains any TCP port.
func hasAnyTCPPort(ports []networkingv1.NetworkPolicyPort) bool {
	if len(ports) == 0 {
		return true
	}
	for _, p := range ports {
		if p.Protocol != nil && *p.Protocol != corev1.ProtocolTCP {
			continue
		}
		return true
	}
	return false
}

// restoreNetworkPolicyWithTimeout deletes a NetworkPolicy and waits for it to be restored with a custom timeout.
func restoreNetworkPolicyWithTimeout(ctx context.Context, client kubernetes.Interface, expected *networkingv1.NetworkPolicy, timeout time.Duration) {
	g.GinkgoHelper()
	namespace := expected.Namespace
	name := expected.Name
	originalUID := expected.UID
	g.GinkgoWriter.Printf("deleting NetworkPolicy %s/%s (UID: %s)\n", namespace, name, originalUID)
	err := client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to delete NetworkPolicy %s/%s", namespace, name)

	// Track whether deletion has been observed by checking UID change
	deletionObserved := false
	err = wait.PollUntilContextTimeout(ctx, defaultPollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		current, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				deletionObserved = true
				g.GinkgoWriter.Printf("NetworkPolicy %s/%s deletion observed (NotFound)\n", namespace, name)
				return false, nil
			}
			return false, err
		}
		// Check if this is a new instance (different UID) indicating recreation
		if current.UID != originalUID {
			deletionObserved = true
			g.GinkgoWriter.Printf("NetworkPolicy %s/%s recreated with new UID: %s\n", namespace, name, current.UID)
		}
		// Only check for restoration after we've observed the deletion
		if !deletionObserved {
			return false, nil
		}
		return equality.Semantic.DeepEqual(expected.Spec, current.Spec), nil
	})
	o.Expect(err).NotTo(o.HaveOccurred(), "timed out waiting for NetworkPolicy %s/%s spec to be restored after %v", namespace, name, timeout)
	g.GinkgoWriter.Printf("NetworkPolicy %s/%s spec restored after delete\n", namespace, name)
}

// mutateAndRestoreNetworkPolicyWithTimeout mutates a NetworkPolicy and waits for it to be reconciled with a custom timeout.
func mutateAndRestoreNetworkPolicyWithTimeout(ctx context.Context, client kubernetes.Interface, namespace, name string, timeout time.Duration) {
	g.GinkgoHelper()
	original := getNetworkPolicy(ctx, client, namespace, name)
	g.GinkgoWriter.Printf("mutating NetworkPolicy %s/%s (podSelector override)\n", namespace, name)
	patch := []byte(`{"spec":{"podSelector":{"matchLabels":{"np-reconcile":"mutated"}}}}`)
	_, err := client.NetworkingV1().NetworkPolicies(namespace).Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to patch NetworkPolicy %s/%s", namespace, name)

	err = wait.PollUntilContextTimeout(ctx, defaultPollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		current := getNetworkPolicy(ctx, client, namespace, name)
		return equality.Semantic.DeepEqual(original.Spec, current.Spec), nil
	})
	o.Expect(err).NotTo(o.HaveOccurred(), "timed out waiting for NetworkPolicy %s/%s spec to be restored after %v", namespace, name, timeout)
	g.GinkgoWriter.Printf("NetworkPolicy %s/%s spec restored\n", namespace, name)
}

// waitForPodsReadyByLabel waits for all pods matching a label selector to be ready.
func waitForPodsReadyByLabel(ctx context.Context, client kubernetes.Interface, namespace, labelSelector string) {
	g.GinkgoHelper()
	g.GinkgoWriter.Printf("waiting for pods ready in %s with selector %s\n", namespace, labelSelector)
	err := wait.PollUntilContextTimeout(ctx, defaultPollInterval, podReadyTimeout, true, func(ctx context.Context) (bool, error) {
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
	o.Expect(err).NotTo(o.HaveOccurred(), "timed out waiting for pods in %s with selector %s to be ready", namespace, labelSelector)
}

// getReadyPodIPs returns IPs of all ready pods matching a label selector.
func getReadyPodIPs(ctx context.Context, client kubernetes.Interface, namespace, labelSelector string) []string {
	g.GinkgoHelper()
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to list pods in %s with selector %s", namespace, labelSelector)

	var ips []string
	for _, pod := range pods.Items {
		if !isPodReady(&pod) {
			continue
		}
		ips = append(ips, podIPs(&pod)...)
	}
	return ips
}

// getActualPodLabels returns the labels from the first ready pod matching a label selector,
// filtered to include only application labels relevant for NetworkPolicy validation.
// This ensures test pods use the same labels as actual production pods for NetworkPolicy validation
// while avoiding system labels that would interfere with pod creation/scheduling.
func getActualPodLabels(ctx context.Context, client kubernetes.Interface, namespace, labelSelector string) map[string]string {
	g.GinkgoHelper()
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to list pods in %s with selector %s", namespace, labelSelector)

	for _, pod := range pods.Items {
		if isPodReady(&pod) {
			return filterPodLabels(pod.Labels)
		}
	}
	return nil
}

// filterPodLabels filters out system and controller-managed labels that should not be copied to test pods.
// Keeps only application-specific labels that NetworkPolicy selectors care about.
func filterPodLabels(labels map[string]string) map[string]string {
	filtered := make(map[string]string)

	// List of label prefixes/keys to exclude
	excludePrefixes := []string{
		"topology.kubernetes.io/",            // Node topology labels (zone, region)
		"kubernetes.io/hostname",             // Node hostname
		"pod-template-hash",                  // ReplicaSet/Deployment hash (causes ReplicaSet to delete test pods!)
		"controller-revision-hash",           // StatefulSet/DaemonSet hash
		"statefulset.kubernetes.io/pod-name", // StatefulSet pod name
		"batch.kubernetes.io/",               // Job/CronJob labels
	}

	for key, value := range labels {
		exclude := false
		for _, prefix := range excludePrefixes {
			if key == prefix || strings.HasPrefix(key, prefix) {
				exclude = true
				break
			}
		}
		if !exclude {
			filtered[key] = value
		}
	}

	return filtered
}

// isPodReady checks if a pod is ready.
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// createServerPod creates a test server pod and returns its IPs and a cleanup function.
func createServerPod(ctx context.Context, kubeClient kubernetes.Interface, namespace, name string, labels map[string]string, port int32) ([]string, func()) {
	g.GinkgoHelper()

	g.GinkgoWriter.Printf("creating server pod %s/%s port=%d labels=%v\n", namespace, name, port, labels)
	pod := netexecPod(name, namespace, labels, port)
	_, err := kubeClient.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to create server pod")
	o.Expect(waitForPodReady(ctx, kubeClient, namespace, name)).NotTo(o.HaveOccurred())

	created, err := kubeClient.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get created server pod")
	o.Expect(created.Status.PodIPs).NotTo(o.BeEmpty())

	ips := podIPs(created)
	g.GinkgoWriter.Printf("server pod %s/%s ips=%v\n", namespace, name, ips)

	return ips, func() {
		g.GinkgoWriter.Printf("deleting server pod %s/%s\n", namespace, name)
		if err := kubeClient.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
			g.GinkgoWriter.Printf("failed to delete server pod %s/%s: %v\n", namespace, name, err)
		}
	}
}

// netexecPod creates a pod spec running the agnhost netexec server.
func netexecPod(name, namespace string, labels map[string]string, port int32) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"openshift.io/required-scc": "nonroot-v2",
			},
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot:   boolptr(true),
				RunAsUser:      int64ptr(1001),
				SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
			},
			Containers: []corev1.Container{
				{
					Name:                     "netexec",
					Image:                    agnhostImage,
					TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: boolptr(false),
						Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
						RunAsNonRoot:             boolptr(true),
						RunAsUser:                int64ptr(1001),
					},
					Command: []string{"/agnhost"},
					Args:    []string{"netexec", fmt.Sprintf("--http-port=%d", port)},
					Ports: []corev1.ContainerPort{
						{ContainerPort: port},
					},
				},
			},
		},
	}
}

// expectConnectivity checks connectivity to all provided IPs (dual-stack aware).
func expectConnectivity(ctx context.Context, kubeClient kubernetes.Interface, namespace string, clientLabels map[string]string, serverIPs []string, port int32, shouldSucceed bool) {
	g.GinkgoHelper()

	for _, ip := range serverIPs {
		family := "IPv4"
		if isIPv6(ip) {
			family = "IPv6"
		}
		g.GinkgoWriter.Printf("checking %s connectivity %s -> %s expected=%t\n", family, namespace, formatIPPort(ip, port), shouldSucceed)
		expectConnectivityForIP(ctx, kubeClient, namespace, clientLabels, ip, port, shouldSucceed)
	}
}

// expectConnectivityForIP checks connectivity to a single IP address.
func expectConnectivityForIP(ctx context.Context, kubeClient kubernetes.Interface, namespace string, clientLabels map[string]string, serverIP string, port int32, shouldSucceed bool) {
	g.GinkgoHelper()
	podName, cleanup, err := createConnectivityClientPod(ctx, kubeClient, namespace, clientLabels, serverIP, port)
	o.Expect(err).NotTo(o.HaveOccurred())
	g.DeferCleanup(cleanup)

	err = wait.PollUntilContextTimeout(ctx, defaultPollInterval, connectivityTimeout, true, func(ctx context.Context) (bool, error) {
		succeeded, err := readConnectivityResult(ctx, kubeClient, namespace, podName)
		if err != nil {
			// Only swallow expected transient errors (pod not ready, logs not available yet)
			if strings.Contains(err.Error(), "no connectivity result yet") {
				g.GinkgoWriter.Printf("waiting for connectivity result: %v\n", err)
				return false, nil
			}
			// Fail fast on unexpected errors (RBAC, apiserver, permission issues)
			return false, err
		}
		return succeeded == shouldSucceed, nil
	})
	o.Expect(err).NotTo(o.HaveOccurred())
	g.GinkgoWriter.Printf("connectivity %s/%s expected=%t\n", namespace, formatIPPort(serverIP, port), shouldSucceed)
}

// createConnectivityClientPod creates a long-running pod that continuously probes
// TCP connectivity and writes results to stdout. Callers read the pod's logs
// to determine the latest result, avoiding per-poll pod create/delete overhead.
func createConnectivityClientPod(ctx context.Context, kubeClient kubernetes.Interface, namespace string, labels map[string]string, serverIP string, port int32) (string, func(), error) {
	name := fmt.Sprintf("np-client-%s", rand.String(5))
	target := formatIPPort(serverIP, port)

	g.GinkgoWriter.Printf("creating client pod %s/%s to probe %s\n", namespace, name, target)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"openshift.io/required-scc": "nonroot-v2",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot:   boolptr(true),
				RunAsUser:      int64ptr(1001),
				SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
			},
			Containers: []corev1.Container{
				{
					Name:                     "connect",
					Image:                    agnhostImage,
					TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: boolptr(false),
						Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
						RunAsNonRoot:             boolptr(true),
						RunAsUser:                int64ptr(1001),
					},
					Command: []string{"/bin/sh", "-c"},
					Args: []string{
						fmt.Sprintf("while true; do if /agnhost connect --protocol=tcp --timeout=5s %s 2>/dev/null; then echo CONN_OK; else echo CONN_FAIL; fi; sleep 3; done", target),
					},
				},
			},
		},
	}

	_, err := kubeClient.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", nil, err
	}

	if err := waitForPodReady(ctx, kubeClient, namespace, name); err != nil {
		if delErr := kubeClient.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{}); delErr != nil {
			g.GinkgoWriter.Printf("failed to delete failed client pod %s/%s: %v\n", namespace, name, delErr)
		}
		return "", nil, fmt.Errorf("client pod %s/%s never became ready: %w", namespace, name, err)
	}

	cleanup := func() {
		g.GinkgoWriter.Printf("deleting client pod %s/%s\n", namespace, name)
		if err := kubeClient.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
			g.GinkgoWriter.Printf("failed to delete client pod %s/%s: %v\n", namespace, name, err)
		}
	}

	return name, cleanup, nil
}

// readConnectivityResult reads the last connectivity result from a client pod's logs.
func readConnectivityResult(ctx context.Context, kubeClient kubernetes.Interface, namespace, podName string) (bool, error) {
	tailLines := int64(1)
	raw, err := kubeClient.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tailLines,
	}).DoRaw(ctx)
	if err != nil {
		return false, err
	}

	line := strings.TrimSpace(string(raw))
	if line == "" {
		return false, fmt.Errorf("no connectivity result yet from pod %s/%s", namespace, podName)
	}

	g.GinkgoWriter.Printf("client pod %s/%s result=%s\n", namespace, podName, line)
	return line == "CONN_OK", nil
}

// nsMatch checks if a namespace selector matches a namespace name.
func nsMatch(selector *metav1.LabelSelector, namespace string) bool {
	if selector == nil {
		return true
	}
	if selector.MatchLabels != nil {
		if selector.MatchLabels["kubernetes.io/metadata.name"] == namespace {
			return true
		}
	}
	for _, expr := range selector.MatchExpressions {
		if expr.Key != "kubernetes.io/metadata.name" {
			continue
		}
		if expr.Operator != metav1.LabelSelectorOpIn {
			continue
		}
		if slices.Contains(expr.Values, namespace) {
			return true
		}
	}
	return false
}

// podMatch checks if a pod selector matches a set of labels.
func podMatch(selector *metav1.LabelSelector, labels map[string]string) bool {
	if selector == nil {
		return true
	}
	for key, value := range selector.MatchLabels {
		if labels[key] != value {
			return false
		}
	}
	return true
}

// waitForPodReady waits for a pod to be ready.
func waitForPodReady(ctx context.Context, kubeClient kubernetes.Interface, namespace, name string) error {
	return wait.PollUntilContextTimeout(ctx, fastPollInterval, connectivityTimeout, true, func(ctx context.Context) (bool, error) {
		pod, err := kubeClient.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}

// logNetworkPolicyEvents logs events related to NetworkPolicies.
func logNetworkPolicyEvents(ctx context.Context, client kubernetes.Interface, namespaces []string, policyName string) {
	g.GinkgoHelper()
	found := false
	_ = wait.PollUntilContextTimeout(ctx, defaultPollInterval, connectivityTimeout, true, func(ctx context.Context) (bool, error) {
		for _, namespace := range namespaces {
			events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				g.GinkgoWriter.Printf("unable to list events in %s: %v\n", namespace, err)
				continue
			}
			for _, event := range events.Items {
				// Check if event is directly on a NetworkPolicy object with matching name
				if event.InvolvedObject.Kind == "NetworkPolicy" && event.InvolvedObject.Name == policyName {
					g.GinkgoWriter.Printf("event in %s: %s %s %s\n", namespace, event.Type, event.Reason, event.Message)
					found = true
				}
				// Also check if the event message mentions the policy name
				// (operator emits events on Deployment with policy name in message)
				if event.Message != "" && strings.Contains(event.Message, policyName) {
					g.GinkgoWriter.Printf("event in %s: %s %s %s\n", namespace, event.Type, event.Reason, event.Message)
					found = true
				}
			}
		}
		if found {
			return true, nil
		}
		g.GinkgoWriter.Printf("no NetworkPolicy events yet for %s (namespaces: %v)\n", policyName, namespaces)
		return false, nil
	})
	if !found {
		g.GinkgoWriter.Printf("no NetworkPolicy events observed for %s (best-effort)\n", policyName)
	}
}

// logNetworkPolicySummary logs a summary of a NetworkPolicy.
func logNetworkPolicySummary(label string, policy *networkingv1.NetworkPolicy) {
	g.GinkgoWriter.Printf("networkpolicy %s namespace=%s name=%s podSelector=%v policyTypes=%v ingress=%d egress=%d\n",
		label,
		policy.Namespace,
		policy.Name,
		policy.Spec.PodSelector.MatchLabels,
		policy.Spec.PolicyTypes,
		len(policy.Spec.Ingress),
		len(policy.Spec.Egress),
	)
}

// logNetworkPolicyDetails logs detailed information about a NetworkPolicy.
func logNetworkPolicyDetails(label string, policy *networkingv1.NetworkPolicy) {
	g.GinkgoHelper()
	g.GinkgoWriter.Printf("networkpolicy %s details:\n", label)
	g.GinkgoWriter.Printf("  podSelector=%v policyTypes=%v\n", policy.Spec.PodSelector.MatchLabels, policy.Spec.PolicyTypes)
	for i, rule := range policy.Spec.Ingress {
		g.GinkgoWriter.Printf("  ingress[%d]: ports=%s from=%s\n", i, formatPorts(rule.Ports), formatPeers(rule.From))
	}
	for i, rule := range policy.Spec.Egress {
		g.GinkgoWriter.Printf("  egress[%d]: ports=%s to=%s\n", i, formatPorts(rule.Ports), formatPeers(rule.To))
	}
}

// formatPorts formats NetworkPolicy ports for logging.
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
	return fmt.Sprintf("[%s]", strings.Join(out, ", "))
}

// formatPeers formats NetworkPolicy peers for logging.
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
	return fmt.Sprintf("[%s]", strings.Join(out, ", "))
}

// formatSelector formats a label selector for logging.
func formatSelector(sel *metav1.LabelSelector) string {
	if sel == nil {
		return ""
	}
	if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
		return "{}"
	}
	return fmt.Sprintf("labels=%v exprs=%v", sel.MatchLabels, sel.MatchExpressions)
}

// podIPs returns all IP addresses assigned to a pod (dual-stack aware).
func podIPs(pod *corev1.Pod) []string {
	var ips []string
	for _, podIP := range pod.Status.PodIPs {
		if podIP.IP != "" {
			ips = append(ips, podIP.IP)
		}
	}
	if len(ips) == 0 && pod.Status.PodIP != "" {
		ips = append(ips, pod.Status.PodIP)
	}
	return ips
}

// isIPv6 returns true if the given IP string is an IPv6 address.
func isIPv6(ip string) bool {
	return net.ParseIP(ip) != nil && strings.Contains(ip, ":")
}

// formatIPPort formats an IP:port pair, using brackets for IPv6 addresses.
func formatIPPort(ip string, port int32) string {
	if isIPv6(ip) {
		return fmt.Sprintf("[%s]:%d", ip, port)
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

// serviceClusterIPs returns all ClusterIPs for a service (dual-stack aware).
func serviceClusterIPs(svc *corev1.Service) []string {
	if len(svc.Spec.ClusterIPs) > 0 {
		return svc.Spec.ClusterIPs
	}
	if svc.Spec.ClusterIP != "" {
		return []string{svc.Spec.ClusterIP}
	}
	return nil
}

// protocolPtr returns a pointer to a Protocol.
func protocolPtr(protocol corev1.Protocol) *corev1.Protocol {
	return &protocol
}

// boolptr returns a pointer to a bool.
func boolptr(value bool) *bool {
	return &value
}

// int64ptr returns a pointer to an int64.
func int64ptr(value int64) *int64 {
	return &value
}
