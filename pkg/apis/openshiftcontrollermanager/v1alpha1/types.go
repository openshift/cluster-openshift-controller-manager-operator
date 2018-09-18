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

	// kubeApiserverConfig holds a sparse config that the user wants for this component.  It only needs to be the overrides from the defaults
	// it will end up overlaying in the following order:
	// 1. hardcoded default
	// 2. this config
	OpenShiftControllerManagerConfig runtime.RawExtension `json:"kubeApiserverConfig" protobuf:"bytes,2,opt,name=kubeApiserverConfig"`
}

type OpenShiftControllerManagerOperatorConfigStatus struct {
	operatorsv1alpha1api.OperatorStatus `json:",inline" protobuf:"bytes,1,opt,name=operatorStatus"`
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
