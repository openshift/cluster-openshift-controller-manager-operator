FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.25-openshift-4.22 AS builder
WORKDIR /go/src/github.com/openshift/cluster-openshift-controller-manager-operator
COPY . .
RUN GO_COMPLIANCE_INFO=0 make build \
    && gzip cluster-openshift-controller-manager-operator-tests-ext

FROM registry.ci.openshift.org/ocp/4.22:base-rhel9
COPY --from=builder /go/src/github.com/openshift/cluster-openshift-controller-manager-operator/cluster-openshift-controller-manager-operator /usr/bin/
COPY --from=builder /go/src/github.com/openshift/cluster-openshift-controller-manager-operator/cluster-openshift-controller-manager-operator-tests-ext.gz /usr/bin/
COPY manifests /manifests
COPY empty-resources /manifests
LABEL io.openshift.release.operator true
