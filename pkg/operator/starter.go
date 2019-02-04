package operator

import (
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/dynamic"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	configv1 "github.com/openshift/api/config/v1"
	operatorapiv1 "github.com/openshift/api/operator/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	configobservationcontroller "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/configobservation/configobservercontroller"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/util"
)

func RunOperator(ctx *controllercmd.ControllerContext) error {
	kubeClient, err := kubernetes.NewForConfig(ctx.KubeConfig)
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
	configClient, err := configv1client.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}

	v1helpers.EnsureOperatorConfigExists(
		dynamicClient,
		v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/operator-config.yaml"),
		schema.GroupVersionResource{Group: operatorapiv1.GroupName, Version: "v1", Resource: "openshiftcontrollermanagers"},
	)

	operatorConfigInformers := operatorinformers.NewSharedInformerFactory(operatorclient, 10*time.Minute)
	kubeInformersForOpenshiftControllerManagerNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(targetNamespaceName))
	kubeInformersForOperatorNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(util.OperatorNamespaceName))
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

	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		"openshift-controller-manager",
		[]configv1.ObjectReference{},
		configClient.ConfigV1(),
		opClient,
		status.NewVersionGetter(),
		ctx.EventRecorder,
	)

	operatorConfigInformers.Start(ctx.Context.Done())
	kubeInformersForOpenshiftControllerManagerNamespace.Start(ctx.Context.Done())
	kubeInformersForOperatorNamespace.Start(ctx.Context.Done())
	configInformers.Start(ctx.Context.Done())

	go operator.Run(1, ctx.Context.Done())
	go configObserver.Run(1, ctx.Context.Done())
	go clusterOperatorStatus.Run(1, ctx.Context.Done())

	<-ctx.Context.Done()
	return fmt.Errorf("stopped")
}
