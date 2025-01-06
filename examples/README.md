### Metrics-exporter
metrics-exporter is a dummy pod which exposes a `/metrics` endpoint for Prometheus to scrape. It is used to demonstrate how Zonekeeper adds availability zone labels to pods.

It can be deployed using the following command:
```bash
kubectl apply -f examples/metrics-exporter/
```

### Zonekeeper
Zonekeeper is a kubernetes controller that adds (patches) availability zone label to pods. Its main goal is to identify the availability zone of the node where the pod is running and add this information to the pod's labels.

This is useful so that metrics collection agents such as Prometheus or victoria-metrics can scrape metrics from pods running in different availability zones and hence reduce inter-available zone traffic and cost.

### Running Zonekeeper
It can be deployed using the following command:
```bash
kubectl apply -f examples/zonekeeper/
```

### VMstack (vmagent and vmsingle)
vmstack is a set of components that can be used to collect metrics from pods running in different availability zones. It consists of:
- vmagent: A Prometheus-compatible metrics collection agent that can scrape metrics from pods and remote write them to a victoria-metrics instance.
- vmsingle: A single-node victoria-metrics instance that can store metrics scraped by vmagent.

vmstack can be deployed using the following command:
```bash
kubectl apply -f examples/vmstack/
```
This demonstrates collection of metrics from zone `us-west-2a`. Same can be created for other zones.

Cluster can be created with 4 nodes along with labels for availability zones for 2 nodes using the following command:
```bash
k3d cluster create multinode \
  --agents 4 \
  --k3s-node-label "topology.kubernetes.io/zone=us-west-2a@agent:0,1" 
```
