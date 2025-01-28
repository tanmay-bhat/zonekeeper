package controller

import (
	"context"
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	processedNodes   = make(map[string]struct{})
	processedNodesMu sync.Mutex
)

type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	podZoneLabel = "topology.kubernetes.io/zone"
)

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.Log.WithName("zonekeeper")

	ZonekeeperReconciliationsTotal.WithLabelValues("Pod", "corev1", "v1", req.Namespace).Inc()

	// Fetch the Pod
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Skip if pod hasn't been scheduled yet
	if pod.Spec.NodeName == "" {
		logger.V(1).Info("pod not yet scheduled",
			"pod", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	// Fetch the Node
	var node corev1.Node
	if err := r.Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
		logger.Error(err, "failed to get node",
			"node", pod.Spec.NodeName,
			"pod", req.NamespacedName)
		return ctrl.Result{}, fmt.Errorf("failed to get node %s: %w", pod.Spec.NodeName, err)
	}

	processedNodesMu.Lock()
	defer processedNodesMu.Unlock()
	if _, exists := processedNodes[node.Name]; !exists {
		ZonekeeperNodesWatchCount.WithLabelValues(node.Labels[podZoneLabel]).Inc()
		processedNodes[node.Name] = struct{}{}
	}

	// Get zone from node labels
	zone, exists := node.Labels[podZoneLabel]

	if !exists {
		logger.Error(nil, "node missing required zone label",
			"node", node.Name,
			"required_label", podZoneLabel)
		return ctrl.Result{}, nil
	}

	// Check if update is needed
	if pod.Labels[podZoneLabel] == zone {
		logger.V(2).Info("pod zone label already up to date",
			"pod", req.NamespacedName,
			"zone", zone)
		return ctrl.Result{}, nil
	}

	// Create patch of original pod
	podCopy := pod.DeepCopy()

	// Ensure labels map exists and update zone
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[podZoneLabel] = zone

	logger.Info("updating pod zone label",
		"pod", req.NamespacedName,
		"zone", zone)

	// Use Patch instead of Update to handle conflicts
	if err := r.Patch(ctx, &pod, client.MergeFrom(podCopy)); err != nil {
		logger.Error(err, "failed to patch pod",
			"pod", req.NamespacedName,
			"zone", zone)
		ZonekeeperLabelUpdateFailedCountTotal.WithLabelValues(zone, req.Namespace).Inc()
		return ctrl.Result{}, fmt.Errorf("failed to patch pod: %w", err)
	}
	ZonekeeperLabelUpdateCountTotal.WithLabelValues(zone, req.Namespace).Inc()
	return ctrl.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(object client.Object) bool {
			pod := object.(*corev1.Pod)
			return pod.Spec.NodeName != ""
		})).
		Complete(r)
}
