# kube-review

**WIP**

Simple utility to transform a Kubernetes resource into a Kubernetes AdmissionReview request, as it would be sent from
the Kubernetes API server. This is useful for testing Kubernetes admission webhook receivers without having to set up
Kubernetes plus admission controller configurations, or to quickly be able to write policies for admission control.

TODO:
---
* Command line flags
* Read from stdin
* Make useful as a library
* Document examples
* Kubectl plugin