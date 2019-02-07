package builds

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

func TestObserveBuildControllerConfig(t *testing.T) {
	memLimit, err := resource.ParseQuantity("1G")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		buildConfig *configv1.Build
		expectError bool
	}{
		{
			name: "no build config",
		},
		{
			name: "valid build config",
			buildConfig: &configv1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.BuildSpec{
					BuildDefaults: configv1.BuildDefaults{
						DefaultProxy: &configv1.ProxySpec{
							HTTPProxy:  "http://user:pass@someproxy.net",
							HTTPSProxy: "https://user:pass@someproxy.net",
							NoProxy:    "image-resgistry.cluster.svc.local",
						},
						GitProxy: &configv1.ProxySpec{
							HTTPProxy:  "http://my-proxy",
							HTTPSProxy: "https://my-proxy",
							NoProxy:    "https://no-proxy",
						},
						Env: []corev1.EnvVar{
							{
								Name:  "FOO",
								Value: "BAR",
							},
						},
						ImageLabels: []configv1.ImageLabel{
							{
								Name:  "build.openshift.io",
								Value: "test",
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: memLimit,
							},
						},
					},
					BuildOverrides: configv1.BuildOverrides{
						ImageLabels: []configv1.ImageLabel{
							{
								Name:  "build.openshift.io",
								Value: "teset2",
							},
						},
						NodeSelector: map[string]string{
							"foo": "bar",
						},
						Tolerations: []corev1.Toleration{
							{
								Key:      "somekey",
								Operator: corev1.TolerationOpExists,
								Effect:   corev1.TaintEffectNoSchedule,
							},
						},
					},
				},
			},
		},
		{
			name: "empty proxy values",
			buildConfig: &configv1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.BuildSpec{
					BuildDefaults: configv1.BuildDefaults{
						DefaultProxy: &configv1.ProxySpec{
							HTTPProxy:  "",
							HTTPSProxy: "https://user:pass@someproxy.net",
							NoProxy:    "",
						},
						GitProxy: &configv1.ProxySpec{
							HTTPProxy:  "http://my-proxy",
							HTTPSProxy: "",
							NoProxy:    "https://no-proxy",
						},
					},
				},
			},
		},
		{
			name: "default proxy",
			buildConfig: &configv1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.BuildSpec{
					BuildDefaults: configv1.BuildDefaults{
						DefaultProxy: &configv1.ProxySpec{
							HTTPProxy:  "http://user:pass@someproxy.net",
							HTTPSProxy: "https://user:pass@someproxy.net",
							NoProxy:    "image-resgistry.cluster.svc.local",
						},
					},
				},
			},
		},
		{
			name: "default proxy with env vars",
			buildConfig: &configv1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.BuildSpec{
					BuildDefaults: configv1.BuildDefaults{
						DefaultProxy: &configv1.ProxySpec{
							HTTPProxy:  "http://user:pass@someproxy.net",
							HTTPSProxy: "https://user:pass@someproxy.net",
							NoProxy:    "image-resgistry.cluster.svc.local",
						},
						Env: []corev1.EnvVar{
							{
								Name:  "HTTP_PROXY",
								Value: "http://other:user@otherproxy.com",
							},
							{
								Name:  "HTTPS_PROXY",
								Value: "https://other:user@otherproxy.com",
							},
							{
								Name:  "NO_PROXY",
								Value: "somedomain",
							},
						},
					},
				},
			},
		},
		{
			name: "git proxy",
			buildConfig: &configv1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.BuildSpec{
					BuildDefaults: configv1.BuildDefaults{
						GitProxy: &configv1.ProxySpec{
							HTTPProxy:  "http://user:pass@someproxy.net",
							HTTPSProxy: "https://user:pass@someproxy.net",
							NoProxy:    "image-resgistry.cluster.svc.local",
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			if test.buildConfig != nil {
				indexer.Add(test.buildConfig)
			}
			listers := configobservation.Listers{
				BuildConfigLister: configlistersv1.NewBuildLister(indexer),
			}
			config := map[string]interface{}{}
			observed, err := ObserveBuildControllerConfig(listers, events.NewInMemoryRecorder(""), config)
			if err != nil {
				if !test.expectError {
					t.Fatalf("unexpected error observing build controller config: %v", err)
				}
			}
			if test.expectError {
				if err == nil {
					t.Error("expected error to be thrown, but was not")
				}
				if len(observed) > 0 {
					t.Error("expected returned config to be empty")
				}
				return
			}
			if test.buildConfig == nil {
				if len(observed) > 0 {
					t.Errorf("expected empty observed config, got %v", observed)
				}
				return
			}

			expectedEnv := test.buildConfig.Spec.BuildDefaults.Env
			testNestedField(observed, expectedEnv, "build.buildDefaults.env", false, t)
			testNestedField(observed, test.buildConfig.Spec.BuildDefaults.ImageLabels, "build.buildDefaults.imageLabels", false, t)
			testNestedField(observed, test.buildConfig.Spec.BuildOverrides.ImageLabels, "build.buildOverrides.imageLabels", false, t)
			testNestedField(observed, test.buildConfig.Spec.BuildOverrides.NodeSelector, "build.buildOverrides.nodeSelector", false, t)
			testNestedField(observed, test.buildConfig.Spec.BuildOverrides.Tolerations, "build.buildOverrides.tolerations", false, t)

			expectedGitProxy := test.buildConfig.Spec.BuildDefaults.DefaultProxy
			if test.buildConfig.Spec.BuildDefaults.GitProxy != nil {
				expectedGitProxy = test.buildConfig.Spec.BuildDefaults.GitProxy
			}
			if expectedGitProxy != nil {
				testNestedField(observed, expectedGitProxy.HTTPProxy, "build.buildDefaults.gitHTTPProxy", true, t)
				testNestedField(observed, expectedGitProxy.HTTPSProxy, "build.buildDefaults.gitHTTPSProxy", true, t)
				testNestedField(observed, expectedGitProxy.NoProxy, "build.buildDefaults.gitNoProxy", true, t)
			} else {
				testNestedField(observed, nil, "build.buildDefaults.gitHTTPProxy", false, t)
				testNestedField(observed, nil, "build.buildDefaults.gitHTTPSProxy", false, t)
				testNestedField(observed, nil, "build.buildDefaults.gitNoProxy", false, t)
			}
		})
	}
}

