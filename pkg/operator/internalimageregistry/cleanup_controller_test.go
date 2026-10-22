package internalimageregistry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openshift/library-go/pkg/controller/factory"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

const (
	testNamespace            = "test"
	testServiceAccountName   = "builder"
	testImagePullSecretName  = "builder-dockercfg-abc12"
	internalRegistryPullRef  = "openshift.io/internal-registry-pull-secret-ref"
	internalRegistryAuthAnno = "openshift.io/internal-registry-auth-token.service-account"
)

// TestCleanupRemovesImagePullSecretReferences verifies the happy path: a successful sync removes
// the managed image pull secret and strips every reference to it from the service account, while
// leaving the shared lister cache untouched.
func TestCleanupRemovesImagePullSecretReferences(t *testing.T) {
	cachedServiceAccount := newServiceAccountWithInternalRegistryPullSecret(
		testNamespace,
		testServiceAccountName,
		testImagePullSecretName,
	)
	imagePullSecret := newManagedImagePullSecret(testNamespace, testImagePullSecretName, testServiceAccountName)

	// Keep a pristine copy to prove the lister cache is never mutated in place.
	expectedCachedServiceAccount := cachedServiceAccount.DeepCopy()

	kubeClient := fake.NewSimpleClientset(cachedServiceAccount.DeepCopy(), imagePullSecret.DeepCopy())

	serviceAccountIndexer := newServiceAccountIndexer(t, cachedServiceAccount)
	secretIndexer := newSecretIndexer(t, imagePullSecret)

	controller := newCleanupControllerForTest(
		kubeClient,
		corelistersv1.NewServiceAccountLister(serviceAccountIndexer),
		corelistersv1.NewSecretLister(secretIndexer),
		corelistersv1.NewPodLister(newEmptyIndexer()),
	)

	if err := controller.cleanup(context.Background()); err != nil {
		t.Fatalf("expected cleanup to succeed, got %v", err)
	}

	// The object sent to the API server must have every reference to the pull secret removed.
	updated, err := kubeClient.CoreV1().ServiceAccounts(testNamespace).Get(context.Background(), testServiceAccountName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get updated service account: %v", err)
	}
	if _, ok := updated.Annotations[internalRegistryPullRef]; ok {
		t.Errorf("expected annotation %q to be removed, but it is still present", internalRegistryPullRef)
	}
	if name := serviceAccountReferencesSecret(updated, testImagePullSecretName); name {
		t.Errorf("expected service account to no longer reference pull secret %q", testImagePullSecretName)
	}

	// The managed image pull secret must be deleted.
	if _, err := kubeClient.CoreV1().Secrets(testNamespace).Get(context.Background(), testImagePullSecretName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Errorf("expected image pull secret %q to be deleted, got err=%v", testImagePullSecretName, err)
	}

	// The shared lister cache must not have been mutated in place.
	assertServiceAccountHasInternalRegistryPullSecret(t, serviceAccountIndexer, expectedCachedServiceAccount)
}

// TestCleanupIsIdempotent verifies that once the references have been cleaned up, a subsequent sync
// (after the informer caches have observed the cleaned state) issues no further updates.
func TestCleanupIsIdempotent(t *testing.T) {
	cachedServiceAccount := newServiceAccountWithInternalRegistryPullSecret(
		testNamespace,
		testServiceAccountName,
		testImagePullSecretName,
	)
	imagePullSecret := newManagedImagePullSecret(testNamespace, testImagePullSecretName, testServiceAccountName)

	kubeClient := fake.NewSimpleClientset(cachedServiceAccount.DeepCopy(), imagePullSecret.DeepCopy())

	serviceAccountIndexer := newServiceAccountIndexer(t, cachedServiceAccount)
	secretIndexer := newSecretIndexer(t, imagePullSecret)

	controller := newCleanupControllerForTest(
		kubeClient,
		corelistersv1.NewServiceAccountLister(serviceAccountIndexer),
		corelistersv1.NewSecretLister(secretIndexer),
		corelistersv1.NewPodLister(newEmptyIndexer()),
	)

	if err := controller.cleanup(context.Background()); err != nil {
		t.Fatalf("expected first cleanup to succeed, got %v", err)
	}

	// Simulate the informer caches catching up with the first sync: the cleaned service account
	// replaces the stale one and the deleted pull secret leaves the cache.
	cleanedServiceAccount, err := kubeClient.CoreV1().ServiceAccounts(testNamespace).Get(context.Background(), testServiceAccountName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get cleaned service account: %v", err)
	}
	if err := serviceAccountIndexer.Update(cleanedServiceAccount); err != nil {
		t.Fatalf("failed to update service account in indexer: %v", err)
	}
	if err := secretIndexer.Delete(imagePullSecret); err != nil {
		t.Fatalf("failed to delete secret from indexer: %v", err)
	}

	kubeClient.ClearActions()

	if err := controller.cleanup(context.Background()); err != nil {
		t.Fatalf("expected second cleanup to succeed, got %v", err)
	}

	for _, action := range kubeClient.Actions() {
		if action.Matches("update", "serviceaccounts") || action.Matches("delete", "secrets") {
			t.Errorf("expected no further writes on idempotent sync, got %s %s", action.GetVerb(), action.GetResource().Resource)
		}
	}
}

