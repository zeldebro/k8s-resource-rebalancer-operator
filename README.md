# Kubernetes Resource Rebalancer Operator

[![Go Version](https://img.shields.io/github/go-mod/go-version/zeldebro/k8s-resource-rebalancer-operator)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![CI](https://github.com/zeldebro/k8s-resource-rebalancer-operator/actions/workflows/ci.yml/badge.svg)](https://github.com/zeldebro/k8s-resource-rebalancer-operator/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zeldebro/k8s-resource-rebalancer-operator)](https://goreportcard.com/report/github.com/zeldebro/k8s-resource-rebalancer-operator)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](CONTRIBUTING.md)

> A Kubernetes operator that monitors pod resource usage in real time and automatically scales down idle workloads to free cluster resources.

---

## The Problem

In shared Kubernetes clusters (especially Kubeflow, dev environments, and multi-tenant setups):

- 😴 Users start pods/notebooks and **forget to stop them**
- 🔒 Idle pods continue **reserving CPU and memory**
- ❌ New pods stay in **Pending** state due to insufficient resources
- ⏰ Kubernetes TTL controllers can delete pods after a fixed time, but resources remain **blocked until TTL expires**
- 🔧 Manual cleanup is **risky and time-consuming**

## The Solution

This operator **watches pod metrics in real time** and automatically scales down deployments whose pods are idle — freeing resources immediately for workloads that actually need them.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                        │
│                                                               │
│  ┌──────────┐     ┌──────────────────────────────────────┐   │
│  │  User     │     │   Resource Rebalancer Operator        │   │
│  │  applies  │────▶│                                       │   │
│  │  CR YAML  │     │   ┌────────────┐                     │   │
│  └──────────┘     │   │ Controller  │ (Reconcile Loop)    │   │
│                    │   └──────┬─────┘                     │   │
│                    │          │ starts                      │   │
│                    │          ▼                             │   │
│  ┌─────────────┐  │   ┌──────────┐  idle   ┌──────────┐  │   │
│  │metrics-server│◀─┤   │ Scanner  │───────▶│  Queue   │  │   │
│  └─────────────┘  │   └──────────┘         └────┬─────┘  │   │
│                    │                             │         │   │
│                    │                             ▼         │   │
│  ┌─────────────┐  │                       ┌──────────┐    │   │
│  │ Deployments │◀─┤───────────────────────│  Worker  │    │   │
│  │ (scale → 0) │  │                       └──────────┘    │   │
│  └─────────────┘  └──────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
```

**Flow**: CR Applied → Controller validates config → Scanner polls metrics-server → Idle pods detected → Added to rate-limited queue → Worker finds owner Deployment → Scales to zero

> 📖 For a deep dive, see [docs/architecture.md](docs/architecture.md)

---

## Features

- ✅ **Real-time monitoring** — Polls metrics-server continuously (not time-based TTL)
- ✅ **Configurable thresholds** — Set CPU and memory idle thresholds per namespace
- ✅ **Safe scale-down** — Traces Pod → ReplicaSet → Deployment ownership chain before scaling
- ✅ **Rate-limited queue** — Prevents thundering herd; handles retries with backoff
- ✅ **CRD-driven** — Fully declarative configuration via Kubernetes Custom Resource
- ✅ **Namespace-scoped** — Monitor specific namespaces, skip system namespaces automatically
- ✅ **Minimal RBAC** — Principle of least privilege

---

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.28+)
- [metrics-server](https://github.com/kubernetes-sigs/metrics-server) installed
- `kubectl` configured

### Install

```bash
# Install CRDs
kubectl apply -f https://raw.githubusercontent.com/zeldebro/k8s-resource-rebalancer-operator/main/config/crd/bases/rebalancer.dev_resourcerebalancers.yaml

# Deploy the operator
kubectl apply -f https://raw.githubusercontent.com/zeldebro/k8s-resource-rebalancer-operator/main/dist/install.yaml
```

### Configure

Create a `ResourceRebalancer` CR to start monitoring:

```yaml
apiVersion: rebalancer.dev/v1
kind: ResourceRebalancer
metadata:
  name: rebalance-sample
spec:
  userNamespace: "default"      # Namespace to monitor
  cpuThreshold: 50              # CPU threshold in millicores (below = idle)
  memoryThreshold: 500          # Memory threshold in MiB (below = idle)
  enableCleanup: true           # Enable automatic scale-down
```

```bash
kubectl apply -f config/samples/smart_v1_resourcerebalancer.yaml
```

### Verify

```bash
# Check operator logs
kubectl logs -f deploy/k8s-resource-rebalancer-controller-manager \
  -n k8s-resource-rebalancer-operator-system

# Check CR status
kubectl get resourcerebalancers
```

---

## Local Development

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.25+ | [go.dev/dl](https://go.dev/dl/) |
| Docker | Latest | [docs.docker.com](https://docs.docker.com/get-docker/) |
| Kind | Latest | [kind.sigs.k8s.io](https://kind.sigs.k8s.io/) |
| kubectl | Latest | [kubernetes.io](https://kubernetes.io/docs/tasks/tools/) |

### Setup

```bash
# Clone the repo
git clone https://github.com/zeldebro/k8s-resource-rebalancer-operator.git
cd k8s-resource-rebalancer-operator

# Create Kind cluster
kind create cluster --name rebalancer-dev

# Install metrics-server (required)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl patch deployment metrics-server -n kube-system \
  --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

# Install CRDs
make install

# Run operator locally
make run
```

### Build & Deploy to Kind

```bash
# Build Docker image
docker build -t rebalancer:latest .

# Load into Kind
kind load docker-image rebalancer:latest --name rebalancer-dev

# Deploy
make deploy IMG=rebalancer:latest

# Apply sample CR
kubectl apply -f config/samples/smart_v1_resourcerebalancer.yaml
```

### Run Tests

```bash
make test          # Unit tests
make lint          # Linting
make test-e2e      # End-to-end tests (creates a Kind cluster)
```

---

## Configuration Reference

### CRD Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userNamespace` | `string` | ✅ | Kubernetes namespace to monitor |
| `cpuThreshold` | `int64` | ✅ | CPU usage threshold in millicores. Pods below this are considered idle |
| `memoryThreshold` | `int64` | ✅ | Memory usage threshold in MiB. Pods below this are considered idle |
| `enableCleanup` | `bool` | ✅ | Set to `true` to enable automatic scale-down of idle deployments |

---

## Tech Stack

| Technology | Purpose |
|------------|---------|
| [Go](https://go.dev/) | Language |
| [Kubebuilder](https://book.kubebuilder.io/) | Operator framework / scaffolding |
| [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) | Controller lifecycle, reconciliation |
| [client-go](https://github.com/kubernetes/client-go) | Kubernetes API client, workqueue |
| [metrics-server](https://github.com/kubernetes-sigs/metrics-server) | Real-time pod resource metrics |

---

## Roadmap

We welcome contributions for any of these! Check [issues](https://github.com/zeldebro/k8s-resource-rebalancer-operator/issues) for the latest.

- [ ] **StatefulSet support** — Detect and handle idle StatefulSets
- [ ] **DaemonSet awareness** — Skip DaemonSet-managed pods
- [ ] **Configurable scan interval** — Allow users to set polling frequency
- [ ] **Dry-run mode** — Log what would be scaled down without taking action
- [ ] **Status reporting** — Write idle pod count and last action to CR status
- [ ] **Multi-namespace support** — Monitor multiple namespaces from a single CR
- [ ] **Webhook validation** — Validate CR spec with an admission webhook
- [ ] **Prometheus metrics** — Expose idle pod count, scale-down events as metrics
- [ ] **Grace period** — Wait N minutes before scaling down (avoid premature cleanup)
- [ ] **Notification support** — Send alerts (Slack, webhook) before scaling down
- [ ] **Helm chart** — Package operator for easy installation via Helm
- [ ] **Pod annotations** — Allow users to opt-out specific pods from cleanup

---

## Contributing

We love contributions! Whether it's a bug fix, feature, documentation improvement, or just a typo fix — every contribution matters.

👉 **[Read the Contributing Guide](CONTRIBUTING.md)** to get started.

### Good First Issues

Look for issues labeled [`good first issue`](https://github.com/zeldebro/k8s-resource-rebalancer-operator/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) — these are great for newcomers!

### Quick Links

| Resource | Link |
|----------|------|
| 🐛 Report a bug | [Open an issue](https://github.com/zeldebro/k8s-resource-rebalancer-operator/issues/new?template=bug_report.md) |
| ✨ Request a feature | [Open an issue](https://github.com/zeldebro/k8s-resource-rebalancer-operator/issues/new?template=feature_request.md) |
| 📖 Contributing guide | [CONTRIBUTING.md](CONTRIBUTING.md) |
| 🏗️ Architecture docs | [docs/architecture.md](docs/architecture.md) |
| 📜 Code of Conduct | [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| 🔒 Security policy | [SECURITY.md](SECURITY.md) |

---

## License

This project is licensed under the Apache License 2.0 — see the [LICENSE](LICENSE) file for details.

---

## Star History

If you find this project useful, please consider giving it a ⭐ — it helps others discover it!

