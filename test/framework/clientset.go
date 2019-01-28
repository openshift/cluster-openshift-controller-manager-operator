package framework

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	clientopenshiftcontrollermanagerv1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned/typed/openshiftcontrollermanager/v1"
)

// Clientset is a set of Kubernetes clients.
type Clientset struct {
	clientcorev1.CoreV1Interface
	clientappsv1.AppsV1Interface
	clientconfigv1.ConfigV1Interface
	clientopenshiftcontrollermanagerv1.OpenshiftcontrollermanagerV1Interface
}

// NewClientset creates a set of Kubernetes clients. The default kubeconfig is
// used if not provided.
func NewClientset(kubeconfig *restclient.Config) (clientset *Clientset, err error) {
	if kubeconfig == nil {
		kubeconfig, err = getConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to get kubeconfig: %s", err)
		}
	}

	clientset = &Clientset{}
	clientset.CoreV1Interface, err = clientcorev1.NewForConfig(kubeconfig)
	if err != nil {
		return
	}
	clientset.AppsV1Interface, err = clientappsv1.NewForConfig(kubeconfig)
	if err != nil {
		return
	}
	clientset.ConfigV1Interface, err = clientconfigv1.NewForConfig(kubeconfig)
	if err != nil {
		return
	}
	clientset.OpenshiftcontrollermanagerV1Interface, err = clientopenshiftcontrollermanagerv1.NewForConfig(kubeconfig)
	if err != nil {
		return
	}
	return
}

// MustNewClientset is like NewClienset but aborts the test if clienset cannot
// be constructed.
func MustNewClientset(t *testing.T, kubeconfig *restclient.Config) *Clientset {
	clientset, err := NewClientset(kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	return clientset
}

// getConfig creates a *rest.Config for talking to a Kubernetes apiserver.
// Otherwise will assume running in cluster and use the cluster provided kubeconfig.
//
// Config precedence
//
// * KUBECONFIG environment variable pointing at a file
//
// * In-cluster config if running in cluster
//
// * $HOME/.kube/config if exists
func getConfig() (*restclient.Config, error) {
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := restclient.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}
