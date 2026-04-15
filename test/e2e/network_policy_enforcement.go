package e2e

import (
	"context"
	"fmt"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

var _ = g.Describe("[Operator][NetworkPolicy] Controller Manager NetworkPolicy Enforcement", func() {
	g.It("should enforce production NetworkPolicy ingress and egress rules [Skipped:MicroShift]", func(ctx context.Context) {
		testControllerManagerNetworkPolicyEnforcement(ctx)
	})
	g.It("should allow metrics port 8443 from multiple namespaces but block unauthorized ports [Skipped:MicroShift]", func(ctx context.Context) {
		testMetricsPortAccessControl(ctx)
	})
})

// testControllerManagerNetworkPolicyEnforcement validates that the production NetworkPolicies
// for controller manager components correctly enforce ingress and egress rules.
func testControllerManagerNetworkPolicyEnforcement(ctx context.Context) {
	g.GinkgoHelper()
	kubeConfig, err := getKubeConfig()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get kubeconfig")
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to create kubernetes client")

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

	g.By("Verifying controller manager NetworkPolicies exist")
	_, err = kubeClient.NetworkingV1().NetworkPolicies(util.TargetNamespace).Get(ctx, controllerManagerPolicyName, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get controller manager NetworkPolicy")
	_, err = kubeClient.NetworkingV1().NetworkPolicies(util.RouteControllerTargetNamespace).Get(ctx, routeControllerManagerPolicyName, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get route controller manager NetworkPolicy")
	_, err = kubeClient.NetworkingV1().NetworkPolicies(util.OperatorNamespace).Get(ctx, operatorPolicyName, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get operator NetworkPolicy")

	g.By("Creating test pods in openshift-controller-manager-operator for allow/deny checks")
	g.GinkgoWriter.Printf("creating operator server pods in %s\n", util.OperatorNamespace)
	allowedServerIPs, cleanupAllowed := createServerPod(ctx, kubeClient, util.OperatorNamespace, fmt.Sprintf("np-operator-allowed-%s", rand.String(5)), operatorLabels, 8443)
	g.DeferCleanup(cleanupAllowed)
	deniedServerIPs, cleanupDenied := createServerPod(ctx, kubeClient, util.OperatorNamespace, fmt.Sprintf("np-operator-denied-%s", rand.String(5)), operatorLabels, 12345)
	g.DeferCleanup(cleanupDenied)

	g.By("Verifying allowed port 8443 ingress to operator")
	expectConnectivity(ctx, kubeClient, util.OperatorNamespace, operatorLabels, allowedServerIPs, 8443, true)

	g.By("Verifying denied port 12345 (not in NetworkPolicy)")
	expectConnectivity(ctx, kubeClient, util.OperatorNamespace, operatorLabels, deniedServerIPs, 12345, false)

	g.By("Verifying operator egress to DNS")
	dnsSvc, err := kubeClient.CoreV1().Services("openshift-dns").Get(ctx, "dns-default", metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get DNS service")
	dnsIPs := serviceClusterIPs(dnsSvc)
	g.GinkgoWriter.Printf("expecting allow from %s to DNS %v:53\n", util.OperatorNamespace, dnsIPs)
	expectConnectivity(ctx, kubeClient, util.OperatorNamespace, operatorLabels, dnsIPs, 53, true)

	g.By("Verifying controller manager pods egress to DNS")
	expectConnectivity(ctx, kubeClient, util.TargetNamespace, controllerManagerLabels, dnsIPs, 53, true)

	g.By("Verifying route controller manager pods egress to DNS")
	expectConnectivity(ctx, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerLabels, dnsIPs, 53, true)
}

// testMetricsPortAccessControl validates that metrics endpoints (port 8443) are accessible
// from various namespaces, while unauthorized ports are blocked by the default-deny policy.
// Uses test pods with labels matching actual operator pods to ensure NetworkPolicy selectors work correctly.
func testMetricsPortAccessControl(ctx context.Context) {
	g.GinkgoHelper()
	kubeConfig, err := getKubeConfig()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get kubeconfig")
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to create kubernetes client")

	g.By("Getting labels from actual operator pods")
	operatorLabels := getActualPodLabels(ctx, kubeClient, util.OperatorNamespace, "app=openshift-controller-manager-operator")
	o.Expect(operatorLabels).NotTo(o.BeEmpty(), "no operator pod labels found")
	g.GinkgoWriter.Printf("using operator labels from actual pods: %v\n", operatorLabels)

	g.By("Getting labels from actual controller-manager pods")
	controllerManagerLabels := getActualPodLabels(ctx, kubeClient, util.TargetNamespace, "controller-manager=true")
	o.Expect(controllerManagerLabels).NotTo(o.BeEmpty(), "no controller-manager pod labels found")
	g.GinkgoWriter.Printf("using controller-manager labels from actual pods: %v\n", controllerManagerLabels)

	g.By("Getting labels from actual route-controller-manager pods")
	routeControllerManagerLabels := getActualPodLabels(ctx, kubeClient, util.RouteControllerTargetNamespace, "route-controller-manager=true")
	o.Expect(routeControllerManagerLabels).NotTo(o.BeEmpty(), "no route-controller-manager pod labels found")
	g.GinkgoWriter.Printf("using route-controller-manager labels from actual pods: %v\n", routeControllerManagerLabels)

	g.By("Creating test server pods with actual pod labels for connectivity checks")
	operatorServerIPs, cleanupOperator := createServerPod(ctx, kubeClient, util.OperatorNamespace, fmt.Sprintf("np-xns-operator-%s", rand.String(5)), operatorLabels, 8443)
	g.DeferCleanup(cleanupOperator)
	controllerServerIPs, cleanupController := createServerPod(ctx, kubeClient, util.TargetNamespace, fmt.Sprintf("np-xns-controller-%s", rand.String(5)), controllerManagerLabels, 8443)
	g.DeferCleanup(cleanupController)
	routeControllerServerIPs, cleanupRouteController := createServerPod(ctx, kubeClient, util.RouteControllerTargetNamespace, fmt.Sprintf("np-xns-route-%s", rand.String(5)), routeControllerManagerLabels, 8443)
	g.DeferCleanup(cleanupRouteController)

	g.By("Verifying cross-namespace access: monitoring -> operator:8443")
	expectConnectivity(ctx, kubeClient, "openshift-monitoring", map[string]string{"app.kubernetes.io/name": "prometheus"}, operatorServerIPs, 8443, true)

	g.By("Verifying cross-namespace access: monitoring -> controller-manager:8443")
	expectConnectivity(ctx, kubeClient, "openshift-monitoring", map[string]string{"app.kubernetes.io/name": "prometheus"}, controllerServerIPs, 8443, true)

	g.By("Verifying cross-namespace access: monitoring -> route-controller-manager:8443")
	expectConnectivity(ctx, kubeClient, "openshift-monitoring", map[string]string{"app.kubernetes.io/name": "prometheus"}, routeControllerServerIPs, 8443, true)

	g.By("Verifying cross-namespace access: default -> operator:8443 (any namespace can access metrics)")
	expectConnectivity(ctx, kubeClient, "default", map[string]string{"test": "client"}, operatorServerIPs, 8443, true)

	g.By("Verifying cross-namespace access: openshift-console -> controller-manager:8443")
	expectConnectivity(ctx, kubeClient, "openshift-console", map[string]string{"app": "console"}, controllerServerIPs, 8443, true)

	g.By("Verifying unauthorized ports are blocked on test operator pods")
	for _, port := range []int32{80, 443, 9090, 9999, 22, 3306, 5432, 6379} {
		g.GinkgoWriter.Printf("expecting deny from openshift-monitoring to operator:%d (unauthorized port)\n", port)
		expectConnectivity(ctx, kubeClient, "openshift-monitoring", map[string]string{"app.kubernetes.io/name": "prometheus"}, operatorServerIPs, port, false)
	}

	g.By("Verifying unauthorized ports are blocked from default namespace")
	for _, port := range []int32{9090, 9999} {
		g.GinkgoWriter.Printf("expecting deny from default to operator:%d (unauthorized port)\n", port)
		expectConnectivity(ctx, kubeClient, "default", map[string]string{"test": "any-pod"}, operatorServerIPs, port, false)
	}

	g.By("Creating server on non-privileged port 8080 and verifying it is blocked")
	unauthorizedServerIPs, cleanupUnauth := createServerPod(ctx, kubeClient, util.OperatorNamespace, fmt.Sprintf("np-port-8080-%s", rand.String(5)), operatorLabels, 8080)
	g.DeferCleanup(cleanupUnauth)
	g.GinkgoWriter.Printf("expecting deny from openshift-monitoring to operator:8080 (unauthorized non-privileged port)\n")
	expectConnectivity(ctx, kubeClient, "openshift-monitoring", map[string]string{"app.kubernetes.io/name": "prometheus"}, unauthorizedServerIPs, 8080, false)
	g.GinkgoWriter.Printf("expecting deny from default to operator:8080 (unauthorized non-privileged port)\n")
	expectConnectivity(ctx, kubeClient, "default", map[string]string{"test": "any-pod"}, unauthorizedServerIPs, 8080, false)
}
