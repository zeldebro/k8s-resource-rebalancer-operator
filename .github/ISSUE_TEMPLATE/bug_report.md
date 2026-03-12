---
name: Bug Report
about: Report a bug to help us improve
title: "[BUG] "
labels: bug
assignees: ''
---

## Describe the Bug

A clear and concise description of what the bug is.

## To Reproduce

Steps to reproduce the behavior:

1. Apply CR with config '...'
2. Deploy workload '...'
3. Wait for '...'
4. See error

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

What actually happened instead.

## Logs

<details>
<summary>Operator Logs</summary>

```
Paste operator logs here:
kubectl logs -f deploy/k8s-resource-rebalancer-controller-manager -n k8s-resource-rebalancer-operator-system
```

</details>

## Environment

- **Kubernetes Version**: (e.g., 1.31)
- **Operator Version/Commit**: (e.g., v0.1.0 or commit SHA)
- **Go Version**: (e.g., 1.25.3)
- **Cluster Type**: (e.g., Kind, Minikube, EKS, GKE, AKS)
- **metrics-server installed**: Yes/No
- **OS**: (e.g., macOS, Linux)

## CR Configuration

```yaml
# Paste your ResourceRebalancer CR here
```

## Additional Context

Add any other context about the problem here.

