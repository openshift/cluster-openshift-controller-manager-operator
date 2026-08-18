# Architecture: cluster-openshift-controller-manager-operator

## Scope

This operator manages two operand Deployments:

- **openshift-controller-manager** — runs OpenShift-specific controllers (builds, deployments, image triggers, service accounts, templates, unidling)
- **route-controller-manager** — runs the ingress-to-route and ingress IP controllers (split from the controller-manager but still sharing the same operator CR)

The operator does **not** manage the upstream kube-controller-manager (that is `cluster-kube-controller-manager-operator`).

## Namespace Map

| Namespace | Purpose |
|-----------|---------|
| `openshift-controller-manager-operator` | Operator Deployment runs here |
| `openshift-controller-manager` | Operand: openshift-controller-manager Deployment, ConfigMaps, ServiceAccount |
| `openshift-route-controller-manager` | Operand: route-controller-manager Deployment, ConfigMaps, ServiceAccount |
| `openshift-config` | Source for user-specified global configuration (proxy CA, build defaults) |
| `openshift-config-managed` | Source for machine-managed global configuration |

## Component Overview

The operator binary starts a set of library-go controllers that collectively reconcile the desired state:

```
RunOperator()
  ├── OpenShiftControllerManagerOperator  — main sync loop for both operand Deployments
  ├── StaticResourceController            — reconciles RBAC, Namespaces, Services, NetworkPolicies from bindata/
  ├── ConfigObserver                      — watches cluster config and transforms it into operand configuration
  ├── UserCAObservationController         — watches proxy config and syncs trusted CA bundles
  ├── ResourceSyncController              — copies Secrets/ConfigMaps between namespaces
  ├── ClusterOperatorStatusController     — aggregates conditions into the ClusterOperator object
  ├── LogLevelController                  — reconciles operator log verbosity from CR spec
  └── ImagePullSecretCleanupController    — cleans up internal-registry pull secrets when registry is disabled
```

## Reconciliation Flow

The main sync loop (`syncOpenShiftControllerManager_v311_00_to_latest`) runs on every change to the operator CR, operand namespaces, or watched cluster config:

1. **ConfigMaps** — merge default config with observed config and operator overrides, compute input hashes for rollout detection
2. **Client CA** — sync the kube-apiserver client CA bundle into each operand namespace
3. **Trust bundles** — ensure service-ca and global-ca ConfigMaps exist with injection annotations
4. **Deployments** — apply the operand Deployment with correct image, replica count (one per master node), log level, proxy env vars, and spec annotations
5. **Status** — set Available/Progressing/Degraded/Upgradeable conditions on the operator CR, then let the ClusterOperatorStatusController aggregate them

## Capabilities Integration

The operator dynamically disables operand controllers when cluster capabilities are not enabled:

| Capability | Disabled Controllers |
|-----------|---------------------|
| `Build` | build, build-config-change, builder-sa, builder-rolebindings |
| `DeploymentConfig` | deployer, deployer-sa, deployment-config, deployer-rolebindings |
| `ImageRegistry` | service-account-pull-secrets, image-puller-rolebindings |

When the Build capability is initially disabled, the operator polls every 5 minutes and exits (triggering a restart) if the capability becomes enabled — ensuring a full re-initialization with build controllers active.

## Config Observation

Config observers watch cluster-level resources and merge their state into the operand ConfigMap:

| Observer | Watches | Configures |
|---------|---------|-----------|
| `builds` | `builds.config.openshift.io/cluster` | Build defaults and overrides (only when Build capability is enabled) |
| `images` | `images.config.openshift.io/cluster` | Internal/external registry hostnames |
| `network` | `networks.config.openshift.io/cluster` | Cluster network CIDR for egress/ingress |
| `deployimages` | Operator env vars | Deployer and builder image pull specs |
| `controllers` | `clusterversions.config.openshift.io/version` | Controller enable/disable list based on capabilities |
| `featuregates` | `featuregates.config.openshift.io/cluster` | Feature flags (e.g., `BuildCSIVolumes`) |
| `apiserver` (TLS) | `apiservers.config.openshift.io/cluster` | TLS security profile (cipher suites, min TLS version) propagated to operand |

