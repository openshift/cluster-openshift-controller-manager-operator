package operator

import (
	"context"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	workloadcontroller "github.com/openshift/library-go/pkg/operator/apiserver/controller/workload"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	"github.com/openshift/library-go/pkg/operator/staticresourcecontroller"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	configobservationcontroller "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/configobservercontroller"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/usercaobservation"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
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
	)

	opClient := &genericClient{
		informers: operatorConfigInformers,
		client:    operatorClient.OperatorV1(),
	}

	// resourceSyncer synchronizes Secrets and ConfigMaps from one namespace to another.
	// Bug 1826183: this will sync the proxy trustedCA ConfigMap to the
	// openshift-controller-manager's user-ca ConfigMap.
	resourceSyncer := resourcesynccontroller.NewResourceSyncController(
		opClient,
		kubeInformers,
		v1helpers.CachedSecretGetter(kubeClient.CoreV1(), kubeInformers),
		v1helpers.CachedConfigMapGetter(kubeClient.CoreV1(), kubeInformers),
		controllerConfig.EventRecorder,
	)

	// ConfigObserver observes the configuration state from cluster config objects and transforms
	// them into configuration used by openshift-controller-manager
	configObserver := configobservationcontroller.NewConfigObserver(
		opClient,
		operatorConfigInformers,
		configInformers,
		kubeInformers.InformersFor(util.OperatorNamespace),
		controllerConfig.EventRecorder,
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
		v311_00_assets.Asset,
		[]string{
			"v3.11.0/openshift-controller-manager/informer-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-role.yaml",
			"v3.11.0/openshift-controller-manager/leader-ingress-to-route-controller-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/leader-role.yaml",
			"v3.11.0/openshift-controller-manager/leader-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/ns.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-informer-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-informer-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-leader-role.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-leader-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-ns.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-sa.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-separate-sa-role.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-separate-sa-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-servicemonitor-role.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-servicemonitor-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-svc.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-tokenreview-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-tokenreview-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/route-controller-svc.yaml",
			"v3.11.0/openshift-controller-manager/old-leader-role.yaml",
			"v3.11.0/openshift-controller-manager/old-leader-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/separate-sa-role.yaml",
			"v3.11.0/openshift-controller-manager/separate-sa-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/sa.yaml",
			"v3.11.0/openshift-controller-manager/svc.yaml",
			"v3.11.0/openshift-controller-manager/servicemonitor-role.yaml",
			"v3.11.0/openshift-controller-manager/servicemonitor-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/buildconfigstatus-clusterrolebinding.yaml",
		},
		resourceapply.NewKubeClientHolder(kubeClient),
		opClient,
		controllerConfig.EventRecorder,
	).AddKubeInformers(kubeInformers)

	operatorConfigInformers.Start(ctx.Done())
	kubeInformers.Start(ctx.Done())
	configInformers.Start(ctx.Done())

	go staticResourceController.Run(ctx, 1)
	go operator.Run(ctx, 1)
	go resourceSyncer.Run(ctx, 1)
	go configObserver.Run(ctx, 1)
	go userCAObserver.Run(ctx, 1)
	go clusterOperatorStatus.Run(ctx, 1)

	<-ctx.Done()
	return fmt.Errorf("stopped")
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
