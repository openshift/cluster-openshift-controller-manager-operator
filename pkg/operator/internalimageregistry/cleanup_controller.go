package internalimageregistry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/client-go/config/informers/externalversions"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/build/naming"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

type imagePullSecretCleanupController struct {
	factory.Controller
	kubeClient            *kubernetes.Clientset
	serviceAccountLister  corelistersv1.ServiceAccountLister
	secretLister          corelistersv1.SecretLister
	podLister             corelistersv1.PodLister
	clusterVersionLister  configlistersv1.ClusterVersionLister
	clusterOperatorLister configlistersv1.ClusterOperatorLister
	rateLimiter           flowcontrol.RateLimiter
}

const (
	// batchSize limits how many ServiceAccounts to process per sync cycle
	// to avoid overwhelming the API server, especially on SNO clusters
	batchSize = 10
	// maxProcessingTime limits how long a single sync can take
	// to ensure the controller remains responsive
	maxProcessingTime = 30 * time.Second
)

func NewImagePullSecretCleanupController(kubeClient *kubernetes.Clientset, informers v1helpers.KubeInformersForNamespaces, configInformers externalversions.SharedInformerFactory, recorder events.Recorder) *imagePullSecretCleanupController {
	c := &imagePullSecretCleanupController{
		kubeClient:            kubeClient,
		serviceAccountLister:  informers.InformersFor(metav1.NamespaceAll).Core().V1().ServiceAccounts().Lister(),
		secretLister:          informers.InformersFor(metav1.NamespaceAll).Core().V1().Secrets().Lister(),
		podLister:             informers.InformersFor(metav1.NamespaceAll).Core().V1().Pods().Lister(),
		clusterVersionLister:  configInformers.Config().V1().ClusterVersions().Lister(),
		clusterOperatorLister: configInformers.Config().V1().ClusterOperators().Lister(),
		// Rate limiter: 1 request per 2 seconds with burst of 5 to be conservative on SNO
		rateLimiter: flowcontrol.NewTokenBucketRateLimiter(0.5, 5),
	}
	c.Controller = factory.New().
		WithInformers(
			informers.InformersFor(metav1.NamespaceAll).Core().V1().ServiceAccounts().Informer(),
			informers.InformersFor(metav1.NamespaceAll).Core().V1().Secrets().Informer(),
			informers.InformersFor(metav1.NamespaceAll).Core().V1().Pods().Informer(),
			configInformers.Config().V1().ClusterVersions().Informer(),
			configInformers.Config().V1().ClusterOperators().Informer(),
		).
		WithSync(c.sync).
		ToController("ImagePullSecretCleanupController", recorder.WithComponentSuffix("image-pull-secret-cleanup-controller"))
	return c
}

func (c *imagePullSecretCleanupController) sync(ctx context.Context, controllerContext factory.SyncContext) error {
	imageRegistryEnabled, err := ImageRegistryIsEnabled(c.clusterVersionLister, c.clusterOperatorLister)
	if err != nil {
		return err
	}
	if imageRegistryEnabled {
		return nil
	}
	return c.cleanup(ctx)
}

func (c *imagePullSecretCleanupController) cleanup(ctx context.Context) error {
	// Use a timeout context to ensure cleanup doesn't run indefinitely
	cleanupCtx, cancel := context.WithTimeout(ctx, maxProcessingTime)
	defer cancel()

	// Get all service accounts
	serviceAccounts, err := c.serviceAccountLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("unable to list ServiceAccounts: %w", err)
	}

	klog.V(4).InfoS("Starting image pull secret cleanup", "totalServiceAccounts", len(serviceAccounts))

	// Process ServiceAccounts in batches to avoid overwhelming the API server
	processed := 0
	for i := 0; i < len(serviceAccounts); i += batchSize {
		// Check context cancellation between batches
		select {
		case <-cleanupCtx.Done():
			klog.V(2).InfoS("Image pull secret cleanup cancelled or timed out", "processed", processed, "total", len(serviceAccounts))
			return nil
		default:
		}

		end := i + batchSize
		if end > len(serviceAccounts) {
			end = len(serviceAccounts)
		}

		batch := serviceAccounts[i:end]
		if err := c.processBatch(cleanupCtx, batch); err != nil {
			// Log error but continue with next batch to make progress
			klog.ErrorS(err, "Error processing batch", "batchStart", i, "batchSize", len(batch))
			// Return error only for critical failures, continue for transient API errors
			if !c.isRetriableError(err) {
				return err
			}
		}

		processed += len(batch)
		klog.V(4).InfoS("Processed batch", "processed", processed, "total", len(serviceAccounts))

		// Rate limit between batches to avoid overwhelming the API server
		if i+batchSize < len(serviceAccounts) {
			c.rateLimiter.Accept()
		}
	}

	klog.V(4).InfoS("Completed image pull secret cleanup", "totalProcessed", processed)
	return nil
}

