package operator

import (
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	configv1 "github.com/openshift/api/config/v1"
	operatorapiv1 "github.com/openshift/api/operator/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	configobservationcontroller "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/configobservercontroller"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

func RunOperator(ctx *controllercmd.ControllerContext) error {
	kubeClient, err := kubernetes.NewForConfig(ctx.ProtoKubeConfig)
	if err != nil {
		return err
	}
	operatorclient, err := operatorclient.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	dynamicClient, err := dynamic.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	configClient, err := configclient.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}

	v1helpers.EnsureOperatorConfigExists(
		dynamicClient,
		v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/operator-config.yaml"),
		schema.GroupVersionResource{Group: operatorapiv1.GroupName, Version: "v1", Resource: "openshiftcontrollermanagers"},
	)

	operatorConfigInformers := operatorinformers.NewSharedInformerFactory(operatorclient, 10*time.Minute)
	kubeInformersForOpenshiftControllerManagerNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(util.TargetNamespace))
	kubeInformersForOperatorNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(util.OperatorNamespace))
	configInformers := configinformers.NewSharedInformerFactory(configClient, 10*time.Minute)

	operator := NewOpenShiftControllerManagerOperator(
		os.Getenv("IMAGE"),
		operatorConfigInformers.Operator().V1().OpenShiftControllerManagers(),
		kubeInformersForOpenshiftControllerManagerNamespace,
		operatorclient.OperatorV1(),
		kubeClient,
		ctx.EventRecorder,
	)

	opClient := &operatorClient{
		informers: operatorConfigInformers,
		client:    operatorclient.OperatorV1(),
	}

	configObserver := configobservationcontroller.NewConfigObserver(
		opClient,
		configInformers,
		kubeInformersForOperatorNamespace,
		ctx.EventRecorder,
	)

	versionGetter := &versionGetter{
		openshiftControllerManagers: operatorclient.OperatorV1().OpenShiftControllerManagers(),
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
		opClient,
		versionGetter,
		ctx.EventRecorder,
	)

	operatorConfigInformers.Start(ctx.Done())
	kubeInformersForOpenshiftControllerManagerNamespace.Start(ctx.Done())
	kubeInformersForOperatorNamespace.Start(ctx.Done())
	configInformers.Start(ctx.Done())

	go operator.Run(1, ctx.Done())
	go configObserver.Run(1, ctx.Done())
	go clusterOperatorStatus.Run(1, ctx.Done())

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
	co, err := v.openshiftControllerManagers.Get("cluster", metav1.GetOptions{})
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
