package operator

import (
	"context"
	"fmt"
	workloadcontroller "github.com/openshift/library-go/pkg/operator/apiserver/controller/workload"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"strconv"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"k8s.io/client-go/tools/cache"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	operatorv1 "github.com/openshift/api/operator/v1"
	operatorfake "github.com/openshift/client-go/operator/clientset/versioned/fake"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/events"
	operatorv1helpers "github.com/openshift/library-go/pkg/operator/v1helpers"
)

func TestProgressingCondition(t *testing.T) {
	rcReplicas := int32(3)

	happyPathDaemonSet := func() *appsv1.DaemonSet {
		return &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "controller-manager",
				Namespace:  "openshift-controller-manager",
				Generation: 100,
			},
			Status: appsv1.DaemonSetStatus{
				NumberAvailable:        3,
				CurrentNumberScheduled: 3,
				DesiredNumberScheduled: 3,
				UpdatedNumberScheduled: 3,
				ObservedGeneration:     100,
			},
		}
	}

	happyPathDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "route-controller-manager",
				Namespace:  "openshift-route-controller-manager",
				Generation: 100,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &rcReplicas,
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas:  3,
				ReadyReplicas:      3,
				Replicas:           3,
				UpdatedReplicas:    3,
				ObservedGeneration: 100,
			},
		}
	}

	testCases := []struct {
		name                     string
		daemonSet                *appsv1.DaemonSet
		deployment               *appsv1.Deployment
		configGeneration         int64
		configObservedGeneration int64
		expectedStatus           operatorv1.ConditionStatus
		expectedMessage          string
		version                  string
	}{
		{
			name:                     "HappyPath",
			daemonSet:                happyPathDaemonSet(),
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionFalse,
			version:                  "v1",
		},
		{
			name:                     "ControllerManagerDaemonSetMissing",
			daemonSet:                nil,
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "daemonset/controller-manager: number available is 0, desired number available > 1",
			version:                  "v1",
		},

		{
			name:                     "RouteControllerDeploymentMissing",
			daemonSet:                happyPathDaemonSet(),
			deployment:               nil,
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/route-controller-manager: available replicas is 0, desired available replicas > 1\ndeployment/route-controller-manager: updated replicas is 0, desired replicas is 3",
			version:                  "v1",
		},
		{
			name: "ControllerManagerDaemonSetObservedAhead",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        3,
					CurrentNumberScheduled: 3,
					DesiredNumberScheduled: 3,
					UpdatedNumberScheduled: 3,
					ObservedGeneration:     101,
				},
			},
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "daemonset/controller-manager: observed generation is 101, desired generation is 100.",
			version:                  "v1",
		},
		{
			name:      "RouteControllerDeploymentObservedAhead",
			daemonSet: happyPathDaemonSet(),
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 101,
				},
			},
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/route-controller-manager: observed generation is 101, desired generation is 100.",
			version:                  "v1",
		},
		{
			name: "ControllerManagerDaemonSetObservedBehind",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 101,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        3,
					CurrentNumberScheduled: 3,
					DesiredNumberScheduled: 3,
					UpdatedNumberScheduled: 3,
					ObservedGeneration:     100,
				},
			},
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "daemonset/controller-manager: observed generation is 100, desired generation is 101.",
			version:                  "v1",
		},
		{
			name:      "RouteControllerDeploymentObservedBehind",
			daemonSet: happyPathDaemonSet(),
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 101,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 100,
				},
			},
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/route-controller-manager: observed generation is 100, desired generation is 101.",
			version:                  "v1",
		},
		{
			name:                     "ConfigObservedAhead",
			daemonSet:                happyPathDaemonSet(),
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 101,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 101, desired generation is 100.",
			version:                  "v1",
		},
		{
			name:                     "ConfigObservedBehind",
			daemonSet:                happyPathDaemonSet(),
			deployment:               happyPathDeployment(),
			configGeneration:         101,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 100, desired generation is 101.",
			version:                  "v1",
		},
		{
			name: "MultipleObservedAhead",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        3,
					CurrentNumberScheduled: 3,
					DesiredNumberScheduled: 3,
					UpdatedNumberScheduled: 3,
					ObservedGeneration:     101,
				},
			},
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 101,
				},
			},
			configGeneration:         100,
			configObservedGeneration: 101,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "daemonset/controller-manager: observed generation is 101, desired generation is 100.\ndeployment/route-controller-manager: observed generation is 101, desired generation is 100.\nopenshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 101, desired generation is 100.",
			version:                  "v1",
		},
		{
			name:                     "ConfigAndDaemonSetGenerationMismatch",
			daemonSet:                happyPathDaemonSet(),
			deployment:               happyPathDeployment(),
			configGeneration:         101,
			configObservedGeneration: 101,
			expectedStatus:           operatorv1.ConditionFalse,
			version:                  "v1",
		},
		{
			name: "ControllerManagerNoneAvailable",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        0,
					CurrentNumberScheduled: 3,
					DesiredNumberScheduled: 3,
					UpdatedNumberScheduled: 3,
					ObservedGeneration:     100,
				},
			},
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "daemonset/controller-manager: number available is 0, desired number available > 1",
			version:                  "v1",
		},
		{
			name:      "RouteControllerDeploymentNoneAvailable",
			daemonSet: happyPathDaemonSet(),
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  0,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 100,
				},
			},
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/route-controller-manager: available replicas is 0, desired available replicas > 1",
			version:                  "v1",
		},
		{
			name: "UpgradeInProgress",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        3,
					CurrentNumberScheduled: 3,
					DesiredNumberScheduled: 3,
					UpdatedNumberScheduled: 2,
					ObservedGeneration:     100,
				},
			},
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "daemonset/controller-manager: updated number scheduled is 2, desired number scheduled is 3",
			version:                  "v1",
		},
		{
			name:      "RouteControllerDeploymentUpgradeInProgress",
			daemonSet: happyPathDaemonSet(),
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    2,
					ObservedGeneration: 100,
				},
			},
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/route-controller-manager: updated replicas is 2, desired replicas is 3",
			version:                  "v1",
		},
		{
			name: "MultipleUpgradeInProgress",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        3,
					CurrentNumberScheduled: 3,
					DesiredNumberScheduled: 3,
					UpdatedNumberScheduled: 2,
					ObservedGeneration:     100,
				},
			},
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    2,
					ObservedGeneration: 100,
				},
			},
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "daemonset/controller-manager: updated number scheduled is 2, desired number scheduled is 3\ndeployment/route-controller-manager: updated replicas is 2, desired replicas is 3",
			version:                  "v1",
		},
		{
			name:                     "UpgradeInProgressVersionMissing",
			daemonSet:                happyPathDaemonSet(),
			deployment:               happyPathDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          fmt.Sprintf("daemonset/controller-manager: version annotation %s missing.\ndeployment/route-controller-manager: version annotation release.openshift.io/version missing.", util.VersionAnnotation),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if len(tc.version) > 0 {
				os.Setenv("RELEASE_VERSION", tc.version)
			} else {
				os.Unsetenv("RELEASE_VERSION")
			}

			objects := []runtime.Object{
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "serving-cert", Namespace: "openshift-controller-manager"}},
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "etcd-client", Namespace: "kube-system"}},
				&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "client-ca", Namespace: "openshift-kube-apiserver"}},
			}
			if tc.daemonSet != nil {
				objects = append(objects, tc.daemonSet)
			}
			if tc.deployment != nil {
				objects = append(objects, tc.deployment)
			}

			kubeClient := fake.NewSimpleClientset(objects...)

			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			proxyLister := configlistersv1.NewProxyLister(indexer)

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
				configMapsGetter:     kubeClient.CoreV1(),
				proxyLister:          proxyLister,
				recorder:             events.NewInMemoryRecorder(""),
				operatorConfigClient: controllerManagerOperatorClient.OperatorV1(),
			}

			countNodes := func(nodeSelector map[string]string) (*int32, error) {
				result := int32(3)
				return &result, nil
			}

			_, _ = syncOpenShiftControllerManager_v311_00_to_latest(operator, operatorConfig, countNodes, workloadcontroller.EnsureAtMostOnePodPerNode)

			result, err := controllerManagerOperatorClient.OperatorV1().OpenShiftControllerManagers().Get(context.TODO(), "cluster", metav1.GetOptions{})
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

