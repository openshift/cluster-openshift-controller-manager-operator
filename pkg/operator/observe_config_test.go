package operator

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
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
