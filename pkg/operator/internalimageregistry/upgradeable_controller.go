package internalimageregistry

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	informers "github.com/openshift/client-go/operator/informers/externalversions/operator/v1"
	listers "github.com/openshift/client-go/operator/listers/operator/v1"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

type authTokenTypeUpgradeableController struct {
	factory.Controller
	operatorClient v1helpers.OperatorClient
	ocms           listers.OpenShiftControllerManagerLister
}

// NewAuthTokenTypeUpgradeableController creates a controller that blocks upgrades if the
// imageRegistryAuthTokenType field is not explicitly set to `Legacy`.
//
// This controller is no longer needed after v4.15.
func NewAuthTokenTypeUpgradeableController(operatorClient v1helpers.OperatorClient,
	ocms informers.OpenShiftControllerManagerInformer, recorder events.Recorder) *authTokenTypeUpgradeableController {
	c := &authTokenTypeUpgradeableController{
		operatorClient: operatorClient,
		ocms:           ocms.Lister(),
	}
	c.Controller = factory.New().
		WithInformers(ocms.Informer()).
		WithSync(c.sync).
		ToController("AuthTokenTypeUpgradeableController", recorder)
	return c
}

func (c *authTokenTypeUpgradeableController) sync(ctx context.Context, controllerContext factory.SyncContext) error {
	cfg, err := c.ocms.Get("cluster")
	if err != nil {
		return nil
	}
	condition := operatorv1.OperatorCondition{
		Type:   "ImageRegistryAuthTokenTypeUpgradeable",
		Status: operatorv1.ConditionTrue,
	}
	if cfg.Spec.ImageRegistryAuthTokenType != operatorv1.ServiceAccountLegacyTokenType {
		condition.Status = operatorv1.ConditionFalse
		condition.Reason = "ImageRegistryAuthTokenTypeNotSet"
	}
	_, _, err = v1helpers.UpdateStatus(ctx, c.operatorClient, v1helpers.UpdateConditionFn(condition))
	return err
}
