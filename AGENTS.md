# AI Agent Instructions for cluster-openshift-controller-manager-operator

> Also read [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions and [CONTRIBUTING.md](CONTRIBUTING.md) for workflow.

## What This Repo Is

This is the **operator** for the openshift-controller-manager and route-controller-manager on OpenShift. It reconciles the `openshiftcontrollermanagers.operator.openshift.io/cluster` CR and manages two operand Deployments, their ConfigMaps, RBAC, network policies, and CA bundles. It reports status via the `openshift-controller-manager` ClusterOperator object.

The **operand** (the controller-manager binary itself) lives in [openshift/openshift-controller-manager](https://github.com/openshift/openshift-controller-manager). This repo only manages how that binary is deployed and configured.

## Repository Layout

```text
cmd/                                    # Entry points
  cluster-openshift-controller-manager-operator/  # Operator binary
  cluster-openshift-controller-manager-operator-tests-ext/  # OTE test binary
pkg/
  operator/                             # Core operator logic
    starter.go                          # RunOperator — wires all controllers
    operator.go                         # Main reconciler struct and work queue
    sync_openshiftcontrollermanager_v311_00.go  # Sync loop for both operands
    configobservation/                  # Config observers (builds, images, network, controllers)
    usercaobservation/                  # Proxy CA bundle sync controller
    internalimageregistry/              # Image pull secret cleanup controller
  util/consts.go                        # Namespace constants
bindata/assets/                         # Static manifests applied by the operator
manifests/                              # CVO-managed manifests (install/upgrade)
test/e2e/                               # E2E tests (Ginkgo)
```

## Build and Test Commands

```bash
make build          # Build operator and OTE test binaries
make test           # Unit tests
make verify         # Lint and vet
make test-e2e       # E2E tests (requires KUBECONFIG)
```

## Critical Rules

1. **Network policy ordering matters.** In `starter.go`, allow-rules must appear before default-deny rules in the StaticResourceController asset list. Reversing this order causes traffic interruption during reconciliation.

2. **Never remove the "always managed" and "not removable" flags.** The operator calls `management.SetOperatorAlwaysManaged()` and `management.SetOperatorNotRemovable()` because the openshift-controller-manager is a core platform component.

3. **Config observers must return previous config on error.** If an observer encounters an error, it must return `previousConfig` (not empty config) to prevent flapping the operand configuration.

4. **All informer caches must sync before controllers start.** The operator explicitly calls `WaitForCacheSync` for every informer group and returns an error if any fail. Skipping this check causes transiently incorrect operations from partially-synced caches (see OCPBUGS-81472).

5. **Vendor is committed.** After dependency changes, run `go mod tidy && go mod vendor` and commit the vendor update separately from code changes.

## Key Patterns

- **library-go operator framework**: Uses `StaticResourceController`, `ConfigObserver`, `ClusterOperatorStatusController`, `ResourceSyncController`, and `LogLevelController` from `openshift/library-go`.
- **Dual-operand sync**: One sync function manages both the openshift-controller-manager and route-controller-manager Deployments. Errors and status conditions are tracked separately with the `RouteControllerManager` prefix for the route-controller-manager.
- **Capability-gated controllers**: Controllers are disabled by prepending `-` to their name in the operand ConfigMap (e.g., `-openshift.io/build`). The mapping from controllers to capabilities is in `controllerCapabilities` in `sync_openshiftcontrollermanager_v311_00.go`.
- **Input hash-driven rollouts**: ConfigMap resource versions and content hashes are embedded as annotations on the Deployment spec template, triggering a rolling update when upstream inputs change.
- **TLS security profile propagation**: The cluster-wide TLS profile from `apiservers.config.openshift.io` is observed and propagated into the operand config. TLS config changes trigger an operator restart.

## What NOT to Do

- Do not add new operand controllers here — controllers belong in `openshift/openshift-controller-manager`. This repo only manages deployment and configuration.
- Do not modify `manifests/` without coordinating with the release team — these are CVO-managed and affect install/upgrade.
- Do not change the rate limiter parameters in `operator.go` without understanding the impact on API server load during rapid config changes.
- Do not add bindata assets after default-deny network policies — always add allow rules first.

## Test Suites

- **Unit tests**: Colocated `_test.go` files in `pkg/`. Focus on config observer logic and deployment generation.
- **E2E tests**: `test/e2e/` using Ginkgo. Validate network policy enforcement and operator availability on a live cluster.
- **OTE**: The test extension binary (`cluster-openshift-controller-manager-operator-tests-ext`) is built alongside the operator and included in the production image.
