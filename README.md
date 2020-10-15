# helm-generate
Recursively generates Kubernetes manifests from a folder using Helm. This project began as a [generator command](https://docs.fluxcd.io/en/1.20.0/references/fluxyaml-config-files/) for using with Flux.

## Docker
Available at: [Docker Hub](https://hub.docker.com/r/tfgco/helm-generate) and [Quay.io](https://quay.io/repository/tfgco/helm-generate).

## Requirements

* Helm

## Overview [![GoDoc](https://godoc.org/github.com/topfreegames/helm-generate?status.svg)](https://godoc.org/github.com/topfreegames/helm-generate)

Helm-generate renders Helm Charts recursively from a folder. It don't actually install any Chart, only render them, similar to `helm template`. Helm repositories must be managed through helm cli, as helm-generate uses the same configuration as your helm binary.
How it works:
* Helm-generate transverse folders and subfolders searching for `values.yaml` files.
* If a values.yaml file is found, the Chart configuration is defined by the following precende:
* 1) Existing .helm.yaml file on the same folder as the values.yaml.
* 2) --default-chart and --default-chart-version flags.
* 3) HELM_DEFAULT_CHART and HELM_DEFAULT_CHART_VERSION environment variables.
* The Namespace manifest for that chart is rendered.
* The Chart is rendered using the `values.yaml` file.
* All generated manifests are printed to STDOUT.

Helm-generate also handles the namespace injection on the manifests, as `helm template` don't handle this this and we don't want to have a requirement for charts to have namespace defined.

There are two required keys on `values.yaml`: namespace and releaseName. Those are internally used by helm-generate to correctly render the desired charts.

It is possible to inject `(key -> value)` pairs to the top level of the values map through the CLI, using the flag `--set my_key=my_value`. This flag is parsed as `[]string`, therefore it can be passed multiple times to inject multiple pairs. This flag overrides the values from `values.yaml`.

## .helm.yaml
This is a special control file designed to change the behavior of helm-generate for a specific folder, this don't apply to any subfolders.
The current keys available at .helm.yaml are:
```
chart: repository/chart-name
chartVersion: 1.x.x
postRenderBinary: path-to-binary
```
If no `.helm.yaml` is present at the same folder as a `values.yaml` file, the default values are used.

## Install

```
Go module:
go get github.com/topfreegames/helm-generate

Build from source:
git clone github.com/topfreegames/helm-generate
cd helm-generate
make all
```

## Example

```
Usage:
  helm-generate docs/examples/multiple-apps --default-chart=example-chart --default-chart-version=1.0.0
```
