# vpa-creator

Helm chart for the `vpa-creator` controller.

Upstream source:

- https://github.com/cloudoutloud/vpa-creator

This chart is sourced from the upstream repository's `charts/vpa-creator`
directory. The upstream project does not publish controller images, and this
chart repository does not publish replacement images. Build the image from the
upstream source and install the chart with the image location you pushed:

```sh
make docker-build docker-push IMG=<some-registry>/vpa-creator:<tag>

helm install vpa-creator ./charts/vpa-creator \
  --set image.repository=<some-registry>/vpa-creator \
  --set image.tag=<tag> \
  --set vpa.enabled.statefulsets=false \
  --set vpa.updateMode.deployments=Auto \
  --set vpa.controlledValues=RequestsOnly
```

`image.repository` is intentionally empty by default. Helm rendering fails until
it is set, because there is no known upstream image repository to pull from.

The chart defaults all created VPAs to update mode `Off`. Set
`vpa.updateMode.deployments`, `vpa.updateMode.statefulsets`, or
`vpa.updateMode.daemonsets` to `Auto` per workload type. Set
`vpa.controlledValues=RequestsOnly` to have auto-mode VPAs control requests
without controlling limits.

Set `vpa.enabled.deployments`, `vpa.enabled.statefulsets`, or
`vpa.enabled.daemonsets` to `false` to disable VPA creation and watches for that
workload type.
