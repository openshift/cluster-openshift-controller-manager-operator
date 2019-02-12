package operator

import (
	"k8s.io/client-go/tools/cache"

	operatorapiv1 "github.com/openshift/api/operator/v1"
	operatorclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
)

type operatorClient struct {
	informers operatorinformers.SharedInformerFactory
	client    operatorclientv1.OperatorV1Interface
}

func (p *operatorClient) Informer() cache.SharedIndexInformer {
	return p.informers.Operator().V1().OpenShiftControllerManagers().Informer()
}

func (p *operatorClient) CurrentStatus() (operatorapiv1.OperatorStatus, error) {
	instance, err := p.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return operatorapiv1.OperatorStatus{}, err
	}

	return instance.Status.OperatorStatus, nil
}

func (c *operatorClient) GetOperatorState() (*operatorapiv1.OperatorSpec, *operatorapiv1.OperatorStatus, string, error) {
	instance, err := c.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return nil, nil, "", err
	}

	return &instance.Spec.OperatorSpec, &instance.Status.OperatorStatus, instance.ResourceVersion, nil
}

func (c *operatorClient) UpdateOperatorSpec(resourceVersion string, spec *operatorapiv1.OperatorSpec) (*operatorapiv1.OperatorSpec, string, error) {
	original, err := c.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return nil, "", err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Spec.OperatorSpec = *spec

	ret, err := c.client.OpenShiftControllerManagers().Update(copy)
	if err != nil {
		return nil, "", err
	}

	return &ret.Spec.OperatorSpec, ret.ResourceVersion, nil
}
func (c *operatorClient) UpdateOperatorStatus(resourceVersion string, status *operatorapiv1.OperatorStatus) (*operatorapiv1.OperatorStatus, error) {
	original, err := c.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return nil, err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Status.OperatorStatus = *status

	ret, err := c.client.OpenShiftControllerManagers().UpdateStatus(copy)
	if err != nil {
		return nil, err
	}

	return &ret.Status.OperatorStatus, nil
}
