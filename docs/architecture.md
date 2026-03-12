# Architecture

This document describes the internal architecture of the Kubernetes Resource Rebalancer Operator for contributors who want to understand the codebase.

## High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                         │
│                                                              │
│  ┌──────────────┐    ┌──────────────────────────────────┐   │
│  │ User applies  │    │  Resource Rebalancer Operator     │   │
│  │ CR (YAML)     │───▶│                                   │   │
│  └──────────────┘    │  ┌────────────┐                   │   │
│                       │  │ Controller │ (Reconcile Loop)  │   │
│                       │  └─────┬──────┘                   │   │
│                       │        │ starts                    │   │
│                       │        ▼                           │   │
│  ┌──────────────┐    │  ┌────────────┐  idle   ┌───────┐ │   │
│  │metrics-server │◀───│──│  Scanner   │───────▶│ Queue  ││   │
│  └──────────────┘    │  └────────────┘         └───┬───┘ │   │
│                       │                             │      │   │
│                       │                             ▼      │   │
│  ┌──────────────┐    │                        ┌────────┐  │   │
│  │ Deployments   │◀───│────────────────────────│ Worker │  │   │
│  │ (scale → 0)   │    │                        └────────┘  │   │
│  └──────────────┘    └──────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Controller (`internal/controller/`)

The **Reconciler** is the entry point. When a `ResourceRebalancer` CR is created or updated:

1. Validates the spec (namespace, thresholds, cleanup flag)
2. Initializes the Kubernetes + Metrics clientsets
3. Starts the **Scanner** in a goroutine
4. Starts the **Worker** in a goroutine

**Key file**: `resourcerebalancer_controller.go`

### 2. Scanner (`internal/scanner/`)

The **Scanner** runs in an infinite loop:

1. Queries `metrics-server` for pod metrics in the configured namespace
2. Sums CPU (millicores) and Memory (MiB) across all containers in each pod
3. If both CPU and Memory are below the configured thresholds → pod is **idle**
4. Adds idle pod key (`namespace/podName`) to the work queue

**Key file**: `scan_cluster.go`

### 3. Queue (`internal/queue/`)

A **rate-limited work queue** (from `client-go`) that decouples the scanner from the worker:

- Prevents duplicate processing
- Handles retries with backoff
- Thread-safe

**Key file**: `queue.go`

### 4. Worker (`internal/worker/`)

The **Worker** processes items from the queue:

1. Retrieves the pod by namespace/name
2. Checks `OwnerReferences` to find the parent ReplicaSet
3. From ReplicaSet, finds the parent Deployment
4. Scales the Deployment to **zero** replicas
5. Handles errors with rate-limited retries

**Key file**: `idle_worker.go`

### 5. Kube Client (`internal/kube/`)

Utility to create Kubernetes clientsets:

- Tries **in-cluster config** first (when running as a pod)
- Falls back to **local kubeconfig** (for development)
- Creates both standard Kubernetes client and metrics client

**Key file**: `client.go`

## CRD Schema

```yaml
apiVersion: rebalancer.dev/v1
kind: ResourceRebalancer
spec:
  userNamespace: "default"     # Namespace to monitor
  cpuThreshold: 50             # CPU threshold in millicores
  memoryThreshold: 500         # Memory threshold in MiB
  enableCleanup: true          # Enable automatic scale-down
```

## Data Flow

```
CR Applied
    │
    ▼
Controller.Reconcile()
    │
    ├──▶ Validate Spec
    │
    ├──▶ Start Scanner goroutine
    │       │
    │       ▼
    │    metrics-server API
    │       │
    │       ▼
    │    Compare usage vs thresholds
    │       │
    │       ▼ (if idle)
    │    queue.Q.Add(key)
    │
    └──▶ Start Worker goroutine
            │
            ▼
         queue.Q.Get()
            │
            ▼
         Pod → ReplicaSet → Deployment
            │
            ▼
         Scale Deployment to 0
```

## Adding New Features

### Supporting new workload types (e.g., StatefulSets)

1. Add RBAC markers in `resourcerebalancer_controller.go`
2. Extend the worker to check for StatefulSet owner references
3. Run `make manifests` to regenerate RBAC
4. Add tests

### Adding new metrics sources

1. Extend the scanner to query additional metrics
2. Add new threshold fields in `api/v1/resourcerebalancer_types.go`
3. Run `make manifests generate`
4. Update the controller validation logic

### Adding status reporting

1. Define status fields in `api/v1/resourcerebalancer_types.go`
2. Update the controller to write status conditions
3. Run `make manifests generate`

