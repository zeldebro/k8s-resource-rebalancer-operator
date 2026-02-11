package internal

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"strings"
)

func StartIdleCleanerWorker(clientset *kubernetes.Clientset) {

	fmt.Println("Idle cleaner worker started")

	for {

		item, shutdown := queue.Get()
		if shutdown {
			fmt.Println("Queue shutdown")
			return
		}

		key := item.(string)
		fmt.Println("Processing:", key)

		// FIX: correct split
		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			fmt.Println("Invalid key:", key)
			queue.Done(item)
			continue
		}

		namespace := parts[0]
		podName := parts[1]

		ctx := context.Background()

		// 1 get pod
		pod, err := clientset.CoreV1().
			Pods(namespace).
			Get(ctx, podName, metav1.GetOptions{})

		if err != nil {
			fmt.Println("Pod not found:", err)
			queue.Done(item)
			continue
		}

		// 2 find ReplicaSet
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" {

				rs, err := clientset.AppsV1().
					ReplicaSets(namespace).
					Get(ctx, owner.Name, metav1.GetOptions{})
				if err != nil {
					continue
				}

				// 3 find deployment
				for _, rsOwner := range rs.OwnerReferences {
					if rsOwner.Kind == "Deployment" {

						deployName := rsOwner.Name
						fmt.Println("Scaling deployment:", deployName)

						deploy, err := clientset.AppsV1().
							Deployments(namespace).
							Get(ctx, deployName, metav1.GetOptions{})

						if err != nil {
							fmt.Println("Deployment fetch error:", err)
							queue.AddRateLimited(key)
							queue.Done(item)
							continue
						}

						// production action → scale to zero
						replicas := int32(0)
						deploy.Spec.Replicas = &replicas

						_, err = clientset.AppsV1().
							Deployments(namespace).
							Update(ctx, deploy, metav1.UpdateOptions{})

						if err != nil {
							fmt.Println("Scale failed:", err)
							queue.AddRateLimited(key)
							queue.Done(item)
							continue
						}

						fmt.Println("Scaled to zero:", deployName)
					}
				}
			}
		}

		queue.Forget(item)
		queue.Done(item)
	}
}
