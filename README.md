🚀 Kubernetes Resource Rebalancer Operator

A production-ready Kubernetes operator built with Kubebuilder + Golang that automatically detects idle pods consuming CPU/memory and rebalances cluster resources by cleaning them safely.

This operator helps prevent cluster resource starvation and improves scheduling efficiency — especially useful for Kubeflow, ML workloads, and multi-tenant clusters.

⸻

📌 Problem Statement

In many Kubernetes clusters (especially ML/Kubeflow environments):
•	Users leave notebook pods running after work
•	Idle pods keep reserving CPU & memory
•	New pods go Pending (Insufficient CPU/Memory)
•	Cluster resources remain blocked

Manual cleanup is painful.

⸻

💡 Solution (What this operator does)

This operator automatically:
1.	Monitors cluster resource usage (CPU + memory)
2.	Detects idle pods (below threshold usage)
3.	Adds them to internal work queue
4.	Safely scales or deletes idle workloads
5.	Frees resources for pending pods
6.	Improves cluster utilization automatically

All logic runs continuously using:
•	Informers
•	Workqueue
•	Controller-runtime
•	Metrics-server API

⸻

🧠 Architecture

CRD (User YAML)
↓
Controller (Reconcile loop)
↓
Operator Logic
↓
Metrics API → Detect idle pods
↓
Queue (FIFO)
↓
Worker → Clean resources


⸻

✨ Features
•	⚡ Real-time pod monitoring using Informers
•	📊 CPU & memory based cleanup
•	🔁 Workqueue based safe processing
•	🧠 Smart namespace filtering
•	🛡 Production-safe retry logic
•	☁ Works with Kind / EKS / AKS / GKE
•	🔥 Built fully in Golang

⸻

📦 Example Use Case

Kubeflow notebook environment:
•	Notebook pods run for 24 hours
•	Users forget to stop notebooks
•	Cluster becomes full
•	New training jobs stay Pending

👉 Operator detects low CPU usage notebooks
👉 Deletes/scales them safely
👉 New jobs get scheduled

⸻

🛠 Tech Stack
•	Golang
•	Kubebuilder
•	client-go
•	controller-runtime
•	metrics-server
•	Kubernetes CRD
•	Workqueue pattern

⸻

⚙️ CRD Example

User deploys operator using this YAML:

apiVersion: rebalancer.dev/v1
kind: ResourceRebalancer
metadata:
name: rebalance-sample
spec:
userNamespace: "mc-"
cpuThreshold: 50
memoryThreshold: 500
enableCleanup: true

Parameters

Field	Description
userNamespace	Monitor only these namespaces
cpuThreshold	CPU usage below this = idle
memoryThreshold	Memory below this = idle
enableCleanup	Enable auto cleanup


⸻

🚀 Getting Started

Prerequisites
•	Go 1.22+
•	Docker
•	Kubernetes cluster (kind/minikube/EKS)
•	metrics-server installed

⸻

🐳 Build Docker Image

make docker-build docker-push IMG=<dockerhub-user>/rebalancer:latest

For local kind cluster:

docker build -t rebalancer:latest .
kind load docker-image rebalancer:latest


⸻

📦 Install CRD

make install


⸻

🚀 Deploy Operator

make deploy IMG=rebalancer:latest


⸻

📄 Apply CR

kubectl apply -f config/samples/


⸻

📊 Check Logs

kubectl logs -f deploy/k8s-resource-rebalancer-operator-controller-manager -n k8s-resource-rebalancer

You should see:

Idle pod detected
Added to queue
Cleaning resources


⸻

🧪 Test Scenario

Create dummy pod:

kubectl run test --image=nginx --requests='cpu=500m,memory=512Mi'

Reduce usage → operator detects idle → cleanup triggered.

⸻

🧠 Interview Ready Explanation

If interviewer asks:

What does your operator do?

Answer:

I built a Kubernetes operator using Golang and Kubebuilder that monitors cluster resource usage in real-time.
It detects idle workloads consuming CPU/memory and automatically rebalances cluster resources using a workqueue-based controller.
This prevents pod scheduling failures and improves cluster efficiency in multi-tenant environments like Kubeflow.

⸻

🔥 Why this project is strong

This project demonstrates:
•	Kubernetes operator development
•	CRD design
•	Controller-runtime
•	Informers & Workqueue
•	Production-grade logic
•	Golang concurrency
•	Metrics API usage

👉 Enough to clear any DevOps/SRE interview

⸻

🛣 Roadmap

Future improvements:
•	Slack/Email alerts
•	Auto scaling integration
•	GPU idle detection
•	Web UI dashboard
•	Multi-cluster support

⸻

🤝 Contributing

PRs welcome.

If you want to improve:
•	add GPU cleanup
•	add Slack alerts
•	add Prometheus metrics

⸻

⭐ If this helped you

Give a ⭐ on GitHub and connect with me on LinkedIn.

⸻

📜 License

Apache 2.0 License
Copyright 2026