func TestDeploymentWithProxy(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "controller-manager",
				Namespace:  "openshift-controller-manager",
				Generation: 2,
			},
		},
	)
	dsClient := kubeClient.AppsV1()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	proxyConfig := &configv1.Proxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.ProxySpec{
			NoProxy:    "no-proxy",
			HTTPProxy:  "http://my-proxy",
			HTTPSProxy: "https://my-proxy",
		},
		Status: configv1.ProxyStatus{
			NoProxy:    "no-proxy",
			HTTPProxy:  "http://my-proxy",
			HTTPSProxy: "https://my-proxy"},
	}
	indexer.Add(proxyConfig)
	proxyLister := configlistersv1.NewProxyLister(indexer)
	recorder := events.NewInMemoryRecorder("")
	operatorConfig := &operatorv1.OpenShiftControllerManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "cluster",
			Generation: 2,
		},
		Spec: operatorv1.OpenShiftControllerManagerSpec{
			OperatorSpec: operatorv1.OperatorSpec{},
		},
		Status: operatorv1.OpenShiftControllerManagerStatus{
			OperatorStatus: operatorv1.OperatorStatus{
				ObservedGeneration: 2,
			},
		},
	}

	specAnnotations := map[string]string{
		"openshiftcontrollermanagers.operator.openshift.io/cluster": strconv.FormatInt(operatorConfig.ObjectMeta.Generation, 10),
		"configmaps/config":               "54587",
		"configmaps/client-ca":            "12515",
		"configmaps/openshift-service-ca": "45789",
		"configmaps/openshift-global-ca":  "56784",
	}

	ds, rcBool, err := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(dsClient, recorder, operatorConfig, "my.co/repo/img:latest", operatorConfig.Status.Generations, proxyLister, specAnnotations)

	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if !rcBool {
		t.Fatal("apply daemon set does not think a changes was made")
	}

	if ds == nil {
		t.Fatalf("nil daemonset returned")
	}

	foundNoProxy := false
	foundHTTPProxy := false
	foundHTTPSProxy := false
	for _, c := range ds.Spec.Template.Spec.Containers {
		for _, e := range c.Env {
			switch e.Name {
			case "NO_PROXY":
				if e.Value == proxyConfig.Status.NoProxy {
					foundNoProxy = true
				}
			case "HTTP_PROXY":
				if e.Value == proxyConfig.Status.HTTPProxy {
					foundHTTPProxy = true
				}
			case "HTTPS_PROXY":
				if e.Value == proxyConfig.Status.HTTPSProxy {
					foundHTTPSProxy = true
				}
			}
		}
	}

	if !foundNoProxy {
		t.Fatalf("NO_PROXY not found: %#v", ds.Spec.Template.Spec.Containers)
	}
	if !foundHTTPProxy {
		t.Fatalf("HTTP_PROXY not found: %#v", ds.Spec.Template.Spec.Containers)
	}
	if !foundHTTPSProxy {
		t.Fatalf("HTTPS_PROXY not found: %#v", ds.Spec.Template.Spec.Containers)
	}
}