// processBatch handles a batch of ServiceAccounts with proper error handling
func (c *imagePullSecretCleanupController) processBatch(ctx context.Context, serviceAccounts []*corev1.ServiceAccount) error {
	for _, serviceAccount := range serviceAccounts {
		// Check context cancellation frequently
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := c.processServiceAccount(ctx, serviceAccount); err != nil {
			if c.isRetriableError(err) {
				klog.V(4).InfoS("Retriable error processing ServiceAccount, continuing",
					"serviceAccount", serviceAccount.Name,
					"namespace", serviceAccount.Namespace,
					"error", err)
				continue
			}
			return err
		}
	}
	return nil
}

// processServiceAccount handles cleanup for a single ServiceAccount with retry logic
func (c *imagePullSecretCleanupController) processServiceAccount(ctx context.Context, serviceAccount *corev1.ServiceAccount) error {
	imagePullSecretName, imagePullSecret, err := c.imagePullSecretForServiceAccount(serviceAccount)
	if err != nil {
		return fmt.Errorf("unable to retrieve the managed image pull secret for the service account %q (ns=%q): %w", serviceAccount.Name, serviceAccount.Namespace, err)
	}

	if len(imagePullSecretName) == 0 {
		// no managed image pull secret reference by current service account
		return nil
	}

	if imagePullSecret != nil && imagePullSecret.CreationTimestamp.After(time.Now().Add(-10*time.Minute)) {
		// managed image pull secret was created within the last 10 minutes, skip for now to avoid fighting with OCM
		klog.V(4).InfoS("Skipping recently created image pull secret",
			"serviceAccount", serviceAccount.Name,
			"namespace", serviceAccount.Namespace,
			"secret", imagePullSecretName,
			"age", time.Since(imagePullSecret.CreationTimestamp.Time))
		return nil
	}

	if imagePullSecret != nil && c.imagePullSecretInUse(imagePullSecret) {
		// managed image pull secret is referenced by a pod, skip for now
		klog.V(4).InfoS("Skipping image pull secret in use",
			"serviceAccount", serviceAccount.Name,
			"namespace", serviceAccount.Namespace,
			"secret", imagePullSecretName)
		return nil
	}

	// Delete secrets with retry logic for better resilience
	var tokenSecret *corev1.Secret
	if imagePullSecret != nil {
		tokenSecret, err = c.tokenSecretForImagePullSecret(imagePullSecret)
		if err != nil {
			return fmt.Errorf("unable to retrieve the managed legacy service account API token secret for the managed image pull secret %q (ns=%q): %w", imagePullSecret.Name, imagePullSecret.Namespace, err)
		}
	}

	// Delete token secret with retry
	if tokenSecret != nil {
		if err := c.deleteSecretWithRetry(ctx, tokenSecret); err != nil {
			return fmt.Errorf("unable to delete the service account token secret %q (ns=%q): %w", tokenSecret.Name, tokenSecret.Namespace, err)
		}
	}

	// Delete image pull secret with retry
	if imagePullSecret != nil {
		if err := c.deleteSecretWithRetry(ctx, imagePullSecret); err != nil {
			return fmt.Errorf("unable to delete image pull secret %q (ns=%q): %w", imagePullSecret.Name, imagePullSecret.Namespace, err)
		}
	}

	// Update ServiceAccount to remove references with retry
	if len(imagePullSecretName) != 0 {
		if err := c.updateServiceAccountWithRetry(ctx, serviceAccount, imagePullSecretName); err != nil {
			return fmt.Errorf("unable to clean up references to the image pull secret %q (ns=%q) from the service account %q: %w", imagePullSecretName, serviceAccount.Namespace, serviceAccount.Name, err)
		}
	}

	klog.V(4).InfoS("Successfully cleaned up image pull secret",
		"serviceAccount", serviceAccount.Name,
		"namespace", serviceAccount.Namespace,
		"secret", imagePullSecretName)

	return nil
}

