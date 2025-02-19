# Zonekeeper
Zonekeeper is a kubernetes controller that adds (patches) availability zone label to pods. Its main goal is to identify the availability zone of the node where the pod is running and add this information to the pod's labels.

This is useful so that metrics collection agents such as Prometheus or victoria-metrics can scrape metrics from pods running in different availability zones and hence reduce inter-available zone traffic and cost.

### Running Zonekeeper

Zonekeeper can be run as a deployment in your kubernetes cluster. We can apply the deployment and required RBAC by running the following command:

```bash
kubectl apply -f examples/zonekeeper/
```

For testing, zonekeeper can also be run locally using the following command and it will use the current kube context to connect to the cluster:

```bash
go run main.go
```

The docker image is available at `tanmaybhat/zonekeeper:1.0.1`

Once the deployment is running, Zonekeeper will start adding availability zone labels to pods in the cluster, for example:

```bash
│ 2025-01-06T05:17:22Z    INFO    zonekeeper    Updating pod kube-system/svclb-traefik-685270de-q2dnf with zone label 'us-west-2b'
│ 2025-01-06T05:17:22Z    INFO    zonekeeper    Updating pod monitoring/metrics-exporter-59f4ddd48-b6p9g with zone label 'us-west-2a'
│ 2025-01-06T05:17:22Z    INFO    zonekeeper    Updating pod kube-system/helm-install-traefik-crd-2x99z with zone label 'us-west-2b'
│ 2025-01-06T05:17:22Z    INFO    zonekeeper    Updating pod kube-system/traefik-d7c9c5778-4dkg9 with zone label 'us-west-2a'
...
```


### Namespaces Filtering
It is possible to watch over only certain namespace(s) by specifying them with env WATCH_NAMESPACE(comma seperated). By default it will watch over all namespaces.

### Label Selectors
Zonekeeper can be configured to watch only pods with specific labels by specifying them with with argument `pod-label-selector`. For example, to watch only pods with label `app=nginx`:
```
./zonekeeper --pod-label-selector=app=nginx
```
Multiple labels can be specified by separating them with comma. For example, to watch only pods with labels `app=nginx` and `env=prod`:
```
./zonekeeper --pod-label-selector=app=nginx,env=prod
```

### Running The Example Stack

Refer to `examples/README.md` for instructions on running the example stack which includes `metrics-exporter`, `vmstack` and `zonekeeper`.

### Metrics
Zonekeeper exposes metrics the below on `/metrics` endpoint by default at port 8080 : 
- `zonekeeper_label_updates_failed_total` : The total number of pod label updates that failed.
- `zonekeeper_label_updates_total` : The total number of pod label updates that succeeded.
- `zonekeeper_nodes_watched` : The total number nodes that are being watched by zonekeeper.
- `zonekeeper_k8s_reconciliations_total` : The total number of kubernetes reconciliations that have been performed by zonekeeper.

