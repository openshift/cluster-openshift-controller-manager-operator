package operator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	operatorv1 "github.com/openshift/api/operator/v1"
	operatorfake "github.com/openshift/client-go/operator/clientset/versioned/fake"
	"github.com/openshift/library-go/pkg/operator/events"
	operatorv1helpers "github.com/openshift/library-go/pkg/operator/v1helpers"
)

func TestProgressingCondition(t *testing.T) {

	testCases := []struct {
		name                        string
		daemonSetGeneration         int64
		daemonSetObservedGeneration int64
		daemonSetNumAvailable       int32
		daemonSetNumDesired         int32
		daemonSetNumUpdated         int32
		configGeneration            int64
		configObservedGeneration    int64
		expectedStatus              operatorv1.ConditionStatus
		expectedMessage             string
	}{
		{
			name:                        "HappyPath",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 100,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            100,
			configObservedGeneration:    100,
			expectedStatus:              operatorv1.ConditionFalse,
		},
		{
			name:                        "DaemonSetObservedAhead",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 101,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            100,
			configObservedGeneration:    100,
			expectedStatus:              operatorv1.ConditionTrue,
			expectedMessage:             "daemonset/controller-manager: observed generation is 101, desired generation is 100.",
		},
		{
			name:                        "DaemonSetObservedBehind",
			daemonSetGeneration:         101,
			daemonSetObservedGeneration: 100,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            100,
			configObservedGeneration:    100,
			expectedStatus:              operatorv1.ConditionTrue,
			expectedMessage:             "daemonset/controller-manager: observed generation is 100, desired generation is 101.",
		},
		{
			name:                        "ConfigObservedAhead",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 100,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            100,
			configObservedGeneration:    101,
			expectedStatus:              operatorv1.ConditionTrue,
			expectedMessage:             "openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 101, desired generation is 100.",
		},
		{
			name:                        "ConfigObservedBehind",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 100,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            101,
			configObservedGeneration:    100,
			expectedStatus:              operatorv1.ConditionTrue,
			expectedMessage:             "openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 100, desired generation is 101.",
		},
		{
			name:                        "MultipleObservedAhead",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 101,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            100,
			configObservedGeneration:    101,
			expectedStatus:              operatorv1.ConditionTrue,
			expectedMessage:             "daemonset/controller-manager: observed generation is 101, desired generation is 100.\nopenshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 101, desired generation is 100.",
		},
		{
			name:                        "ConfigAndDaemonSetGenerationMismatch",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 100,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            101,
			configObservedGeneration:    101,
			expectedStatus:              operatorv1.ConditionFalse,
		},
		{
			name:                        "NoneAvailable",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 100,
			daemonSetNumAvailable:       0,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         3,
			configGeneration:            100,
			configObservedGeneration:    100,
			expectedStatus:              operatorv1.ConditionTrue,
			expectedMessage:             "daemonset/controller-manager: number available is 0, desired number available > 1",
		},
		{
			name:                        "UpgradeInProgress",
			daemonSetGeneration:         100,
			daemonSetObservedGeneration: 100,
			daemonSetNumAvailable:       3,
			daemonSetNumDesired:         3,
			daemonSetNumUpdated:         2,
			configGeneration:            100,
			configObservedGeneration:    100,
			expectedStatus:              operatorv1.ConditionTrue,
			expectedMessage:             "daemonset/controller-manager: updated number scheduled is 2, desired number scheduled is 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			kubeClient := fake.NewSimpleClientset(
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "serving-cert", Namespace: "openshift-controller-manager"}},
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "etcd-client", Namespace: "kube-system"}},
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "controller-manager",
						Namespace:  "openshift-controller-manager",
						Generation: tc.daemonSetGeneration,
					},
					Status: appsv1.DaemonSetStatus{
						NumberAvailable:        tc.daemonSetNumAvailable,
						CurrentNumberScheduled: tc.daemonSetNumDesired,
						DesiredNumberScheduled: tc.daemonSetNumDesired,
						UpdatedNumberScheduled: tc.daemonSetNumUpdated,
						ObservedGeneration:     tc.daemonSetObservedGeneration,
					},
				})

			operatorConfig := &operatorv1.OpenShiftControllerManager{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "cluster",
					Generation: tc.configGeneration,
				},
				Spec: operatorv1.OpenShiftControllerManagerSpec{
					OperatorSpec: operatorv1.OperatorSpec{},
				},
				Status: operatorv1.OpenShiftControllerManagerStatus{
					OperatorStatus: operatorv1.OperatorStatus{
						ObservedGeneration: tc.configObservedGeneration,
					},
				},
			}
			controllerManagerOperatorClient := operatorfake.NewSimpleClientset(operatorConfig)

			operator := OpenShiftControllerManagerOperator{
				kubeClient:           kubeClient,
				recorder:             events.NewInMemoryRecorder(""),
				operatorConfigClient: controllerManagerOperatorClient.OperatorV1(),
			}

			_, _ = syncOpenShiftControllerManager_v311_00_to_latest(operator, operatorConfig)

			result, err := controllerManagerOperatorClient.OperatorV1().OpenShiftControllerManagers().Get("cluster", metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			condition := operatorv1helpers.FindOperatorCondition(result.Status.Conditions, operatorv1.OperatorStatusTypeProgressing)
			if condition == nil {
				t.Fatalf("No %v condition found.", operatorv1.OperatorStatusTypeProgressing)
			}
			if condition.Status != tc.expectedStatus {
				t.Errorf("expected status == %v, actual status == %v", tc.expectedStatus, condition.Status)
			}
			if condition.Message != tc.expectedMessage {
				t.Errorf("expected message:\n%v\nactual message:\n%v", tc.expectedMessage, condition.Message)
			}

		})
	}

}
