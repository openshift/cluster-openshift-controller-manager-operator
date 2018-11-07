package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	operatorsv1alpha1api "github.com/openshift/api/operator/v1alpha1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubeApiserverConfig provides information to configure openshift-controller-manager
type OpenShiftControllerManagerConfig struct {
	metav1.TypeMeta `json:",inline"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubeApiserverOperatorConfig provides information to configure an operator to manage openshift-controller-manager.
type OpenShiftControllerManagerOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	Spec   OpenShiftControllerManagerOperatorConfigSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status OpenShiftControllerManagerOperatorConfigStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

type OpenShiftControllerManagerOperatorConfigSpec struct {
	operatorsv1alpha1api.OperatorSpec `json:",inline" protobuf:"bytes,1,opt,name=operatorSpec"`

	// userConfig holds a sparse config that the user wants for this component.  It only needs to be the overrides from the defaults
	// it will end up overlaying in the following order:
	// 1. hardcoded default
	// 2. this config
	UserConfig runtime.RawExtension `json:"userConfig"`

	// observedConfig holds a sparse config that controller has observed from the cluster state.  It exists in spec because
	// it causes action for the operator
	ObservedConfig runtime.RawExtension `json:"observedConfig"`

	// additionalTrustedCA references the additional trusted certificate authorities that operator should attempt to configure for the build controller.
	AdditionalTrustedCA *AdditionalTrustedCA `json:"additionalTrustedCA,omitempty"`
}

type AdditionalTrustedCA struct {
	// sha1Hash contains the sha1 hash of the CA bundle data
	SHA1Hash string `json:"sha1Hash,omitempty" protobuf:"bytes,1,opt,name=sha1Hash"`

	// configMapName is the name of the ConfigMap in the openshift-config namespace containing the additional trusted CAs for the build controller.
	ConfigMapName string `json:"configMap,omitempty" protobuf:"bytes,2,opt,name=configMap"`
}

type OpenShiftControllerManagerOperatorConfigStatus struct {
	operatorsv1alpha1api.OperatorStatus `json:",inline" protobuf:"bytes,1,opt,name=operatorStatus"`

	// additionalTrustedCA references the additional trusted certificate authorities that operator configured for the build controller.
	AdditionalTrustedCA *AdditionalTrustedCA `json:"additionalTrustedCA,omitempty" protobuf:"bytes,2,opt,name=additionalTrustedCAHash"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubeApiserverOperatorConfigList is a collection of items
type OpenShiftControllerManagerOperatorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items contains the items
	Items []OpenShiftControllerManagerOperatorConfig `json:"items" protobuf:"bytes,2,rep,name=items"`
}
