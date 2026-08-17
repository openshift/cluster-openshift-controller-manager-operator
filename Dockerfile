FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.26-openshift-5.0 AS builder
WORKDIR /go/src/github.com/openshift/cluster-openshift-controller-manager-operator
COPY . .
RUN GO_COMPLIANCE_INFO=0 make build \
    && gzip cluster-openshift-controller-manager-operator-tests-ext

FROM registry.ci.openshift.org/ocp/5.0:base-rhel9
COPY --from=builder /go/src/github.com/openshift/cluster-openshift-controller-manager-operator/cluster-openshift-controller-manager-operator /usr/bin/
COPY --from=builder /go/src/github.com/openshift/cluster-openshift-controller-manager-operator/cluster-openshift-controller-manager-operator-tests-ext.gz /usr/bin/
COPY manifests /manifests
COPY empty-resources /manifests
LABEL io.openshift.release.operator true
