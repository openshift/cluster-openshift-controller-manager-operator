package internalimageregistry

import (
	"context"
	errs "errors"
	"fmt"
	"net/http"
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
	"k8s.io/client-go/kubernetes"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
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
}

func NewImagePullSecretCleanupController(kubeClient *kubernetes.Clientset, informers v1helpers.KubeInformersForNamespaces, configInformers externalversions.SharedInformerFactory, recorder events.Recorder) *imagePullSecretCleanupController {
	c := &imagePullSecretCleanupController{
		kubeClient:            kubeClient,
		serviceAccountLister:  informers.InformersFor(metav1.NamespaceAll).Core().V1().ServiceAccounts().Lister(),
		secretLister:          informers.InformersFor(metav1.NamespaceAll).Core().V1().Secrets().Lister(),
		podLister:             informers.InformersFor(metav1.NamespaceAll).Core().V1().Pods().Lister(),
		clusterVersionLister:  configInformers.Config().V1().ClusterVersions().Lister(),
		clusterOperatorLister: configInformers.Config().V1().ClusterOperators().Lister(),
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
	// cleanup service accounts
	serviceAccounts, err := c.serviceAccountLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("unable to list ServiceAccounts: %w", err)
	}
	for _, serviceAccount := range serviceAccounts {
		imagePullSecretName, imagePullSecret, err := c.imagePullSecretForServiceAccount(serviceAccount)
		if err != nil {
			return fmt.Errorf("unable to retrieve the managed image pull secret for the service account %q (ns=%q): %w", serviceAccount.Name, serviceAccount.Namespace, err)
		}

		if len(imagePullSecretName) == 0 {
			// no managed image pull secret reference by current service account
			continue
		}

		if imagePullSecret != nil && imagePullSecret.CreationTimestamp.After(time.Now().Add(-10*time.Minute)) {
			// managed image pull secret was created within the last 10 minutes, skip for now to avoid fighting with OCM
			continue
		}

		if imagePullSecret != nil && c.imagePullSecretInUse(imagePullSecret) {
			// managed image pull secret is referenced by a pod, skip for now
			continue
		}

		var tokenSecret *corev1.Secret
		if imagePullSecret != nil {
			tokenSecret, err = c.tokenSecretForImagePullSecret(imagePullSecret)
			if err != nil {
				return fmt.Errorf("unable to retrive the managed legacy service account API token secret for the managed image pull secret %q (ns=%q): %w", imagePullSecret.Name, imagePullSecret.Namespace, err)
			}
		}

		// if there is a corresponding legacy service account API token, delete it
		if tokenSecret != nil {
			err = c.kubeClient.CoreV1().Secrets(tokenSecret.Namespace).Delete(ctx, tokenSecret.Name, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("unable to delete the service account token secret %q (ns=%q): %w", tokenSecret.Name, tokenSecret.Namespace, err)
			}
		}

		// delete the image pull secret
		if imagePullSecret != nil {
			err = c.kubeClient.CoreV1().Secrets(imagePullSecret.Namespace).Delete(ctx, imagePullSecret.Name, metav1.DeleteOptions{})
			// the image pull secret might of been deleted via cascade if its corresponding API token was deleted, so ignore NotFound error
			if err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("unable to delete image pull secret %q (ns=%q): %w", imagePullSecret.Name, imagePullSecret.Namespace, err)
			}
		}

		// cleanup the refs in the service account
		if len(imagePullSecretName) != 0 {
			var secretRefs []corev1.ObjectReference
			for _, secretRef := range serviceAccount.Secrets {
				if secretRef.Name != imagePullSecretName {
					secretRefs = append(secretRefs, secretRef)
				}
			}
			serviceAccount.Secrets = secretRefs

			var imagePullSecretRefs []corev1.LocalObjectReference = []corev1.LocalObjectReference{}
			for _, imagePullSecretRef := range serviceAccount.ImagePullSecrets {
				if imagePullSecretRef.Name != imagePullSecretName {
					imagePullSecretRefs = append(imagePullSecretRefs, imagePullSecretRef)
				}
			}
			serviceAccount.ImagePullSecrets = imagePullSecretRefs

			_, err = c.kubeClient.CoreV1().ServiceAccounts(serviceAccount.Namespace).Update(ctx, serviceAccount, metav1.UpdateOptions{})
			if err != nil {
				var statusErr *errors.StatusError
				if errs.As(err, &statusErr) && statusErr.Status().Code == http.StatusConflict {
					return factory.SyntheticRequeueError
				}
				return fmt.Errorf("unable to clean up references to the image pull secret %q (ns=%q) from the service accout %q: %w", imagePullSecretName, serviceAccount.Namespace, serviceAccount.Name, err)
			}
		}
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
	return nil
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
