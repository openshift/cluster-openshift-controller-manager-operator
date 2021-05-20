package operator

import (
	"context"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
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
	kubeClient, err := kubernetes.NewForConfig(controllerConfig.ProtoKubeConfig)
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

	// Empty namespace provides informers for cluster-scoped resources
	kubeInformers := v1helpers.NewKubeInformersForNamespaces(kubeClient,
		"",
		util.TargetNamespace,
		util.OperatorNamespace,
		util.UserSpecifiedGlobalConfigNamespace,
		util.InfraNamespace)
	operatorConfigInformers := operatorinformers.NewSharedInformerFactory(operatorClient, 10*time.Minute)
	configInformers := configinformers.NewSharedInformerFactory(configClient, 10*time.Minute)

	operator := NewOpenShiftControllerManagerOperator(
		os.Getenv("IMAGE"),
		operatorConfigInformers.Operator().V1().OpenShiftControllerManagers(),
		configInformers.Config().V1().Proxies(),
		kubeInformers.InformersFor(util.TargetNamespace),
		operatorClient.OperatorV1(),
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
	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		util.ClusterOperatorName,
		[]configv1.ObjectReference{
			{Group: "operator.openshift.io", Resource: "openshiftcontrollermanagers", Name: "cluster"},
			{Resource: "namespaces", Name: util.UserSpecifiedGlobalConfigNamespace},
			{Resource: "namespaces", Name: util.MachineSpecifiedGlobalConfigNamespace},
			{Resource: "namespaces", Name: util.OperatorNamespace},
			{Resource: "namespaces", Name: util.TargetNamespace},
		},
		configClient.ConfigV1(),
		configInformers.Config().V1().ClusterOperators(),
		opClient,
		versionGetter,
		controllerConfig.EventRecorder,
	)

	staticResourceController := staticresourcecontroller.NewStaticResourceController(
		"OpenshiftControllerManagerStaticResources",
		v311_00_assets.Asset,
		[]string{
			"v3.11.0/openshift-controller-manager/informer-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/informer-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/tokenreview-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/tokenreview-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/leader-role.yaml",
			"v3.11.0/openshift-controller-manager/leader-rolebinding.yaml",
			"v3.11.0/openshift-controller-manager/ns.yaml",
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
			"v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/serviceaccount-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/build-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/build-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/build-config-change-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/deployer-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/deployer-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/deploymentconfig-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/template-instance-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding-admin.yaml",
			"v3.11.0/openshift-controller-manager/template-instance-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding-admin.yaml",
			"v3.11.0/openshift-controller-manager/template-instance-finalizer-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/origin-namespace-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/serviceaccount-pull-secrets-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/image-trigger-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/image-import-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/image-import-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/unidling-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/unidling-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/service-ingress-ip-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrole.yaml",
			"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding.yaml",
			"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-deployer.yaml",
			"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-builder.yaml",
			"v3.11.0/openshift-controller-manager/default-rolebindings-controller-clusterrolebinding-image-puller.yaml",
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
