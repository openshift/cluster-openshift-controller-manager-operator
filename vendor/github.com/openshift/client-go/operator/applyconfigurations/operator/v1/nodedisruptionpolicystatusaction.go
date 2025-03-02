// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

import (
	operatorv1 "github.com/openshift/api/operator/v1"
)

// NodeDisruptionPolicyStatusActionApplyConfiguration represents a declarative configuration of the NodeDisruptionPolicyStatusAction type for use
// with apply.
type NodeDisruptionPolicyStatusActionApplyConfiguration struct {
	Type    *operatorv1.NodeDisruptionPolicyStatusActionType `json:"type,omitempty"`
	Reload  *ReloadServiceApplyConfiguration                 `json:"reload,omitempty"`
	Restart *RestartServiceApplyConfiguration                `json:"restart,omitempty"`
}

// NodeDisruptionPolicyStatusActionApplyConfiguration constructs a declarative configuration of the NodeDisruptionPolicyStatusAction type for use with
// apply.
func NodeDisruptionPolicyStatusAction() *NodeDisruptionPolicyStatusActionApplyConfiguration {
	return &NodeDisruptionPolicyStatusActionApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *NodeDisruptionPolicyStatusActionApplyConfiguration) WithType(value operatorv1.NodeDisruptionPolicyStatusActionType) *NodeDisruptionPolicyStatusActionApplyConfiguration {
	b.Type = &value
	return b
}

// WithReload sets the Reload field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Reload field is set to the value of the last call.
func (b *NodeDisruptionPolicyStatusActionApplyConfiguration) WithReload(value *ReloadServiceApplyConfiguration) *NodeDisruptionPolicyStatusActionApplyConfiguration {
	b.Reload = value
	return b
}

// WithRestart sets the Restart field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Restart field is set to the value of the last call.
func (b *NodeDisruptionPolicyStatusActionApplyConfiguration) WithRestart(value *RestartServiceApplyConfiguration) *NodeDisruptionPolicyStatusActionApplyConfiguration {
	b.Restart = value
	return b
}
