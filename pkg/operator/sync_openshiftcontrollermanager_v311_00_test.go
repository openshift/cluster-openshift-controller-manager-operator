package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/diff"

	workloadcontroller "github.com/openshift/library-go/pkg/operator/apiserver/controller/workload"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	configv1 "github.com/openshift/api/config/v1"
	v1 "github.com/openshift/api/config/v1"
	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	configlisterv1 "github.com/openshift/client-go/config/listers/config/v1"
	"k8s.io/client-go/tools/cache"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/clock"

	operatorv1 "github.com/openshift/api/operator/v1"
	operatorfake "github.com/openshift/client-go/operator/clientset/versioned/fake"
	"github.com/openshift/cluster-openshift-controller-manager-operator/bindata"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/events"
	operatorv1helpers "github.com/openshift/library-go/pkg/operator/v1helpers"
)

func TestExpectedConfigMap(t *testing.T) {

	objects := []runtime.Object{
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "serving-cert", Namespace: "openshift-controller-manager"}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "etcd-client", Namespace: "kube-system"}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "client-ca", Namespace: "openshift-kube-apiserver"}},
	}
	kubeClient := fake.NewSimpleClientset(objects...)
	cv := &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
		Status: configv1.ClusterVersionStatus{
			Capabilities: configv1.ClusterVersionCapabilitiesStatus{
				EnabledCapabilities: []configv1.ClusterVersionCapability{},
				KnownCapabilities: []configv1.ClusterVersionCapability{
					configv1.ClusterVersionCapabilityBuild,
				},
			},
		},
	}
	expectedConfig := &openshiftcontrolplanev1.OpenShiftControllerManagerConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "openshiftcontrolplane.config.openshift.io/v1",
			Kind:       "OpenShiftControllerManagerConfig",
		},
		LeaderElection: configv1.LeaderElection{
			Name: "openshift-master-controllers",
		},
		Controllers: []string{"*",
			"-openshift.io/build",
			"-openshift.io/build-config-change",
			"-openshift.io/builder-rolebindings",
			"-openshift.io/builder-serviceaccount",
			"-openshift.io/default-rolebindings",
		},
		ServiceServingCert: openshiftcontrolplanev1.ServiceServingCert{},
	}
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	proxyLister := configlistersv1.NewProxyLister(indexer)
	operatorConfig := &operatorv1.OpenShiftControllerManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "cluster",
			Generation: cv.Generation,
		},
		Spec: operatorv1.OpenShiftControllerManagerSpec{
			OperatorSpec: operatorv1.OperatorSpec{},
		},
		Status: operatorv1.OpenShiftControllerManagerStatus{
			OperatorStatus: operatorv1.OperatorStatus{},
		},
	}
	kubeInformers := v1helpers.NewKubeInformersForNamespaces(kubeClient, "", util.TargetNamespace)
	configMapsGetter := v1helpers.CachedConfigMapGetter(kubeClient.CoreV1(), kubeInformers)
	controllerManagerOperatorClient := operatorfake.NewSimpleClientset(operatorConfig)
	indexer.Add(cv)
	clusterVersionLister := configlistersv1.NewClusterVersionLister(indexer)
	operator := OpenShiftControllerManagerOperator{
		kubeClient:           kubeClient,
		configMapsGetter:     kubeClient.CoreV1(),
		proxyLister:          proxyLister,
		recorder:             events.NewInMemoryRecorder("", clock.RealClock{}),
		operatorConfigClient: controllerManagerOperatorClient.OperatorV1(),
		clusterVersionLister: clusterVersionLister,
	}
	resultConfigMap, _, err := manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(clusterVersionLister, kubeClient, configMapsGetter, operator.recorder, operatorConfig)
	scheme := runtime.NewScheme()
	utilruntime.Must(openshiftcontrolplanev1.Install(scheme))
	codecs := serializer.NewCodecFactory(scheme)
	obj, err := runtime.Decode(codecs.UniversalDecoder(openshiftcontrolplanev1.GroupVersion, configv1.GroupVersion), []byte(resultConfigMap.Data["config.yaml"]))
	if err != nil {
		t.Fatalf("Unable to decode OpenShiftControllerManagerConfig: %v", err)
	}
	config := obj.(*openshiftcontrolplanev1.OpenShiftControllerManagerConfig)
	if err != nil {
		t.Fatalf("unable to generate ConfigMap")
	}
	if err == nil && !equality.Semantic.DeepEqual(config, expectedConfig) {
		t.Errorf("Results are not deep equal. mismatch (-want +got):\n%s", diff.ObjectDiff(config, expectedConfig))
	}
}

