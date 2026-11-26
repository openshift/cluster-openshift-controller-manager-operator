package configobservercontroller

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/clock"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	configfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorfake "github.com/openshift/client-go/operator/clientset/versioned/fake"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

// fakeFeatureGateAccessor implements featuregates.FeatureGateAccess for testing
type fakeFeatureGateAccessor struct {
	featureGate *configv1.FeatureGate
}

func (f *fakeFeatureGateAccessor) CurrentFeatureGates() (featuregates.FeatureGate, error) {
	enabled := []configv1.FeatureGateName{}
	disabled := []configv1.FeatureGateName{}

	// When feature set is TechPreviewNoUpgrade, TLSGroupPreferences is enabled
	if f.featureGate != nil && f.featureGate.Spec.FeatureSet == configv1.TechPreviewNoUpgrade {
		enabled = append(enabled, "TLSGroupPreferences")
	} else {
		disabled = append(disabled, "TLSGroupPreferences")
	}

	return featuregates.NewFeatureGate(enabled, disabled), nil
}

func (f *fakeFeatureGateAccessor) SetChangeHandler(handler featuregates.FeatureGateChangeHandlerFunc) {
}

func (f *fakeFeatureGateAccessor) AreInitialFeatureGatesObserved() bool {
	return true
}

func (f *fakeFeatureGateAccessor) InitialFeatureGatesObserved() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (f *fakeFeatureGateAccessor) Run(ctx context.Context) {
}

// tlsGroupToCurveIDOrDie converts a TLS group to curve ID as float64 or panics
func tlsGroupToCurveIDOrDie(group configv1.TLSGroup) float64 {
	curveID, ok := crypto.TLSGroupToCurveID(group)
	if !ok {
		panic("unrecognized TLS group: " + group)
	}
	return float64(curveID)
}

// groupsToCurvePrefs converts TLS groups to curve preference IDs as float64 (JSON unmarshal format)
func groupsToCurvePrefs(t *testing.T, groups []configv1.TLSGroup) []interface{} {
	result := make([]interface{}, len(groups))
	for i, group := range groups {
		result[i] = tlsGroupToCurveIDOrDie(group)
	}
	return result
}

// setupConfigObserver creates a config observer with the given APIServer TLS profile and existing config
func setupConfigObserver(t *testing.T, featureSet configv1.FeatureSet, apiServerTLSProfile *configv1.TLSSecurityProfile, existingConfig map[string]interface{}) (factory.Controller, v1helpers.OperatorClient, events.InMemoryRecorder) {
	featureGate := &configv1.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: configv1.FeatureGateSpec{
			FeatureGateSelection: configv1.FeatureGateSelection{
				FeatureSet: featureSet,
			},
		},
	}

	apiServer := &configv1.APIServer{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: configv1.APIServerSpec{
			TLSSecurityProfile: apiServerTLSProfile,
		},
	}

	var observedConfigRaw []byte
	if existingConfig != nil {
		var err error
		observedConfigRaw, err = json.Marshal(existingConfig)
		if err != nil {
			t.Fatalf("failed to marshal existing config: %v", err)
		}
	} else {
		observedConfigRaw = []byte(`{}`)
	}

	ocm := &operatorv1.OpenShiftControllerManager{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: operatorv1.OpenShiftControllerManagerSpec{
			OperatorSpec: operatorv1.OperatorSpec{
				ObservedConfig: runtime.RawExtension{Raw: observedConfigRaw},
			},
		},
		Status: operatorv1.OpenShiftControllerManagerStatus{
			OperatorStatus: operatorv1.OperatorStatus{},
		},
	}

	operatorClient := operatorfake.NewSimpleClientset(ocm)
	operatorInformers := operatorinformers.NewSharedInformerFactory(operatorClient, 0)

	configClient := configfake.NewSimpleClientset(featureGate, apiServer)
	configInformers := configinformers.NewSharedInformerFactory(configClient, 0)

	kubeClient := fake.NewSimpleClientset()
	kubeInformers := kubeinformers.NewSharedInformerFactory(kubeClient, 0)

	configInformers.Config().V1().FeatureGates().Informer().GetStore().Add(featureGate)
	configInformers.Config().V1().APIServers().Informer().GetStore().Add(apiServer)
	configInformers.Config().V1().Images().Informer().GetStore().Add(&configv1.Image{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}})
	configInformers.Config().V1().Networks().Informer().GetStore().Add(&configv1.Network{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}})
	configInformers.Config().V1().ClusterVersions().Informer().GetStore().Add(&configv1.ClusterVersion{ObjectMeta: metav1.ObjectMeta{Name: "version"}})
	configInformers.Config().V1().ClusterOperators().Informer().GetStore().Add(&configv1.ClusterOperator{ObjectMeta: metav1.ObjectMeta{Name: "openshift-controller-manager"}})
	operatorInformers.Operator().V1().OpenShiftControllerManagers().Informer().GetStore().Add(ocm)

	fakeOperatorClient := v1helpers.NewFakeOperatorClient(
		&operatorv1.OperatorSpec{
			ObservedConfig: runtime.RawExtension{Raw: observedConfigRaw},
		},
		&operatorv1.OperatorStatus{},
		nil,
	)
	recorder := events.NewInMemoryRecorder("test", clock.RealClock{})

	observer := NewConfigObserver(
		fakeOperatorClient,
		operatorInformers,
		configInformers,
		kubeInformers,
		&fakeFeatureGateAccessor{featureGate: featureGate},
		recorder,
		false, // buildEnabled
	)

	return observer, fakeOperatorClient, recorder
}

