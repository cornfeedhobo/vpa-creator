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

The chart defaults all created VPAs to update mode `Off`. Set the
`vpa-creator.cornfeedhobo/update-mode` annotation on a Deployment, StatefulSet,
or DaemonSet object metadata, or on its pod template metadata, to `Initial`,
`Recreate`, `InPlaceOrRecreate`, or `InPlace` to opt that resource into an
applying update mode. Object metadata takes precedence when both locations set
the annotation. Set
`vpa.controlledValues=RequestsOnly` to have applying VPAs control requests
without controlling limits.

Set `vpa.enabled.deployments`, `vpa.enabled.statefulsets`, or
`vpa.enabled.daemonsets` to `false` to disable VPA creation and watches for that
workload type.
