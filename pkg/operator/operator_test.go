package operator

import (
	"context"
	"strings"
	"testing"

	clienttesting "k8s.io/client-go/testing"

	operatorapiv1 "github.com/openshift/api/operator/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	operatorclientfakev1 "github.com/openshift/client-go/operator/clientset/versioned/fake"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/workload"
	"github.com/openshift/library-go/pkg/operator/events"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

// assertUpdateActionsCondition run the informed method against each update action found on slice of
// actions, when probe function is successful the loop is interrupted.
func assertUpdateActionsCondition(
	t *testing.T,
	actions []clienttesting.Action,
	probeFn func(c operatorapiv1.OperatorCondition) bool,
) {
	probeSuccessful := false
	for _, action := range actions {
		updateAction, ok := action.(clienttesting.UpdateAction)
		if !ok {
			continue
		}
		cfg, ok := updateAction.GetObject().(*operatorapiv1.OpenShiftControllerManager)
		if !ok {
			continue
		}
		for _, condition := range cfg.Status.Conditions {
			if probeFn(condition) {
				probeSuccessful = true
				break
			}
		}
	}
	if !probeSuccessful {
		t.Fatal("assertion error, not able to satisfy probe function!")
	}
}

// TestOperator_sync runs test-cases after running "sync". The reconciliation must update the status
// conditions accordingly, thus they are inspected to be emptied out or contain "unsupported" string.
func TestOperator_sync(t *testing.T) {
	type testCase struct {
		managementState       operatorv1.ManagementState // cluster config management state
		expectAmountOfActions int                        // expected amount of actions issued
		emptiedConditions     bool                       // check for emptied out conditions
		unsupportedConditions bool                       // check for unsupported conditions
	}

	testCases := []testCase{{
		managementState:       operatorv1.Managed,
		expectAmountOfActions: 4,
		emptiedConditions:     true,
		unsupportedConditions: false,
	}, {
		managementState:       operatorv1.Unmanaged,
		expectAmountOfActions: 7,
		emptiedConditions:     false,
		unsupportedConditions: true,
	}}

	servingCertSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "serving-cert",
			Namespace: "openshift-controller-manager",
		},
	}
	etcdClientSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-client",
			Namespace: "kube-system",
		},
	}
	operatorConfig := &operatorv1.OpenShiftControllerManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "cluster",
			Generation: 1,
		},
		Spec: operatorv1.OpenShiftControllerManagerSpec{
			OperatorSpec: operatorv1.OperatorSpec{},
		},
		Status: operatorv1.OpenShiftControllerManagerStatus{
			OperatorStatus: operatorv1.OperatorStatus{
				ObservedGeneration: 1,
				Conditions: []operatorv1.OperatorCondition{{
					Type:   operatorv1.OperatorStatusTypeAvailable,
					Status: operatorv1.ConditionFalse,
				}},
			},
		},
	}

	kubeClient := fake.NewSimpleClientset(servingCertSecret, etcdClientSecret)
	operatorConfigClientSet := operatorclientfakev1.NewSimpleClientset(operatorConfig)
	operatorConfigInformer := operatorinformers.NewSharedInformerFactory(operatorConfigClientSet, 0)
	operatorConfigLister := operatorConfigInformer.Operator().V1().OpenShiftControllerManagers().Lister()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	operatorConfigInformer.Start(ctx.Done())
	operatorConfigInformer.WaitForCacheSync(ctx.Done())

	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	proxyLister := configlistersv1.NewProxyLister(indexer)
	recorder := events.NewInMemoryRecorder("")

	c := OpenShiftControllerManagerOperator{
		operatorConfigClient: operatorConfigClientSet.OperatorV1(),
		operatorConfigLister: operatorConfigLister,
		kubeClient:           kubeClient,
		workload:             workload.NewWorkload("", proxyLister, kubeClient, recorder),
	}

	for _, tc := range testCases {
		cfg, err := c.getOperatorConfig()
		if err != nil {
			t.Fatalf("Error not expected got: '%v'", err)
		}

		cfg.Spec.ManagementState = tc.managementState
		cfg, err = c.operatorConfigClient.
			OpenShiftControllerManagers().
			Update(context.TODO(), cfg, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("Error not expected got: '%v'", err)
		}

		err = c.sync()
		if err != nil {
			t.Fatalf("Error not expected got: '%v'", err)
		}

		actions := operatorConfigClientSet.Actions()
		t.Logf("Amount of actions: %d", len(actions))

		if tc.expectAmountOfActions != len(actions) {
			t.Fatalf("[%s] expects '%d' actions, but found '%d'",
				tc.managementState, tc.expectAmountOfActions, len(actions))
		}

		if tc.emptiedConditions {
			t.Log("checking for emptied conditions...")
			assertUpdateActionsCondition(t, actions, func(c operatorv1.OperatorCondition) bool {
				return c.Message == "" && c.Reason == ""
			})
		}

		if tc.unsupportedConditions {
			t.Log("checking for unsupported on conditions...")
			assertUpdateActionsCondition(t, actions, func(c operatorv1.OperatorCondition) bool {
				msg := strings.ToLower(c.Message)
				reason := strings.ToLower(c.Reason)
				return strings.Contains(msg, "unsupported") &&
					strings.Contains(reason, "unsupported")
			})
		}
	}
}
