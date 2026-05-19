package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"

	otecmd "github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	oteextension "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	oteginkgo "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
	"github.com/openshift/cluster-openshift-controller-manager-operator/pkg/version"

	_ "github.com/openshift/cluster-openshift-controller-manager-operator/test/e2e"

	"k8s.io/klog/v2"
)

func main() {
	command, err := newOperatorTestCommand(context.Background())
	if err != nil {
		klog.Fatal(err)
	}
	code := cli.Run(command)
	os.Exit(code)
}

func newOperatorTestCommand(ctx context.Context) (*cobra.Command, error) {
	registry, err := prepareOperatorTestsRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare operator tests registry: %w", err)
	}

	cmd := &cobra.Command{
		Use:   "cluster-openshift-controller-manager-operator-tests-ext",
		Short: "A binary used to run cluster-openshift-controller-manager-operator tests as part of OTE.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				klog.Fatal(err)
			}
		},
	}

	if v := version.Get().String(); len(v) == 0 {
		cmd.Version = "<unknown>"
	} else {
		cmd.Version = v
	}

	cmd.AddCommand(otecmd.DefaultExtensionCommands(registry)...)

	return cmd, nil
}

func prepareOperatorTestsRegistry() (*oteextension.Registry, error) {
	registry := oteextension.NewRegistry()
	extension := oteextension.NewExtension("openshift", "payload", "cluster-openshift-controller-manager-operator")

	// Non-disruptive NetworkPolicy tests
	extension.AddSuite(oteextension.Suite{
		Name:        "openshift/cluster-openshift-controller-manager-operator/operator",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[Operator]") && name.contains("[NetworkPolicy]") && !name.contains("[Serial]")`,
		},
	})

	// Serial NetworkPolicy tests (e.g. reconciliation tests)
	extension.AddSuite(oteextension.Suite{
		Name:        "openshift/cluster-openshift-controller-manager-operator/operator/serial",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[Operator]") && name.contains("[NetworkPolicy]") && name.contains("[Serial]")`,
		},
	})

	specs, err := oteginkgo.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		return nil, fmt.Errorf("couldn't build extension test specs from ginkgo: %w", err)
	}

	extension.AddSpecs(specs)
	registry.Register(extension)
	return registry, nil
}