func TestControllerDisabling(t *testing.T) {

	testCases := []struct {
		name                string
		versionLister       configlisterv1.ClusterVersionLister
		knownCapabilities   []configv1.ClusterVersionCapability
		enabledCapabilities []configv1.ClusterVersionCapability
		result              map[string][]string
	}{
		{
			name: "CapabilitiesEnabled",
			knownCapabilities: []v1.ClusterVersionCapability{
				configv1.ClusterVersionCapabilityBuild,
				configv1.ClusterVersionCapabilityDeploymentConfig,
				configv1.ClusterVersionCapabilityImageRegistry,
			},
			enabledCapabilities: []v1.ClusterVersionCapability{
				configv1.ClusterVersionCapabilityBuild,
				configv1.ClusterVersionCapabilityDeploymentConfig,
				configv1.ClusterVersionCapabilityImageRegistry,
			},
			result: map[string][]string{
				"controllers": {"*",
					"-openshift.io/default-rolebindings",
				}},
		},
		{
			name: "BuildCapDisabled",
			knownCapabilities: []v1.ClusterVersionCapability{
				configv1.ClusterVersionCapabilityBuild,
			},
			enabledCapabilities: []v1.ClusterVersionCapability{},
			result: map[string][]string{
				"controllers": {"*",
					"-openshift.io/build",
					"-openshift.io/build-config-change",
					"-openshift.io/builder-rolebindings",
					"-openshift.io/builder-serviceaccount",
					"-openshift.io/default-rolebindings",
				}},
		},
		{
			name: "DeploymentConfigCapDisabled",
			knownCapabilities: []v1.ClusterVersionCapability{
				configv1.ClusterVersionCapabilityDeploymentConfig,
			},
			enabledCapabilities: []v1.ClusterVersionCapability{},
			result: map[string][]string{
				"controllers": {"*",
					"-openshift.io/default-rolebindings",
					"-openshift.io/deployer",
					"-openshift.io/deployer-rolebindings",
					"-openshift.io/deployer-serviceaccount",
					"-openshift.io/deploymentconfig",
				}},
		},
		{
			name: "ImageRegistryCapDisabled",
			knownCapabilities: []v1.ClusterVersionCapability{
				configv1.ClusterVersionCapabilityImageRegistry,
			},
			enabledCapabilities: []v1.ClusterVersionCapability{},
			result: map[string][]string{
				"controllers": {"*",
					"-openshift.io/default-rolebindings",
					"-openshift.io/image-puller-rolebindings",
					"-openshift.io/serviceaccount-pull-secrets",
				}},
		},
		{
			name: "CapabilitiesDisabled",
			knownCapabilities: []v1.ClusterVersionCapability{
				configv1.ClusterVersionCapabilityBuild,
				configv1.ClusterVersionCapabilityDeploymentConfig,
				configv1.ClusterVersionCapabilityImageRegistry,
			},
			enabledCapabilities: []v1.ClusterVersionCapability{},
			result: map[string][]string{
				"controllers": {"*",
					"-openshift.io/build",
					"-openshift.io/build-config-change",
					"-openshift.io/builder-rolebindings",
					"-openshift.io/builder-serviceaccount",
					"-openshift.io/default-rolebindings",
					"-openshift.io/deployer",
					"-openshift.io/deployer-rolebindings",
					"-openshift.io/deployer-serviceaccount",
					"-openshift.io/deploymentconfig",
					"-openshift.io/image-puller-rolebindings",
					"-openshift.io/serviceaccount-pull-secrets",
				}},
		},
		{
			name:                "CapabilitiesDisabledButUnknown",
			knownCapabilities:   []v1.ClusterVersionCapability{},
			enabledCapabilities: []v1.ClusterVersionCapability{},
			result: map[string][]string{
				"controllers": {"*",
					"-openshift.io/default-rolebindings",
				}},
		},
	}

	for _, tc := range testCases {
		objects := []runtime.Object{
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "serving-cert", Namespace: "openshift-controller-manager"}},
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "etcd-client", Namespace: "kube-system"}},
			&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "client-ca", Namespace: "openshift-kube-apiserver"}},
		}
		kubeClient := fake.NewSimpleClientset(objects...)
		cv := &configv1.ClusterVersion{
			ObjectMeta: metav1.ObjectMeta{Name: "version"},
			Status: configv1.ClusterVersionStatus{
				Capabilities: configv1.ClusterVersionCapabilitiesStatus{
					EnabledCapabilities: tc.enabledCapabilities,
					KnownCapabilities:   tc.knownCapabilities,
				},
			},
		}

		indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
		proxyLister := configlistersv1.NewProxyLister(indexer)

		operatorConfig := &operatorv1.OpenShiftControllerManager{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "cluster",
				Generation: cv.Generation,
			},
			Spec: operatorv1.OpenShiftControllerManagerSpec{
				OperatorSpec: operatorv1.OperatorSpec{},
			},
			Status: operatorv1.OpenShiftControllerManagerStatus{
				OperatorStatus: operatorv1.OperatorStatus{},
			},
		}
		kubeInformers := v1helpers.NewKubeInformersForNamespaces(kubeClient, "", util.TargetNamespace)
		configMapsGetter := v1helpers.CachedConfigMapGetter(kubeClient.CoreV1(), kubeInformers)
		controllerManagerOperatorClient := operatorfake.NewSimpleClientset(operatorConfig)

		indexer.Add(cv)
		clusterVersionLister := configlistersv1.NewClusterVersionLister(indexer)
		operator := OpenShiftControllerManagerOperator{
			kubeClient:           kubeClient,
			configMapsGetter:     kubeClient.CoreV1(),
			proxyLister:          proxyLister,
			recorder:             events.NewInMemoryRecorder("", clock.RealClock{}),
			operatorConfigClient: controllerManagerOperatorClient.OperatorV1(),
			clusterVersionLister: clusterVersionLister,
		}
		result := map[string][]string{}
		resultConfigMap, _, err := manageOpenShiftControllerManagerConfigMap_v311_00_to_latest(clusterVersionLister, kubeClient, configMapsGetter, operator.recorder, operatorConfig)
		if err != nil {
			t.Fatalf("unable to generate ConfigMap")
		} else {
			json.Unmarshal([]byte(resultConfigMap.Data["config.yaml"]), &result)
		}
		sort.Strings(result["controllers"])
		resultControllers := map[string][]string{"controllers": result["controllers"]}
		if err == nil && !reflect.DeepEqual(tc.result, resultControllers) {
			t.Errorf("test '%s' failed. Results are not deep equal. mismatch (-want +got):\n%s\n%v", tc.name, tc.result, resultControllers)
		}
	}
}

