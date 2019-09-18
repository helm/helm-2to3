# Helm 2to3 Plugin

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/helm/helm-2to3)](https://goreportcard.com/report/github.com/helm/helm-2to3)
[![CircleCI](https://circleci.com/gh/helm/helm-2to3/tree/master.svg?style=svg)](https://circleci.com/gh/helm/helm-2to3/tree/master)
[![Release](https://img.shields.io/github/release/helm/helm-2to3.svg?style=flat-square)](https://github.com/helm/helm-2to3/releases/latest)

![diagram](./helm-2to3.png)

**Helm plugin which migrates and cleans up Helm v2 configuration and releases in-place to Helm v3**

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

It will migrate:
- Chart starters
- Repositories
- Plugins 

*Note:* Please check that all Helm v2 plugins work fine with the Helm v3, and remove not working ones.

For migration it uses default Helm v2 home and v3 config and data folders.
To override those folders you need to set environment variables `HELM_V2_HOME`, `HELM_V3_CONFIG` and `HELM_V3_DATA`:

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
      --dry-run                  simulate a convert
  -h, --help                     help for convert
      --delete-v2-releases       v2 releases are deleted after migration. By default, the v2 releases are retained
  -l, --label string             label to select tiller resources by (default "OWNER=TILLER")
  -s, --release-storage string   v2 release storage type/object. It can be 'secrets' or 'configmaps'. This is only used with the 'tiller-out-cluster' flag (default "secrets")
  -t, --tiller-ns string         namespace of Tiller (default "kube-system")
      --tiller-out-cluster       when Tiller is not running in the cluster e.g. Tillerless
```

### Clean up Helm v2 data

Clean up Helm v2 configuration, release data and Tiller deployment:

```console
$ helm 2to3 cleanup [flags]

Flags:
      --dry-run                  simulate a command
  -h, --help                     help for cleanup
  -l, --label string             label to select tiller resources by (default "OWNER=TILLER")
  -s, --release-storage string   v2 release storage type/object. It can be 'secrets' or 'configmaps'. This is only used with the 'tiller-out-cluster' flag (default "secrets")
  -t, --tiller-ns string         namespace of Tiller (default "kube-system")
      --tiller-out-cluster       when  Tiller is not running in the cluster e.g. Tillerless
```

It will clean:
- Configuration (Helm home directory)
- v2 release data
- Tiller deployment

For cleanup it uses the default Helm v2 home folder.
To override this folder you need to set the environment variable `HELM_V2_HOME`:

```console
$ export HELM_V2_HOME=$PWD/.helm2
$ helm 2to3 cleanup
```

*Warning:* The `cleanup` command will remove the Helm v2 Configuration, Release Data and Tiller Deployment.
It cleans up all releases managed by Helm v2. It will not be possible to restore them if you haven't made a backup of the releases.
Helm v2 will not be usable afterwards.

## Install

Based on the version in `plugin.yaml`, release binary will be downloaded from GitHub:

```console
$ helm plugin install https://github.com/helm/helm-2to3
Downloading and installing helm-2to3 v0.1.1 ...
https://github.com/helm/helm-2to3/releases/download/v0.1.1/helm-2to3_0.1.1_darwin_amd64.tar.gz
Installed plugin: 2to3
```

## Developer (From Source) Install

If you would like to handle the build yourself, this is the recommended way to do it.

You must first have [Go v1.13](http://golang.org) installed, and then you run:

```console
$ git clone git@github.com:helm/helm-2to3.git
$ cd helm-2to3
$ make build
$ helm plugin install <your_path>/helm-2to3
```

That last command will use the binary that you built.
