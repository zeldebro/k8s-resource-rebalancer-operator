Below is your complete professional README.md (GitHub ready + interview ready).
You can copy-paste directly.

⸻

🚀 Kubernetes Resource Rebalancer Operator

A production-ready Kubernetes Operator built using Golang + Kubebuilder that automatically detects idle pods consuming CPU and memory and safely frees cluster resources.

This operator continuously monitors workloads and rebalances cluster usage to prevent Pending pods due to insufficient CPU/memory.

Perfect for:
•	Kubeflow environments
•	ML/AI workloads
•	Multi-tenant clusters
•	Shared Kubernetes platforms

⸻

📌 Problem Statement

In many Kubernetes clusters:
•	Users start notebook or application pods and forget to stop them
•	Idle pods keep reserving CPU and memory
•	New pods fail scheduling due to Insufficient CPU/Memory
•	Cluster resources remain blocked
•	Manual cleanup is risky and time-consuming

❗ Why TTL is not enough?

Kubernetes supports TTL (Time To Live) for pods/jobs.

👉 TTL meaning

TTL = Time To Live
It defines how long a pod/job should live before automatic deletion.

Example:
If TTL = 24 hours
Pod will run full 24 hours even if idle
Resources stay reserved until TTL expires

Problem with TTL

Even if TTL exists:
•	Pod continues running until TTL finishes
•	Resources remain blocked during that time
•	New workloads still go Pending
•	Cluster utilization becomes inefficient

TTL frees resources late
But cluster needs resources immediately

⸻

💡 Solution (What this operator does)

This operator solves the problem by working in real-time.

It continuously monitors only the user-defined namespace and:
1.	Monitors CPU & memory usage using metrics-server
2.	Detects idle pods (below threshold usage)
3.	Adds them to internal workqueue
4.	Safely scales down deployments
5.	Frees cluster resources immediately
6.	Prevents Pending pod issues

Unlike TTL-based cleanup, this operator:
•	Works in real-time
•	Frees resources early
•	Improves scheduling efficiency
•	Supports multi-tenant namespace filtering

⸻

🧠 Architecture

CR (User YAML)
↓
Controller (Reconcile loop)
↓
Metrics Scanner
↓
Workqueue (FIFO)
↓
Worker
↓
Scale deployment → Free resources

⸻

✨ Features
•	Real-time CPU & memory monitoring
•	Automatic idle workload detection
•	Namespace-based filtering
•	Workqueue-based safe processing
•	Retry & backoff handling
•	Production-ready logging
•	Scales deployments safely
•	Prevents cluster starvation
•	Works on Kind / EKS / AKS / GKE

⸻

📦 Example Use Case (Kubeflow)

Kubeflow notebook pods run for long hours.

Problem:
•	Users forget to stop notebooks
•	Idle notebooks consume resources
•	New training jobs stay Pending

Solution:
•	Operator detects low CPU usage notebooks
•	Scales them safely
•	Frees CPU & memory
•	New jobs schedule successfully

⸻

🛠 Tech Stack
•	Golang
•	Kubebuilder
•	controller-runtime
•	client-go
•	metrics-server API
•	Kubernetes CRD
•	Workqueue pattern

⸻

⚙️ CRD Example

User deploys operator using this YAML:

apiVersion: rebalancer.dev/v1
kind: ResourceRebalancer
metadata:
name: smart-rebalancer
spec:
userNamespace: "default"
cpuThreshold: 50
memoryThreshold: 500
enableCleanup: true

Parameters

Field	Description
userNamespace	Namespace to monitor
cpuThreshold	CPU below this = idle
memoryThreshold	Memory below this = idle
enableCleanup	Enable auto scaling


⸻

🚀 Getting Started

Prerequisites
•	Go 1.22+
•	Docker
•	Kubernetes cluster (kind/minikube/EKS)
•	metrics-server installed

⸻

🐳 Build Docker Image

docker build -t rebalancer:latest .

For kind cluster:

kind load docker-image rebalancer:latest


⸻

📦 Install CRD

make install


⸻

🚀 Deploy Operator

make deploy IMG=rebalancer:latest


⸻

📄 Apply Custom Resource

kubectl apply -f config/samples/


⸻

📊 Check Logs

kubectl logs -f deploy/k8s-resource-rebalancer-controller-manager -n k8s-resource-rebalancer

You should see:

Cluster scanner started
Idle pod detected
Processing pod
Scaling deployment


⸻

🧪 Test Scenario

Create test workload:

kubectl run test --image=nginx -n default

Keep CPU usage low → operator detects idle → deployment scaled.


🤝 Contributing

PRs welcome.


⭐ If this helped you

Give a ⭐ on GitHub
and connect on LinkedIn.

⸻

📜 License

Apache 2.0 License
Copyright 2026