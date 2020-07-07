package usercaobservation

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	fakeconfig "github.com/openshift/client-go/config/clientset/versioned/fake"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"

	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

type fakeSyncer struct {
	configMaps          map[string]string
	configMapMux        sync.Mutex
	configMapSynced     chan struct{}
	configMapSyncClosed bool
}

func newFakeSyncer() *fakeSyncer {
	return &fakeSyncer{
		configMaps:      make(map[string]string),
		configMapSynced: make(chan struct{}),
	}
}

func (f *fakeSyncer) SyncConfigMap(destination, source resourcesynccontroller.ResourceLocation) error {
	f.configMapMux.Lock()
	defer f.configMapMux.Unlock()
	if f.configMapSyncClosed {
		return nil
	}
	f.configMaps[canonicalLocation(destination)] = canonicalLocation(source)
	close(f.configMapSynced)
	f.configMapSyncClosed = true
	return nil
}

func (f *fakeSyncer) SyncSecret(destination, source resourcesynccontroller.ResourceLocation) error {
	return nil
}

func canonicalLocation(r resourcesynccontroller.ResourceLocation) string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.Name)
}

func TestFindProxyCASource(t *testing.T) {
	destination := canonicalLocation(resourcesynccontroller.ResourceLocation{
		Namespace: util.TargetNamespace,
		Name:      "openshift-user-ca",
	})
	cases := []struct {
		name           string
		proxy          *configv1.Proxy
		expectedSource string
	}{
		{
			name: "default",
			proxy: &configv1.Proxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
			},
			expectedSource: "/",
		},
		{
			name:           "no found",
			expectedSource: "/",
		},
		{
			name: "has user CA",
			proxy: &configv1.Proxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: configv1.ProxySpec{
					TrustedCA: configv1.ConfigMapNameReference{
						Name: "user-ca",
					},
				},
			},
			expectedSource: "openshift-config/user-ca",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeOperatorClient := v1helpers.NewFakeOperatorClient(
				&operatorv1.OperatorSpec{
					ManagementState: operatorv1.Managed,
				},
				&operatorv1.OperatorStatus{},
				nil,
			)

			configObjects := []runtime.Object{}
			if tc.proxy != nil {
				configObjects = append(configObjects, tc.proxy)
			}
			fakeConfigClient := fakeconfig.NewSimpleClientset(configObjects...)
			configInformer := configinformers.NewSharedInformerFactory(fakeConfigClient, 1*time.Minute)
			syncer := newFakeSyncer()
			controller := NewController(fakeOperatorClient,
				configInformer,
				syncer,
				events.NewInMemoryRecorder("test"))

			ctx, ctxCancel := context.WithCancel(context.TODO())
			defer ctxCancel()
			go configInformer.Start(ctx.Done())
			go controller.Run(ctx, 1)

			select {
			case <-syncer.configMapSynced:
			case <-time.After(20 * time.Second):
				t.Fatal("timeout waiting for configMap to by synced")
			}
			actualSource, ok := syncer.configMaps[destination]
			if !ok {
				t.Errorf("Expected source for %s, found none", destination)
			}
			if tc.expectedSource != actualSource {
				t.Errorf("Expected source for %s to be %s, got %s", destination, tc.expectedSource, actualSource)
			}
		})
	}
}
