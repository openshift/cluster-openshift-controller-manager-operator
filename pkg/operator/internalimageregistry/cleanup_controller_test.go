package internalimageregistry

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openshift/library-go/pkg/controller/factory"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	testNamespace           = "test"
	testServiceAccountName  = "builder"
	testImagePullSecretName = "builder-dockercfg-abc12"
	internalRegistryPullRef = "openshift.io/internal-registry-pull-secret-ref"
)

func TestCleanupServiceAccountUpdateConflictDoesNotMutateListerCache(t *testing.T) {
	cachedServiceAccount := newServiceAccountWithInternalRegistryPullSecret(
		testNamespace,
		testServiceAccountName,
		testImagePullSecretName,
	)

	expectedServiceAccount := cachedServiceAccount.DeepCopy()

	indexer := newServiceAccountIndexer(t, cachedServiceAccount)

	controller := newCleanupControllerForTest(
		newKubeClientWithServiceAccountUpdateConflict(t),
		corelistersv1.NewServiceAccountLister(indexer),
		corelistersv1.NewSecretLister(newEmptyIndexer()),
		corelistersv1.NewPodLister(newEmptyIndexer()),
	)

	err := controller.cleanup(context.Background())
	if !errors.Is(err, factory.SyntheticRequeueError) {
		t.Fatalf("expected SyntheticRequeueError, got %v", err)
	}

	assertServiceAccountHasInternalRegistryPullSecret(t, indexer, expectedServiceAccount)
}

func newKubeClientWithServiceAccountUpdateConflict(t *testing.T) *kubernetes.Clientset {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/serviceaccounts/") {
			status := metav1.Status{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Status",
				},
				Status:  metav1.StatusFailure,
				Message: "conflict",
				Reason:  metav1.StatusReasonConflict,
				Code:    http.StatusConflict,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(status); err != nil {
				t.Errorf("failed to encode conflict status: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	t.Cleanup(server.Close)

	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	if err != nil {
		t.Fatalf("failed to create kubernetes client: %v", err)
	}

	return client
}

func newServiceAccountWithInternalRegistryPullSecret(namespace, name, secretName string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				internalRegistryPullRef: secretName,
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{Name: secretName},
		},
		Secrets: []corev1.ObjectReference{
			{Name: secretName},
		},
	}
}

func newServiceAccountIndexer(t *testing.T, serviceAccount *corev1.ServiceAccount) cache.Indexer {
	t.Helper()

	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	})
	if err := indexer.Add(serviceAccount); err != nil {
		t.Fatalf("failed to add service account to indexer: %v", err)
	}

	return indexer
}

func newEmptyIndexer() cache.Indexer {
	return cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	})
}

func newCleanupControllerForTest(
	kubeClient *kubernetes.Clientset,
	serviceAccountLister corelistersv1.ServiceAccountLister,
	secretLister corelistersv1.SecretLister,
	podLister corelistersv1.PodLister,
) *imagePullSecretCleanupController {
	return &imagePullSecretCleanupController{
		kubeClient:           kubeClient,
		serviceAccountLister: serviceAccountLister,
		secretLister:         secretLister,
		podLister:            podLister,
	}
}

func assertServiceAccountHasInternalRegistryPullSecret(t *testing.T, indexer cache.Indexer, expected *corev1.ServiceAccount) {
	t.Helper()

	cached, exists, err := indexer.GetByKey(expected.Namespace + "/" + expected.Name)
	if err != nil {
		t.Fatalf("failed to get cached service account: %v", err)
	}
	if !exists {
		t.Fatal("expected cached service account to exist")
	}

	cachedServiceAccount := cached.(*corev1.ServiceAccount)

	if got := cachedServiceAccount.Annotations[internalRegistryPullRef]; got != expected.Annotations[internalRegistryPullRef] {
		t.Fatalf(
			"expected cached annotation %q to remain %q, got %q",
			internalRegistryPullRef,
			expected.Annotations[internalRegistryPullRef],
			got,
		)
	}

	if len(cachedServiceAccount.ImagePullSecrets) != len(expected.ImagePullSecrets) {
		t.Fatalf("expected %d image pull secret references, got %d", len(expected.ImagePullSecrets), len(cachedServiceAccount.ImagePullSecrets))
	}
	for i, expectedRef := range expected.ImagePullSecrets {
		if cachedServiceAccount.ImagePullSecrets[i].Name != expectedRef.Name {
			t.Fatalf("expected image pull secret reference %q, got %q", expectedRef.Name, cachedServiceAccount.ImagePullSecrets[i].Name)
		}
	}

	if len(cachedServiceAccount.Secrets) != len(expected.Secrets) {
		t.Fatalf("expected %d secret references, got %d", len(expected.Secrets), len(cachedServiceAccount.Secrets))
	}
	for i, expectedRef := range expected.Secrets {
		if cachedServiceAccount.Secrets[i].Name != expectedRef.Name {
			t.Fatalf("expected secret reference %q, got %q", expectedRef.Name, cachedServiceAccount.Secrets[i].Name)
		}
	}
}
