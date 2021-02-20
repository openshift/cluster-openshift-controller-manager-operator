package images

import (
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

func ObserveInternalRegistryHostname(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	var errs []error
	prevObservedConfig := map[string]interface{}{}

	// first observe all the existing config values so that if we get any errors
	// we can at least return those.
	internalRegistryHostnamePath := []string{"dockerPullSecret", "internalRegistryHostname"}
	currentInternalRegistryHostname, _, err := unstructured.NestedString(existingConfig, internalRegistryHostnamePath...)
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if len(currentInternalRegistryHostname) > 0 {
		err := unstructured.SetNestedField(prevObservedConfig, currentInternalRegistryHostname, internalRegistryHostnamePath...)
		if err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	// now gather the cluster config and turn it into the observed config
	observedConfig := map[string]interface{}{}
	configImage, err := listers.ImageConfigLister.Get("cluster")
	if errors.IsNotFound(err) {
		klog.V(2).Infof("images.config.openshift.io/cluster: not found")
		return observedConfig, errs
	}
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}

	internalRegistryHostName := configImage.Status.InternalRegistryHostname
	if len(internalRegistryHostName) > 0 {
		err = unstructured.SetNestedField(observedConfig, internalRegistryHostName, internalRegistryHostnamePath...)
		if err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	return observedConfig, errs
}

// ObserveExternalRegistryHostnames observers information about registry external URLs,
// aka Routes. It retrieves this information from cluster's Image config.
func ObserveExternalRegistryHostnames(
	genericListers configobserver.Listers,
	recorder events.Recorder,
	existingConfig map[string]interface{},
) (map[string]interface{}, []error) {
	var errs []error
	prevObservedConfig := map[string]interface{}{}
	listers := genericListers.(configobservation.Listers)

	// first observe all the existing config values so that if we get any errors
	// we can at least return those.
	cfgpath := []string{"dockerPullSecret", "registryURLs"}
	currentURLs, _, err := unstructured.NestedStringSlice(existingConfig, cfgpath...)
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if len(currentURLs) > 0 {
		if err := unstructured.SetNestedStringSlice(
			prevObservedConfig, currentURLs, cfgpath...,
		); err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	observedConfig := map[string]interface{}{}
	configImage, err := listers.ImageConfigLister.Get("cluster")
	if errors.IsNotFound(err) {
		klog.V(2).Infof("images.config.openshift.io/cluster: not found")
		return observedConfig, errs
	} else if err != nil {
		return prevObservedConfig, append(errs, err)
	}

	if len(configImage.Status.ExternalRegistryHostnames) == 0 {
		return observedConfig, errs
	}

	if err := unstructured.SetNestedStringSlice(
		observedConfig, configImage.Status.ExternalRegistryHostnames, cfgpath...,
	); err != nil {
		return prevObservedConfig, append(errs, err)
	}
	return observedConfig, errs
}