func testNestedField(obj map[string]interface{}, expectedVal interface{}, field string, existIfEmpty bool, t *testing.T) {
	nestedField := strings.Split(field, ".")
	switch expected := expectedVal.(type) {
	case string:
		value, found, err := unstructured.NestedString(obj, nestedField...)
		if err != nil {
			t.Fatalf("failed to read nested string %s: %v", field, err)
		}
		if expected != value {
			t.Errorf("expected field %s to be %s, got %s", field, expectedVal, value)
		}
		if existIfEmpty && !found {
			t.Errorf("expected field %s to exist, even if empty", field)
		}
	case map[string]string:
		value, found, err := unstructured.NestedStringMap(obj, nestedField...)
		if err != nil {
			t.Fatal(err)
		}
		if !equality.Semantic.DeepEqual(value, expected) {
			t.Errorf("expected %s to be %v, got %v", field, expected, value)
		}
		if existIfEmpty && !found {
			t.Errorf("expected field %s to exist, even if empty", field)
		}
	case []corev1.EnvVar:
		value, found, err := unstructured.NestedSlice(obj, nestedField...)
		if err != nil {
			t.Fatal(err)
		}
		rawExpected, err := configobservation.ConvertJSON(expected)
		if err != nil {
			t.Fatalf("unable to test field %s: %v", field, err)
		}
		if !equality.Semantic.DeepEqual(value, rawExpected) {
			t.Errorf("expected %s to be %v, got %v", field, rawExpected, value)
		}
		if existIfEmpty && !found {
			t.Errorf("expected field %s to exist, even if empty", field)
		}
	case []corev1.Toleration:
		value, found, err := unstructured.NestedSlice(obj, nestedField...)
		if err != nil {
			t.Fatal(err)
		}
		rawExpected, err := configobservation.ConvertJSON(expected)
		if err != nil {
			t.Fatalf("unable to test field %s: %v", field, err)
		}
		if !equality.Semantic.DeepEqual(value, rawExpected) {
			t.Errorf("expected %s to be %v, got %v", field, expected, value)
		}
		if existIfEmpty && !found {
			t.Errorf("expected field %s to exist, even if empty", field)
		}
	case []configv1.ImageLabel:
		value, found, err := unstructured.NestedSlice(obj, nestedField...)
		if err != nil {
			t.Fatal(err)
		}
		rawExpected, err := configobservation.ConvertJSON(expected)
		if err != nil {
			t.Fatalf("unable to test field %s: %v", field, err)
		}
		if !equality.Semantic.DeepEqual(value, rawExpected) {
			t.Errorf("expected %s to be %v, got %v", field, expected, value)
		}
		if existIfEmpty && !found {
			t.Errorf("expected field %s to exist, even if empty", field)
		}
	case []interface{}:
		value, found, err := unstructured.NestedSlice(obj, nestedField...)
		if err != nil {
			t.Fatalf("unable to test field %s: %v", field, err)
		}
		rawExpected, err := configobservation.ConvertJSON(expected)
		if err != nil {
			t.Fatalf("unable to test field %s: %v", field, err)
		}
		if !equality.Semantic.DeepEqual(value, rawExpected) {
			t.Errorf("expected %s to be %v, got %v", field, expected, value)
		}
		if existIfEmpty && !found {
			t.Errorf("expected field %s to exist, even if empty", field)
		}
	default:
		value, found, err := unstructured.NestedFieldCopy(obj, nestedField...)
		if err != nil {
			t.Fatalf("unable to test field %s: %v", field, err)
		}
		rawExpected, err := configobservation.ConvertJSON(expected)
		if err != nil {
			t.Fatalf("unable to test field %s: %v", field, err)
		}
		if !equality.Semantic.DeepEqual(rawExpected, value) {
			t.Errorf("expected %s to be %v; got %v", field, expected, value)
		}
		if existIfEmpty && !found {
			t.Errorf("expected field %s to exist, even if empty", field)
		}
	}
}
