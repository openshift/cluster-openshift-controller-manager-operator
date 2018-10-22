package operator

import (
	"fmt"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	operatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/apis/openshiftcontrollermanager/v1alpha1"
	operatorconfigclient "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned"

	operatorclientinformers "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/informers/externalversions"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator/v311_00_assets"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1alpha1helpers"
)

func RunOperator(clientConfig *rest.Config, stopCh <-chan struct{}) error {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}
	operatorConfigClient, err := operatorconfigclient.NewForConfig(clientConfig)
	if err != nil {
		return err
	}
	dynamicClient, err := dynamic.NewForConfig(clientConfig)
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

	operator := NewOpenShiftControllerManagerOperator(
		operatorConfigInformers.Openshiftcontrollermanager().V1alpha1().OpenShiftControllerManagerOperatorConfigs(),
		kubeInformersForOpenshiftControllerManagerNamespace,
		operatorConfigClient.OpenshiftcontrollermanagerV1alpha1(),
		kubeClient,
	)

	configObserver := NewConfigObserver(
		operatorConfigInformers.Openshiftcontrollermanager().V1alpha1().OpenShiftControllerManagerOperatorConfigs(),
		operatorConfigClient.OpenshiftcontrollermanagerV1alpha1(),
		kubeInformersForOpenshiftCoreOperatorsNamespace,
		kubeClient,
		clientConfig,
	)

	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		"openshift-core-operators",
		"openshift-controller-manager",
		dynamicClient,
		&operatorStatusProvider{informers: operatorConfigInformers},
	)

	operatorConfigInformers.Start(stopCh)
	kubeInformersForOpenshiftControllerManagerNamespace.Start(stopCh)
	kubeInformersForOpenshiftCoreOperatorsNamespace.Start(stopCh)

	go operator.Run(1, stopCh)
	go configObserver.Run(1, stopCh)
	go clusterOperatorStatus.Run(1, stopCh)

	<-stopCh
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
