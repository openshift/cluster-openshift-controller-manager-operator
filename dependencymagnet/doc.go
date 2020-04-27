// +build tools

// dependencymagnet imports code that is not an explicit go dependency, but is used in things
// like Makefile targets.
package dependencymagnet

import (
	_ "github.com/go-bindata/go-bindata/go-bindata"
	_ "github.com/openshift/build-machinery-go"
)
