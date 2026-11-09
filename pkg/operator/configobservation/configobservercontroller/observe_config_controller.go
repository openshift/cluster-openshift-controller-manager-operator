package configobservercontroller

import (
	"fmt"

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
	tempOCMGroupsField = "tempOCMServingInfoGroupsIgnore"
)

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
				if _, err := configobserver.CurvePreferencesToTLSGroups[int64](existingConfig, []string{"servingInfo", "curvePreferences"}, []string{tempOCMGroupsField}); err != nil {
					// keep going on read error from existing config
					errs = append(errs, err)
				}

				defer func() {
					unstructured.RemoveNestedField(existingConfig, tempOCMGroupsField)
					unstructured.RemoveNestedField(observedConfig, tempOCMGroupsField)
				}()

				var observeErrs []error
				observedConfig, observeErrs = apiserver.ObserveTLSSecurityProfileWithGroupPaths(
					listers,
					recorder,
					existingConfig,
					[]string{"servingInfo", "minTLSVersion"},
					[]string{"servingInfo", "cipherSuites"},
					[]string{tempOCMGroupsField},
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
				groups, found, err := unstructured.NestedStringSlice(observedConfig, tempOCMGroupsField)
				if err == nil && found {
					if len(groups) == 0 {
						// Empty groups list - set curvePreferences to empty
						unstructured.SetNestedSlice(observedConfig, []interface{}{}, "servingInfo", "curvePreferences")
					} else if unrecognizedGroups, err := configobserver.TLSGroupsToCurvePreferences[int64](observedConfig, []string{tempOCMGroupsField}, []string{"servingInfo", "curvePreferences"}); err != nil {
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
