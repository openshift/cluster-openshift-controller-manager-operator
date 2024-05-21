//go:build tools
// +build tools

// dependencymagnet imports code that is not an explicit go dependency, but is used in things
// like Makefile targets.
package dependencymagnet

import (
	_ "github.com/openshift/api/config/v1/zz_generated.crd-manifests"
	_ "github.com/openshift/api/operator/v1/zz_generated.crd-manifests"
	_ "github.com/openshift/build-machinery-go"
)
