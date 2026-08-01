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
  --set image.tag=<tag>
```

`image.repository` is intentionally empty by default. Helm rendering fails until
it is set, because there is no known upstream image repository to pull from.
