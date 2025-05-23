# kube-review

Simple command line utility to transform a provided Kubernetes resource into a Kubernetes AdmissionReview request, as
sent from the Kubernetes API server when [dynamic admission control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
(i.e. webhook) is configured.

**deployment.yaml**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 8080
```
**Command**
```shell
$ kube-review create deployment.yaml
```
**Output**
```json
{
  "kind": "AdmissionReview",
  "apiVersion": "admission.k8s.io/v1",
  "request": {
    "uid": "2024ee9c-c374-413c-838d-e62bcb4826be",
    "kind": {
      "group": "apps",
      "version": "v1",
      "kind": "Deployment"
    },
    "resource": {
      "group": "apps",
      "version": "v1",
      "resource": "deployments"
    },
    "requestKind": {
      "group": "apps",
      "version": "v1",
      "kind": "Deployment"
    },
    "requestResource": {
      "group": "apps",
      "version": "v1",
      "resource": "deployments"
    },
    "name": "nginx",
    "operation": "CREATE",
    "userInfo": {
      "username": "kube-review",
      "uid": "611a19d7-6aa5-47d2-bba3-8c5df2bffbc7"
    },
    "object": {
      "kind": "Deployment",
      "apiVersion": "apps/v1",
      "metadata": {
        "name": "nginx",
        "creationTimestamp": null,
        "labels": {
          "app": "nginx"
        }
      },
      "spec": {
        "selector": {
          "matchLabels": {
            "app": "nginx"
          }
        },
        "template": {
          "metadata": {
            "creationTimestamp": null,
            "labels": {
              "app": "nginx"
            }
          },
          "spec": {
            "containers": [
              {
                "name": "nginx",
                "image": "nginx",
                "ports": [
                  {
                    "containerPort": 8080
                  }
                ],
                "resources": {}
              }
            ]
          }
        },
        "strategy": {}
      },
      "status": {}
    },
    "oldObject": null,
    "dryRun": true,
    "options": {
      "kind": "CreateOptions",
      "apiVersion": "meta.k8s.io/v1"
    }
  }
}
```

## Why?

* Testing Kubernetes admission webhook receivers without Kubernetes (CI/CD pipelines, faster integration tests, etc.)
* Quickly be able to author, and test, admission control policies with tools like [Open Policy Agent](https://www.openpolicyagent.org/)

## Installation

Find the latest release for your platform at the [release page](https://github.com/anderseknert/kube-review/releases/latest).
Once downloaded, rename it to `kube-review` (or `kube-review.exe` for Windows if not using WSL),
allow it to be executed, and put it somewhere on your `$PATH`.

## Running kube-review

`kube-review create` can either be provided a filename with a resource to create an admission review for, or can read
data from stdin. This allows easily piping resources from a kube cluster and into kube-review.

**Command**
```shell
$ kubectl get service gatekeeper-webhook-service -o yaml | kube-review create --action update
```
**Output**
```json
{
    "kind": "AdmissionReview",
    "apiVersion": "admission.k8s.io/v1",
    "request": {
        "uid": "b42420d7-5cc2-4644-992f-72ff67dc2889",
        "kind": {
            "group": "",
            "version": "v1",
            "kind": "Service"
        },
        "name": "gatekeeper-webhook-service",
        "namespace": "gatekeeper-system",
        "operation": "UPDATE",
        "userInfo": {
            "username": "kube-review",
            "uid": "42eac911-a8ec-4d72-9eb1-e6c466328085"
        },
        "...": "..."
    }
}
```
## Command line options

| Name         | Type   | Default     | Description                                                                                |
|--------------|--------|-------------|--------------------------------------------------------------------------------------------|
| `--action`   | string | create      | Type of operation to apply in admission review (create, update, delete, connect)           |
| `--as`       | string | kube-review | Name of user or service account for userInfo attributes                                    |
| `--as-group` | string | none        | Name of group this user or service account belongs to. May be repeated for multiple groups |
| `--indent`   | int    | 2           | Number of spaces to indent JSON output                                                     |

The `action` provided has the following effects on the produced `AdmissionReview` object:

* `create` and `connect`: `request.oldObject` is `null`
* `delete`: `request.object` is `null`
* All actions change the value of `request.operation` and `request.options`

## Using with Open Policy Agent

Assuming we have a policy that denies any deployment where the number of replicas is either undefined or below two:

```rego
package admission

deny contains "Deployment must have at least 2 replicas" if {
    input.request.object.spec.replicas < 2
}

deny contains "Deployment must define number of replicas explicitly" if {
    not input.request.object.spec.replicas
}
```

We could either run kube-review with a deployment from disk, and pipe the output into `opa eval`:

**Command**
```shell
$ kube-review create deployment.yaml \
| opa eval --format pretty --stdin-input --data policy.rego data.admission.deny
```
**Output**
```json
[
  "Deployment must define number of replicas explicitly"
]
```
Or we could run the policy against any resource in our cluster in the same manner:

**Command**
```shell
$ kubectl get deployment my-microservice -o yaml \
| kube-review create \
| opa eval --format pretty --stdin-input --data policy.rego data.admission.deny
```
**Output**
```json
[
  "Deployment must have at least 2 replicas"
]
```
Alternatively, we could use `curl` to send the data into a running OPA server:

**Command**
```shell
$ kubectl get deployment my-microservice -o yaml \
| kube-review create \
| curl --data-binary "@-" http://localhost:8181/v0/data/admission/deny
```
**Output**
```json
[
    "Deployment must define number of replicas explicitly"
]
```

If your policies are written for [OPA Gatekeeper](https://github.com/open-policy-agent/gatekeeper), simply rename the
`request` object in the admission request to `review`:

```shell
$ kube-review create deployment.yaml \
| opa eval --format pretty --stdin-input '{"review": input.request}' \
| opa eval --format pretty --stdin-input --data policy.rego data.admission.deny
```

## Limitations

* kube-review can create `AdmissionReview` objects from CRDs, or any "Kubernetes-like" manifest, but it makes no attempt
  to verify that they conform to the schema of the corresponding `CustomResourceDefinition`.
* Currently, kube-review doesn't handle subresource requests.
