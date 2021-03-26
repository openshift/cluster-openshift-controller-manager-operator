package images

import (
	"reflect"
	"strings"
	"testing"

	"github.com/openshift/library-go/pkg/operator/events"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

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

	listers := configobservation.Listers{
		ImageConfigLister: configlistersv1.NewImageLister(indexer),
	}

	result, errs := ObserveInternalRegistryHostname(listers, events.NewInMemoryRecorder(""), map[string]interface{}{})
	if len(errs) != 0 {
		t.Errorf("expected no errors: %v", errs)
	}
	internalRegistryHostname, _, err := unstructured.NestedString(result, "dockerPullSecret", "internalRegistryHostname")
	if err != nil {
		t.Fatal(err)
	}
	if internalRegistryHostname != expectedInternalRegistryHostname {
		t.Errorf("expected internal registry hostname: %s, got %s", expectedInternalRegistryHostname, internalRegistryHostname)
	}
}

func TestObserveRegistryExternalHostnames(t *testing.T) {
	for _, tt := range []struct {
		name     string
		err      string
		config   *configv1.Image
		expected map[string]interface{}
		existing map[string]interface{}
	}{
		{
			name:     "broken pre existing config",
			err:      "accessor error: broken is of the type string",
			config:   nil,
			expected: map[string]interface{}{},
			existing: map[string]interface{}{
				"dockerPullSecret": "broken",
			},
		},
		{
			name:     "empty if no image config found",
			err:      "",
			config:   nil,
			expected: map[string]interface{}{},
			existing: map[string]interface{}{
				"dockerPullSecret": map[string]interface{}{
					"registryURLs": []interface{}{
						"old-registry.openshift.io",
					},
				},
			},
		},
		{
			name: "empty if no external hostnames in image config status",
			err:  "",
			config: &configv1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
			},
			expected: map[string]interface{}{},
			existing: map[string]interface{}{
				"dockerPullSecret": map[string]interface{}{
					"registryURLs": []interface{}{
						"old-registry.openshift.io",
					},
				},
			},
		},
		{
			name: "empty if external hostnames only in image config spec",
			err:  "",
			config: &configv1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.ImageSpec{
					ExternalRegistryHostnames: []string{
						"hostname-in-spec.openshift.io",
					},
				},
			},
			expected: map[string]interface{}{},
			existing: map[string]interface{}{
				"dockerPullSecret": map[string]interface{}{
					"registryURLs": []interface{}{
						"old-registry.openshift.io",
					},
				},
			},
		},
		{
			name: "using external hostnames from image config status",
			err:  "",
			config: &configv1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Status: configv1.ImageStatus{
					ExternalRegistryHostnames: []string{
						"hostname-0-in-status.openshift.io",
						"hostname-1-in-status.openshift.io",
					},
				},
			},
			expected: map[string]interface{}{
				"dockerPullSecret": map[string]interface{}{
					"registryURLs": []interface{}{
						"hostname-0-in-status.openshift.io",
						"hostname-1-in-status.openshift.io",
					},
				},
			},
			existing: map[string]interface{}{
				"dockerPullSecret": map[string]interface{}{
					"registryURLs": []interface{}{
						"old-registry.openshift.io",
					},
				},
			},
		},
		{
			name: "ignoring hostnames from image config's spec",
			err:  "",
			config: &configv1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.ImageSpec{
					ExternalRegistryHostnames: []string{
						"hostname-0-in-status.openshift.io",
						"hostname-1-in-status.openshift.io",
					},
				},
				Status: configv1.ImageStatus{
					ExternalRegistryHostnames: []string{
						"hostname-2-in-status.openshift.io",
						"hostname-3-in-status.openshift.io",
					},
				},
			},
			expected: map[string]interface{}{
				"dockerPullSecret": map[string]interface{}{
					"registryURLs": []interface{}{
						"hostname-2-in-status.openshift.io",
						"hostname-3-in-status.openshift.io",
					},
				},
			},
			existing: map[string]interface{}{
				"dockerPullSecret": map[string]interface{}{
					"registryURLs": []interface{}{
						"old-registry.openshift.io",
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			if tt.config != nil {
				indexer.Add(tt.config)
			}
			listers := configobservation.Listers{
				ImageConfigLister: configlistersv1.NewImageLister(indexer),
			}

			result, errs := ObserveExternalRegistryHostnames(
				listers, events.NewInMemoryRecorder(""), tt.existing,
			)
			if len(errs) > 0 && len(tt.err) == 0 {
				t.Errorf("unexpected error: %v", errs)
			} else if len(errs) > 0 {
				errstr := ""
				for _, err := range errs {
					errstr += err.Error()
				}
				if !strings.Contains(errstr, tt.err) {
					t.Errorf(
						"expecting error to have %q, %v received instead",
						tt.err, errs,
					)
				}
			} else if len(tt.err) > 0 {
				t.Errorf("expecting errors to contain %q, nil received instead", tt.err)
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf(
					"expected config %+v, got %+v",
					tt.expected, result,
				)
			}
		})
	}
}