// TestCleanupServiceAccountUpdateConflictDoesNotMutateListerCache verifies that a conflict on the
// service account update is surfaced as a requeue and never mutates the shared lister cache.
func TestCleanupServiceAccountUpdateConflictDoesNotMutateListerCache(t *testing.T) {
	cachedServiceAccount := newServiceAccountWithInternalRegistryPullSecret(
		testNamespace,
		testServiceAccountName,
		testImagePullSecretName,
	)
	imagePullSecret := newManagedImagePullSecret(testNamespace, testImagePullSecretName, testServiceAccountName)

	expectedServiceAccount := cachedServiceAccount.DeepCopy()

	kubeClient := fake.NewSimpleClientset(cachedServiceAccount.DeepCopy(), imagePullSecret.DeepCopy())
	kubeClient.PrependReactor("update", "serviceaccounts", func(action clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewConflict(
			corev1.Resource("serviceaccounts"),
			testServiceAccountName,
			errors.New("conflict"),
		)
	})

	serviceAccountIndexer := newServiceAccountIndexer(t, cachedServiceAccount)
	secretIndexer := newSecretIndexer(t, imagePullSecret)

	controller := newCleanupControllerForTest(
		kubeClient,
		corelistersv1.NewServiceAccountLister(serviceAccountIndexer),
		corelistersv1.NewSecretLister(secretIndexer),
		corelistersv1.NewPodLister(newEmptyIndexer()),
	)

	err := controller.cleanup(context.Background())
	if !errors.Is(err, factory.SyntheticRequeueError) {
		t.Fatalf("expected SyntheticRequeueError, got %v", err)
	}

	assertServiceAccountHasInternalRegistryPullSecret(t, serviceAccountIndexer, expectedServiceAccount)
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

// newManagedImagePullSecret returns an image pull secret that looks like one generated by the
// openshift-controller-manager: it carries the auth-token annotation and was created long enough
// ago that the cleanup controller will not skip it.
func newManagedImagePullSecret(namespace, name, serviceAccountName string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				internalRegistryAuthAnno: serviceAccountName,
			},
			CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Hour)),
		},
		Type: corev1.SecretTypeDockercfg,
	}
}

func serviceAccountReferencesSecret(serviceAccount *corev1.ServiceAccount, secretName string) bool {
	for _, ref := range serviceAccount.Secrets {
		if ref.Name == secretName {
			return true
		}
	}
	for _, ref := range serviceAccount.ImagePullSecrets {
		if ref.Name == secretName {
			return true
		}
	}
	return false
}

func newServiceAccountIndexer(t *testing.T, serviceAccount *corev1.ServiceAccount) cache.Indexer {
	t.Helper()

	indexer := newEmptyIndexer()
	if err := indexer.Add(serviceAccount); err != nil {
		t.Fatalf("failed to add service account to indexer: %v", err)
	}

	return indexer
}

func newSecretIndexer(t *testing.T, secret *corev1.Secret) cache.Indexer {
	t.Helper()

	indexer := newEmptyIndexer()
	if err := indexer.Add(secret); err != nil {
		t.Fatalf("failed to add secret to indexer: %v", err)
	}

	return indexer
}

func newEmptyIndexer() cache.Indexer {
	return cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	})
}

func newCleanupControllerForTest(
	kubeClient kubernetes.Interface,
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
