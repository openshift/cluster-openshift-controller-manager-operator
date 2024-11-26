package operator

import (
	"github.com/spf13/cobra"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/operator"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"

	"k8s.io/utils/clock"
)

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("openshift-controller-manager-operator", version.Get(), operator.RunOperator, clock.RealClock{}).
		NewCommand()
	cmd.Use = "operator"
	cmd.Short = "Start the Cluster openshift-controller-manager Operator"

	return cmd
}
