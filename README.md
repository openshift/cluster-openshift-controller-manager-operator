# cluster-openshift-controller-manager-operator
The openshift-controller-manager operator is an 
[OpenShift ClusterOperator](https://github.com/openshift/enhancements/blob/master/dev-guide/operators.md#what-is-an-openshift-clusteroperator).
It installs and maintains the OpenShiftControllerManager [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) in a cluster and can be viewed with:     
```
oc get clusteroperator openshift-controller-manager -o yaml
```

The [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions)
`openshiftcontrollermanagers.operator.openshift.io`    
can be viewed in a cluster with:

```console
$ oc get crd openshiftcontrollermanagers.operator.openshift.io -o yaml
```

## Test
Many OpenShift ClusterOperators share common build, test, deployment, and update methods.    
For more information about how to build, deploy, test, update, and develop OpenShift ClusterOperators, see      
[OpenShift ClusterOperator and Operand Developer Document](https://github.com/openshift/enhancements/blob/master/dev-guide/operators.md#how-do-i-buildupdateverifyrun-unit-tests)

This section explains how to deploy OpenShift with your test openshift-controller-manager image:        
[Testing a ClusterOperator/Operand image in a cluster](https://github.com/openshift/enhancements/blob/master/dev-guide/operators.md#how-can-i-test-changes-to-an-openshift-operatoroperandrelease-component)

## Rebase
Follow this checklist and copy into the PR:

- [ ] Select the desired [kubernetes release branch](https://github.com/kubernetes/kubernetes/branches), and use its `go.mod` and `CHANGELOG` as references for the rest of the work.
- [ ] Bump go version, all `k8s.io/`, `github.com/openshift/`, and any other relevant dependencies as needed.
- [ ] Run `go mod vendor && go mod tidy`, commit that separately from all other changes.
- [ ] Bump image versions (Dockerfile, ci...) if needed.
- [ ] Run `make build verify test`.
- [ ] Make code changes as needed until the above pass.
- [ ] Any other minor update, like documentation.

## OpenShift Tests Extension (OTE)

This repository is compatible with the "OpenShift Tests Extension (OTE)" framework.

### Building the test binary
```bash
make build
```

### Running test suites and tests
```bash
# Run a specific test suite or test
./cluster-openshift-controller-manager-operator-tests-ext run-suite openshift/openshift-controller-manager-operator/all
./cluster-openshift-controller-manager-operator-tests-ext run-test "test-name"

# Run with JUnit output
./cluster-openshift-controller-manager-operator-tests-ext run-suite openshift/openshift-controller-manager-operator/all --junit-path=/tmp/junit-results/junit.xml
./cluster-openshift-controller-manager-operator-tests-ext run-test "test-name" --junit-path=/tmp/junit-results/junit.xml
```

### Listing available tests and suites
```bash
# List all test suites
./cluster-openshift-controller-manager-operator-tests-ext list-suites

# List tests in a specific suite
./cluster-openshift-controller-manager-operator-tests-ext list-tests openshift/openshift-controller-manager-operator/all
```

The test extension binary is included in the production image for CI/CD integration.