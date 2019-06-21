IMAGE ?= docker.io/openshift/origin-cluster-openshift-controller-manager-operator
TAG ?= latest
PROG  := cluster-openshift-controller-manager-operator
GOFLAGS :=

all: build build-image verify
.PHONY: all
build:
	go build $(GOFLAGS) ./cmd/cluster-openshift-controller-manager-operator
.PHONY: build

image:
	docker build -t "$(IMAGE):$(TAG)" .
.PHONY: build-image

test: test-unit test-e2e
.PHONY: test

test-unit:
ifndef JUNITFILE
	go test $(GOFLAGS) -race ./pkg/... ./cmd/...
else
ifeq (, $(shell which gotest2junit 2>/dev/null))
$(error gotest2junit not found! Get it by `go get -u github.com/openshift/release/tools/gotest2junit`.)
endif
	go test $(GOFLAGS) -race -json ./... | gotest2junit > $(JUNITFILE)
endif
.PHONY: test-unit

test-e2e:
	go test -count=1 -v ./test/e2e/...
.PHONY: test-e2e

verify: verify-govet
	hack/verify-gofmt.sh
	hack/verify-generated-bindata.sh
.PHONY: verify

verify-govet:
	go vet $(GOFLAGS) ./...
.PHONY: verify-govet

clean:
	rm -- "$(PROG)"
.PHONY: clean

update-codegen-crds:
	go run ./vendor/github.com/openshift/library-go/cmd/crd-schema-gen/main.go --domain openshift.io --apis-dir vendor/github.com/openshift/api
update-codegen: update-codegen-crds
verify-codegen-crds:
	go run ./vendor/github.com/openshift/library-go/cmd/crd-schema-gen/main.go --domain openshift.io --apis-dir vendor/github.com/openshift/api --verify-only
verify-codegen: verify-codegen-crds
verify: verify-codegen

.PHONY: update-codegen-crds update-codegen verify-codegen-crds verify-codegen verify