func TestNewConfigObserver_TLSGroups(t *testing.T) {
	modernProfile := configv1.TLSProfiles[configv1.TLSProfileModernType]

	tests := []struct {
		name                 string
		featureSet           configv1.FeatureSet
		apiServerTLSProfile  *configv1.TLSSecurityProfile
		expectedCurvePrefs   []interface{}
		expectedMinVersion   string
		expectedCipherSuites []string
	}{
		{
			name:       "Modern profile with feature gate enabled",
			featureSet: configv1.TechPreviewNoUpgrade,
			apiServerTLSProfile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			},
			expectedCurvePrefs:   groupsToCurvePrefs(t, modernProfile.Groups),
			expectedMinVersion:   string(modernProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(modernProfile.Ciphers),
		},
		{
			name:       "Modern profile with feature gate disabled - no curve preferences",
			featureSet: configv1.Default,
			apiServerTLSProfile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			},
			expectedCurvePrefs:   nil, // curvePreferences should not be set when feature gate is disabled
			expectedMinVersion:   string(modernProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(modernProfile.Ciphers),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer, fakeOperatorClient, recorder := setupConfigObserver(t, tt.featureSet, tt.apiServerTLSProfile, nil)

			err := observer.Sync(context.TODO(), factory.NewSyncContext("test", recorder))
			if err != nil {
				t.Fatalf("sync failed: %v", err)
			}

			spec, _, _, err := fakeOperatorClient.GetOperatorState()
			if err != nil {
				t.Fatalf("failed to get operator state: %v", err)
			}

			// The fake client stores observed config in .Object field
			// We need to convert it to map[string]interface{} for our checks
			var observedConfig map[string]interface{}
			if spec.ObservedConfig.Object != nil {
				// Convert the runtime.Object to map by marshaling and unmarshaling
				objBytes, err := json.Marshal(spec.ObservedConfig.Object)
				if err != nil {
					t.Fatalf("failed to marshal ObservedConfig.Object: %v", err)
				}
				if err := json.Unmarshal(objBytes, &observedConfig); err != nil {
					t.Fatalf("failed to unmarshal ObservedConfig.Object: %v", err)
				}
			} else {
				observedConfig = map[string]interface{}{}
			}

			if tt.expectedCurvePrefs != nil {
				curvePrefs, found, err := unstructured.NestedSlice(observedConfig, "servingInfo", "curvePreferences")
				if err != nil {
					t.Fatalf("failed to read curvePreferences: %v", err)
				}
				if !found {
					t.Error("curvePreferences should be set")
				} else if !reflect.DeepEqual(curvePrefs, tt.expectedCurvePrefs) {
					t.Errorf("curve preferences mismatch:\nexpected: %v\nactual:   %v",
						tt.expectedCurvePrefs, curvePrefs)
				}
			} else {
				_, found, _ := unstructured.NestedSlice(observedConfig, "servingInfo", "curvePreferences")
				if found {
					t.Error("curvePreferences should not be set when feature gate is disabled")
				}
			}

			if tt.expectedMinVersion != "" {
				minVersion, found, err := unstructured.NestedString(observedConfig, "servingInfo", "minTLSVersion")
				if err != nil {
					t.Fatalf("failed to read minTLSVersion: %v", err)
				}
				if !found {
					t.Error("minTLSVersion should be set")
				} else if minVersion != tt.expectedMinVersion {
					t.Errorf("expected minTLSVersion=%s, got: %v", tt.expectedMinVersion, minVersion)
				}
			}

			if tt.expectedCipherSuites != nil {
				cipherSuites, found, err := unstructured.NestedStringSlice(observedConfig, "servingInfo", "cipherSuites")
				if err != nil {
					t.Fatalf("failed to read cipherSuites: %v", err)
				}
				if !found {
					t.Error("cipherSuites should be set")
				} else if !reflect.DeepEqual(cipherSuites, tt.expectedCipherSuites) {
					t.Errorf("cipher suites mismatch:\nexpected: %v\nactual:   %v",
						tt.expectedCipherSuites, cipherSuites)
				}
			}
		})
	}
}
