package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	ZonekeeperLabelUpdateFailedCountTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zonekeeper_label_updates_failed_total",
			Help: "Total number of pods the controller failed to apply zone labels to",
		},
		[]string{"zone", "namespace"},
	)
	ZonekeeperLabelUpdateCountTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zonekeeper_label_updates_total",
			Help: "Total number of pods the controller successfully updated zone labels for",
		},
		[]string{"zone", "namespace"},
	)
	ZonekeeperNodesWatchCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zonekeeper_nodes_watched",
			Help: "Total number of nodes in watch by the controller",
		},
		[]string{"zone"},
	)
	ZonekeeperReconciliationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zonekeeper_k8s_reconciliations_total",
			Help: "Total number of reconciliations performed by the controller",
		},
		[]string{"kind", "group", "version", "namespace"},
	)
)

func RegisterMetrics() {
	metrics.Registry.MustRegister(ZonekeeperLabelUpdateFailedCountTotal)
	metrics.Registry.MustRegister(ZonekeeperLabelUpdateCountTotal)
	metrics.Registry.MustRegister(ZonekeeperNodesWatchCount)
	metrics.Registry.MustRegister(ZonekeeperReconciliationsTotal)
}
