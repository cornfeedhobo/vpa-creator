# vpa-creator

Helm chart for the `vpa-creator` controller.

Upstream source:

- https://github.com/cloudoutloud/vpa-creator

This chart installs the controller image published from this repository. The
default image tag follows the chart `appVersion`.

```sh
helm install vpa-creator ./charts/vpa-creator \
  --set vpa.enabled.statefulsets=false \
  --set vpa.updateMode.deployments=Recreate \
  --set vpa.controlledValues=RequestsOnly
```

Set `image.repository` and `image.tag` only when installing a custom controller
image.

The chart defaults all created VPAs to update mode `Off`. Set
`vpa.updateMode.deployments`, `vpa.updateMode.statefulsets`, or
`vpa.updateMode.daemonsets` to `Initial`, `Recreate`, `InPlaceOrRecreate`,
or `InPlace` per workload type. Set
`vpa.controlledValues=RequestsOnly` to have applying VPAs control requests
without controlling limits.

Set `vpa.enabled.deployments`, `vpa.enabled.statefulsets`, or
`vpa.enabled.daemonsets` to `false` to disable VPA creation and watches for that
workload type.
