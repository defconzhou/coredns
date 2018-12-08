# external

## Name

*external* - resolve ingress and load balance IPs from kubernetes clusters.

## Description

Expand this massively; why is this useful. How to enable.

Any zone enabled here, will resolve to load balanced and ingress IP addresses as defined in the
Kubernetes cluster.

## Syntax

~~~
external [ZONE...]
~~~

* **ZONES** zones *external* should be authoritative for.

# Examples

~~~
external
~~~

# Also See

For some background see [resolve external IP address](https://github.com/kubernetes/dns/issues/242).
And [A records for services with Load Balancer IP](https://github.com/coredns/coredns/issues/1851).
