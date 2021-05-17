package framework

import (
	"testing"

	"github.com/openshift/api"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
)

var (
	encoderFactory serializer.CodecFactory
	scheme         = runtime.NewScheme()
)

func init() {
	utilruntime.Must(api.Install(scheme))
	utilruntime.Must(api.InstallKube(scheme))
	encoderFactory = serializer.NewCodecFactory(scheme)
}

// MustDecodeClusterRole decodes the YAML file into a ClusterRole, or fails the test
func MustDecodeClusterRole(t *testing.T, asset resourceapply.AssetFunc, file string) *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{}
	mustDecodeIntoObject(t, asset, file, clusterRole)
	return clusterRole
}

// MustDecodeClusterRoleBinding decodes the YAML file into a ClusterRoleBinding, or fails the test
func MustDecodeClusterRoleBinding(t *testing.T, asset resourceapply.AssetFunc, file string) *rbacv1.ClusterRoleBinding {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	mustDecodeIntoObject(t, asset, file, clusterRoleBinding)
	return clusterRoleBinding
}

func mustDecodeIntoObject(t *testing.T, asset resourceapply.AssetFunc, file string, object runtime.Object) {
	rawYAML, err := asset(file)
	if err != nil {
		t.Fatalf("failed to decode asset %s: %v", file, err)
	}

	err = runtime.DecodeInto(encoderFactory.UniversalDeserializer(), rawYAML, object)
	if err != nil {
		t.Fatalf("failed to decode object: %v", err)
	}
}
