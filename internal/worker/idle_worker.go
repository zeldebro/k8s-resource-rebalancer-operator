package worker

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"time"
)

// global queue shared between scanner + worker
var queue = workqueue.NewRateLimitingQueue(
	workqueue.DefaultControllerRateLimiter(),
)

// This scans cluster every 30s and pushes idle pods into queue
func ScanCluster(
	metricsClient *metricsclient.Clientset,
	cpuThreshold int64,
	memThreshold int64,
	userNamespace string,
	cleanup bool,
) {

	fmt.Println("Cluster scanner started...")

	for {

		// get metrics ONLY for given namespace
		metrics, err := metricsClient.MetricsV1beta1().
			PodMetricses(userNamespace).
			List(context.TODO(), metav1.ListOptions{})

		if err != nil {
			fmt.Println("Metrics error:", err)
			time.Sleep(20 * time.Second)
			continue
		}

		for _, pod := range metrics.Items {

			ns := pod.Namespace

			// safety: skip system namespaces
			if ns == "kube-system" ||
				ns == "kube-public" ||
				ns == "kube-node-lease" ||
				ns == "local-path-storage" {
				continue
			}

			// ensure only user namespace monitored
			if ns != userNamespace {
				continue
			}

			var totalCPU int64
			var totalMem int64

			// sum container usage
			for _, c := range pod.Containers {
				totalCPU += c.Usage.Cpu().MilliValue()
				totalMem += c.Usage.Memory().Value() / (1024 * 1024) // Mi
			}

			// 🔍 idle detection
			if totalCPU < cpuThreshold && totalMem < memThreshold {

				key := ns + "/" + pod.Name
				fmt.Println("💤 Idle pod detected:", key,
					"CPU:", totalCPU, "m",
					"MEM:", totalMem, "Mi")

				// add to queue only if cleanup enabled
				if cleanup {
					queue.Add(key)
				}
			}
		}

		time.Sleep(30 * time.Second)
	}
}
