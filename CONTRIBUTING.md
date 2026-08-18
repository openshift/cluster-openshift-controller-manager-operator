# Contributing to cluster-openshift-controller-manager-operator

## Development Workflow

1. Fork the repo and clone your fork
2. Create a feature branch from `master`
3. Make your changes, add or update tests
4. Run verification locally:
   ```bash
   make build verify test
   ```
5. If you changed dependencies: `go mod tidy && go mod vendor`, commit separately from code changes
6. Push your branch and open a PR

## Pull Request Guidelines

- Keep PRs focused — one logical change per PR
- Reference JIRA tickets in PR title: `OCPBUGS-XXXXX: description` or `CNTRLPLANE-XXXXX: description`
- Include tests for new functionality
- PRs require `/lgtm` from a reviewer and `/approve` from an approver (see OWNERS)
- All PRs require `/verified` before merge (see OpenShift verification workflow)

## Testing

| Command | What It Runs |
|---------|-------------|
| `make build` | Builds the operator and OTE test binaries |
| `make test` | Unit tests in `pkg/...` and `cmd/...` |
| `make verify` | Linting, formatting, vet checks |
| `make test-e2e` | E2E tests against a live cluster (requires `KUBECONFIG`) |

### E2E Tests

E2E tests use Ginkgo and require a running OpenShift cluster. They validate network policy enforcement and operator health.

```bash
make test-e2e
```

## Code Conventions

- **Error handling**: Use `fmt.Errorf` with `%w` for wrapping. Accumulate errors in slices when syncing multiple resources, then report all at once.
- **Status conditions**: Use library-go `v1helpers.SetOperatorCondition` helpers. Prefix condition types with `RouteControllerManager` for route-controller-manager conditions to distinguish from openshift-controller-manager conditions.
- **Config observation**: Each observer returns `(observedConfig, errors)`. On error, return the previous config to avoid flapping.
- **Static resources**: RBAC and namespace manifests go in `bindata/assets/`. Network policy allow rules must be listed before default-deny rules in the StaticResourceController to avoid traffic interruption during reconciliation.

## Areas Requiring Extra Care

- **`bindata/assets/`**: Changes to static manifests affect what the operator reconciles. Network policy ordering matters (allow before deny).
- **`manifests/`**: CVO-managed resources. Changes here affect install/upgrade — coordinate with the release team.
- **Capability integration**: Adding or removing controllers from the capability-disable lists affects which controllers run on clusters with reduced capability sets. Test with capabilities both enabled and disabled.
- **Vendor directory**: Committed. Run `go mod tidy && go mod vendor` and commit separately from code changes.

## CI Pipeline

CI is managed by Prow via [ci-operator](https://docs.ci.openshift.org/docs/architecture/ci-operator/). The build root image is configured in `.ci-operator.yaml`. Required checks include unit tests, build verification, and e2e tests on a live cluster.

## Review and Approval

Reviews and approvals follow the [OWNERS](OWNERS) file. The primary approver groups are `control-plane-approvers` and build-related maintainers. See [OWNERS_ALIASES](OWNERS_ALIASES) for the full list of control-plane approvers.

## Rebase Checklist

When rebasing to a new Kubernetes release:

- [ ] Reference the target [kubernetes release branch](https://github.com/kubernetes/kubernetes/branches) `go.mod` and `CHANGELOG`
- [ ] Bump Go version, all `k8s.io/`, `github.com/openshift/`, and other relevant dependencies
- [ ] Run `go mod vendor && go mod tidy`, commit separately
- [ ] Bump image versions (Dockerfile, `.ci-operator.yaml`) if needed
- [ ] Run `make build verify test`
- [ ] Make code changes as needed until the above pass
