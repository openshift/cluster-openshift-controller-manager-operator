package operator

import (
	"github.com/spf13/cobra"

	_ "github.com/openshift/api/config/v1/zz_generated.crd-manifests"
	_ "github.com/openshift/api/operator/v1/zz_generated.crd-manifests"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("openshift-controller-manager-operator", version.Get(), operator.RunOperator).
		NewCommand()
	cmd.Use = "operator"
	cmd.Short = "Start the Cluster openshift-controller-manager Operator"

	return cmd
}
