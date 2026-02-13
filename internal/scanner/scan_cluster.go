package scanner

import (
	"context"
	"github.com/zeldebro/k8s-resource-rebalancer-operator/internal/worker"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func StartIdleCleanerWorker(clientset *kubernetes.Clientset) {

	log := ctrl.Log.WithName("idle-cleaner-worker")
	log.Info("Worker started")

	for {

		item, shutdown := scan_cluster_go.queue.Get()
		if shutdown {
			log.Info("Queue shutdown received")
			return
		}

		key, ok := item.(string)
		if !ok {
			log.Error(nil, "Invalid queue item")
			scan_cluster_go.queue.Done(item)
			continue
		}

		log.Info("Processing", "podKey", key)

		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			log.Error(nil, "Invalid key format", "key", key)
			scan_cluster_go.queue.Done(item)
			continue
		}

		namespace := parts[0]
		podName := parts[1]
		ctx := context.Background()

		// get pod
		pod, err := clientset.CoreV1().
			Pods(namespace).
			Get(ctx, podName, metav1.GetOptions{})

		if err != nil {
			log.Error(err, "Pod not found", "pod", podName, "namespace", namespace)
			scan_cluster_go.queue.Forget(item)
			scan_cluster_go.queue.Done(item)
			continue
		}

		// standalone pod skip
		if len(pod.OwnerReferences) == 0 {
			log.Info("Standalone pod. Skipping", "pod", podName)
			scan_cluster_go.queue.Forget(item)
			scan_cluster_go.queue.Done(item)
			continue
		}

		scaled := false

		for _, owner := range pod.OwnerReferences {

			if owner.Kind != "ReplicaSet" {
				continue
			}

			rs, err := clientset.AppsV1().
				ReplicaSets(namespace).
				Get(ctx, owner.Name, metav1.GetOptions{})

			if err != nil {
				log.Error(err, "ReplicaSet fetch failed", "replicaset", owner.Name)
				continue
			}

			for _, rsOwner := range rs.OwnerReferences {

				if rsOwner.Kind != "Deployment" {
					continue
				}

				deployName := rsOwner.Name
				log.Info("Scaling deployment", "deployment", deployName)

				deploy, err := clientset.AppsV1().
					Deployments(namespace).
					Get(ctx, deployName, metav1.GetOptions{})

				if err != nil {
					log.Error(err, "Deployment fetch failed", "deployment", deployName)
					scan_cluster_go.queue.AddRateLimited(key)
					scan_cluster_go.queue.Done(item)
					continue
				}

				// already scaled check
				if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas == 0 {
					log.Info("Already scaled", "deployment", deployName)
					scaled = true
					break
				}

				replicas := int32(0)
				deploy.Spec.Replicas = &replicas

				_, err = clientset.AppsV1().
					Deployments(namespace).
					Update(ctx, deploy, metav1.UpdateOptions{})

				if err != nil {
					log.Error(err, "Scaling failed", "deployment", deployName)
					scan_cluster_go.queue.AddRateLimited(key)
					scan_cluster_go.queue.Done(item)
					continue
				}

				log.Info("Scaled deployment to zero", "deployment", deployName)
				scaled = true
				break
			}

			if scaled {
				break
			}
		}

		if !scaled {
			log.Info("No deployment owner found", "pod", podName)
		}

		scan_cluster_go.queue.Forget(item)
		scan_cluster_go.queue.Done(item)
	}
}
