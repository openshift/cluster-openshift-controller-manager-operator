package e2e

import (
	"context"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

const (
	defaultDenyAllPolicyName         = "default-deny"
	controllerManagerPolicyName      = "allow-controller-manager"
	routeControllerManagerPolicyName = "allow-route-controller-manager"
	operatorPolicyName               = "allow-operator"
)

var _ = g.Describe("[Operator][NetworkPolicy] NetworkPolicy Validation", func() {
	g.It("should have properly configured NetworkPolicies across all namespaces [Skipped:MicroShift]", func(ctx context.Context) {
		testControllerManagerNetworkPolicies(ctx)
	})
})

var _ = g.Describe("[Serial][Operator][NetworkPolicy] NetworkPolicy Reconciliation[Timeout:30m]", func() {
	g.It("should restore NetworkPolicies after deletion or modification [Skipped:MicroShift]", func(ctx context.Context) {
		testControllerManagerNetworkPolicyReconcile(ctx)
	})
})

// testControllerManagerNetworkPolicies validates that NetworkPolicies are properly configured
// across controller manager, route controller manager, and operator namespaces.
func testControllerManagerNetworkPolicies(ctx context.Context) {
	g.GinkgoHelper()
	g.GinkgoWriter.Printf("Creating Kubernetes clients\n")
	kubeConfig, err := getKubeConfig()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get kubeconfig")
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to create kubernetes client")

	g.GinkgoWriter.Printf("Validating NetworkPolicies in openshift-controller-manager\n")
	controllerManagerDefaultDeny := getNetworkPolicy(ctx, kubeClient, util.TargetNamespace, defaultDenyAllPolicyName)
	logNetworkPolicySummary("controller-manager/default-deny-all", controllerManagerDefaultDeny)
	logNetworkPolicyDetails("controller-manager/default-deny-all", controllerManagerDefaultDeny)
	requireDefaultDenyAll(controllerManagerDefaultDeny)

	controllerManagerPolicy := getNetworkPolicy(ctx, kubeClient, util.TargetNamespace, controllerManagerPolicyName)
	logNetworkPolicySummary("controller-manager/allow-controller-manager", controllerManagerPolicy)
	logNetworkPolicyDetails("controller-manager/allow-controller-manager", controllerManagerPolicy)
	requirePodSelectorLabel(controllerManagerPolicy, "controller-manager", "true")
	requireIngressPort(controllerManagerPolicy, corev1.ProtocolTCP, 8443)
	requireEgressAllowAllTCP(controllerManagerPolicy)

	g.GinkgoWriter.Printf("Validating NetworkPolicies in openshift-route-controller-manager\n")
	routeControllerManagerDefaultDeny := getNetworkPolicy(ctx, kubeClient, util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	logNetworkPolicySummary("route-controller-manager/default-deny-all", routeControllerManagerDefaultDeny)
	logNetworkPolicyDetails("route-controller-manager/default-deny-all", routeControllerManagerDefaultDeny)
	requireDefaultDenyAll(routeControllerManagerDefaultDeny)

	routeControllerManagerPolicy := getNetworkPolicy(ctx, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	logNetworkPolicySummary("route-controller-manager/allow-route-controller-manager", routeControllerManagerPolicy)
	logNetworkPolicyDetails("route-controller-manager/allow-route-controller-manager", routeControllerManagerPolicy)
	requirePodSelectorLabel(routeControllerManagerPolicy, "route-controller-manager", "true")
	requireIngressPort(routeControllerManagerPolicy, corev1.ProtocolTCP, 8443)
	requireEgressAllowAllTCP(routeControllerManagerPolicy)

	g.GinkgoWriter.Printf("Validating NetworkPolicies in openshift-controller-manager-operator\n")
	operatorDefaultDeny := getNetworkPolicy(ctx, kubeClient, util.OperatorNamespace, defaultDenyAllPolicyName)
	logNetworkPolicySummary("operator/default-deny-all", operatorDefaultDeny)
	logNetworkPolicyDetails("operator/default-deny-all", operatorDefaultDeny)
	requireDefaultDenyAll(operatorDefaultDeny)

	operatorPolicy := getNetworkPolicy(ctx, kubeClient, util.OperatorNamespace, operatorPolicyName)
	logNetworkPolicySummary("operator/allow-operator", operatorPolicy)
	logNetworkPolicyDetails("operator/allow-operator", operatorPolicy)
	requirePodSelectorLabel(operatorPolicy, "app", "openshift-controller-manager-operator")
	requireIngressPort(operatorPolicy, corev1.ProtocolTCP, 8443)
	requireEgressAllowAllTCP(operatorPolicy)

	g.GinkgoWriter.Printf("Verifying pods are ready in controller manager namespaces\n")
	waitForPodsReadyByLabel(ctx, kubeClient, util.TargetNamespace, "controller-manager=true")
	waitForPodsReadyByLabel(ctx, kubeClient, util.RouteControllerTargetNamespace, "route-controller-manager=true")
	waitForPodsReadyByLabel(ctx, kubeClient, util.OperatorNamespace, "app=openshift-controller-manager-operator")
}

// testControllerManagerNetworkPolicyReconcile validates that NetworkPolicies are automatically
// restored after deletion or modification, ensuring the reconciliation loop is working correctly.
func testControllerManagerNetworkPolicyReconcile(ctx context.Context) {
	g.GinkgoHelper()
	g.GinkgoWriter.Printf("Creating Kubernetes clients\n")
	kubeConfig, err := getKubeConfig()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get kubeconfig")
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to create kubernetes client")

	g.GinkgoWriter.Printf("Capturing expected NetworkPolicy specs\n")
	expectedControllerManagerDefaultDeny := getNetworkPolicy(ctx, kubeClient, util.TargetNamespace, defaultDenyAllPolicyName)
	expectedControllerManagerPolicy := getNetworkPolicy(ctx, kubeClient, util.TargetNamespace, controllerManagerPolicyName)
	expectedRouteControllerManagerDefaultDeny := getNetworkPolicy(ctx, kubeClient, util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	expectedRouteControllerManagerPolicy := getNetworkPolicy(ctx, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	expectedOperatorDefaultDeny := getNetworkPolicy(ctx, kubeClient, util.OperatorNamespace, defaultDenyAllPolicyName)
	expectedOperatorPolicy := getNetworkPolicy(ctx, kubeClient, util.OperatorNamespace, operatorPolicyName)

	g.GinkgoWriter.Printf("Deleting main policies and waiting for restoration\n")
	g.GinkgoWriter.Printf("deleting NetworkPolicy %s/%s\n", util.TargetNamespace, controllerManagerPolicyName)
	restoreNetworkPolicyWithTimeout(ctx, kubeClient, expectedControllerManagerPolicy, reconcileTimeout)
	g.GinkgoWriter.Printf("deleting NetworkPolicy %s/%s\n", util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	restoreNetworkPolicyWithTimeout(ctx, kubeClient, expectedRouteControllerManagerPolicy, reconcileTimeout)
	g.GinkgoWriter.Printf("deleting NetworkPolicy %s/%s (operator namespace may need longer to reconcile)\n", util.OperatorNamespace, operatorPolicyName)
	restoreNetworkPolicyWithTimeout(ctx, kubeClient, expectedOperatorPolicy, operatorReconcileTimeout)

	g.GinkgoWriter.Printf("Deleting default-deny-all policies and waiting for restoration\n")
	g.GinkgoWriter.Printf("deleting NetworkPolicy %s/%s\n", util.TargetNamespace, defaultDenyAllPolicyName)
	restoreNetworkPolicyWithTimeout(ctx, kubeClient, expectedControllerManagerDefaultDeny, reconcileTimeout)
	g.GinkgoWriter.Printf("deleting NetworkPolicy %s/%s\n", util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	restoreNetworkPolicyWithTimeout(ctx, kubeClient, expectedRouteControllerManagerDefaultDeny, reconcileTimeout)
	g.GinkgoWriter.Printf("deleting NetworkPolicy %s/%s (operator namespace may need longer to reconcile)\n", util.OperatorNamespace, defaultDenyAllPolicyName)
	restoreNetworkPolicyWithTimeout(ctx, kubeClient, expectedOperatorDefaultDeny, operatorReconcileTimeout)

	g.GinkgoWriter.Printf("Mutating main policies and waiting for reconciliation\n")
	g.GinkgoWriter.Printf("mutating NetworkPolicy %s/%s\n", util.TargetNamespace, controllerManagerPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(ctx, kubeClient, util.TargetNamespace, controllerManagerPolicyName, reconcileTimeout)
	g.GinkgoWriter.Printf("mutating NetworkPolicy %s/%s\n", util.RouteControllerTargetNamespace, routeControllerManagerPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(ctx, kubeClient, util.RouteControllerTargetNamespace, routeControllerManagerPolicyName, reconcileTimeout)
	g.GinkgoWriter.Printf("mutating NetworkPolicy %s/%s (operator namespace may need longer to reconcile)\n", util.OperatorNamespace, operatorPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(ctx, kubeClient, util.OperatorNamespace, operatorPolicyName, operatorReconcileTimeout)

	g.GinkgoWriter.Printf("Mutating default-deny-all policies and waiting for reconciliation\n")
	g.GinkgoWriter.Printf("mutating NetworkPolicy %s/%s\n", util.TargetNamespace, defaultDenyAllPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(ctx, kubeClient, util.TargetNamespace, defaultDenyAllPolicyName, reconcileTimeout)
	g.GinkgoWriter.Printf("mutating NetworkPolicy %s/%s\n", util.RouteControllerTargetNamespace, defaultDenyAllPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(ctx, kubeClient, util.RouteControllerTargetNamespace, defaultDenyAllPolicyName, reconcileTimeout)
	g.GinkgoWriter.Printf("mutating NetworkPolicy %s/%s (operator namespace may need longer to reconcile)\n", util.OperatorNamespace, defaultDenyAllPolicyName)
	mutateAndRestoreNetworkPolicyWithTimeout(ctx, kubeClient, util.OperatorNamespace, defaultDenyAllPolicyName, operatorReconcileTimeout)

	g.GinkgoWriter.Printf("Checking NetworkPolicy-related events (best-effort)\n")
	logNetworkPolicyEvents(ctx, kubeClient, []string{util.OperatorNamespace, util.TargetNamespace, util.RouteControllerTargetNamespace}, controllerManagerPolicyName)
}
