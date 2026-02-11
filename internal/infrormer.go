package internal

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
func ScanCluster(metricsClient *metricsclient.Clientset, CpuThreshold int64, MemThreshold int64, NamespacePrefix string) {

	fmt.Println("Cluster scanner started...")

	for {
		metrics, err := metricsClient.MetricsV1beta1().
			PodMetricses("").
			List(context.TODO(), metav1.ListOptions{})

		if err != nil {
			fmt.Println("metrics error:", err)
			time.Sleep(20 * time.Second)
			continue
		}

		for _, pod := range metrics.Items {

			var totalCPU int64
			var totalMem int64

			for _, c := range pod.Containers {
				totalCPU += c.Usage.Cpu().MilliValue()
				totalMem += c.Usage.Memory().Value() / (1024 * 1024)
			}

			//  idle condition (tune later)
			if totalCPU < CpuThreshold && totalMem < MemThreshold {

				key := NamespacePrefix + "/" + pod.Name
				fmt.Println("Idle pod detected:", key)

				// add to queue
				queue.Add(key)
			}
		}

		time.Sleep(30 * time.Second)
	}
}