// deleteSecretWithRetry deletes a secret with exponential backoff retry for API server resilience
func (c *imagePullSecretCleanupController) deleteSecretWithRetry(ctx context.Context, secret *corev1.Secret) error {
	backoff := wait.Backoff{
		Steps:    3,
		Duration: 100 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
		Cap:      5 * time.Second,
	}

	return retry.OnError(backoff, c.isRetriableError, func() error {
		err := c.kubeClient.CoreV1().Secrets(secret.Namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if errors.IsNotFound(err) {
			// Secret already deleted, consider this success
			return nil
		}
		return err
	})
}

// updateServiceAccountWithRetry updates a ServiceAccount with exponential backoff retry
func (c *imagePullSecretCleanupController) updateServiceAccountWithRetry(ctx context.Context, serviceAccount *corev1.ServiceAccount, imagePullSecretName string) error {
	backoff := wait.Backoff{
		Steps:    3,
		Duration: 100 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
		Cap:      5 * time.Second,
	}

	return retry.OnError(backoff, func(err error) bool {
		// Always retry conflicts for ServiceAccount updates
		if errors.IsConflict(err) {
			return true
		}
		return c.isRetriableError(err)
	}, func() error {
		// Get fresh copy to avoid conflicts
		fresh, err := c.kubeClient.CoreV1().ServiceAccounts(serviceAccount.Namespace).Get(ctx, serviceAccount.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Remove references to the image pull secret
		var secretRefs []corev1.ObjectReference
		for _, secretRef := range fresh.Secrets {
			if secretRef.Name != imagePullSecretName {
				secretRefs = append(secretRefs, secretRef)
			}
		}
		fresh.Secrets = secretRefs

		var imagePullSecretRefs []corev1.LocalObjectReference
		for _, imagePullSecretRef := range fresh.ImagePullSecrets {
			if imagePullSecretRef.Name != imagePullSecretName {
				imagePullSecretRefs = append(imagePullSecretRefs, imagePullSecretRef)
			}
		}
		fresh.ImagePullSecrets = imagePullSecretRefs

		_, err = c.kubeClient.CoreV1().ServiceAccounts(fresh.Namespace).Update(ctx, fresh, metav1.UpdateOptions{})
		return err
	})
}

// isRetriableError determines if an error is worth retrying
func (c *imagePullSecretCleanupController) isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Retry on temporary network errors, rate limiting, or server errors
	if errors.IsTimeout(err) || errors.IsServerTimeout(err) ||
		errors.IsTooManyRequests(err) || errors.IsInternalError(err) ||
		errors.IsServiceUnavailable(err) {
		return true
	}

	// Retry on connection refused (the main issue on SNO)
	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "connect: connection refused") ||
		strings.Contains(err.Error(), "dial tcp") {
		return true
	}

	return false
}

func (c *imagePullSecretCleanupController) imagePullSecretInUse(imagePullSecret *corev1.Secret) bool {
	pods, err := c.podLister.Pods(imagePullSecret.Namespace).List(labels.Everything())
	if err != nil {
		runtime.HandleError(err)
		return true // play it safe
	}
	for _, pod := range pods {
		for _, imagePullSecretRef := range pod.Spec.ImagePullSecrets {
			if imagePullSecret.Name == imagePullSecretRef.Name {
				klog.V(4).InfoS("Image pull secret in use", "ns", imagePullSecret.Namespace, "secret", imagePullSecret.Name, "pod ns", pod.Namespace, "pod", pod.Name)
				return true
			}
		}
	}
	return false
}

func (c *imagePullSecretCleanupController) imagePullSecretForServiceAccount(serviceAccount *corev1.ServiceAccount) (string, *corev1.Secret, error) {
	// look in the annotation added in 4.16
	imagePullSecretName := serviceAccount.Annotations["openshift.io/internal-registry-pull-secret-ref"]

	imagePullSecretNamePrefix := naming.GetName(serviceAccount.Name, "dockercfg-", 58)
	if len(imagePullSecretName) == 0 {
		// look in the list of image pull secrets
		for _, imagePullSecretRef := range serviceAccount.ImagePullSecrets {
			if strings.HasPrefix(imagePullSecretRef.Name, imagePullSecretNamePrefix) {
				imagePullSecretName = imagePullSecretRef.Name
				break
			}
		}
	}
	if len(imagePullSecretName) == 0 {
		// look in the list of mountable secrets
		for _, secretRef := range serviceAccount.Secrets {
			if strings.HasPrefix(secretRef.Name, imagePullSecretNamePrefix) {
				imagePullSecretName = secretRef.Name
				break
			}
		}
	}
	if len(imagePullSecretName) == 0 {
		return "", nil, nil
	}
	imagePullSecret, err := c.secretLister.Secrets(serviceAccount.Namespace).Get(imagePullSecretName)
	if errors.IsNotFound(err) {
		return imagePullSecretName, nil, nil
	}
	if err != nil {
		return "", nil, err
	}
	// more confirmation that this was generated by ocm
	if _, ok := imagePullSecret.Annotations["openshift.io/internal-registry-auth-token.service-account"]; ok {
		return imagePullSecretName, imagePullSecret, nil
	}
	if _, ok := imagePullSecret.Annotations["openshift.io/token-secret.name"]; ok {
		return imagePullSecretName, imagePullSecret, nil
	}
	return "", nil, nil
}

func (c *imagePullSecretCleanupController) tokenSecretForImagePullSecret(secret *corev1.Secret) (*corev1.Secret, error) {
	tokenSecretName := secret.Annotations["openshift.io/token-secret.name"]
	if len(tokenSecretName) == 0 {
		return nil, nil
	}
	tokenSecret, err := c.secretLister.Secrets(secret.Namespace).Get(tokenSecretName)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	// more confirmation that this was generated by ocm
	value, ok := tokenSecret.Annotations["kubernetes.io/created-by"]
	if !ok || value != "openshift.io/create-dockercfg-secrets" {
		return nil, nil
	}
	return tokenSecret, err
}
