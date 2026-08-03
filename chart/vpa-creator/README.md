# vpa-creator

Helm chart for the `vpa-creator` controller.

Source:

- https://github.com/cornfeedhobo/vpa-creator

This chart installs the controller image published from this repository. The
default image tag follows the chart `appVersion`.

```sh
helm repo add vpa-creator https://cornfeedhobo.github.io/vpa-creator
helm repo update
helm upgrade --install vpa-creator vpa-creator/vpa-creator \
  --set vpa.enabled.statefulsets=false \
  --set vpa.controlledValues=RequestsOnly
```

Set `image.repository` and `image.tag` only when installing a custom controller
image.

The chart defaults all created VPAs to update mode `Off`. VPA behavior can be
configured per resource with annotations on a Deployment, StatefulSet, or
DaemonSet object metadata, or on its pod template metadata. Object metadata
takes precedence for the same annotation key.

| Annotation | Format | Generated VPA field | Notes |
| --- | --- | --- | --- |
| `vpa-creator.cornfeedhobo/update-mode` | `Off`, `Initial`, `Recreate`, `InPlaceOrRecreate`, or `InPlace` | `spec.updatePolicy.updateMode` | Resources without this annotation remain `Off`. |
| `vpa-creator.cornfeedhobo/min-replicas` | Positive integer, for example `1` | `spec.updatePolicy.minReplicas` | Overrides the VPA updater's global minimum replica safety check for this VPA. |
| `vpa-creator.cornfeedhobo/controlled-resources` | Comma-separated resource names, for example `cpu,memory` | `spec.resourcePolicy.containerPolicies[0].controlledResources` | Applies to the wildcard `*` container policy. |
| `vpa-creator.cornfeedhobo/min-allowed` | Comma-separated resource list, for example `cpu=100m,memory=128Mi` | `spec.resourcePolicy.containerPolicies[0].minAllowed` | Applies to the wildcard `*` container policy. |
| `vpa-creator.cornfeedhobo/max-allowed` | Comma-separated resource list, for example `cpu=2,memory=4Gi` | `spec.resourcePolicy.containerPolicies[0].maxAllowed` | Applies to the wildcard `*` container policy. |

Set `vpa.controlledValues=RequestsOnly` to have applying VPAs control requests
without controlling limits.

Set `vpa.enabled.deployments`, `vpa.enabled.statefulsets`, or
`vpa.enabled.daemonsets` to `false` to disable VPA creation and watches for that
workload type.
