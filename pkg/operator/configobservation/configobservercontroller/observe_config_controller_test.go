package configobservercontroller

import (
	"encoding/json"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/clock"

	configv1 "github.com/openshift/api/config/v1"
	configfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

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

// makeTLSProfile creates a TLS security profile for testing
func makeTLSProfile(profileType configv1.TLSProfileType, minTLSVersion configv1.TLSProtocolVersion, ciphers []string, groups []configv1.TLSGroup) *configv1.TLSSecurityProfile {
	profile := &configv1.TLSSecurityProfile{
		Type: profileType,
	}
	if profileType == configv1.TLSProfileCustomType {
		profile.Custom = &configv1.CustomTLSProfile{
			TLSProfileSpec: configv1.TLSProfileSpec{
				MinTLSVersion: minTLSVersion,
				Ciphers:       ciphers,
				Groups:        groups,
			},
		}
	}
	return profile
}

// setupTLSGroupsTest creates listers and recorder for testing ObserveTLSSecurityProfileWithGroupsConversion
func setupTLSGroupsTest(t *testing.T, apiServerTLSProfile *configv1.TLSSecurityProfile) (configobservation.Listers, events.InMemoryRecorder) {
	apiServer := &configv1.APIServer{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: configv1.APIServerSpec{
			TLSSecurityProfile: apiServerTLSProfile,
		},
	}

	configClient := configfake.NewSimpleClientset(apiServer)
	configInformers := configinformers.NewSharedInformerFactory(configClient, 0)

	configInformers.Config().V1().APIServers().Informer().GetStore().Add(apiServer)

	listers := configobservation.Listers{
		APIServerLister_: configInformers.Config().V1().APIServers().Lister(),
	}
	recorder := events.NewInMemoryRecorder("test", clock.RealClock{})

	return listers, recorder
}

func TestNewConfigObserver_TLSGroups(t *testing.T) {
	defaultProfile := configv1.TLSProfiles[crypto.DefaultTLSProfileType]
	oldProfile := configv1.TLSProfiles[configv1.TLSProfileOldType]
	modernProfile := configv1.TLSProfiles[configv1.TLSProfileModernType]

	tests := []struct {
		name                          string
		apiServerTLSProfile           *configv1.TLSSecurityProfile
		existingConfig                map[string]interface{}
		expectedCurvePrefs            []interface{}
		expectedMinVersion            string
		expectedCipherSuites          []string
		expectError                   bool
		expectedErrorStrings          []string
		mockObserveTLSSecurityProfile observeTLSSecurityProfileWithGroupPathsFunc
	}{
		{
			name:                 "missing TLS config defaults to default profile",
			apiServerTLSProfile:  nil,
			expectedCurvePrefs:   groupsToCurvePrefs(t, defaultProfile.Groups),
			expectedMinVersion:   string(defaultProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(defaultProfile.Ciphers),
		},
		{
			name:                 "Old profile",
			apiServerTLSProfile:  makeTLSProfile(configv1.TLSProfileOldType, "", nil, nil),
			expectedCurvePrefs:   groupsToCurvePrefs(t, oldProfile.Groups),
			expectedMinVersion:   string(oldProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(oldProfile.Ciphers),
		},
		{
			name:                 "Modern profile",
			apiServerTLSProfile:  makeTLSProfile(configv1.TLSProfileModernType, "", nil, nil),
			expectedCurvePrefs:   groupsToCurvePrefs(t, modernProfile.Groups),
			expectedMinVersion:   string(modernProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(modernProfile.Ciphers),
		},
		{
			name: "Custom profile with specific groups",
			apiServerTLSProfile: makeTLSProfile(
				configv1.TLSProfileCustomType,
				configv1.VersionTLS12,
				[]string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				[]configv1.TLSGroup{configv1.TLSGroupX25519, configv1.TLSGroupSecP256r1},
			),
			expectedCurvePrefs: groupsToCurvePrefs(t, []configv1.TLSGroup{
				configv1.TLSGroupX25519,
				configv1.TLSGroupSecP256r1,
			}),
			expectedMinVersion: string(configv1.VersionTLS12),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites([]string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
			}),
		},
		{
			name:                "Unknown curve ID in existing curvePreferences before reconciling - silently dropped",
			apiServerTLSProfile: makeTLSProfile(configv1.TLSProfileModernType, "", nil, nil),
			existingConfig: map[string]interface{}{
				"servingInfo": map[string]interface{}{
					"curvePreferences": []interface{}{
						tlsGroupToCurveIDOrDie(configv1.TLSGroupX25519),
						float64(9999),
						tlsGroupToCurveIDOrDie(configv1.TLSGroupSecP256r1),
					},
				},
			},
			expectedCurvePrefs:   groupsToCurvePrefs(t, modernProfile.Groups),
			expectedMinVersion:   string(modernProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(modernProfile.Ciphers),
		},
		{
			name:                "Corrupted minTLSVersion in existing config - all TLS settings still injected",
			apiServerTLSProfile: makeTLSProfile(configv1.TLSProfileModernType, "", nil, nil),
			existingConfig: map[string]interface{}{
				"servingInfo": map[string]interface{}{
					"minTLSVersion": []string{"not-a-string"},
				},
			},
			expectedCurvePrefs:   groupsToCurvePrefs(t, modernProfile.Groups),
			expectedMinVersion:   string(modernProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(modernProfile.Ciphers),
			expectError:          true,
			expectedErrorStrings: []string{
				"failed to retrieve spec.servingInfo.minTLSVersion: .servingInfo.minTLSVersion accessor error: [not-a-string] is of the type []interface {}, expected string",
			},
		},
		{
			name:                "Corrupted cipherSuites in existing config - all TLS settings still injected",
			apiServerTLSProfile: makeTLSProfile(configv1.TLSProfileModernType, "", nil, nil),
			existingConfig: map[string]interface{}{
				"servingInfo": map[string]interface{}{
					"cipherSuites": "not-a-slice",
				},
			},
			expectedCurvePrefs:   groupsToCurvePrefs(t, modernProfile.Groups),
			expectedMinVersion:   string(modernProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(modernProfile.Ciphers),
			expectError:          true,
			expectedErrorStrings: []string{
				"failed to retrieve spec.servingInfo.cipherSuites: .servingInfo.cipherSuites accessor error: not-a-slice is of the type string, expected []interface{}",
			},
		},
		{
			name:                "Corrupted curvePreferences in existing config - all TLS settings still injected",
			apiServerTLSProfile: makeTLSProfile(configv1.TLSProfileModernType, "", nil, nil),
			existingConfig: map[string]interface{}{
				"servingInfo": map[string]interface{}{
					"curvePreferences": "not-a-slice",
				},
			},
			expectedCurvePrefs:   groupsToCurvePrefs(t, modernProfile.Groups),
			expectedMinVersion:   string(modernProfile.MinTLSVersion),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites(modernProfile.Ciphers),
			expectError:          true,
			expectedErrorStrings: []string{
				"failed to read curve preferences from [servingInfo curvePreferences]: .servingInfo.curvePreferences accessor error: not-a-slice is of the type string, expected []interface{}",
				".servingInfo.curvePreferences accessor error: not-a-slice is of the type string, expected []interface{}",
			},
		},
		{
			name: "Unrecognized TLS groups - recognized groups converted, error for unrecognized",
			apiServerTLSProfile: makeTLSProfile(
				configv1.TLSProfileCustomType,
				configv1.VersionTLS12,
				[]string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				[]configv1.TLSGroup{configv1.TLSGroupX25519, "UnknownGroup", configv1.TLSGroupSecP256r1},
			),
			expectedCurvePrefs: groupsToCurvePrefs(t, []configv1.TLSGroup{
				configv1.TLSGroupX25519,
				configv1.TLSGroupSecP256r1,
			}),
			expectedMinVersion: string(configv1.VersionTLS12),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites([]string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
			}),
			expectError: true,
			expectedErrorStrings: []string{
				"unrecognized TLS groups: [UnknownGroup]",
			},
		},
		{
			name: "Unrecognized TLS groups with corrupted curvePreferences - all errors reported, all TLS settings injected",
			apiServerTLSProfile: makeTLSProfile(
				configv1.TLSProfileCustomType,
				configv1.VersionTLS12,
				[]string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				[]configv1.TLSGroup{configv1.TLSGroupX25519, "UnknownGroup", configv1.TLSGroupSecP256r1},
			),
			existingConfig: map[string]interface{}{
				"servingInfo": map[string]interface{}{
					"curvePreferences": "not-a-slice",
				},
			},
			expectedCurvePrefs: groupsToCurvePrefs(t, []configv1.TLSGroup{
				configv1.TLSGroupX25519,
				configv1.TLSGroupSecP256r1,
			}),
			expectedMinVersion: string(configv1.VersionTLS12),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites([]string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
			}),
			expectError: true,
			expectedErrorStrings: []string{
				"failed to read curve preferences from [servingInfo curvePreferences]: .servingInfo.curvePreferences accessor error: not-a-slice is of the type string, expected []interface{}",
				".servingInfo.curvePreferences accessor error: not-a-slice is of the type string, expected []interface{}",
				"unrecognized TLS groups: [UnknownGroup]",
			},
		},
		{
			name: "All TLS groups unrecognized - curvePreferences set to empty, minTLSVersion and cipherSuites set",
			apiServerTLSProfile: makeTLSProfile(
				configv1.TLSProfileCustomType,
				configv1.VersionTLS12,
				[]string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				[]configv1.TLSGroup{"UnknownGroup1", "UnknownGroup2", "UnknownGroup3"},
			),
			expectedCurvePrefs: []interface{}{},
			expectedMinVersion: string(configv1.VersionTLS12),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites([]string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
			}),
			expectError: true,
			expectedErrorStrings: []string{
				"unrecognized TLS groups: [UnknownGroup1 UnknownGroup2 UnknownGroup3]",
			},
		},
		{
			name: "Custom profile with zero groups but existing curvePreferences - curvePreferences set to empty",
			apiServerTLSProfile: makeTLSProfile(
				configv1.TLSProfileCustomType,
				configv1.VersionTLS12,
				[]string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				[]configv1.TLSGroup{},
			),
			existingConfig: map[string]interface{}{
				"servingInfo": map[string]interface{}{
					"curvePreferences": []interface{}{
						tlsGroupToCurveIDOrDie(configv1.TLSGroupX25519),
						tlsGroupToCurveIDOrDie(configv1.TLSGroupSecP256r1),
					},
				},
			},
			expectedCurvePrefs: []interface{}{},
			expectedMinVersion: string(configv1.VersionTLS12),
			expectedCipherSuites: crypto.OpenSSLToIANACipherSuites([]string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
			}),
		},
		{
			name: "FIPS filtering - non-compliant groups filtered out, error reported",
			apiServerTLSProfile: makeTLSProfile(
				configv1.TLSProfileCustomType,
				configv1.VersionTLS12,
				[]string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				[]configv1.TLSGroup{
					configv1.TLSGroupX25519,         // Not FIPS compliant (filtered out)
					configv1.TLSGroupSecP256r1,      // FIPS compliant (kept)
					configv1.TLSGroupSecP384r1,      // FIPS compliant (kept)
					configv1.TLSGroupX25519MLKEM768, // Not FIPS compliant (filtered out)
				},
			),
			// Mock to simulate FIPS filtering
			mockObserveTLSSecurityProfile: func(
				listers configobserver.Listers,
				recorder events.Recorder,
				existingConfig map[string]interface{},
				minTLSVersionPath []string,
				cipherSuitesPath []string,
				groupsPath []string,
			) (map[string]interface{}, []error) {
				// Simulate FIPS filtering: only FIPS-compliant groups remain
				return map[string]interface{}{
					"servingInfo": map[string]interface{}{
						"minTLSVersion": string(configv1.VersionTLS12),
						"cipherSuites": []interface{}{
							"TLS_AES_128_GCM_SHA256",
							"TLS_AES_256_GCM_SHA384",
						},
					},
					tempTLSGroupsField: []interface{}{
						string(configv1.TLSGroupSecP256r1),
						string(configv1.TLSGroupSecP384r1),
					},
				}, nil
			},
			expectedCurvePrefs: groupsToCurvePrefs(t, []configv1.TLSGroup{
				configv1.TLSGroupSecP256r1,
				configv1.TLSGroupSecP384r1,
			}),
			expectedMinVersion: string(configv1.VersionTLS12),
			expectedCipherSuites: []string{
				"TLS_AES_128_GCM_SHA256",
				"TLS_AES_256_GCM_SHA384",
			},
			expectError: true,
			expectedErrorStrings: []string{
				"TLS groups filtered out from custom profile: [X25519 X25519MLKEM768]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listers, recorder := setupTLSGroupsTest(t, tt.apiServerTLSProfile)

			// Normalize existing config through JSON marshaling/unmarshaling to match real behavior.
			// In production, existingConfig comes from the CR's spec.observedConfig field, which is stored
			// in etcd as JSON. When retrieved, JSON numbers become float64 and concrete Go types like []string
			// become []interface{}. This round-trip simulates that normalization.
			existingConfig := tt.existingConfig
			if existingConfig == nil {
				existingConfig = map[string]interface{}{}
			} else {
				jsonBytes, err := json.Marshal(existingConfig)
				if err != nil {
					t.Fatalf("failed to marshal existing config: %v", err)
				}
				if err := json.Unmarshal(jsonBytes, &existingConfig); err != nil {
					t.Fatalf("failed to unmarshal existing config: %v", err)
				}
			}

			var observedConfig map[string]interface{}
			var errs []error

			// Use mock if provided, otherwise use the real function
			if tt.mockObserveTLSSecurityProfile != nil {
				observedConfig, errs = innerObserveTLSSecurityProfileWithGroupsConversion(listers, recorder, existingConfig, tt.mockObserveTLSSecurityProfile)
			} else {
				observedConfig, errs = ObserveTLSSecurityProfileWithGroupsConversion(listers, recorder, existingConfig)
			}

			// Normalize observed config through JSON marshaling/unmarshaling to match real behavior.
			// In production, observedConfig is written to the CR's spec.observedConfig field and persisted
			// to etcd as JSON. When the operator reads it back (as existingConfig in the next reconcile),
			// the JSON round-trip converts int64 → float64 and other type normalizations. This simulates
			// that behavior so our expected values match what would actually be seen.
			if observedConfig != nil {
				jsonBytes, err := json.Marshal(observedConfig)
				if err != nil {
					t.Fatalf("failed to marshal observed config: %v", err)
				}
				if err := json.Unmarshal(jsonBytes, &observedConfig); err != nil {
					t.Fatalf("failed to unmarshal observed config: %v", err)
				}
			}

			if tt.expectError {
				if len(errs) == 0 {
					t.Fatal("expected errors, but got none")
				}
				if len(tt.expectedErrorStrings) > 0 {
					var actualErrorStrings []string
					for _, err := range errs {
						actualErrorStrings = append(actualErrorStrings, err.Error())
					}
					if !reflect.DeepEqual(actualErrorStrings, tt.expectedErrorStrings) {
						t.Errorf("error mismatch:\nexpected: %v\nactual:   %v", tt.expectedErrorStrings, actualErrorStrings)
					}
				}
			} else {
				if len(errs) > 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
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
					t.Error("curvePreferences should not be set when not expected")
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
			} else {
				_, found, _ := unstructured.NestedString(observedConfig, "servingInfo", "minTLSVersion")
				if found {
					t.Error("minTLSVersion should not be set when not expected")
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
			} else {
				_, found, _ := unstructured.NestedStringSlice(observedConfig, "servingInfo", "cipherSuites")
				if found {
					t.Error("cipherSuites should not be set when not expected")
				}
			}

			if tempField, found, err := unstructured.NestedFieldNoCopy(observedConfig, tempTLSGroupsField); err != nil {
				t.Errorf("error checking temp field %s: %v", tempTLSGroupsField, err)
			} else if found {
				t.Errorf("temp field %s should be removed from observedConfig, got: %v", tempTLSGroupsField, tempField)
			}

			for _, event := range recorder.Events() {
				if event.Reason == "ObservationFailed" {
					t.Errorf("unexpected observation failure: %s", event.Message)
				}
			}
		})
	}
}
