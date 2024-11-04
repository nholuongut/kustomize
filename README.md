# Kustomize

![](https://i.imgur.com/waxVImv.png)
### [View all Roadmaps](https://github.com/nholuongut/all-roadmaps) &nbsp;&middot;&nbsp; [Best Practices](https://github.com/nholuongut/all-roadmaps/blob/main/public/best-practices/) &nbsp;&middot;&nbsp; [Questions](https://www.linkedin.com/in/nholuong/)
<br/>

`kustomize` lets you customize raw, template-free YAML
files for multiple purposes, leaving the original YAML
untouched and usable as is.

`kustomize` targets kubernetes; it understands and can
patch [kubernetes style] API objects.  It's like
[`make`], in that what it does is declared in a file,
and it's like [`sed`], in that it emits edited text.

This tool is sponsored by [sig-cli] ([KEP]).

 - [Installation instructions](https://kubectl.docs.kubernetes.io/installation/kustomize/)
 - [General documentation](https://kubectl.docs.kubernetes.io/references/kustomize/)
 - [Examples](examples)

[![Build Status](https://prow.k8s.io/badge.svg?jobs=kustomize-presubmit-master)](https://prow.k8s.io/job-history/kubernetes-jenkins/pr-logs/directory/kustomize-presubmit-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/nholuongut/kustomize)](https://goreportcard.com/report/github.com/nholuongut/kustomize)

## kubectl integration

To find the kustomize version embedded in recent versions of kubectl, run `kubectl version`:

```sh
> kubectl version --client
Client Version: v1.31.0
Kustomize Version: v5.4.2
```

The kustomize build flow at [v2.0.3] was added
to [kubectl v1.14][kubectl announcement].  The kustomize
flow in kubectl remained frozen at v2.0.3 until kubectl v1.21,
which [updated it to v4.0.5][kust-in-kubectl update]. It will
be updated on a regular basis going forward, and such updates
will be reflected in the Kubernetes release notes.

| Kubectl version | Kustomize version |
| --------------- | ----------------- |
| < v1.14         | n/a               |
| v1.14-v1.20     | v2.0.3            |
| v1.21           | v4.0.5            |
| v1.22           | v4.2.0            |
| v1.23           | v4.4.1            |
| v1.24           | v4.5.4            |
| v1.25           | v4.5.7            |
| v1.26           | v4.5.7            |
| v1.27           | v5.0.1            |

[v2.0.3]: https://github.com/nholuongut/kustomize/releases/tag/v2.0.3
[#2506]: https://github.com/nholuongut/kustomize/issues/2506
[#1500]: https://github.com/nholuongut/kustomize/issues/1500
[kust-in-kubectl update]: https://github.com/kubernetes/kubernetes/blob/4d75a6238a6e330337526e0513e67d02b1940b63/CHANGELOG/CHANGELOG-1.21.md#kustomize-updates-in-kubectl

For examples and guides for using the kubectl integration please
see the [kubernetes documentation].

## Usage


### 1) Make a [kustomization] file

In some directory containing your YAML [resource]
files (deployments, services, configmaps, etc.), create a
[kustomization] file.

This file should declare those resources, and any
customization to apply to them, e.g. _add a common
label_.

```

base: kustomization + resources

kustomization.yaml                                      deployment.yaml                                                 service.yaml
+---------------------------------------------+         +-------------------------------------------------------+       +-----------------------------------+
| apiVersion: kustomize.config.k8s.io/v1beta1 |         | apiVersion: apps/v1                                   |       | apiVersion: v1                    |
| kind: Kustomization                         |         | kind: Deployment                                      |       | kind: Service                     |
| commonLabels:                               |         | metadata:                                             |       | metadata:                         |
|   app: myapp                                |         |   name: myapp                                         |       |   name: myapp                     |
| resources:                                  |         | spec:                                                 |       | spec:                             |
|   - deployment.yaml                         |         |   selector:                                           |       |   selector:                       |
|   - service.yaml                            |         |     matchLabels:                                      |       |     app: myapp                    |
| configMapGenerator:                         |         |       app: myapp                                      |       |   ports:                          |
|   - name: myapp-map                         |         |   template:                                           |       |     - port: 6060                  |
|     literals:                               |         |     metadata:                                         |       |       targetPort: 6060            |
|       - KEY=value                           |         |       labels:                                         |       +-----------------------------------+
+---------------------------------------------+         |         app: myapp                                    |
                                                        |     spec:                                             |
                                                        |       containers:                                     |
                                                        |         - name: myapp                                 |
                                                        |           image: myapp                                |
                                                        |           resources:                                  |
                                                        |             limits:                                   |
                                                        |               memory: "128Mi"                         |
                                                        |               cpu: "500m"                             |
                                                        |           ports:                                      |
                                                        |             - containerPort: 6060                     |
                                                        +-------------------------------------------------------+

```

File structure:

> ```
> ~/someApp
> ├── deployment.yaml
> ├── kustomization.yaml
> └── service.yaml
> ```

The resources in this directory could be a fork of
someone else's configuration.  If so, you can easily
rebase from the source material to capture
improvements, because you don't modify the resources
directly.

Generate customized YAML with:

```
kustomize build ~/someApp
```

The YAML can be directly [applied] to a cluster:

> ```
> kustomize build ~/someApp | kubectl apply -f -
> ```


### 2) Create [variants] using [overlays]

Manage traditional [variants] of a configuration - like
_development_, _staging_ and _production_ - using
[overlays] that modify a common [base].

```

overlay: kustomization + patches

kustomization.yaml                                      replica_count.yaml                      cpu_count.yaml
+-----------------------------------------------+       +-------------------------------+       +------------------------------------------+
| apiVersion: kustomize.config.k8s.io/v1beta1   |       | apiVersion: apps/v1           |       | apiVersion: apps/v1                      |
| kind: Kustomization                           |       | kind: Deployment              |       | kind: Deployment                         |
| commonLabels:                                 |       | metadata:                     |       | metadata:                                |  
|   variant: prod                               |       |   name: myapp                 |       |   name: myapp                            |
| resources:                                    |       | spec:                         |       | spec:                                    |
|   - ../../base                                |       |   replicas: 80                |       |  template:                               |
| patches:                                      |       +-------------------------------+       |     spec:                                |
|   - path: replica_count.yaml                  |                                               |       containers:                        |
|   - path: cpu_count.yaml                      |                                               |         - name: myapp                    |  
+-----------------------------------------------+                                               |           resources:                     |
                                                                                                |             limits:                      |
                                                                                                |               memory: "128Mi"            |
                                                                                                |               cpu: "7000m"               |
                                                                                                +------------------------------------------+
```


File structure:
> ```
> ~/someApp
> ├── base
> │   ├── deployment.yaml
> │   ├── kustomization.yaml
> │   └── service.yaml
> └── overlays
>     ├── development
>     │   ├── cpu_count.yaml
>     │   ├── kustomization.yaml
>     │   └── replica_count.yaml
>     └── production
>         ├── cpu_count.yaml
>         ├── kustomization.yaml
>         └── replica_count.yaml
> ```

Take the work from step (1) above, move it into a
`someApp` subdirectory called `base`, then
place overlays in a sibling directory.

An overlay is just another kustomization, referring to
the base, and referring to patches to apply to that
base.

This arrangement makes it easy to manage your
configuration with `git`.  The base could have files
from an upstream repository managed by someone else.
The overlays could be in a repository you own.
Arranging the repo clones as siblings on disk avoids
the need for git submodules (though that works fine, if
you are a submodule fan).

Generate YAML with

```sh
kustomize build ~/someApp/overlays/production
```

The YAML can be directly [applied] to a cluster:

> ```sh
> kustomize build ~/someApp/overlays/production | kubectl apply -f -
> ```

# 🚀 I'm are always open to your feedback.  Please contact as bellow information:
### [Contact ]
* [Name: Nho Luong]
* [Skype](luongutnho_skype)
* [Github](https://github.com/nholuongut/)
* [Linkedin](https://www.linkedin.com/in/nholuong/)
* [Email Address](luongutnho@hotmail.com)
* [PayPal.me](https://www.paypal.com/paypalme/nholuongut)

![](https://i.imgur.com/waxVImv.png)
![](Donate.png)
[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/nholuong)

# License
* Nho Luong (c). All Rights Reserved.🌟

