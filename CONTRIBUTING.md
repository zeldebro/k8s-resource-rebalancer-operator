# Contributing to Kubernetes Resource Rebalancer Operator

First off, **thank you** for considering contributing! 🎉

This project is open source and we welcome contributions of all kinds — bug fixes, new features, documentation improvements, test coverage, and more.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Your First Code Contribution](#your-first-code-contribution)
  - [Pull Requests](#pull-requests)
- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites)
  - [Local Development](#local-development)
  - [Running Tests](#running-tests)
- [Project Structure](#project-structure)
- [Coding Guidelines](#coding-guidelines)
- [Commit Messages](#commit-messages)
- [Community](#community)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior by opening an issue.

---

## How Can I Contribute?

### Reporting Bugs

If you find a bug, please [open an issue](https://github.com/zeldebro/k8s-resource-rebalancer-operator/issues/new?template=bug_report.md) with:

- A clear title and description
- Steps to reproduce
- Expected vs. actual behavior
- Kubernetes version, Go version, and environment details
- Relevant logs (operator logs, `kubectl describe` output)

### Suggesting Features

Have an idea? [Open a feature request](https://github.com/zeldebro/k8s-resource-rebalancer-operator/issues/new?template=feature_request.md) with:

- A clear description of the problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

### Your First Code Contribution

Not sure where to start? Look for issues labeled:

- 🏷️ **`good first issue`** — Great for newcomers
- 🏷️ **`help wanted`** — Looking for community help
- 🏷️ **`documentation`** — Improve docs, README, or code comments

### Pull Requests

1. **Fork** the repository
2. **Clone** your fork:
   ```bash
   git clone https://github.com/<your-username>/k8s-resource-rebalancer-operator.git
   cd k8s-resource-rebalancer-operator
   ```
3. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/my-awesome-feature
   ```
4. **Make your changes** (see [Development Setup](#development-setup))
5. **Run tests and linting**:
   ```bash
   make test
   make lint
   ```
6. **Commit** with a meaningful message (see [Commit Messages](#commit-messages))
7. **Push** to your fork:
   ```bash
   git push origin feat/my-awesome-feature
   ```
8. **Open a Pull Request** against `main`

---

## Development Setup

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.25+ | Language runtime |
| [Docker](https://docs.docker.com/get-docker/) | Latest | Building container images |
| [Kind](https://kind.sigs.k8s.io/) | Latest | Local Kubernetes cluster |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | Latest | Kubernetes CLI |
| [metrics-server](https://github.com/kubernetes-sigs/metrics-server) | Latest | Required for pod metrics |

### Local Development

```bash
# 1. Clone and enter the project
git clone https://github.com/zeldebro/k8s-resource-rebalancer-operator.git
cd k8s-resource-rebalancer-operator

# 2. Install dependencies
go mod download

# 3. Generate manifests and code
make manifests generate

# 4. Create a local Kind cluster
kind create cluster --name rebalancer-dev

# 5. Install metrics-server (required)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
# For Kind, you may need to patch metrics-server:
kubectl patch deployment metrics-server -n kube-system \
  --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

# 6. Install CRDs
make install

# 7. Run the operator locally (outside the cluster)
make run

# 8. In another terminal, apply a sample CR
kubectl apply -f config/samples/smart_v1_resourcerebalancer.yaml
```

### Running Tests

```bash
# Unit tests
make test

# Linting
make lint

# Fix lint issues automatically
make lint-fix

# End-to-end tests (requires Kind)
make test-e2e
```

---

## Project Structure

```
├── api/v1/                     # CRD type definitions (ResourceRebalancer)
├── cmd/main.go                 # Operator entry point
├── internal/
│   ├── controller/             # Reconciliation logic (main operator loop)
│   ├── kube/                   # Kubernetes client setup
│   ├── queue/                  # Internal work queue for idle pods
│   ├── scanner/                # Cluster scanner (reads metrics, detects idle pods)
│   └── worker/                 # Worker that processes queue & scales down deployments
├── config/
│   ├── crd/                    # Generated CRD manifests (DO NOT EDIT)
│   ├── rbac/                   # Generated RBAC roles (DO NOT EDIT)
│   ├── manager/                # Operator deployment manifests
│   └── samples/                # Example CR YAML files
├── test/                       # E2E test suite
├── Dockerfile                  # Multi-stage container build
├── Makefile                    # Build, test, deploy commands
└── PROJECT                     # Kubebuilder metadata
```

### Key Components

| Component | File | Description |
|-----------|------|-------------|
| **Controller** | `internal/controller/resourcerebalancer_controller.go` | Watches `ResourceRebalancer` CRs, validates spec, starts scanner + worker |
| **Scanner** | `internal/scanner/scan_cluster.go` | Polls metrics-server, detects idle pods below CPU/memory thresholds |
| **Queue** | `internal/queue/queue.go` | Rate-limited work queue bridging scanner → worker |
| **Worker** | `internal/worker/idle_worker.go` | Processes queued pods, finds owner Deployment, scales to zero |
| **Kube Client** | `internal/kube/client.go` | Initializes Kubernetes + metrics clientsets |

---

## Coding Guidelines

- **Language**: Go — follow [Effective Go](https://go.dev/doc/effective_go) and standard library conventions
- **Formatting**: Run `go fmt ./...` or `make fmt` before committing
- **Linting**: Run `make lint` — all code must pass `golangci-lint`
- **Testing**: Add unit tests for new functionality. Place tests next to the code they test
- **Error handling**: Always handle errors explicitly. Use structured logging via `controller-runtime/pkg/log`
- **CRD changes**: After modifying `api/v1/*_types.go`, always run:
  ```bash
  make manifests generate
  ```
- **Do NOT edit auto-generated files**:
  - `config/crd/bases/*.yaml`
  - `config/rbac/role.yaml`
  - `**/zz_generated.*.go`

---

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `test` | Adding or updating tests |
| `refactor` | Code refactoring (no feature/fix) |
| `chore` | Build, CI, tooling changes |
| `perf` | Performance improvement |

### Examples

```
feat(scanner): add configurable scan interval
fix(worker): handle orphaned ReplicaSets gracefully
docs: add architecture diagram to README
test(controller): add reconciliation edge case tests
chore(ci): add GitHub Actions workflow
```

---

## Community

- 💬 **Discussions**: Use [GitHub Discussions](https://github.com/zeldebro/k8s-resource-rebalancer-operator/discussions) for questions and ideas
- 🐛 **Issues**: [GitHub Issues](https://github.com/zeldebro/k8s-resource-rebalancer-operator/issues) for bugs and feature requests
- ⭐ **Star** the repo if you find it useful!

---

Thank you for helping make Kubernetes resource management better! 🚀

