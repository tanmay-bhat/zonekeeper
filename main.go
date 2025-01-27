package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"

	controller "github.com/tanmay-bhat/zonekeeper/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func getWatchNamespace() string {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		setupLog.Info(fmt.Sprintf("%s is not set, the manager will watch and manage resources in all namespaces", watchNamespaceEnvVar))
		return ""
	}
	return ns
}

func main() {
	var probeAddr string
	var podLabelSelector string
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&podLabelSelector, "pod-label-selector", "", "The label selector for pods to watch, key=value, multiple can be separated by comma")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	watchNamespace := getWatchNamespace()
	namespaceConfigs := make(map[string]cache.Config)
	labelSelectorConfig := make(map[string]string)

	if watchNamespace == "" {
		namespaceConfigs[""] = cache.Config{}
	}

	namespaces := strings.Split(watchNamespace, ",")
	setupLog.Info("manager set up with namespaces", "namespaces", watchNamespace)

	for _, ns := range namespaces {
		ns = strings.TrimSpace(ns)
		if ns != "" {
			namespaceConfigs[ns] = cache.Config{}
		}
	}

	if podLabelSelector != "" {
		setupLog.Info("manager set up with pod label selector", "label", podLabelSelector)

		labels := strings.Split(podLabelSelector, ",")
		for _, label := range labels {
			label = strings.TrimSpace(label)
			kv := strings.Split(label, "=")
			if len(kv) != 2 {
				setupLog.Error(fmt.Errorf("invalid label selector"), "invalid label selector", "label", label)
				os.Exit(1)
			}
			labelSelectorConfig[kv[0]] = kv[1]
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		Cache: cache.Options{
			DefaultNamespaces: namespaceConfigs,
			ByObject: map[client.Object]cache.ByObject{
				&corev1.Pod{}: {
					Label: labels.SelectorFromSet(labelSelectorConfig),
				},
			},
		},
		LeaderElection:   true,
		LeaderElectionID: "zonekeeper-leader-election",
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.PodReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pod")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
