package operator

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/cluster-openshift-controller-manager-operator/bindata"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/internalimageregistry"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	workloadcontroller "github.com/openshift/library-go/pkg/operator/apiserver/controller/workload"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/loglevel"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	"github.com/openshift/library-go/pkg/operator/staticresourcecontroller"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	configobservationcontroller "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/configobservercontroller"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/usercaobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

func RunOperator(ctx context.Context, controllerConfig *controllercmd.ControllerContext) error {
	// Increase QPS and burst to avoid client-side rate limits when reconciling RBAC API objects.
	// See TODO below for the StaticResourceController
	highRateLimitProtoKubeConfig := rest.CopyConfig(controllerConfig.ProtoKubeConfig)
	if highRateLimitProtoKubeConfig.QPS < 50 {
		highRateLimitProtoKubeConfig.QPS = 50
	}
	if highRateLimitProtoKubeConfig.Burst < 100 {
		highRateLimitProtoKubeConfig.Burst = 100
	}
	kubeClient, err := kubernetes.NewForConfig(highRateLimitProtoKubeConfig)
	if err != nil {
		return err
	}

	operatorClient, err := operatorclient.NewForConfig(controllerConfig.KubeConfig)
	if err != nil {
		return err
	}
	configClient, err := configclient.NewForConfig(controllerConfig.KubeConfig)
	if err != nil {
		return err
	}

	// Create kube informers for namespaces that the operator reconciles content from or to.
	// The empty string "" adds informers for cluster-scoped resources.
	kubeInformers := v1helpers.NewKubeInformersForNamespaces(kubeClient,
		"",
		util.TargetNamespace,
		util.RouteControllerTargetNamespace,
		util.OperatorNamespace,
		util.UserSpecifiedGlobalConfigNamespace,
		util.InfraNamespace,
		metav1.NamespaceSystem,
	)
	operatorConfigInformers := operatorinformers.NewSharedInformerFactory(operatorClient, 10*time.Minute)
	configInformers := configinformers.NewSharedInformerFactory(configClient, 10*time.Minute)

	// OpenShiftControlllerManagerOperator reconciles the state of the openshift-controller-manager
	// DaemonSet and associated ConfigMaps.
	operator := NewOpenShiftControllerManagerOperator(
		os.Getenv("IMAGE"),
		os.Getenv("ROUTE_CONTROLLER_MANAGER_IMAGE"),
		operatorConfigInformers.Operator().V1().OpenShiftControllerManagers(),
		configInformers.Config().V1().Proxies(),
		kubeInformers,
		operatorClient.OperatorV1(),
		workloadcontroller.CountNodesFuncWrapper(kubeInformers.InformersFor("").Core().V1().Nodes().Lister()),
		workloadcontroller.EnsureAtMostOnePodPerNode,
		kubeClient,
		controllerConfig.EventRecorder,
		configInformers.Config().V1().ClusterVersions().Lister(),
	)

	opClient := &genericClient{
		clock:     clock.RealClock{},
		informers: operatorConfigInformers,
		client:    operatorClient.OperatorV1(),
	}

	desiredVersion := status.VersionForOperatorFromEnv()
	missingVersion := "0.0.1-snapshot"

	// By default, this will exit(0) the process if the featuregates ever change to a different set of values.
	featureGateAccessor := featuregates.NewFeatureGateAccess(
		desiredVersion, missingVersion,
		configInformers.Config().V1().ClusterVersions(), configInformers.Config().V1().FeatureGates(),
		controllerConfig.EventRecorder,
	)
	go featureGateAccessor.Run(ctx)
	go configInformers.Start(ctx.Done())

	select {
	case <-featureGateAccessor.InitialFeatureGatesObserved():
		featureGates, _ := featureGateAccessor.CurrentFeatureGates()
		klog.Infof("FeatureGates initialized: knownFeatureGates=%v", featureGates.KnownFeatures())
	case <-time.After(1 * time.Minute):
		klog.Errorf("timed out waiting for FeatureGate detection")
		return fmt.Errorf("timed out waiting for FeatureGate detection")
	}

	// resourceSyncer synchronizes Secrets and ConfigMaps from one namespace to another.
	// Bug 1826183: this will sync the proxy trustedCA ConfigMap to the
	// openshift-controller-manager's user-ca ConfigMap.
	resourceSyncer := resourcesynccontroller.NewResourceSyncController(
		"openshift-controller-manager",
		opClient,
		kubeInformers,
		v1helpers.CachedSecretGetter(kubeClient.CoreV1(), kubeInformers),
		v1helpers.CachedConfigMapGetter(kubeClient.CoreV1(), kubeInformers),
		controllerConfig.EventRecorder,
	)

	if !cache.WaitForCacheSync(ctx.Done(), configInformers.Config().V1().ClusterVersions().Informer().HasSynced) {
		klog.Errorf("timed out waiting for configInformers ClusterVersions")
		return fmt.Errorf("timed out waiting for configInformers ClusterVersions")
	}

	buildCapabilityEnabled := false
	cv, err := configInformers.Config().V1().ClusterVersions().Lister().Get("version")
	if err != nil {
		return err
	}

	for _, capability := range cv.Status.Capabilities.EnabledCapabilities {
		if capability == configv1.ClusterVersionCapabilityBuild {
			buildCapabilityEnabled = true
			break
		}
	}

	// ConfigObserver observes the configuration state from cluster config objects and transforms
	// them into configuration used by openshift-controller-manager
	configObserver := configobservationcontroller.NewConfigObserver(
		opClient,
		operatorConfigInformers,
		configInformers,
		kubeInformers.InformersFor(util.OperatorNamespace),
		featureGateAccessor,
		controllerConfig.EventRecorder,
		buildCapabilityEnabled,
	)

	// userCAObserver watches the cluster proxy config and updates the resourceSyncer.
	userCAObserver := usercaobservation.NewController(
		opClient,
		configInformers,
		resourceSyncer,
		controllerConfig.EventRecorder,
	)

	versionGetter := &versionGetter{
		openshiftControllerManagers: operatorClient.OperatorV1().OpenShiftControllerManagers(),
		version:                     os.Getenv("RELEASE_VERSION"),
	}

	// ClusterOperatorStatusController aggregates the conditions in our openshiftcontrollermanager
	// object to the corresponding ClusterOperator object.
	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		util.ClusterOperatorName,
		[]configv1.ObjectReference{
			{Group: "operator.openshift.io", Resource: "openshiftcontrollermanagers", Name: "cluster"},
			{Resource: "namespaces", Name: util.UserSpecifiedGlobalConfigNamespace},
			{Resource: "namespaces", Name: util.MachineSpecifiedGlobalConfigNamespace},
			{Resource: "namespaces", Name: util.OperatorNamespace},
			{Resource: "namespaces", Name: util.TargetNamespace},
			{Resource: "namespaces", Name: util.RouteControllerTargetNamespace},
		},
		configClient.ConfigV1(),
		configInformers.Config().V1().ClusterOperators(),
		opClient,
		versionGetter,
		controllerConfig.EventRecorder,
	)

	// StaticResourceController uses library-go's resourceapply package to reconcile a set of YAML
	// manifests against a cluster.
	// TODO: enhance resourceapply to use listers for RBAC APIs.
	staticResourceController := staticresourcecontroller.NewStaticResourceController(
		"OpenshiftControllerManagerStaticResources",
		bindata.Asset,
		[]string{
			"assets/openshift-controller-manager/informer-clusterrole.yaml",
			"assets/openshift-controller-manager/informer-clusterrolebinding.yaml",
			"assets/openshift-controller-manager/tokenreview-clusterrole.yaml",
			"assets/openshift-controller-manager/tokenreview-clusterrolebinding.yaml",
			"assets/openshift-controller-manager/leader-role.yaml",
			"assets/openshift-controller-manager/leader-rolebinding.yaml",
			"assets/openshift-controller-manager/ns.yaml",
			"assets/openshift-controller-manager/route-controller-manager-clusterrole.yaml",
			"assets/openshift-controller-manager/route-controller-manager-clusterrolebinding.yaml",
			"assets/openshift-controller-manager/route-controller-manager-leader-role.yaml",
			"assets/openshift-controller-manager/route-controller-manager-leader-rolebinding.yaml",
			"assets/openshift-controller-manager/route-controller-manager-ns.yaml",
			"assets/openshift-controller-manager/route-controller-manager-sa.yaml",
			"assets/openshift-controller-manager/route-controller-manager-separate-sa-role.yaml",
			"assets/openshift-controller-manager/route-controller-manager-separate-sa-rolebinding.yaml",
			"assets/openshift-controller-manager/route-controller-manager-servicemonitor-role.yaml",
			"assets/openshift-controller-manager/route-controller-manager-servicemonitor-rolebinding.yaml",
			"assets/openshift-controller-manager/route-controller-manager-svc.yaml",
			"assets/openshift-controller-manager/route-controller-manager-tokenreview-clusterrole.yaml",
			"assets/openshift-controller-manager/route-controller-manager-tokenreview-clusterrolebinding.yaml",
			"assets/openshift-controller-manager/route-controller-manager-svc.yaml",
			"assets/openshift-controller-manager/route-controller-manager-ingress-to-route-controller-clusterrole.yaml",
			"assets/openshift-controller-manager/route-controller-manager-ingress-to-route-controller-clusterrolebinding.yaml",
			"assets/openshift-controller-manager/old-leader-role.yaml",
			"assets/openshift-controller-manager/old-leader-rolebinding.yaml",
			"assets/openshift-controller-manager/separate-sa-role.yaml",
			"assets/openshift-controller-manager/separate-sa-rolebinding.yaml",
			"assets/openshift-controller-manager/sa.yaml",
			"assets/openshift-controller-manager/svc.yaml",
			"assets/openshift-controller-manager/servicemonitor-role.yaml",
			"assets/openshift-controller-manager/servicemonitor-rolebinding.yaml",
			"assets/openshift-controller-manager/buildconfigstatus-clusterrole.yaml",
			"assets/openshift-controller-manager/buildconfigstatus-clusterrolebinding.yaml",
			"assets/openshift-controller-manager/deployer-clusterrole.yaml",
			"assets/openshift-controller-manager/deployer-clusterrolebinding.yaml",
			"assets/openshift-controller-manager/image-trigger-controller-clusterrole.yaml",
			"assets/openshift-controller-manager/image-trigger-controller-clusterrolebinding.yaml",
		},
		resourceapply.NewKubeClientHolder(kubeClient),
		opClient,
		controllerConfig.EventRecorder,
	).WithConditionalResources(
		bindata.Asset,
		[]string{
			// TODO: remove all of these ingress-to-route leader-election entries and files in 4.14
			"assets/openshift-controller-manager/leader-ingress-to-route-controller-role.yaml",
			"assets/openshift-controller-manager/leader-ingress-to-route-controller-rolebinding.yaml",
		},
		func() bool {
			return false
		},
		func() bool {
			return true
		},
	).AddKubeInformers(kubeInformers)

	logLevelController := loglevel.NewClusterOperatorLoggingController(opClient, controllerConfig.EventRecorder)

	imagePullSecretCleanupController := internalimageregistry.NewImagePullSecretCleanupController(
		kubeClient,
		kubeInformers,
		configInformers,
		controllerConfig.EventRecorder,
	)

	ensureDaemonSetCleanup(ctx, kubeClient, controllerConfig.EventRecorder)

	operatorConfigInformers.Start(ctx.Done())
	kubeInformers.Start(ctx.Done())
	configInformers.Start(ctx.Done())

	go staticResourceController.Run(ctx, 1)
	go operator.Run(ctx, 1)
	go resourceSyncer.Run(ctx, 1)
	go configObserver.Run(ctx, 1)
	go userCAObserver.Run(ctx, 1)
	go clusterOperatorStatus.Run(ctx, 1)
	go logLevelController.Run(ctx, 1)
	go imagePullSecretCleanupController.Run(ctx, 1)

	capabilityChangedCh := make(chan struct{})
	if !buildCapabilityEnabled {
		// check capability periodically and close chan and return
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					cv, err := configInformers.Config().V1().ClusterVersions().Lister().Get("version")
					if err != nil {
						klog.Errorf("capability checker error %v", err)
						continue
					}

					for _, capability := range cv.Status.Capabilities.EnabledCapabilities {
						if capability == configv1.ClusterVersionCapabilityBuild {
							close(capabilityChangedCh)
							return
						}
					}
				}
			}
		}()
	}

	select {
	case <-capabilityChangedCh:
		return fmt.Errorf("capability is enabled, stopping")
	case <-ctx.Done():
		return fmt.Errorf("stopped")
	}
}

