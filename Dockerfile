#
# This is the integrated OpenShift Service Serving Cert Signer.  It signs serving certificates for use inside the platform.
#
# The standard name for this image is openshift/origin-cluster-openshift-controller-manager-operator
#
FROM openshift/origin-release:golang-1.10
COPY . /go/src/github.com/openshift/cluster-openshift-controller-manager-operator
RUN cd /go/src/github.com/openshift/cluster-openshift-controller-manager-operator && go build ./cmd/cluster-openshift-controller-manager-operator

FROM centos:7
COPY --from=0 /go/src/github.com/openshift/cluster-openshift-controller-manager-operator/cluster-openshift-controller-manager-operator /usr/bin/cluster-openshift-controller-manager-operator
