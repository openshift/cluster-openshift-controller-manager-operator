package operator

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/clock"

	operatorapiv1 "github.com/openshift/api/operator/v1"
	v1 "github.com/openshift/client-go/operator/applyconfigurations/operator/v1"
	operatorclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

type genericClient struct {
	clock     clock.PassiveClock
	informers operatorinformers.SharedInformerFactory
	client    operatorclientv1.OperatorV1Interface
}

func (p *genericClient) Informer() cache.SharedIndexInformer {
	return p.informers.Operator().V1().OpenShiftControllerManagers().Informer()
}

func (p *genericClient) CurrentStatus() (operatorapiv1.OperatorStatus, error) {
	instance, err := p.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return operatorapiv1.OperatorStatus{}, err
	}

	return instance.Status.OperatorStatus, nil
}

func (p *genericClient) GetOperatorState() (*operatorapiv1.OperatorSpec, *operatorapiv1.OperatorStatus, string, error) {
	instance, err := p.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return nil, nil, "", err
	}

	return &instance.Spec.OperatorSpec, &instance.Status.OperatorStatus, instance.ResourceVersion, nil
}

func (p *genericClient) GetObjectMeta() (*metav1.ObjectMeta, error) {
	resource, err := p.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return nil, err
	}
	return &resource.ObjectMeta, nil
}

func (c *genericClient) GetOperatorStateWithQuorum(ctx context.Context) (*operatorapiv1.OperatorSpec, *operatorapiv1.OperatorStatus, string, error) {
	instance, err := c.client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return nil, nil, "", err
	}

	return &instance.Spec.OperatorSpec, &instance.Status.OperatorStatus, instance.GetResourceVersion(), nil
}

func (p *genericClient) UpdateOperatorSpec(ctx context.Context, resourceVersion string, spec *operatorapiv1.OperatorSpec) (*operatorapiv1.OperatorSpec, string, error) {
	resource, err := p.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return nil, "", err
	}
	resourceCopy := resource.DeepCopy()
	resourceCopy.ResourceVersion = resourceVersion
	resourceCopy.Spec.OperatorSpec = *spec

	ret, err := p.client.OpenShiftControllerManagers().Update(context.TODO(), resourceCopy, metav1.UpdateOptions{})
	if err != nil {
		return nil, "", err
	}

	return &ret.Spec.OperatorSpec, ret.ResourceVersion, nil
}
func (p *genericClient) UpdateOperatorStatus(ctx context.Context, resourceVersion string, status *operatorapiv1.OperatorStatus) (*operatorapiv1.OperatorStatus, error) {
	resource, err := p.informers.Operator().V1().OpenShiftControllerManagers().Lister().Get("cluster")
	if err != nil {
		return nil, err
	}
	resourceCopy := resource.DeepCopy()
	resourceCopy.ResourceVersion = resourceVersion
	resourceCopy.Status.OperatorStatus = *status

	ret, err := p.client.OpenShiftControllerManagers().UpdateStatus(context.TODO(), resourceCopy, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return &ret.Status.OperatorStatus, nil
}

func (p *genericClient) ApplyOperatorSpec(ctx context.Context, fieldManager string, desiredConfiguration *v1.OperatorSpecApplyConfiguration) (err error) {
	if desiredConfiguration == nil {
		return fmt.Errorf("applyConfiguration must have a value")
	}

	desiredSpec := &v1.OpenShiftControllerManagerSpecApplyConfiguration{
		OperatorSpecApplyConfiguration: *desiredConfiguration,
	}
	desired := v1.OpenShiftControllerManager("cluster")
	desired.WithSpec(desiredSpec)

	instance, err := p.client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
	// do nothing and proceed with the apply
	case err != nil:
		return fmt.Errorf("unable to get operator configuration: %w", err)
	default:
		original, err := v1.ExtractOpenShiftControllerManager(instance, fieldManager)
		if err != nil {
			return fmt.Errorf("unable to extract operator configuration: %w", err)
		}
		if equality.Semantic.DeepEqual(original, desired) {
			return nil
		}
	}

	_, err = p.client.OpenShiftControllerManagers().Apply(ctx, desired, metav1.ApplyOptions{
		Force:        true,
		FieldManager: fieldManager,
	})
	if err != nil {
		return fmt.Errorf("unable to Apply for operator using fieldManager %q: %w", fieldManager, err)
	}

	return nil
}

func (p *genericClient) ApplyOperatorStatus(ctx context.Context, fieldManager string, desiredConfiguration *v1.OperatorStatusApplyConfiguration) (err error) {
	if desiredConfiguration == nil {
		return fmt.Errorf("applyConfiguration must have a value")
	}

	desired := v1.OpenShiftControllerManager("cluster")
	instance, err := p.client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		// do nothing and proceed with the apply
		v1helpers.SetApplyConditionsLastTransitionTime(p.clock, &desiredConfiguration.Conditions, nil)
		desiredStatus := &v1.OpenShiftControllerManagerStatusApplyConfiguration{
			OperatorStatusApplyConfiguration: *desiredConfiguration,
		}
		desired.WithStatus(desiredStatus)
	case err != nil:
		return fmt.Errorf("unable to get operator configuration: %w", err)
	default:
		previous, err := v1.ExtractOpenShiftControllerManagerStatus(instance, fieldManager)
		if err != nil {
			return fmt.Errorf("unable to extract operator configuration: %w", err)
		}

		operatorStatus := &v1.OperatorStatusApplyConfiguration{}
		if previous.Status != nil {
			jsonBytes, err := json.Marshal(previous.Status)
			if err != nil {
				return fmt.Errorf("unable to serialize operator configuration: %w", err)
			}
			if err := json.Unmarshal(jsonBytes, operatorStatus); err != nil {
				return fmt.Errorf("unable to deserialize operator configuration: %w", err)
			}
		}

		switch {
		case desiredConfiguration.Conditions != nil && operatorStatus != nil:
			v1helpers.SetApplyConditionsLastTransitionTime(p.clock, &desiredConfiguration.Conditions, operatorStatus.Conditions)
		case desiredConfiguration.Conditions != nil && operatorStatus == nil:
			v1helpers.SetApplyConditionsLastTransitionTime(p.clock, &desiredConfiguration.Conditions, nil)
		}

		v1helpers.CanonicalizeOperatorStatus(desiredConfiguration)
		v1helpers.CanonicalizeOperatorStatus(operatorStatus)

		original := v1.OpenShiftControllerManager("cluster")
		if operatorStatus != nil {
			originalStatus := &v1.OpenShiftControllerManagerStatusApplyConfiguration{
				OperatorStatusApplyConfiguration: *operatorStatus,
			}
			original.WithStatus(originalStatus)
		}

		desiredStatus := &v1.OpenShiftControllerManagerStatusApplyConfiguration{
			OperatorStatusApplyConfiguration: *desiredConfiguration,
		}
		desired.WithStatus(desiredStatus)

		if equality.Semantic.DeepEqual(original, desired) {
			return nil
		}
	}

	_, err = p.client.OpenShiftControllerManagers().ApplyStatus(ctx, desired, metav1.ApplyOptions{
		Force:        true,
		FieldManager: fieldManager,
	})
	if err != nil {
		return fmt.Errorf("unable to Apply for operator using fieldManager %q: %w", fieldManager, err)
	}

	return nil
}
