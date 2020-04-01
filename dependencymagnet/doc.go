// +build tools

// dependencymagnet imports code that is not an explicit go dependency, but is used in things
// like Makefile targets.
package dependencymagnet

import (
	_ "github.com/jteeuwen/go-bindata/go-bindata"
	_ "github.com/openshift/build-machinery-go/make"
	_ "github.com/openshift/build-machinery-go/make/lib"
	_ "github.com/openshift/build-machinery-go/make/targets"
	_ "github.com/openshift/build-machinery-go/make/targets/golang"
	_ "github.com/openshift/build-machinery-go/make/targets/openshift"
)
