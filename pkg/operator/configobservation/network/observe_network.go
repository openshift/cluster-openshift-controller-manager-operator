package network

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/configobserver/network"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

// ObserveExternalIPAutoAssignCIDRs watches the
// config.openshift.io/v1/Network.Spec.ExternalIP.AutoAssignCIDRs field and
// configures the IngressController accordingly.
// The IngressController is a bit of a misnomer - it automatically allocates ExternalIPs
// to Services of type LoadBalancer. It is useful for ingressing traffic to bare-metal.
func ObserveExternalIPAutoAssignCIDRs(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	out := map[string]interface{}{}

	configPath := []string{"ingress", "ingressIPNetworkCIDR"}
	prevValue, ok, err := unstructured.NestedString(existingConfig, configPath...)

	// Preserve the existing value, so we can return it if we encounter an error
	if err == nil && ok {
		err = unstructured.SetNestedField(out, prevValue, configPath...)
	}
	if err != nil { // unlikely
		return out, []error{err}
	}

	cidrs, err := network.GetExternalIPAutoAssignCIDRs(listers.NetworkLister, recorder)
	if err != nil {
		return out, []error{err}
	}

	if len(cidrs) == 0 {
		err = unstructured.SetNestedField(out, "", configPath...)
	} else if len(cidrs) == 1 {
		err = unstructured.SetNestedField(out, cidrs[0], configPath...)
	} else {
		err = fmt.Errorf("error reading networks.%s/cluster Spec.ExternalIP.AutoAssignCIDRs: only one CIDR is supported", configv1.GroupName)
	}

	if err != nil {
		return out, []error{err}
	}
	return out, nil
}
