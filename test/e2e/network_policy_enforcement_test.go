package e2e

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

const (
	agnhostImage = "registry.k8s.io/e2e-test-images/agnhost:2.45"
)

// Import constants from network_policy_test.go - these are defined there:
// - defaultDenyAllPolicyName
// - controllerManagerPolicyName
// - routeControllerManagerPolicyName
// - operatorPolicyName

func TestGenericNetworkPolicyEnforcement(t *testing.T) {
	kubeConfig, err := getKubeConfig()
	if err != nil {
		t.Fatalf("failed to get kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Fatalf("failed to create kubernetes client: %v", err)
	}

	t.Log("Creating a temporary namespace for policy enforcement checks")
	nsName := "np-enforcement-" + rand.String(5)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
	_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create test namespace: %v", err)
	}
	defer func() {
		t.Logf("deleting test namespace %s", nsName)
		_ = kubeClient.CoreV1().Namespaces().Delete(context.TODO(), nsName, metav1.DeleteOptions{})
	}()

	serverName := "np-server"
	clientLabels := map[string]string{"app": "np-client"}
	serverLabels := map[string]string{"app": "np-server"}

	t.Logf("creating netexec server pod %s/%s", nsName, serverName)
	serverPod := netexecPod(serverName, nsName, serverLabels, 8080)
	_, err = kubeClient.CoreV1().Pods(nsName).Create(context.TODO(), serverPod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create server pod: %v", err)
	}
	if err := waitForPodReadyT(t, kubeClient, nsName, serverName); err != nil {
		t.Fatalf("server pod not ready: %v", err)
	}

	server, err := kubeClient.CoreV1().Pods(nsName).Get(context.TODO(), serverName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get server pod: %v", err)
	}
	if len(server.Status.PodIPs) == 0 {
		t.Fatalf("server pod has no IPs")
	}
	serverIPs := podIPs(server)
	t.Logf("server pod %s/%s ips=%v", nsName, serverName, serverIPs)

	t.Log("Verifying allow-all when no policies select the pod")
	expectConnectivity(t, kubeClient, nsName, clientLabels, serverIPs, 8080, true)

	t.Log("Applying default deny and verifying traffic is blocked")
	t.Logf("creating default-deny policy in %s", nsName)
	_, err = kubeClient.NetworkingV1().NetworkPolicies(nsName).Create(context.TODO(), defaultDenyPolicy("default-deny", nsName), metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create default-deny policy: %v", err)
	}

	t.Log("Adding ingress allow only and verifying traffic is still blocked")
	t.Logf("creating allow-ingress policy in %s", nsName)
	_, err = kubeClient.NetworkingV1().NetworkPolicies(nsName).Create(context.TODO(), allowIngressPolicy("allow-ingress", nsName, serverLabels, clientLabels, 8080), metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create allow-ingress policy: %v", err)
	}
	expectConnectivity(t, kubeClient, nsName, clientLabels, serverIPs, 8080, false)

	t.Log("Adding egress allow and verifying traffic is permitted")
	t.Logf("creating allow-egress policy in %s", nsName)
	_, err = kubeClient.NetworkingV1().NetworkPolicies(nsName).Create(context.TODO(), allowEgressPolicy("allow-egress", nsName, clientLabels, serverLabels, 8080), metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create allow-egress policy: %v", err)
	}
	expectConnectivity(t, kubeClient, nsName, clientLabels, serverIPs, 8080, true)
}

