package workload

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/operator/events"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestDeploymentWithProxy(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "controller-manager",
				Namespace:  "openshift-controller-manager",
				Generation: 2,
			},
		},
	)
	dsClient := kubeClient.AppsV1()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	proxyConfig := &configv1.Proxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: configv1.ProxySpec{
			NoProxy:    "no-proxy",
			HTTPProxy:  "http://my-proxy",
			HTTPSProxy: "https://my-proxy",
		},
		Status: configv1.ProxyStatus{
			NoProxy:    "no-proxy",
			HTTPProxy:  "http://my-proxy",
			HTTPSProxy: "https://my-proxy",
		},
	}
	indexer.Add(proxyConfig)
	proxyLister := configlistersv1.NewProxyLister(indexer)
	recorder := events.NewInMemoryRecorder("")
	operatorConfig := &operatorv1.OpenShiftControllerManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "cluster",
			Generation: 2,
		},
		Spec: operatorv1.OpenShiftControllerManagerSpec{
			OperatorSpec: operatorv1.OperatorSpec{},
		},
		Status: operatorv1.OpenShiftControllerManagerStatus{
			OperatorStatus: operatorv1.OperatorStatus{
				ObservedGeneration: 2,
			},
		},
	}

	ds, rcBool, err := manageOpenShiftControllerManagerDeployment_v311_00_to_latest(dsClient, recorder, operatorConfig, "my.co/repo/img:latest", operatorConfig.Status.Generations, false, proxyLister)

	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if !rcBool {
		t.Fatal("apply daemon set does not think a changes was made")
	}

	if ds == nil {
		t.Fatalf("nil daemonset returned")
	}

	foundNoProxy := false
	foundHTTPProxy := false
	foundHTTPSProxy := false
	for _, c := range ds.Spec.Template.Spec.Containers {
		for _, e := range c.Env {
			switch e.Name {
			case "NO_PROXY":
				if e.Value == proxyConfig.Status.NoProxy {
					foundNoProxy = true
				}
			case "HTTP_PROXY":
				if e.Value == proxyConfig.Status.HTTPProxy {
					foundHTTPProxy = true
				}
			case "HTTPS_PROXY":
				if e.Value == proxyConfig.Status.HTTPSProxy {
					foundHTTPSProxy = true
				}
			}
		}
	}

	if !foundNoProxy {
		t.Fatalf("NO_PROXY not found: %#v", ds.Spec.Template.Spec.Containers)
	}
	if !foundHTTPProxy {
		t.Fatalf("HTTP_PROXY not found: %#v", ds.Spec.Template.Spec.Containers)
	}
	if !foundHTTPSProxy {
		t.Fatalf("HTTPS_PROXY not found: %#v", ds.Spec.Template.Spec.Containers)
	}
}