func TestProgressingCondition(t *testing.T) {
	ocmReplicas := int32(3)
	rcmReplicas := int32(3)

	happyPathOCMDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "controller-manager",
				Namespace:  "openshift-controller-manager",
				Generation: 100,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &ocmReplicas,
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

	happyPathRCMDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "route-controller-manager",
				Namespace:  "openshift-route-controller-manager",
				Generation: 100,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &rcmReplicas,
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
		ocmDeployment            *appsv1.Deployment
		rcmDeployment            *appsv1.Deployment
		configGeneration         int64
		configObservedGeneration int64
		expectedStatus           operatorv1.ConditionStatus
		expectedMessage          string
		version                  string
	}{
		{
			name:                     "HappyPath",
			ocmDeployment:            happyPathOCMDeployment(),
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionFalse,
			version:                  "v1",
		},
		{
			name:                     "ControllerManagerDeploymentMissing",
			ocmDeployment:            nil,
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/controller-manager: available replicas is 0, desired available replicas > 1\ndeployment/controller-manager: updated replicas is 0, desired replicas is 3",
			version:                  "v1",
		},

		{
			name:                     "RouteControllerDeploymentMissing",
			ocmDeployment:            happyPathOCMDeployment(),
			rcmDeployment:            nil,
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/route-controller-manager: available replicas is 0, desired available replicas > 1\ndeployment/route-controller-manager: updated replicas is 0, desired replicas is 3",
			version:                  "v1",
		},
		{
			name: "ControllerManagerDeploymentObservedAhead",
			ocmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &ocmReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 101,
				},
			},
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/controller-manager: observed generation is 101, desired generation is 100.",
			version:                  "v1",
		},
		{
			name:          "RouteControllerDeploymentObservedAhead",
			ocmDeployment: happyPathOCMDeployment(),
			rcmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcmReplicas,
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
			name: "ControllerManagerDeploymentObservedBehind",
			ocmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 101,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &ocmReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 100,
				},
			},
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/controller-manager: observed generation is 100, desired generation is 101.",
			version:                  "v1",
		},
		{
			name:          "RouteControllerDeploymentObservedBehind",
			ocmDeployment: happyPathOCMDeployment(),
			rcmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 101,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcmReplicas,
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
			ocmDeployment:            happyPathOCMDeployment(),
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 101,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 101, desired generation is 100.",
			version:                  "v1",
		},
		{
			name:                     "ConfigObservedBehind",
			ocmDeployment:            happyPathOCMDeployment(),
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         101,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "openshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 100, desired generation is 101.",
			version:                  "v1",
		},
		{
			name: "MultipleObservedAhead",
			ocmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &ocmReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 101,
				},
			},
			rcmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcmReplicas,
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
			expectedMessage:          "deployment/controller-manager: observed generation is 101, desired generation is 100.\ndeployment/route-controller-manager: observed generation is 101, desired generation is 100.\nopenshiftcontrollermanagers.operator.openshift.io/cluster: observed generation is 101, desired generation is 100.",
			version:                  "v1",
		},
		{
			name:                     "ConfigAndDeploymentGenerationMismatch",
			ocmDeployment:            happyPathOCMDeployment(),
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         101,
			configObservedGeneration: 101,
			expectedStatus:           operatorv1.ConditionFalse,
			version:                  "v1",
		},
		{
			name: "ControllerManagerNoneAvailable",
			ocmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &ocmReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  0,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    3,
					ObservedGeneration: 100,
				},
			},
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/controller-manager: available replicas is 0, desired available replicas > 1",
			version:                  "v1",
		},
		{
			name:          "RouteControllerDeploymentNoneAvailable",
			ocmDeployment: happyPathOCMDeployment(),
			rcmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcmReplicas,
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
			ocmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &ocmReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    2,
					ObservedGeneration: 100,
				},
			},
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          "deployment/controller-manager: updated replicas is 2, desired replicas is 3",
			version:                  "v1",
		},
		{
			name:          "RouteControllerDeploymentUpgradeInProgress",
			ocmDeployment: happyPathOCMDeployment(),
			rcmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcmReplicas,
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
			ocmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "controller-manager",
					Namespace:  "openshift-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &ocmReplicas,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas:  3,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    2,
					ObservedGeneration: 100,
				},
			},
			rcmDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "route-controller-manager",
					Namespace:  "openshift-route-controller-manager",
					Generation: 100,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &rcmReplicas,
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
			expectedMessage:          "deployment/controller-manager: updated replicas is 2, desired replicas is 3\ndeployment/route-controller-manager: updated replicas is 2, desired replicas is 3",
			version:                  "v1",
		},
		{
			name:                     "UpgradeInProgressVersionMissing",
			ocmDeployment:            happyPathOCMDeployment(),
			rcmDeployment:            happyPathRCMDeployment(),
			configGeneration:         100,
			configObservedGeneration: 100,
			expectedStatus:           operatorv1.ConditionTrue,
			expectedMessage:          fmt.Sprintf("deployment/controller-manager: version annotation %s missing.\ndeployment/route-controller-manager: version annotation release.openshift.io/version missing.", util.VersionAnnotation),
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
			if tc.ocmDeployment != nil {
				objects = append(objects, tc.ocmDeployment)
			}
			if tc.rcmDeployment != nil {
				objects = append(objects, tc.rcmDeployment)
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

			cv := &configv1.ClusterVersion{}
			indexer.Add(cv)
			operator := OpenShiftControllerManagerOperator{
				kubeClient:           kubeClient,
				configMapsGetter:     kubeClient.CoreV1(),
				proxyLister:          proxyLister,
				recorder:             events.NewInMemoryRecorder("", clock.RealClock{}),
				operatorConfigClient: controllerManagerOperatorClient.OperatorV1(),
				clusterVersionLister: configlistersv1.NewClusterVersionLister(indexer),
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
	tests := []struct {
		name          string
		mustLoadAsset func(name string) []byte
		proxyConfig   *configv1.Proxy
		expectedEnv   []corev1.EnvVar
	}{
		{
			name:          "default deployment template",
			mustLoadAsset: bindata.MustAsset,
			proxyConfig: &configv1.Proxy{
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
					HTTPSProxy: "https://my-proxy",
				},
			},
			expectedEnv: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
				{
					Name:  "HTTPS_PROXY",
					Value: "https://my-proxy",
				},
				{
					Name:  "HTTP_PROXY",
					Value: "http://my-proxy",
				},
				{
					Name:  "NO_PROXY",
					Value: "no-proxy",
				},
			},
		},
		{
			name: "template proxy variables replaced",
			mustLoadAsset: func(path string) []byte {
				required := resourceread.ReadDeploymentV1OrDie(bindata.MustAsset(path))
				for i := range required.Spec.Template.Spec.Containers {
					required.Spec.Template.Spec.Containers[i].Env = []corev1.EnvVar{
						{
							Name:  "NO_PROXY",
							Value: "1.2.3,4",
						},
						{
							Name:  "HTTP_PROXY",
							Value: "6.7.8.9",
						},
						{
							Name:  "POD_NAME",
							Value: "my-pod",
						},
					}
				}

				scheme := runtime.NewScheme()
				appsv1.AddToScheme(scheme)
				codecs := serializer.NewCodecFactory(scheme)
				return []byte(runtime.EncodeOrDie(codecs.LegacyCodec(appsv1.SchemeGroupVersion), required))
			},
			proxyConfig: &configv1.Proxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.ProxySpec{
					HTTPProxy:  "http://my-proxy",
					HTTPSProxy: "https://my-proxy",
				},
				Status: configv1.ProxyStatus{
					HTTPProxy:  "http://my-proxy",
					HTTPSProxy: "https://my-proxy",
				},
			},
			expectedEnv: []corev1.EnvVar{
				// POD_NAME is kept as it was in the template.
				{
					Name:  "POD_NAME",
					Value: "my-pod",
				},
				// HTTPS_PROXY is added as it isn't in the template, but it's present in the proxy config.
				{
					Name:  "HTTPS_PROXY",
					Value: "https://my-proxy",
				},
				// HTTP_PROXY value is replaced with the proxy config value.
				{
					Name:  "HTTP_PROXY",
					Value: "http://my-proxy",
				},
				// NO_PROXY is removed from the env since it's not present in the proxy config.
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			kubeClient := fake.NewSimpleClientset(
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "controller-manager",
						Namespace:  "openshift-controller-manager",
						Generation: 2,
					},
				},
			)
			deployClient := kubeClient.AppsV1()

			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			indexer.Add(tt.proxyConfig)
			proxyLister := configlistersv1.NewProxyLister(indexer)

			recorder := events.NewInMemoryRecorder("", clock.RealClock{})

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

			countNodes := func(nodeSelector map[string]string) (*int32, error) {
				result := int32(3)
				return &result, nil
			}

			// Sync
			ds, rcBool, err := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(
				tt.mustLoadAsset,
				deployClient,
				countNodes,
				workloadcontroller.EnsureAtMostOnePodPerNode,
				recorder,
				operatorConfig,
				"my.co/repo/img:latest",
				operatorConfig.Status.Generations,
				proxyLister,
				specAnnotations,
			)

			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if !rcBool {
				t.Fatal("apply deployment does not think a changes was made")
			}
			if ds == nil {
				t.Fatalf("nil deployment returned")
			}

			// Check the resulting env
			for _, c := range ds.Spec.Template.Spec.Containers {
				if !cmp.Equal(c.Env, tt.expectedEnv) {
					t.Error("Unexpected container environment definition encountered:\n", cmp.Diff(tt.expectedEnv, c.Env))
				}
			}
		})
	}
}
