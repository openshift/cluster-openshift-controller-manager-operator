FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.22-openshift-4.17 AS builder
WORKDIR /go/src/github.com/openshift/cluster-openshift-controller-manager-operator
COPY . .
RUN GO_COMPLIANCE_INFO=0 make

FROM registry.ci.openshift.org/ocp/4.17:base-rhel9
COPY --from=builder /go/src/github.com/openshift/cluster-openshift-controller-manager-operator/cluster-openshift-controller-manager-operator /usr/bin/
COPY manifests /manifests
COPY vendor/github.com/openshift/api/config/v1/zz_generated.crd-manifests/*_openshift-controller-manager_*.yaml /manifests
COPY vendor/github.com/openshift/api/operator/v1/zz_generated.crd-manifests/0000_50_openshift-controller-manager_02_openshiftcontrollermanagers.crd.yaml /manifests
COPY empty-resources /manifests
LABEL io.openshift.release.operator true
