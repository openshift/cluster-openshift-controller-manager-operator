package builds

import (
	"fmt"

	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
)

// ObserveBuildControllerConfig reads the cluster-wide build controller configuration as provided by the cluster admin.
func ObserveBuildControllerConfig(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {

	listers := genericListers.(configobservation.Listers)
	var errs []error
	prevObservedConfig := map[string]interface{}{}

	// first observe all the existing config values so that if we get any errors
	// we can at least return those.
	gitHTTPProxyPath := []string{"build", "buildDefaults", "gitHTTPProxy"}
	currentGitHTTPProxy, _, err := unstructured.NestedString(existingConfig, gitHTTPProxyPath...)
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if len(currentGitHTTPProxy) > 0 {
		err := unstructured.SetNestedField(prevObservedConfig, currentGitHTTPProxy, gitHTTPProxyPath...)
		if err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	gitHTTPSProxyPath := []string{"build", "buildDefaults", "gitHTTPSProxy"}
	currentGitHTTPSProxy, _, err := unstructured.NestedString(existingConfig, gitHTTPSProxyPath...)
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if len(currentGitHTTPSProxy) > 0 {
		err := unstructured.SetNestedField(prevObservedConfig, currentGitHTTPSProxy, gitHTTPSProxyPath...)
		if err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	gitNoProxyPath := []string{"build", "buildDefaults", "gitNoProxy"}
	currentGitNoProxy, _, err := unstructured.NestedString(existingConfig, gitNoProxyPath...)
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}
	if len(currentGitNoProxy) > 0 {
		err := unstructured.SetNestedField(prevObservedConfig, currentGitNoProxy, gitNoProxyPath...)
		if err != nil {
			return prevObservedConfig, append(errs, err)
		}
	}

	// now gather the cluster config and turn it into the observed config
	observedConfig := map[string]interface{}{}
	buildConfig, err := listers.BuildConfigLister.Get("cluster")
	if errors.IsNotFound(err) {
		klog.V(2).Infof("builds.config.openshift.io/cluster: not found")
		return observedConfig, errs
	}
	if err != nil {
		return prevObservedConfig, append(errs, err)
	}

	// NOTE proxies are now entirely handled by the build controller itself;
	// but we still process the other defaults/overrides cluster config for builds here

	if len(buildConfig.Spec.BuildDefaults.Env) > 0 {
		if err = configobservation.ObserveField(observedConfig, buildConfig.Spec.BuildDefaults.Env, "build.buildDefaults.env", true); err != nil {
			return nil, append(errs, fmt.Errorf("failed to observe %s: %v", "build.buildDefaults.env", err))
		}
	}
	if len(buildConfig.Spec.BuildDefaults.ImageLabels) > 0 {
		if err = configobservation.ObserveField(observedConfig, buildConfig.Spec.BuildDefaults.ImageLabels, "build.buildDefaults.imageLabels", true); err != nil {
			return nil, append(errs, fmt.Errorf("failed to observe %s: %v", "build.buildDefaults.imageLabels", err))
		}
	}

	// set build overrides
	if len(buildConfig.Spec.BuildOverrides.ImageLabels) > 0 {
		if err = configobservation.ObserveField(observedConfig, buildConfig.Spec.BuildOverrides.ImageLabels, "build.buildOverrides.imageLabels", true); err != nil {
			return nil, append(errs, fmt.Errorf("failed to observe %s: %v", "build.buildOverrides.imageLabels", err))
		}
	}
	if len(buildConfig.Spec.BuildOverrides.NodeSelector) > 0 {
		if err = configobservation.ObserveField(observedConfig, buildConfig.Spec.BuildOverrides.NodeSelector, "build.buildOverrides.nodeSelector", true); err != nil {
			return nil, append(errs, fmt.Errorf("failed to observe %s: %v", "build.buildOverrides.nodeSelector", err))
		}
	}
	if len(buildConfig.Spec.BuildOverrides.Tolerations) > 0 {
		if err = configobservation.ObserveField(observedConfig, buildConfig.Spec.BuildOverrides.Tolerations, "build.buildOverrides.tolerations", true); err != nil {
			return nil, append(errs, fmt.Errorf("failed to observe %s: %v", "build.buildOverrides.tolerations", err))
		}
	}
	return observedConfig, errs
}
