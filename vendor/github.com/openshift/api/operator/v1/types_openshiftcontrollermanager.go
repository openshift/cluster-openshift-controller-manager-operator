package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenShiftControllerManager provides information to configure an operator to manage openshift-controller-manager.
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
type OpenShiftControllerManager struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	// +kubebuilder:validation:Required
	// +required
	Spec OpenShiftControllerManagerSpec `json:"spec"`
	// +optional
	Status OpenShiftControllerManagerStatus `json:"status"`
}

type OpenShiftControllerManagerSpec struct {
	OperatorSpec `json:",inline"`

	// imageRegistryAuthTokenType directs the openshift-controller-manager to use either a
	// legacy,(unbound, long-lived) service acccount tokens or a bound service account
	// token when generating image pull secrets for the integrated image registry.
	// +kubebuilder:default=Legacy
	// +kubebuilder:validation:Enum=Legacy;Bound
	// +optional
	ImageRegistryAuthTokenType ServiceAccountTokenType `json:"imageRegistryAuthTokenType,omitempty"`
}

type ServiceAccountTokenType string

const (
	ServiceAccountLegacyTokenType ServiceAccountTokenType = "Legacy"
	ServiceAccountBoundTokenType  ServiceAccountTokenType = "Bound"
)

type OpenShiftControllerManagerStatus struct {
	OperatorStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenShiftControllerManagerList is a collection of items
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
type OpenShiftControllerManagerList struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard list's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items contains the items
	Items []OpenShiftControllerManager `json:"items"`
}
