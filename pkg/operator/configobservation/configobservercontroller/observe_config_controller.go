package configobservercontroller

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/api/features"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorv1informers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/configobserver/apiserver"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/builds"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/controllers"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/deployimages"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/images"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/network"
)

const (
	tempTLSGroupsField = "tempServingInfoGroupsIgnore"
)

// getFilteredOutGroups checks if any TLS groups from a Custom profile were filtered out (e.g., by FIPS)
// and returns the list of filtered groups. Returns an error if the APIServer CR cannot be retrieved (NotFound is acceptable).
func getFilteredOutGroups(listers configobserver.Listers, remainingGroups []string) ([]configv1.TLSGroup, error) {
	configListers, ok := listers.(configobservation.Listers)
	if !ok {
		panic("listers must be of type configobservation.Listers")
	}

	apiServer, err := configListers.APIServerLister_.Get("cluster")
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get apiserver/cluster: %w", err)
	}

	if apiServer == nil || apiServer.Spec.TLSSecurityProfile == nil {
		return nil, nil
	}

	if apiServer.Spec.TLSSecurityProfile.Type != configv1.TLSProfileCustomType ||
		apiServer.Spec.TLSSecurityProfile.Custom == nil ||
		len(apiServer.Spec.TLSSecurityProfile.Custom.Groups) <= len(remainingGroups) {
		return nil, nil
	}

	groupsSet := make(map[string]bool)
	for _, g := range remainingGroups {
		groupsSet[g] = true
	}

	filteredOut := []configv1.TLSGroup{}
	for _, originalGroup := range apiServer.Spec.TLSSecurityProfile.Custom.Groups {
		if !groupsSet[string(originalGroup)] {
			filteredOut = append(filteredOut, originalGroup)
		}
	}

	return filteredOut, nil
}

// observeTLSSecurityProfileWithGroupPathsFunc is the signature for observing TLS security profile with groups.
type observeTLSSecurityProfileWithGroupPathsFunc func(
	listers configobserver.Listers,
	recorder events.Recorder,
	existingConfig map[string]interface{},
	minTLSVersionPath []string,
	cipherSuitesPath []string,
	groupsPath []string,
) (map[string]interface{}, []error)

// innerObserveTLSSecurityProfileWithGroupsConversion is the internal implementation that accepts a custom
// observeTLSSecurityProfile function for testability (e.g., to mock FIPS filtering).
func innerObserveTLSSecurityProfileWithGroupsConversion(
	listers configobserver.Listers,
	recorder events.Recorder,
	existingConfig map[string]interface{},
	observeTLSSecurityProfile observeTLSSecurityProfileWithGroupPathsFunc,
) (map[string]interface{}, []error) {
	var errs []error
	var observedConfig map[string]interface{}

	if _, err := configobserver.CurvePreferencesToTLSGroups[int64](existingConfig, []string{"servingInfo", "curvePreferences"}, []string{tempTLSGroupsField}); err != nil {
		// keep going on read error from existing config
		errs = append(errs, err)
	}

	defer func() {
		unstructured.RemoveNestedField(existingConfig, tempTLSGroupsField)
		unstructured.RemoveNestedField(observedConfig, tempTLSGroupsField)
	}()

	var observeErrs []error
	observedConfig, observeErrs = observeTLSSecurityProfile(
		listers,
		recorder,
		existingConfig,
		[]string{"servingInfo", "minTLSVersion"},
		[]string{"servingInfo", "cipherSuites"},
		[]string{tempTLSGroupsField},
	)
	if len(observeErrs) > 0 {
		// accumulate errors but keep going - we still need to convert groups to curves
		errs = append(errs, observeErrs...)
	}

	// Copy existing curve preferences to observedConfig so TLSGroupsToCurvePreferences can compare against them
	existingCurvePreferences, found, err := unstructured.NestedSlice(existingConfig, "servingInfo", "curvePreferences")
	if err != nil {
		// keep going on read error from existing config
		errs = append(errs, err)
	} else if found {
		unstructured.SetNestedSlice(observedConfig, existingCurvePreferences, "servingInfo", "curvePreferences")
	}

	// Check if we need to set curvePreferences to empty
	groups, found, err := unstructured.NestedStringSlice(observedConfig, tempTLSGroupsField)
	if err == nil && found {
		// Check if this is a Custom profile and if any groups were filtered out (e.g., by FIPS)
		if filteredOut, err := getFilteredOutGroups(listers, groups); err != nil {
			errs = append(errs, err)
		} else if len(filteredOut) > 0 {
			errs = append(errs, fmt.Errorf("TLS groups filtered out from custom profile: %v", filteredOut))
		}

		if len(groups) == 0 {
			// Empty groups list - set curvePreferences to empty
			unstructured.SetNestedSlice(observedConfig, []interface{}{}, "servingInfo", "curvePreferences")
		} else if unrecognizedGroups, err := configobserver.TLSGroupsToCurvePreferences[int64](observedConfig, []string{tempTLSGroupsField}, []string{"servingInfo", "curvePreferences"}); err != nil {
			errs = append(errs, err)
		} else if len(unrecognizedGroups) > 0 {
			// If all groups are unrecognized, set curvePreferences to empty
			if len(groups) == len(unrecognizedGroups) {
				unstructured.SetNestedSlice(observedConfig, []interface{}{}, "servingInfo", "curvePreferences")
			}
			errs = append(errs, fmt.Errorf("unrecognized TLS groups: %v", unrecognizedGroups))
		}
	}
	return observedConfig, errs
}

