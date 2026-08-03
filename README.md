# vpa-creator

A controller to automatically deploy VPA to Deployment, StatefulSet, and DaemonSet resources across a cluster.

This is just a controller not an operator, there are no defined custom CRD's.

## Description

Will watch for Deployments, StatefulSets, and DaemonSets and create a Vertical Pod Autoscaler (VPA) for each resource in update mode off.

VPA's are deployed as update mode off, to be used to gain insights on right sizing CPU/Mem requests for workloads.

VPA update mode can be enabled per workload type with
`--deployment-update-mode`, `--statefulset-update-mode`, or
`--daemonset-update-mode`. Supported values are `Off`, `Initial`, `Recreate`,
`InPlaceOrRecreate`, and `InPlace`. By default, all three remain `Off`.

VPA creation can also be disabled per workload type with
`--enable-deployment-vpa=false`, `--enable-statefulset-vpa=false`, or
`--enable-daemonset-vpa=false`.

When VPA is configured to apply resources, set
`--vpa-controlled-values=RequestsOnly` to create VPAs that control requests
without controlling limits. The default is `RequestsAndLimits`.

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

- helm
- kubectl
- Access to a Kubernetes cluster with the Vertical Pod Autoscaler API installed.

### Install

Add the Helm chart repository:

```sh
helm repo add vpa-creator https://cornfeedhobo.github.io/vpa-creator
helm repo update
```

Install the controller with the default controller image:

```sh
helm upgrade --install vpa-creator vpa-creator/vpa-creator \
  --set vpa.enabled.statefulsets=false \
  --set vpa.updateMode.deployments=Recreate \
  --set vpa.controlledValues=RequestsOnly
```

The chart defaults to installing the controller resources into the
`vpa-creator-system` namespace. Enable optional ServiceMonitor or metrics
NetworkPolicy resources with chart values when those APIs are available in
your cluster.

### Uninstall

Remove the controller:

```sh
helm uninstall vpa-creator
```
