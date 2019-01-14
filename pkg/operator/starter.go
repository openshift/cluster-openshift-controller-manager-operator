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
	configv1client "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/apis/openshiftcontrollermanager/v1"
	operatorconfigclient "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned"

	operatorclientinformers "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/informers/externalversions"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

func RunOperator(ctx *controllercmd.ControllerContext) error {
	kubeClient, err := kubernetes.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	operatorConfigClient, err := operatorconfigclient.NewForConfig(ctx.KubeConfig)
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
		schema.GroupVersionResource{Group: v1.GroupName, Version: "v1", Resource: "openshiftcontrollermanageroperatorconfigs"},
	)

	operatorConfigInformers := operatorclientinformers.NewSharedInformerFactory(operatorConfigClient, 10*time.Minute)
	kubeInformersForOpenshiftControllerManagerNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(targetNamespaceName))
	kubeInformersForOperatorNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(operatorNamespaceName))
	configInformers := configinformers.NewSharedInformerFactory(configClient, 10*time.Minute)

	operator := NewOpenShiftControllerManagerOperator(
		os.Getenv("IMAGE"),
		operatorConfigInformers.Openshiftcontrollermanager().V1().OpenShiftControllerManagerOperatorConfigs(),
		kubeInformersForOpenshiftControllerManagerNamespace,
		operatorConfigClient.OpenshiftcontrollermanagerV1(),
		kubeClient,
		ctx.EventRecorder,
	)

	configObserver := NewConfigObserver(
		operatorConfigInformers.Openshiftcontrollermanager().V1().OpenShiftControllerManagerOperatorConfigs(),
		operatorConfigClient.OpenshiftcontrollermanagerV1(),
		kubeInformersForOperatorNamespace,
		configInformers,
	)

	opClient := &operatorClient{
		informers: operatorConfigInformers,
		client:    operatorConfigClient.OpenshiftcontrollermanagerV1(),
	}

	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		"openshift-controller-manager-operator",
		[]configv1.ObjectReference{},
		configClient.ConfigV1(),
		opClient,
		ctx.EventRecorder,
	)

	operatorConfigInformers.Start(ctx.StopCh)
	kubeInformersForOpenshiftControllerManagerNamespace.Start(ctx.StopCh)
	kubeInformersForOperatorNamespace.Start(ctx.StopCh)
	configInformers.Start(ctx.StopCh)

	go operator.Run(1, ctx.StopCh)
	go configObserver.Run(1, ctx.StopCh)
	go clusterOperatorStatus.Run(1, ctx.StopCh)

	<-ctx.StopCh
	return fmt.Errorf("stopped")
}
