FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.20-openshift-4.15 AS builder
WORKDIR /go/src/github.com/openshift/cluster-openshift-controller-manager-operator
COPY . .
RUN make

FROM registry.ci.openshift.org/ocp/4.15:base
COPY --from=builder /go/src/github.com/openshift/cluster-openshift-controller-manager-operator/cluster-openshift-controller-manager-operator /usr/bin/
COPY manifests /manifests
COPY vendor/github.com/openshift/api/config/v1/*_openshift-controller-manager-operator_*.yaml /manifests
COPY vendor/github.com/openshift/api/operator/v1/0000_50_cluster-openshift-controller-manager-operator_02_config.crd.yaml /manifests
COPY empty-resources /manifests
LABEL io.openshift.release.operator true
