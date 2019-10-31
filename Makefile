all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/library-go/alpha-build-machinery/make/, \
	golang.mk \
	targets/openshift/bindata.mk \
	targets/openshift/images.mk \
	targets/openshift/crd-schema-gen.mk \
)

IMAGE_REGISTRY?=registry.svc.ci.openshift.org

# Set crd-schema-gen variables
CONTROLLER_GEN_VERSION :=v0.2.1
CRD_APIS :=./vendor/github.com/openshift/api/operator/v1

GO_TEST_PACKAGES :=./pkg/... ./cmd/...

# This will call a macro called "build-image" which will generate image specific targets based on the parameters:
# $0 - macro name
# $1 - target suffix
# $2 - Dockerfile path
# $3 - context directory for image build
# It will generate target "image-$(1)" for builing the image an binding it as a prerequisite to target "images".
$(call build-image,ocp-cluster-openshift-controller-manager-operator,$(IMAGE_REGISTRY)/ocp/4.3:cluster-openshift-controller-manager-operator,./Dockerfile.rhel7,.)

# This will call a macro called "add-bindata" which will generate bindata specific targets based on the parameters:
# $0 - macro name
# $1 - target suffix
# $2 - input dirs
# $3 - prefix
# $4 - pkg
# $5 - output
# It will generate targets {update,verify}-bindata-$(1) logically grouping them in unsuffixed versions of these targets
# and also hooked into {update,verify}-generated for broader integration.
$(call add-bindata,v3.11.0,./bindata/v3.11.0/...,bindata,v311_00_assets,pkg/operator/v311_00_assets/bindata.go)

# This will call a macro called "add-crd-gen" will will generate crd manifests based on the parameters:
# $1 - target name
# $2 - apis
# $3 - manifests
# $4 - output
$(call add-crd-gen,manifests,$(CRD_APIS),./manifests,./manifests)

test-e2e: GO_TEST_PACKAGES :=./test/e2e/...
test-e2e: GO_TEST_FLAGS += -v -count=1
test-e2e: test-unit
.PHONY: test-e2e

update-codegen: update-codegen-crds
.PHONY: update-codegen

verify-codegen: verify-codegen-crds
.PHONY: verif-codegen
