all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/library-go/alpha-build-machinery/make/, \
	golang.mk \
	targets/openshift/bindata.mk \
	targets/openshift/deps.mk \
	targets/openshift/images.mk \
)

# This will call a macro called "build-image" which will generate image specific targets based on the parameters:
# $0 - macro name
# $1 - target suffix
# $2 - Dockerfile path
# $3 - context directory for image build
# It will generate target "image-$(1)" for builing the image an binding it as a prerequisite to target "images".
$(call build-image,origin-$(notdir $(GO_PACKAGE)),./Dockerfile,.)

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


GO_TEST_PACKAGES :=./pkg/... ./cmd/...


test-e2e: GO_TEST_PACKAGES :=./test/e2e/...
test-e2e: GO_TEST_FLAGS += -v -count=1
test-e2e: test-unit
.PHONY: test-e2e

update-codegen-crds:
	go run ./vendor/github.com/openshift/library-go/cmd/crd-schema-gen/main.go --domain openshift.io --apis-dir vendor/github.com/openshift/api
.PHONY: update-codegen-crds

update-codegen: update-codegen-crds
.PHONY: update-codegen

verify-codegen-crds:
	go run ./vendor/github.com/openshift/library-go/cmd/crd-schema-gen/main.go --domain openshift.io --apis-dir vendor/github.com/openshift/api --verify-only
.PHONY: verify-codegen-crds

verify-codegen: verify-codegen-crds
.PHONY: verify-codegen

verify: verify-codegen
