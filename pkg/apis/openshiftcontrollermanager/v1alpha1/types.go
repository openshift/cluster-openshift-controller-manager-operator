package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorsv1 "github.com/openshift/api/operator/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenShiftControllerManagerConfig provides information to configure openshift-controller-manager
type OpenShiftControllerManagerConfig struct {
	metav1.TypeMeta `json:",inline"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenShiftControllerManagerOperatorConfig provides information to configure an operator to manage openshift-controller-manager.
type OpenShiftControllerManagerOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	Spec   OpenShiftControllerManagerOperatorConfigSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status OpenShiftControllerManagerOperatorConfigStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

type OpenShiftControllerManagerOperatorConfigSpec struct {
	operatorsv1.OperatorSpec `json:",inline" protobuf:"bytes,1,opt,name=operatorSpec"`
}

type OpenShiftControllerManagerOperatorConfigStatus struct {
	operatorsv1.OperatorStatus `json:",inline" protobuf:"bytes,1,opt,name=operatorStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenShiftControllerManagerOperatorConfigList is a collection of items
type OpenShiftControllerManagerOperatorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items contains the items
	Items []OpenShiftControllerManagerOperatorConfig `json:"items" protobuf:"bytes,2,rep,name=items"`
}
