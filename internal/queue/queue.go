package queue

import "k8s.io/client-go/util/workqueue"

var Q = workqueue.NewRateLimitingQueue(
	workqueue.DefaultControllerRateLimiter(),
)