func TestControllerManagerNetworkPolicyEnforcement(t *testing.T) {
	kubeConfig, err := getKubeConfig()
	if err != nil {
		t.Fatalf("failed to get kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Fatalf("failed to create kubernetes client: %v", err)
	}

	// Labels must match the NetworkPolicy pod selectors for egress to work
	controllerManagerLabels := map[string]string{
		"app":                "openshift-controller-manager-a",
		"controller-manager": "true",
	}
	routeControllerManagerLabels := map[string]string{
		"app":                      "route-controller-manager",
		"route-controller-manager": "true",
	}
	operatorLabels := map[string]string{"app": "openshift-controller-manager-operator"}

	t.Log("Verifying controller manager NetworkPolicies exist")
	_, err = kubeClient.NetworkingV1().NetworkPolicies(util.TargetNamespace).Get(context.TODO(), controllerManagerPolicyName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get controller manager NetworkPolicy: %v", err)
	}
	_, err = kubeClient.NetworkingV1().NetworkPolicies(util.RouteControllerTargetNamespace).Get(context.TODO(), routeControllerManagerPolicyName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get route controller manager NetworkPolicy: %v", err)
	}
	_, err = kubeClient.NetworkingV1().NetworkPolicies(util.OperatorNamespace).Get(context.TODO(), operatorPolicyName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get operator NetworkPolicy: %v", err)
	}

	t.Log("Creating test pods in openshift-controller-manager-operator for allow/deny checks")
	t.Logf("creating operator server pods in %s", util.OperatorNamespace)
	allowedServerIPs, cleanupAllowed := createServerPodT(t, kubeClient, util.OperatorNamespace, "np-operator-allowed", operatorLabels, 8443)
	defer cleanupAllowed()
	deniedServerIPs, cleanupDenied := createServerPodT(t, kubeClient, util.OperatorNamespace, "np-operator-denied", operatorLabels, 12345)
	defer cleanupDenied()

	t.Log("Verifying allowed port 8443 ingress to operator")
	expectConnectivity(t, kubeClient, util.OperatorNamespace, operatorLabels, allowedServerIPs, 8443, true)

	t.Log("Verifying denied port 12345 (not in NetworkPolicy)")
	expectConnectivity(t, kubeClient, util.OperatorNamespace, operatorLabels, deniedServerIPs, 12345, false)

	t.Log("Verifying denied ports even from same namespace")
	for _, port := range []int32{12346, 12347, 12348, 12349} {
		ips, cleanup := createServerPodT(t, kubeClient, util.OperatorNamespace, fmt.Sprintf("np-operator-denied-%d", port), operatorLabels, port)
		defer cleanup()
		expectConnectivity(t, kubeClient, util.OperatorNamespace, operatorLabels, ips, port, false)
	}

	t.Log("Verifying operator egress to DNS")
	dnsSvc, err := kubeClient.CoreV1().Services("openshift-dns").Get(context.TODO(), "dns-default", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get DNS service: %v", err)
	}
	dnsIPs := serviceClusterIPs(dnsSvc)
	t.Logf("expecting allow from %s to DNS %v:53", util.OperatorNamespace, dnsIPs)
	expectConnectivity(t, kubeClient, util.OperatorNamespace, operatorLabels, dnsIPs, 53, true)

	t.Log("Verifying controller manager pods egress to DNS")
	expectConnectivity(t, kubeClient, util.TargetNamespace, controllerManagerLabels, dnsIPs, 53, true)

	t.Log("Verifying route controller manager pods egress to DNS")
	expectConnectivity(t, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerLabels, dnsIPs, 53, true)
}

func netexecPod(name, namespace string, labels map[string]string, port int32) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot:   boolptr(true),
				RunAsUser:      int64ptr(1001),
				SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
			},
			Containers: []corev1.Container{
				{
					Name:  "netexec",
					Image: agnhostImage,
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

func createServerPodT(t *testing.T, kubeClient kubernetes.Interface, namespace, name string, labels map[string]string, port int32) ([]string, func()) {
	t.Helper()

	t.Logf("creating server pod %s/%s port=%d labels=%v", namespace, name, port, labels)
	pod := netexecPod(name, namespace, labels, port)
	_, err := kubeClient.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create server pod: %v", err)
	}
	if err := waitForPodReadyT(t, kubeClient, namespace, name); err != nil {
		t.Fatalf("server pod not ready: %v", err)
	}

	created, err := kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get created server pod: %v", err)
	}
	if len(created.Status.PodIPs) == 0 {
		t.Fatalf("server pod has no IPs")
	}

	ips := podIPs(created)
	t.Logf("server pod %s/%s ips=%v", namespace, name, ips)

	return ips, func() {
		t.Logf("deleting server pod %s/%s", namespace, name)
		_ = kubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}
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

func defaultDenyPolicy(name, namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
		},
	}
}

func allowIngressPolicy(name, namespace string, podLabels, fromLabels map[string]string, port int32) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: podLabels},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{PodSelector: &metav1.LabelSelector{MatchLabels: fromLabels}},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: &intstr.IntOrString{Type: intstr.Int, IntVal: port}, Protocol: protocolPtr(corev1.ProtocolTCP)},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

