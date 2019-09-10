# Helm 2to3 Plugin

This is a Helm plugin which migrates Helm v2 configuration and releases in-place to Helm v3

## Usage

### Migrate Helm v2 configuration

Migrate Helm v2 configuration in-place to Helm v3:

```console
$ helm 2to3 move config [flags]
```

Flags:

```
  -h, --help     help for move
```

For migration it uses default Helm v2 home and v3 config  and data folders.
To override those folders you need to set environment variables `HELM_V2_HOME, `HELM_V3_CONFIG` and `HELM_V3_DATA`:

```console
$ export HELM_V2_HOME=$PWD/.helm2
$ export HELM_V3_CONFIG=$PWD/.helm3
$ export HELM_V3_DATA=$PWD/.helm3
$ helm 2to3 move config
```

The `move config` will create the Helm v3 config and data folders if they don't exist, and will override the `repositories.yaml` file if it does exist.

### Migrate Helm v2 releases

Migrate Helm v2 releases in-place to Helm v3

```console
$ helm 2to3 convert [flags] RELEASE
```

Flags:

```
      --dry-run            simulate a convert
  -h, --help               help for convert
      --keep-v2-releases   v2 releases are retained after migration. By default, the v2 releases are removed
  -l, --label string       label to select tiller resources by (default "OWNER=TILLER")
  -s, --release-storage string   v2 release storage type/object. It can be 'configmaps' or 'secrets'. This is only used with the 'tiller-out-cluster' flag (default "configmaps")
  -t, --tiller-ns string   namespace of Tiller (default "kube-system")
      --tiller-out-cluster       when  Tiller is not running in the cluster e.g. Tillerless
```

## Install

Based on the version in `plugin.yaml`, release binary will be downloaded from GitHub:

```console
$ helm plugin install https://github.com/hickeyma/helm-2to3
Downloading and installing helm-2to3 v0.1.0 ...
https://github.com/hickeyma/helm-2to3/releases/download/v0.1.0/helm-2to3_0.1.0_darwin_amd64.tar.gz
Installed plugin: 2to3
```

## Developer (From Source) Install

If you would like to handle the build yourself, this is the recommended way to do it.

You must first have [Go v1.13](http://golang.org) installed, and then you run:

```console
$ git clone git@github.com:hickeyma/helm-2to3.git
$ cd helm-2to3
$ make build
$ helm plugin install <your_path>/helm-2to3
```

That last command will use the binary that you built.
