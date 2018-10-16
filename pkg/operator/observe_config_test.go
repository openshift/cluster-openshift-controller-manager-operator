package operator

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestObserveClusterConfig(t *testing.T) {
	tests := []struct {
		name   string
		cm     *v1.ConfigMap
		expect map[string]interface{}
	}{
		{
			name: "ensure valid configmap is observed and parsed",
			cm: &v1.ConfigMap{
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
			cm: &v1.ConfigMap{
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
			cm: &v1.ConfigMap{
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
			cm: &v1.ConfigMap{
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
			kubeClient := fake.NewSimpleClientset(tc.cm)

			result, err := observeControllerManagerImagesConfig(kubeClient, &rest.Config{}, map[string]interface{}{})
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if !reflect.DeepEqual(result, tc.expect) {
				t.Errorf("expected %v, but got %v", tc.expect, result)
			}
		})
	}
}