func allowEgressPolicy(name, namespace string, podLabels, toLabels map[string]string, port int32) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: podLabels},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					To: []networkingv1.NetworkPolicyPeer{
						{PodSelector: &metav1.LabelSelector{MatchLabels: toLabels}},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: &intstr.IntOrString{Type: intstr.Int, IntVal: port}, Protocol: protocolPtr(corev1.ProtocolTCP)},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

// expectConnectivityForIP checks connectivity to a single IP address.
func expectConnectivityForIP(t *testing.T, kubeClient kubernetes.Interface, namespace string, clientLabels map[string]string, serverIP string, port int32, shouldSucceed bool) {
	t.Helper()

	err := wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		succeeded, err := runConnectivityCheck(t, kubeClient, namespace, clientLabels, serverIP, port)
		if err != nil {
			return false, err
		}
		return succeeded == shouldSucceed, nil
	})
	if err != nil {
		t.Fatalf("connectivity check failed for %s/%s expected=%t: %v", namespace, formatIPPort(serverIP, port), shouldSucceed, err)
	}
	t.Logf("connectivity %s/%s expected=%t", namespace, formatIPPort(serverIP, port), shouldSucceed)
}

// expectConnectivity checks connectivity to all provided IPs (dual-stack aware).
func expectConnectivity(t *testing.T, kubeClient kubernetes.Interface, namespace string, clientLabels map[string]string, serverIPs []string, port int32, shouldSucceed bool) {
	t.Helper()

	for _, ip := range serverIPs {
		family := "IPv4"
		if isIPv6(ip) {
			family = "IPv6"
		}
		t.Logf("checking %s connectivity %s -> %s expected=%t", family, namespace, formatIPPort(ip, port), shouldSucceed)
		expectConnectivityForIP(t, kubeClient, namespace, clientLabels, ip, port, shouldSucceed)
	}
}

func runConnectivityCheck(t *testing.T, kubeClient kubernetes.Interface, namespace string, labels map[string]string, serverIP string, port int32) (bool, error) {
	t.Helper()

	name := fmt.Sprintf("np-client-%s", rand.String(5))
	t.Logf("creating client pod %s/%s to connect %s:%d", namespace, name, serverIP, port)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
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
					Name:  "connect",
					Image: agnhostImage,
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: boolptr(false),
						Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
						RunAsNonRoot:             boolptr(true),
						RunAsUser:                int64ptr(1001),
					},
					Command: []string{"/agnhost"},
					Args: []string{
						"connect",
						"--protocol=tcp",
						"--timeout=5s",
						formatIPPort(serverIP, port),
					},
				},
			},
		},
	}

	_, err := kubeClient.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	defer func() {
		_ = kubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}()

	if err := waitForPodCompletion(kubeClient, namespace, name); err != nil {
		return false, err
	}
	completed, err := kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if len(completed.Status.ContainerStatuses) == 0 {
		return false, fmt.Errorf("no container status recorded for pod %s", name)
	}
	state := completed.Status.ContainerStatuses[0].State
	if state.Terminated == nil {
		return false, fmt.Errorf("pod %s completed without a terminated container state: phase=%s reason=%s", name, completed.Status.Phase, completed.Status.Reason)
	}
	exitCode := state.Terminated.ExitCode
	t.Logf("client pod %s/%s exitCode=%d", namespace, name, exitCode)
	return exitCode == 0, nil
}

func waitForPodReadyT(t *testing.T, kubeClient kubernetes.Interface, namespace, name string) error {
	return wait.PollImmediate(2*time.Second, 2*time.Minute, func() (bool, error) {
		pod, err := kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

func waitForPodCompletion(kubeClient kubernetes.Interface, namespace, name string) error {
	return wait.PollImmediate(2*time.Second, 2*time.Minute, func() (bool, error) {
		pod, err := kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed, nil
	})
}

func protocolPtr(protocol corev1.Protocol) *corev1.Protocol {
	return &protocol
}

func boolptr(value bool) *bool {
	return &value
}

func int64ptr(value int64) *int64 {
	return &value
}
