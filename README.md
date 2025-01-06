# zonekeeper
Zonekeeper is a kubernetes controller that adds (patches) availability zone label to pods. Its main goal is to identify the availability zone of the node where the pod is running and add this information to the pod's labels.

This is useful so that metrics collection agents such as Prometheus or victoria-metrics can scrape metrics from pods running in different availability zones and hence reduce inter-available zone traffic and cost.