type versionGetter struct {
	openshiftControllerManagers operatorclientv1.OpenShiftControllerManagerInterface
	version                     string
}

func (v *versionGetter) SetVersion(operandName, version string) {
	// this versionGetter impl always gets the current version dynamically from operator config object status.
}

func (v *versionGetter) GetVersions() map[string]string {
	co, err := v.openshiftControllerManagers.Get(context.TODO(), "cluster", metav1.GetOptions{})
	if co == nil || err != nil {
		return map[string]string{}
	}
	if len(co.Status.Version) > 0 {
		return map[string]string{"operator": co.Status.Version}
	}
	return map[string]string{}
}

func (v *versionGetter) VersionChangedChannel() <-chan struct{} {
	// this versionGetter never notifies of a version change, getVersion always returns the new version.
	return make(chan struct{})
}

// ensureDaemonSetCleanup continually ensures the removal of the daemonset
// used to manage controller-manager pods in releases prior to 4.12. The daemonset is
// removed once the deployment now managing controller-manager pods reports at least
// one pod available.
// TODO: remove this function in later releases
func ensureDaemonSetCleanup(ctx context.Context, kubeClient *kubernetes.Clientset, eventRecorder events.Recorder) {
	// daemonset and deployment both use the same name
	resourceName := "controller-manager"

	dsClient := kubeClient.AppsV1().DaemonSets(util.TargetNamespace)
	deployClient := kubeClient.AppsV1().Deployments(util.TargetNamespace)

	go wait.PollImmediateUntilWithContext(ctx, time.Minute, func(_ context.Context) (done bool, err error) {
		// This function isn't expected to take long enough to suggest
		// checking that the context is done. The wait method will do that
		// checking.

		// Check whether the legacy daemonset exists and is not marked for deletion
		ds, err := dsClient.Get(ctx, resourceName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			// Done - daemonset does not exist
			return true, nil
		}
		if err != nil {
			klog.Warningf("Error retrieving legacy daemonset: %v", err)
			return false, nil
		}
		if ds.ObjectMeta.DeletionTimestamp != nil {
			// Done - daemonset has been marked for deletion
			return true, nil
		}

		// Check that the deployment managing the controller-manager pods has at last one available replica
		deploy, err := deployClient.Get(ctx, resourceName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			// No available replicas if the deployment doesn't exist
			return false, nil
		}
		if err != nil {
			klog.Warningf("Error retrieving the deployment that manages controller-manager pods: %v", err)
			return false, nil
		}
		if deploy.Status.AvailableReplicas == 0 {
			eventRecorder.Warning("LegacyDaemonSetCleanup", "the deployment replacing the daemonset does not have available replicas yet")
			return false, nil
		}

		// Safe to remove legacy daemonset since the deployment has at least one available replica
		err = dsClient.Delete(ctx, resourceName, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			klog.Warningf("Failed to delete legacy daemonset: %v", err)
			return false, nil
		}
		eventRecorder.Event("LegacyDaemonSetCleanup", "legacy daemonset has been removed")

		return false, nil
	})
}
