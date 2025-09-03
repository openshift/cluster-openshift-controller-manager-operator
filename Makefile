all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/images.mk \
)

IMAGE_REGISTRY?=registry.svc.ci.openshift.org

GO_TEST_PACKAGES :=./pkg/... ./cmd/...

# This will call a macro called "build-image" which will generate image specific targets based on the parameters:
# $0 - macro name
# $1 - target suffix
# $2 - Dockerfile path
# $3 - context directory for image build
# It will generate target "image-$(1)" for builing the image an binding it as a prerequisite to target "images".
$(call build-image,ocp-cluster-openshift-controller-manager-operator,$(IMAGE_REGISTRY)/ocp/4.3:cluster-openshift-controller-manager-operator,./Dockerfile,.)

test-e2e: GO_TEST_PACKAGES :=./test/e2e/...
test-e2e: GO_TEST_FLAGS += -v -count=1
test-e2e: test-unit
.PHONY: test-e2e

# -------------------------------------------------------------------
# OpenShift Tests Extension (Cluster OpenShift Controller Manager Operator)
# -------------------------------------------------------------------
TESTS_EXT_BINARY := cluster-openshift-controller-manager-operator-tests-ext
TESTS_EXT_PACKAGE := ./cmd/cluster-openshift-controller-manager-operator-tests-ext

TESTS_EXT_GIT_COMMIT := $(shell git rev-parse --short HEAD)
TESTS_EXT_BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
TESTS_EXT_GIT_TREE_STATE := $(shell if git diff-index --quiet HEAD --; then echo clean; else echo dirty; fi)

TESTS_EXT_LDFLAGS := \
	-X 'main.CommitFromGit=$(TESTS_EXT_GIT_COMMIT)' \
	-X 'main.BuildDate=$(TESTS_EXT_BUILD_DATE)' \
	-X 'main.GitTreeState=$(TESTS_EXT_GIT_TREE_STATE)'

.PHONY: tests-ext-build
tests-ext-build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) GO_COMPLIANCE_POLICY=exempt_all CGO_ENABLED=0 go build -o $(TESTS_EXT_BINARY) -ldflags "$(TESTS_EXT_LDFLAGS)" $(TESTS_EXT_PACKAGE)

.PHONY: tests-ext-update
tests-ext-update:
	./$(TESTS_EXT_BINARY) update

.PHONY: tests-ext-clean
tests-ext-clean:
	rm -f $(TESTS_EXT_BINARY) $(TESTS_EXT_BINARY).gz