## CRDs and API Types

- **Owned:** `openshiftcontrollermanagers.operator.openshift.io` — the operator's configuration CR (singleton named `cluster`)
- **Consumed:** `clusteroperators.config.openshift.io`, `clusterversions.config.openshift.io`, `proxies.config.openshift.io`, `builds.config.openshift.io`, `images.config.openshift.io`, `networks.config.openshift.io`, `featuregates.config.openshift.io`

## Manifest and Resource Management

- **CVO-managed** (`manifests/`): Operator Deployment, Namespace, ServiceAccount, RBAC, ServiceMonitors, ClusterOperator, NetworkPolicies — applied by the Cluster Version Operator during install/upgrade
- **Operator-applied** (`bindata/assets/`): Operand Namespaces, Deployments, ConfigMaps, RBAC, Services, NetworkPolicies — reconciled by the StaticResourceController and the main sync loop

## Dependencies

| Dependency | Usage |
|-----------|-------|
| `openshift/library-go` | Operator framework: static resource controller, config observer, status aggregation, resource apply, log level controller |
| `openshift/api` | CRD types (`operator/v1`, `config/v1`, `openshiftcontrolplane/v1`) |
| `openshift/client-go` | Typed clients and informers for OpenShift API groups |
| `k8s.io/client-go` | Kubernetes clients, informers, work queues |

## Testing Strategy

- **Unit tests** (`pkg/...`): config observer logic, deployment generation, controller enable/disable based on capabilities
- **E2E tests** (`test/e2e/`): cluster-level validation using Ginkgo — network policy enforcement, operator availability
- **OTE** (`.openshift-tests-extension/`): test extension binary included in the production image for CI/CD integration

## Design Decisions

1. **Two operands, one operator**: The route-controller-manager was split from the openshift-controller-manager but both share the same operator CR and operator binary. This avoids a separate operator for a small set of controllers, but means the sync loop manages two independent Deployments with separate status condition prefixes.

2. **Capability-aware controller disable**: Rather than separate operand images per capability, the operator passes a controller enable/disable list via the operand ConfigMap. The `-controllerName` convention (prefixing with `-`) disables individual controllers without changing the binary.

3. **Rate-limited sync loop**: The main operator uses a token bucket rate limiter (0.05 tokens/sec, burst 4) to avoid hot-looping on rapid config changes. This is in addition to the workqueue's exponential backoff.

4. **Always managed, not removable**: The operator opts out of library-go's "unmanaged" and "removable" states via `management.SetOperatorAlwaysManaged()` and `management.SetOperatorNotRemovable()`. The openshift-controller-manager is a core platform component that must always run.

5. **Informer cache sync before controller start**: All informer caches must be fully synced before any controller starts making decisions. A bug (OCPBUGS-81472) caused transiently incorrect operations when controllers acted on partially-synced caches. The fix adds explicit `WaitForCacheSync` checks with error returns for each informer group (operator, kube, config) before starting any controllers.

6. **TLS security profile propagation**: The operator observes the cluster-wide TLS security profile from `apiservers.config.openshift.io/cluster` and propagates it to the operand configuration via the config observer. When TLS config changes, the operator restarts to pick up the new settings (CNTRLPLANE-2620).

7. **Capability poll-and-restart**: When the Build capability is disabled at startup, the operator skips the build config informer (avoiding a deadlock in the ConfigObserver — OCPBUGS-22956) and polls ClusterVersion every 5 minutes. If the capability becomes enabled, the operator exits to trigger a pod restart with full re-initialization. This is intentional: capabilities can only be enabled post-install (never disabled), and the newly-enabled informers require a clean startup. The [component-selection enhancement](https://github.com/openshift/enhancements/blob/master/enhancements/installer/component-selection.md) keeps capability awareness at the CVO layer and does not prescribe an operator-level reactive pattern.
