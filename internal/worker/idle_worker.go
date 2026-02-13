package worker

import (
	"context"
	"strings"

	"github.com/zeldebro/k8s-resource-rebalancer-operator/internal/queue"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func StartIdleCleanerWorker(clientset *kubernetes.Clientset) {

	log := ctrl.Log.WithName("idle-cleaner")
	log.Info("Idle cleaner worker started")

	for {

		item, shutdown := queue.Q.Get()
		if shutdown {
			log.Info("Queue shutdown received")
			return
		}

		key, ok := item.(string)
		if !ok {
			log.Error(nil, "Invalid queue item")
			queue.Q.Done(item)
			continue
		}

		log.Info("Processing pod", "key", key)

		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			log.Error(nil, "Invalid key format", "key", key)
			queue.Q.Done(item)
			continue
		}

		namespace := parts[0]
		podName := parts[1]
		ctx := context.Background()

		pod, err := clientset.CoreV1().
			Pods(namespace).
			Get(ctx, podName, metav1.GetOptions{})

		if err != nil {
			log.Error(err, "Pod not found", "pod", podName, "namespace", namespace)
			queue.Q.Forget(item)
			queue.Q.Done(item)
			continue
		}

		if len(pod.OwnerReferences) == 0 {
			log.Info("Standalone pod detected, skipping", "pod", podName)
			queue.Q.Forget(item)
			queue.Q.Done(item)
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
				log.Error(err, "Failed fetching ReplicaSet", "rs", owner.Name)
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
					queue.Q.AddRateLimited(key)
					queue.Q.Done(item)
					continue
				}

				if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas == 0 {
					log.Info("Deployment already scaled", "deployment", deployName)
					scaled = true
					break
				}

				replicas := int32(0)
				deploy.Spec.Replicas = &replicas

				_, err = clientset.AppsV1().
					Deployments(namespace).
					Update(ctx, deploy, metav1.UpdateOptions{})

				if err != nil {
					log.Error(err, "Failed scaling deployment", "deployment", deployName)
					queue.Q.AddRateLimited(key)
					queue.Q.Done(item)
					continue
				}

				log.Info("Deployment scaled to zero", "deployment", deployName)
				scaled = true
				break
			}

			if scaled {
				break
			}
		}

		if !scaled {
			log.Info("No deployment owner found for pod", "pod", podName)
		}

		queue.Q.Forget(item)
		queue.Q.Done(item)
	}
}
