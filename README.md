# **Kubernetes Resource Rebalancer Operator**



This project is a custom Kubernetes operator written in Go using Kubebuilder.

It monitors pod CPU and memory usage and automatically frees cluster resources by scaling down idle workloads.



The main goal of this operator is to avoid situations where cluster resources are blocked by idle pods while new pods remain in pending state due to insufficient CPU or memory.

----------

## **Why this operator is needed**



In many Kubernetes environments (especially shared clusters or Kubeflow setups):

-   Users start pods or notebooks and forget to stop them

-   Idle pods continue reserving CPU and memory

-   New pods fail to schedule due to lack of resources

-   Manual cleanup is risky and time-consuming




Kubernetes TTL can delete pods after a fixed time, but resources remain reserved until TTL expires.

This operator solves that by monitoring usage in real time and cleaning idle workloads earlier.

----------

## **What this operator does**



The operator continuously monitors a selected namespace and:

-   Reads CPU and memory usage using metrics-server

-   Detects idle pods based on threshold values

-   Adds them to an internal work queue

-   Finds the deployment owning that pod

-   Scales the deployment to zero safely

-   Frees cluster resources




Everything runs automatically once the CR is applied.

----------

## **How it works (simple flow)**



User creates CR YAML

→ Operator reads config

→ Monitors pod metrics

→ Detects idle pods

→ Adds to queue

→ Worker processes queue

→ Scales deployment to zero

----------

## **Tech used**

-   Golang

-   Kubebuilder

-   controller-runtime

-   client-go

-   metrics-server

-   Kubernetes CRD

-   Workqueue pattern


----------

## **CRD Example**



Apply this YAML to start monitoring:

```
apiVersion: rebalancer.dev/v1
kind: ResourceRebalancer
metadata:
  name: rebalance-sample
spec:
  userNamespace: "default"
  cpuThreshold: 50
  memoryThreshold: 500
  enableCleanup: true
```

### **Fields**

-   userNamespace → namespace to monitor

-   cpuThreshold → below this CPU = idle

-   memoryThreshold → below this memory = idle

-   enableCleanup → enable automatic scaling


----------

## **How to run locally (kind cluster)**



### **Build image**

```
docker build -t rebalancer:latest .
```

### **Load into kind**

```
kind load docker-image rebalancer:latest
```

### **Install CRD**

```
make install
```

### **Deploy operator**

```
make deploy IMG=rebalancer:latest
```

### **Apply CR**

```
kubectl apply -f config/samples/
```

----------

## **Check logs**

```
kubectl logs -f deploy/k8s-resource-rebalancer-controller-manager -n k8s-resource-rebalancer
```

You will see logs when idle pods are detected and scaled.

----------

## **Example test**



Create test deployment:

```
kubectl run test --image=nginx -n default
```

Keep usage low.

Operator will detect idle pod and scale deployment.

----------

## **What I learned from this project**

-   Writing Kubernetes operators using Kubebuilder

-   Using workqueue and controller pattern

-   Reading metrics-server data

-   Handling ownerReferences (pod → replicaset → deployment)

-   RBAC and cluster permissions

-   Production-style retry and logging


----------

## **Future improvements**

-   Slack alerts

-   GPU idle detection

-   Prometheus metrics

-   Web UI

-   Multi-namespace support


----------

## **Author**



Built as a learning + production-ready DevOps project to demonstrate Kubernetes operator development using Golang.