package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PodReconciler reconciles Pod objects
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	podZoneLabel = "topology.kubernetes.io/zone"
)

// +kubebuilder:rbac:groups=core,resources=pods;nodes,verbs=get;list;watch;update;patch

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.Log.WithName("zonekeeper")

	// Fetch the Pod
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Skip if pod hasn't been scheduled yet
	if pod.Spec.NodeName == "" {
		logger.V(1).Info("Pod not yet scheduled")
		return ctrl.Result{}, nil
	}

	// Fetch the Node
	var node corev1.Node
	if err := r.Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get node %s: %w", pod.Spec.NodeName, err)
	}

	// Get zone from node labels
	zone, exists := node.Labels[podZoneLabel]
	if !exists {
		logger.Info(fmt.Sprintf("Zone label not found on node %s", pod.Spec.NodeName))
		return ctrl.Result{}, nil
	}

	// Check if update is needed
	if pod.Labels[podZoneLabel] == zone {
		return ctrl.Result{}, nil
	}

	// Create patch of original pod
	podCopy := pod.DeepCopy()

	// Ensure labels map exists and update zone
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[podZoneLabel] = zone

	logger.Info(fmt.Sprintf("Updating pod %s/%s with zone label '%s'", pod.Namespace, pod.Name, zone))

	// Use Patch instead of Update to handle conflicts
	if err := r.Patch(ctx, &pod, client.MergeFrom(podCopy)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to patch pod: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager and filter only pods that are scheduled
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(object client.Object) bool {
			pod := object.(*corev1.Pod)
			return pod.Spec.NodeName != ""
		})).
		Complete(r)
}
