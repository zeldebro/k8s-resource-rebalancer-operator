package scanner

import (
	"context"
	"time"

	"github.com/zeldebro/k8s-resource-rebalancer-operator/internal/queue"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
)

func ScanCluster(
	metricsClient *metricsclient.Clientset,
	cpuThreshold int64,
	memThreshold int64,
	userNamespace string,
	cleanup bool,
) {

	log := ctrl.Log.WithName("cluster-scanner")
	log.Info("Cluster scanner started",
		"namespace", userNamespace,
		"cpuThreshold", cpuThreshold,
		"memThreshold", memThreshold,
	)

	for {

		metrics, err := metricsClient.MetricsV1beta1().
			PodMetricses(userNamespace).
			List(context.TODO(), metav1.ListOptions{})

		if err != nil {
			log.Error(err, "Unable to fetch pod metrics")
			time.Sleep(20 * time.Second)
			continue
		}

		for _, pod := range metrics.Items {

			ns := pod.Namespace

			// skip system namespaces
			if ns == "kube-system" ||
				ns == "kube-public" ||
				ns == "kube-node-lease" ||
				ns == "local-path-storage" {
				continue
			}

			if ns != userNamespace {
				continue
			}

			var totalCPU int64
			var totalMem int64

			for _, c := range pod.Containers {
				totalCPU += c.Usage.Cpu().MilliValue()
				totalMem += c.Usage.Memory().Value() / (1024 * 1024)
			}

			if totalCPU < cpuThreshold && totalMem < memThreshold {

				key := ns + "/" + pod.Name

				log.Info("Idle pod detected",
					"pod", pod.Name,
					"namespace", ns,
					"cpu(m)", totalCPU,
					"memory(Mi)", totalMem,
				)

				if cleanup {
					queue.Q.Add(key)
				}
			}
		}

		time.Sleep(30 * time.Second)
	}
}