// ObserveTLSSecurityProfileWithGroupsConversion observes TLS groups from APIServer config and converts them to curve preferences.
// It handles the full lifecycle: reading existing curves, observing new groups, checking for filtered groups and converting groups to curves.
func ObserveTLSSecurityProfileWithGroupsConversion(listers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	return innerObserveTLSSecurityProfileWithGroupsConversion(listers, recorder, existingConfig, apiserver.ObserveTLSSecurityProfileWithGroupPaths)
}

// NewConfigObserver initializes a new configuration observer.
func NewConfigObserver(
	operatorClient v1helpers.OperatorClient,
	operatorConfigInformers operatorv1informers.SharedInformerFactory,
	configInformers configinformers.SharedInformerFactory,
	kubeInformersForOperatorNamespace kubeinformers.SharedInformerFactory,
	featureGateAccessor featuregates.FeatureGateAccess,
	eventRecorder events.Recorder,
	buildEnabled bool,
) factory.Controller {
	informersSynced := []cache.InformerSynced{
		configInformers.Config().V1().Images().Informer().HasSynced,
		configInformers.Config().V1().Networks().Informer().HasSynced,
		configInformers.Config().V1().APIServers().Informer().HasSynced,
		configInformers.Config().V1().ClusterVersions().Informer().HasSynced,
		configInformers.Config().V1().ClusterOperators().Informer().HasSynced,
		kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Informer().HasSynced,
		operatorConfigInformers.Operator().V1().OpenShiftControllerManagers().Informer().HasSynced,
	}

	if buildEnabled {
		informersSynced = append(informersSynced, configInformers.Config().V1().Builds().Informer().HasSynced)
	}

	configObservationListers := configobservation.Listers{
		ImageConfigLister:     configInformers.Config().V1().Images().Lister(),
		NetworkLister:         configInformers.Config().V1().Networks().Lister(),
		FeatureGateLister_:    configInformers.Config().V1().FeatureGates().Lister(),
		APIServerLister_:      configInformers.Config().V1().APIServers().Lister(),
		ClusterVersionLister:  configInformers.Config().V1().ClusterVersions().Lister(),
		ClusterOperatorLister: configInformers.Config().V1().ClusterOperators().Lister(),
		ConfigMapLister:       kubeInformersForOperatorNamespace.Core().V1().ConfigMaps().Lister(),
		PreRunCachesSynced:    informersSynced,
	}

	observerFuncs := []configobserver.ObserveConfigFunc{
		images.ObserveInternalRegistryHostname,
		images.ObserveExternalRegistryHostnames,
		network.ObserveExternalIPAutoAssignCIDRs,
		deployimages.ObserveControllerManagerImagesConfig,
		controllers.ObserveControllers,
		featuregates.NewObserveFeatureFlagsFunc(
			sets.New[configv1.FeatureGateName]("BuildCSIVolumes"),
			nil,
			[]string{"featureGates"},
			featureGateAccessor,
		),
		func(listers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (observedConfig map[string]interface{}, errs []error) {
			featureGate, err := featureGateAccessor.CurrentFeatureGates()
			if err != nil {
				return existingConfig, append(errs, err)
			}
			if featureGate.Enabled(features.FeatureGateTLSGroupPreferences) {
				return ObserveTLSSecurityProfileWithGroupsConversion(listers, recorder, existingConfig)
			} else {
				return apiserver.ObserveTLSSecurityProfile(
					listers,
					recorder,
					existingConfig,
				)
			}
		},
	}

	if buildEnabled {
		configObservationListers.BuildConfigLister = configInformers.Config().V1().Builds().Lister()
		observerFuncs = append(observerFuncs, builds.ObserveBuildControllerConfig)
	}

	c := configobserver.NewConfigObserver(
		"openshift-controller-manager",
		operatorClient,
		eventRecorder,
		configObservationListers,
		[]factory.Informer{operatorConfigInformers.Operator().V1().OpenShiftControllerManagers().Informer()},
		observerFuncs...,
	)

	return c
}
