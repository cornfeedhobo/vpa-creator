# vpa-creator

A controller to automatically deploy VPA to Deployment, StatefulSet, and DaemonSet resources across a cluster.

This is just a controller not an operator, there are no defined custom CRD's

## Description

Will watch for Deployments, StatefulSets, and DaemonSets and create a Vertical Pod Autoscaler (VPA) for each resource in update mode off.

VPA's are deployed as update mode off, to be used to gain insights on right sizing CPU/Mem requests for workloads.

VPA update mode can be enabled per workload type with
`--deployment-update-mode=Auto`, `--statefulset-update-mode=Auto`, or
`--daemonset-update-mode=Auto`. By default, all three remain `Off`.

VPA creation can also be disabled per workload type with
`--enable-deployment-vpa=false`, `--enable-statefulset-vpa=false`, or
`--enable-daemonset-vpa=false`.

When auto mode is enabled, set `--vpa-controlled-values=RequestsOnly` to create
VPAs that control requests without controlling limits. The default is
`RequestsAndLimits`.

When a resource is garbage collected, the controller automatically removes the associated VPA through Kubernetes' garbage collection mechanism.

This controller is designed to work with the Kubernetes Vertical Pod Autoscaler API.

It assumes that the Vertical Pod Autoscaler API is installed in the cluster. You would need metric server and VPA CRD's installed.

## What problem can this solve?

Right sizing workloads can be a difficult tasks, often developers and DevOps teams don't know what to set for an application.

As a result of this workload request and limits are often set at the start of the app lifecycle and never re-reviewed by service owners/teams. This requests and limits are often over provisioned and set based on simple assumption.

Over time this can lead to inaccurate settings, by configuring VPA in dry mode for each resource we can track the recommend actual usage and adjust request and limits accordingly.

This in turn could reduce overall compute cost of a cluster.

## Getting Started

### Prerequisites
- go version v1.26.4+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/vpa-creator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.


**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/vpa-creator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/vpa-creator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/vpa-creator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

Install the chart with the image built and published in your registry:

```sh
helm install vpa-creator ./charts/vpa-creator \
  --set image.repository=<some-registry>/vpa-creator \
  --set image.tag=<tag> \
  --set vpa.enabled.statefulsets=false \
  --set vpa.updateMode.deployments=Auto \
  --set vpa.controlledValues=RequestsOnly
```

The chart defaults to installing the controller resources into the
`vpa-creator-system` namespace. Enable optional ServiceMonitor or metrics
NetworkPolicy resources with chart values when those APIs are available in
your cluster.

## Contributing

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
