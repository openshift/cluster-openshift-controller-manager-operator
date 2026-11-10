package internalimageregistry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
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
	unrelatedAnno            = "example.com/unrelated"
	unrelatedAnnoValue       = "keep-me"
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
	if serviceAccountReferencesSecret(updated, testImagePullSecretName) {
		t.Errorf("expected service account to no longer reference pull secret %q", testImagePullSecretName)
	}

	// The managed image pull secret must be deleted.
	if _, err := kubeClient.CoreV1().Secrets(testNamespace).Get(context.Background(), testImagePullSecretName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Errorf("expected image pull secret %q to be deleted, got err=%v", testImagePullSecretName, err)
	}

	// The shared lister cache must not have been mutated in place.
	assertServiceAccountUnchanged(t, serviceAccountIndexer, expectedCachedServiceAccount)
}

// TestCleanupRemovesStaleAnnotationWhenImagePullSecretIsGone verifies that when the managed image
// pull secret and every reference to it are already gone, cleanup strips the leftover internal
// registry annotation and a subsequent sync issues no further updates.
func TestCleanupRemovesStaleAnnotationWhenImagePullSecretIsGone(t *testing.T) {
	cachedServiceAccount := newServiceAccountWithStaleInternalRegistryPullSecretAnnotation(
		testNamespace,
		testServiceAccountName,
		testImagePullSecretName,
	)

	// Keep a pristine copy to prove the lister cache is never mutated in place.
	expectedCachedServiceAccount := cachedServiceAccount.DeepCopy()

	// Neither the fake API server nor the secret cache holds the pull secret: it is already deleted.
	kubeClient := fake.NewSimpleClientset(cachedServiceAccount.DeepCopy())

	serviceAccountIndexer := newServiceAccountIndexer(t, cachedServiceAccount)

	controller := newCleanupControllerForTest(
		kubeClient,
		corelistersv1.NewServiceAccountLister(serviceAccountIndexer),
		corelistersv1.NewSecretLister(newEmptyIndexer()),
		corelistersv1.NewPodLister(newEmptyIndexer()),
	)

	if err := controller.cleanup(context.Background()); err != nil {
		t.Fatalf("expected cleanup to succeed, got %v", err)
	}

	updated, err := kubeClient.CoreV1().ServiceAccounts(testNamespace).Get(context.Background(), testServiceAccountName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get updated service account: %v", err)
	}
	if _, ok := updated.Annotations[internalRegistryPullRef]; ok {
		t.Errorf("expected annotation %q to be removed, but it is still present", internalRegistryPullRef)
	}
	// Only the controller's own annotation may be removed.
	if got := updated.Annotations[unrelatedAnno]; got != unrelatedAnnoValue {
		t.Errorf("expected unrelated annotation %q to remain %q, got %q", unrelatedAnno, unrelatedAnnoValue, got)
	}

	// The shared lister cache must not have been mutated in place.
	assertServiceAccountUnchanged(t, serviceAccountIndexer, expectedCachedServiceAccount)

	// Once the cache observes the stripped annotation, the loop must stop: no further writes.
	if err := serviceAccountIndexer.Update(updated); err != nil {
		t.Fatalf("failed to update service account in indexer: %v", err)
	}

	kubeClient.ClearActions()

	if err := controller.cleanup(context.Background()); err != nil {
		t.Fatalf("expected second cleanup to succeed, got %v", err)
	}

	for _, action := range kubeClient.Actions() {
		if action.Matches("update", "serviceaccounts") {
			t.Errorf("expected no further service account updates once the annotation is gone, got %s %s", action.GetVerb(), action.GetResource().Resource)
		}
	}
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
// service account update is surfaced as an error and never mutates the shared lister cache.
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

	if err := controller.cleanup(context.Background()); err == nil {
		t.Fatal("expected cleanup to fail on the service account update conflict")
	}

	assertServiceAccountUnchanged(t, serviceAccountIndexer, expectedServiceAccount)
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

// newServiceAccountWithStaleInternalRegistryPullSecretAnnotation returns a service account whose
// only trace of the managed image pull secret is the internal registry annotation: both reference
// lists are already empty. The unrelated annotation guards against cleanup removing more than its
// own key.
func newServiceAccountWithStaleInternalRegistryPullSecretAnnotation(namespace, name, secretName string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				internalRegistryPullRef: secretName,
				unrelatedAnno:           unrelatedAnnoValue,
			},
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

// assertServiceAccountUnchanged verifies that the service account held by the shared lister cache
// is still byte-for-byte identical to expected, i.e. that cleanup never mutated it in place.
func assertServiceAccountUnchanged(t *testing.T, indexer cache.Indexer, expected *corev1.ServiceAccount) {
	t.Helper()

	cached, exists, err := indexer.GetByKey(expected.Namespace + "/" + expected.Name)
	if err != nil {
		t.Fatalf("failed to get cached service account: %v", err)
	}
	if !exists {
		t.Fatal("expected cached service account to exist")
	}

	if cachedServiceAccount := cached.(*corev1.ServiceAccount); !equality.Semantic.DeepEqual(expected, cachedServiceAccount) {
		t.Fatalf("cached service account was mutated:\n%s", cmp.Diff(expected, cachedServiceAccount))
	}
}
