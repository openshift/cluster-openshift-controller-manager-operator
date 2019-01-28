package operator

import (
	"k8s.io/client-go/tools/cache"

	operatorv1 "github.com/openshift/api/operator/v1"
	clientv1 "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/clientset/versioned/typed/openshiftcontrollermanager/v1"
	clientinformers "github.com/openshift/cluster-openshift-controller-manager-operator/pkg/generated/informers/externalversions"
)

type operatorClient struct {
	informers clientinformers.SharedInformerFactory
	client    clientv1.OpenshiftcontrollermanagerV1Interface
}

func (p *operatorClient) Informer() cache.SharedIndexInformer {
	return p.informers.Openshiftcontrollermanager().V1().OpenShiftControllerManagerOperatorConfigs().Informer()
}

func (p *operatorClient) CurrentStatus() (operatorv1.OperatorStatus, error) {
	instance, err := p.informers.Openshiftcontrollermanager().V1().OpenShiftControllerManagerOperatorConfigs().Lister().Get("instance")
	if err != nil {
		return operatorv1.OperatorStatus{}, err
	}

	return instance.Status.OperatorStatus, nil
}

func (c *operatorClient) GetOperatorState() (*operatorv1.OperatorSpec, *operatorv1.OperatorStatus, string, error) {
	instance, err := c.informers.Openshiftcontrollermanager().V1().OpenShiftControllerManagerOperatorConfigs().Lister().Get("instance")
	if err != nil {
		return nil, nil, "", err
	}

	return &instance.Spec.OperatorSpec, &instance.Status.OperatorStatus, instance.ResourceVersion, nil
}

func (c *operatorClient) UpdateOperatorSpec(resourceVersion string, spec *operatorv1.OperatorSpec) (*operatorv1.OperatorSpec, string, error) {
	original, err := c.informers.Openshiftcontrollermanager().V1().OpenShiftControllerManagerOperatorConfigs().Lister().Get("instance")
	if err != nil {
		return nil, "", err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Spec.OperatorSpec = *spec

	ret, err := c.client.OpenShiftControllerManagerOperatorConfigs().Update(copy)
	if err != nil {
		return nil, "", err
	}

	return &ret.Spec.OperatorSpec, ret.ResourceVersion, nil
}
func (c *operatorClient) UpdateOperatorStatus(resourceVersion string, status *operatorv1.OperatorStatus) (*operatorv1.OperatorStatus, error) {
	original, err := c.informers.Openshiftcontrollermanager().V1().OpenShiftControllerManagerOperatorConfigs().Lister().Get("instance")
	if err != nil {
		return nil, err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Status.OperatorStatus = *status

	ret, err := c.client.OpenShiftControllerManagerOperatorConfigs().UpdateStatus(copy)
	if err != nil {
		return nil, err
	}

	return &ret.Status.OperatorStatus, nil
}
