package operator

import (
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
)

func TestObserveClusterConfig(t *testing.T) {
	tests := []struct {
		name   string
		cm     *corev1.ConfigMap
		expect map[string]interface{}
	}{
		{
			name: "ensure valid configmap is observed and parsed",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-controller-manager-images",
					Namespace: operatorNamespaceName,
				},
				Data: map[string]string{
					"builderImage":  "quay.io/sample/origin-builder:v4.0",
					"deployerImage": "quay.io/sample/origin-deployer:v4.0",
				},
			},
			expect: map[string]interface{}{
				"build": map[string]interface{}{
					"imageTemplateFormat": map[string]interface{}{
						"format": "quay.io/sample/origin-builder:v4.0",
					},
				},
				"deployer": map[string]interface{}{
					"imageTemplateFormat": map[string]interface{}{
						"format": "quay.io/sample/origin-deployer:v4.0",
					},
				},
			},
		},
		{
			name: "check that extraneous configmap fields are ignored",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-controller-manager-images",
					Namespace: operatorNamespaceName,
				},
				Data: map[string]string{
					"builderImage": "quay.io/sample/origin-builder:v4.0",
					"unknown":      "???",
				},
			},
			expect: map[string]interface{}{
				"build": map[string]interface{}{
					"imageTemplateFormat": map[string]interface{}{
						"format": "quay.io/sample/origin-builder:v4.0",
					},
				},
			},
		},
		{
			name: "expect empty result if no image data is found",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "openshift-controller-manager-images",
					Namespace: operatorNamespaceName,
				},
				Data: map[string]string{
					"unknownField":  "quay.io/sample/origin-builder:v4.0",
					"unknownField2": "quay.io/sample/origin-deployer:v4.0",
				},
			},
			expect: map[string]interface{}{},
		},
		{
			name: "expect empty result if no configmap is found",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shall-not-be-found",
					Namespace: operatorNamespaceName,
				},
				Data: map[string]string{
					"builderImage":  "quay.io/sample/origin-builder:v4.0",
					"deployerImage": "quay.io/sample/origin-deployer:v4.0",
				},
			},
			expect: map[string]interface{}{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			indexer.Add(tc.cm)

			listers := Listers{
				configmapLister: corelistersv1.NewConfigMapLister(indexer),
			}
			result, err := observeControllerManagerImagesConfig(listers, map[string]interface{}{})
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if !reflect.DeepEqual(result, tc.expect) {
				t.Errorf("expected %v, but got %v", tc.expect, result)
			}
		})
	}
}

func TestObserveRegistryConfig(t *testing.T) {
	const (
		expectedInternalRegistryHostname = "docker-registry.openshift-image-registry.svc.cluster.local:5000"
	)

	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	imageConfig := &configv1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: configv1.ImageStatus{
			InternalRegistryHostname: expectedInternalRegistryHostname,
		},
	}
	indexer.Add(imageConfig)

	listers := Listers{
		imageConfigLister: configlistersv1.NewImageLister(indexer),
	}

	result, err := observeInternalRegistryHostname(listers, map[string]interface{}{})
	if err != nil {
		t.Error("expected err == nil")
	}
	internalRegistryHostname, _, err := unstructured.NestedString(result, "dockerPullSecret", "internalRegistryHostname")
	if err != nil {
		t.Fatal(err)
	}
	if internalRegistryHostname != expectedInternalRegistryHostname {
		t.Errorf("expected internal registry hostname: %s, got %s", expectedInternalRegistryHostname, internalRegistryHostname)
	}
}

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
						DefaultProxy: &configv1.ProxyConfig{
							HTTPProxy:  "http://user:pass@someproxy.net",
							HTTPSProxy: "https://user:pass@someproxy.net",
							NoProxy:    "image-resgistry.cluster.svc.local",
						},
						GitProxy: &configv1.ProxyConfig{
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
						NodeSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"foo": "bar",
							},
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
						DefaultProxy: &configv1.ProxyConfig{
							HTTPProxy:  "",
							HTTPSProxy: "https://user:pass@someproxy.net",
							NoProxy:    "",
						},
						GitProxy: &configv1.ProxyConfig{
							HTTPProxy:  "http://my-proxy",
							HTTPSProxy: "",
							NoProxy:    "https://no-proxy",
						},
					},
				},
			},
		},
		{
			name: "match expressions",
			buildConfig: &configv1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.BuildSpec{
					BuildOverrides: configv1.BuildOverrides{
						NodeSelector: metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "mylabel",
									Values:   []string{"foo", "bar"},
									Operator: metav1.LabelSelectorOpIn,
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "default proxy",
			buildConfig: &configv1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.BuildSpec{
					BuildDefaults: configv1.BuildDefaults{
						DefaultProxy: &configv1.ProxyConfig{
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
						DefaultProxy: &configv1.ProxyConfig{
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
						GitProxy: &configv1.ProxyConfig{
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
			listers := Listers{
				buildConfigLister: configlistersv1.NewBuildLister(indexer),
			}
			config := map[string]interface{}{}
			observed, err := observeBuildControllerConfig(listers, config)
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
			testNestedField(observed, test.buildConfig.Spec.BuildOverrides.NodeSelector.MatchLabels, "build.buildOverrides.nodeSelector", false, t)
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
		rawExpected, err := ConvertJSON(expected)
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
		rawExpected, err := ConvertJSON(expected)
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
		rawExpected, err := ConvertJSON(expected)
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
		rawExpected, err := ConvertJSON(expected)
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
		rawExpected, err := ConvertJSON(expected)
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
