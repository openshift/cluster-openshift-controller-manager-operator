package main

import (
	"os"

	"github.com/spf13/cobra"

	"k8s.io/component-base/cli"

	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/cmd/operator"
)

func main() {
	command := NewSSCSCommand()
	code := cli.Run(command)
	os.Exit(code)
}

func NewSSCSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster-openshift-controller-manager-operator",
		Short: "OpenShift cluster openshift-controller-manager operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	cmd.AddCommand(operator.NewOperator())

	return cmd
}
