# cluster-openshift-controller-manager-operator

This operator installs and manages the [openshift-controller-manager](https://github.com/openshift/openshift-controller-manager) and the [route-controller-manager](https://github.com/openshift/route-controller-manager) on an OpenShift cluster. It is an [OpenShift ClusterOperator](https://github.com/openshift/enhancements/blob/master/dev-guide/operators.md#what-is-an-openshift-clusteroperator) that reconciles the `openshiftcontrollermanagers.operator.openshift.io` custom resource, manages operand Deployments, ConfigMaps, RBAC, and reports status via the `openshift-controller-manager` ClusterOperator object.

The operator dynamically enables and disables operand controllers based on [cluster capabilities](https://docs.openshift.com/container-platform/latest/installing/cluster-capabilities.html) (Build, DeploymentConfig, ImageRegistry), and handles proxy configuration, CA bundle injection, and feature gate observation.

## Quick Start

### Prerequisites

- Go 1.25+
- Access to an OpenShift cluster (for e2e tests)

### Building

```bash
make build
```

### Running Tests

```bash
# Unit tests
make test

# E2E tests (requires a running OpenShift cluster)
make test-e2e
```

### Running Locally

This operator runs as a Deployment in the `openshift-controller-manager-operator` namespace. To test a custom image in a cluster, see [Testing a ClusterOperator/Operand image in a cluster](https://github.com/openshift/enhancements/blob/master/dev-guide/operators.md#how-can-i-test-changes-to-an-openshift-operatoroperandrelease-component).

## Inspecting Cluster State

```bash
# View the ClusterOperator status
oc get clusteroperator openshift-controller-manager -o yaml

# View the operator CR
oc get openshiftcontrollermanagers.operator.openshift.io cluster -o yaml
```

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) — Design decisions and component architecture
- [CONTRIBUTING.md](CONTRIBUTING.md) — How to submit changes
- [AGENTS.md](AGENTS.md) — AI agent instructions

## OpenShift Tests Extension (OTE)

This repository includes an OTE-compatible test binary for CI integration.

```bash
# List available test suites
./cluster-openshift-controller-manager-operator-tests-ext list-suites

# Run a test suite
./cluster-openshift-controller-manager-operator-tests-ext run-suite openshift/openshift-controller-manager-operator/all
```

## Related Repositories

| Repository | Relationship |
|-----------|-------------|
| [openshift/openshift-controller-manager](https://github.com/openshift/openshift-controller-manager) | Operand — the controller-manager binary this operator deploys |
| [openshift/route-controller-manager](https://github.com/openshift/route-controller-manager) | Operand — the route-controller-manager binary this operator deploys |
| [openshift/library-go](https://github.com/openshift/library-go) | Shared operator framework (static resource controller, config observer, status controller) |
| [openshift/api](https://github.com/openshift/api) | API types including `OpenShiftControllerManager` operator config |
