package operator

import (
	"fmt"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	operatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/apis/openshiftcontrollermanager/v1alpha1"
	operatorconfigclient "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned"

	operatorclientinformers "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/informers/externalversions"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/v1alpha1helpers"
	status "github.com/openshift/library-go/pkg/operator/v1alpha1status"
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

	v1alpha1helpers.EnsureOperatorConfigExists(
		dynamicClient,
		v311_00_assets.MustAsset("v3.11.0/openshift-controller-manager/operator-config.yaml"),
		schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: "v1alpha1", Resource: "openshiftcontrollermanageroperatorconfigs"},
		v1alpha1helpers.GetImageEnv,
	)

	operatorConfigInformers := operatorclientinformers.NewSharedInformerFactory(operatorConfigClient, 10*time.Minute)
	kubeInformersForOpenshiftControllerManagerNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(targetNamespaceName))
	kubeInformersForOpenshiftCoreOperatorsNamespace := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(operatorNamespaceName))
	configInformers := configinformers.NewSharedInformerFactory(configClient, 10*time.Minute)

	operator := NewOpenShiftControllerManagerOperator(
		operatorConfigInformers.Openshiftcontrollermanager().V1alpha1().OpenShiftControllerManagerOperatorConfigs(),
		kubeInformersForOpenshiftControllerManagerNamespace,
		operatorConfigClient.OpenshiftcontrollermanagerV1alpha1(),
		kubeClient,
		ctx.EventRecorder,
	)

	configObserver := NewConfigObserver(
		operatorConfigInformers.Openshiftcontrollermanager().V1alpha1().OpenShiftControllerManagerOperatorConfigs(),
		operatorConfigClient.OpenshiftcontrollermanagerV1alpha1(),
		kubeInformersForOpenshiftCoreOperatorsNamespace,
		configInformers,
	)

	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		"openshift-cluster-openshift-controller-manager-operator",
		"openshift-cluster-openshift-controller-manager-operator",
		dynamicClient,
		&operatorStatusProvider{informers: operatorConfigInformers},
	)

	operatorConfigInformers.Start(ctx.StopCh)
	kubeInformersForOpenshiftControllerManagerNamespace.Start(ctx.StopCh)
	kubeInformersForOpenshiftCoreOperatorsNamespace.Start(ctx.StopCh)
	configInformers.Start(ctx.StopCh)

	go operator.Run(1, ctx.StopCh)
	go configObserver.Run(1, ctx.StopCh)
	go clusterOperatorStatus.Run(1, ctx.StopCh)

	<-ctx.StopCh
	return fmt.Errorf("stopped")
}

type operatorStatusProvider struct {
	informers operatorclientinformers.SharedInformerFactory
}

func (p *operatorStatusProvider) Informer() cache.SharedIndexInformer {
	return p.informers.Openshiftcontrollermanager().V1alpha1().OpenShiftControllerManagerOperatorConfigs().Informer()
}

func (p *operatorStatusProvider) CurrentStatus() (operatorv1alpha1.OperatorStatus, error) {
	instance, err := p.informers.Openshiftcontrollermanager().V1alpha1().OpenShiftControllerManagerOperatorConfigs().Lister().Get("instance")
	if err != nil {
		return operatorv1alpha1.OperatorStatus{}, err
	}

	return instance.Status.OperatorStatus, nil
}
