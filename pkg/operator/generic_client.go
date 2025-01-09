package operator

import (
	"context"
	"fmt"
	"github.com/openshift/library-go/pkg/apiserver/jsonpatch"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"

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

	for i, val := range desiredConfiguration.Conditions {
		// desired configuration may contain empty string fields.
		// However, they are persisted by API server as nil. This causes a hotloop as empty string and nil
		// are not equal. We explicitly convert empty strings to nil to prevent hotloop.
		// This should be safe, because explicitly setting empty string has no meaning.
		if len(ptr.Deref(val.Message, "")) == 0 {
			desiredConfiguration.Conditions[i].Message = nil
		}
		if len(ptr.Deref(val.Reason, "")) == 0 {
			desiredConfiguration.Conditions[i].Reason = nil
		}
	}

	desiredStatus := &v1.OpenShiftControllerManagerStatusApplyConfiguration{
		OperatorStatusApplyConfiguration: *desiredConfiguration,
	}
	desired := v1.OpenShiftControllerManager("cluster")
	desired.WithStatus(desiredStatus)
	instance, err := p.client.OpenShiftControllerManagers().Get(ctx, "cluster", metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		// do nothing and proceed with the apply
		v1helpers.SetApplyConditionsLastTransitionTime(p.clock, &desired.Status.Conditions, nil)
	case err != nil:
		return fmt.Errorf("unable to get operator configuration: %w", err)
	default:
		original, err := v1.ExtractOpenShiftControllerManagerStatus(instance, fieldManager)
		if err != nil {
			return fmt.Errorf("unable to extract operator configuration: %w", err)
		}

		if equality.Semantic.DeepEqual(original, desired) {
			return nil
		}

		if original.Status != nil {
			v1helpers.SetApplyConditionsLastTransitionTime(clock.RealClock{}, &desired.Status.Conditions, original.Status.Conditions)
		} else {
			v1helpers.SetApplyConditionsLastTransitionTime(clock.RealClock{}, &desired.Status.Conditions, nil)
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

func (p *genericClient) PatchOperatorStatus(ctx context.Context, jsonPatch *jsonpatch.PatchSet) (err error) {
	jsonPatchBytes, err := jsonPatch.Marshal()
	if err != nil {
		return err
	}
	_, err = p.client.OpenShiftControllerManagers().Patch(ctx, "cluster", types.JSONPatchType, jsonPatchBytes, metav1.PatchOptions{}, "/status")
	return err
}
