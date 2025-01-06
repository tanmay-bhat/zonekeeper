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

### Running The Example Stack

Refer to `examples/README.md` for instructions on running the example stack which includes `metrics-exporter`, `vmstack` and `zonekeeper`.