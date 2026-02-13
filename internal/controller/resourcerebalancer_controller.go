/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"github.com/zeldebro/k8s-resource-rebalancer-operator/internal/scanner"
	"github.com/zeldebro/k8s-resource-rebalancer-operator/internal/worker"
	"k8s.io/client-go/kubernetes"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	smartv1 "github.com/zeldebro/k8s-resource-rebalancer-operator/api/v1"
)

// ResourceRebalancerReconciler reconciles a ResourceRebalancer object
type ResourceRebalancerReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Clientset     *kubernetes.Clientset
	MetricsClient *metricsclient.Clientset
}

// POD permission
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// ReplicaSet permission (VERY IMPORTANT)
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch

// Deployment permission (for scaling)
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch

// Metrics permission
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list;watch

// +kubebuilder:rbac:groups=rebalancer.dev,resources=resourcerebalancers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rebalancer.dev,resources=resourcerebalancers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rebalancer.dev,resources=resourcerebalancers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete;patch;update
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceRebalancer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *ResourceRebalancerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)
	var resourceRebalancer smartv1.ResourceRebalancer
	if err := r.Get(ctx, req.NamespacedName, &resourceRebalancer); err != nil {
		// handle not found error (e.g., resource was deleted)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//Validating Namespace
	if resourceRebalancer.Spec.UserNamespace == "" {
		logf.Log.Error(nil, "userNamespace is required in ResourceRebalancer spec")
		return ctrl.Result{}, nil
	}
	// validating CPU thresholds
	if resourceRebalancer.Spec.CpuThreshold <= 0 {
		logf.Log.Error(nil, "CpuThreshold is required in ResourceRebalancer spec")
		return ctrl.Result{}, nil
	}
	// validating Memory thresholds
	if resourceRebalancer.Spec.MemoryThreshold <= 0 {
		logf.Log.Error(nil, "MemoryThreshold is required in ResourceRebalancer spec")
		return ctrl.Result{}, nil
	}
	// validating EnableCleanup
	if resourceRebalancer.Spec.EnableCleanup == false {
		logf.Log.Error(nil, "EnableCleanup is required in ResourceRebalancer spec")
		return ctrl.Result{}, nil
	}

	logf.Log.V(1).Info("CR loaded successfully")
	// start operator logic here (e.g., start scanning cluster, manage resources, etc.)
	userNamespace := resourceRebalancer.Spec.UserNamespace
	cpu := resourceRebalancer.Spec.CpuThreshold
	memory := resourceRebalancer.Spec.MemoryThreshold
	cleanup := resourceRebalancer.Spec.EnableCleanup

	go scanner.ScanCluster(r.MetricsClient, cpu, memory, userNamespace, cleanup)
	go worker.StartIdleCleanerWorker(r.Clientset)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceRebalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&smartv1.ResourceRebalancer{}).
		Named("resourcerebalancer").
		Complete(r)
}
