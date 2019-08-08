# Helm 2to3 Plugin

This is a Helm plugin which migrates Helm v2 releases in-place to Helm v3

## Usage

migrate Helm v2 releases in-place to Helm v3

```
$ helm 2to3 convert [flags] RELEASE
```

### Flags:

```
      --dry-run            simulate a convert
  -h, --help               help for convert
  -l, --label string       label to select tiller resources by (default "OWNER=TILLER")
  -t, --tiller-ns string   namespace of Tiller (default "kube-system")
```

## Install

**TODO**

### Developer (From Source) Install

If you would like to handle the build yourself, this is how recommend doing it.

You must first have [Go](http://golang.org) installed , and then you run:

```
$ git clone git@github.com:hickeyma/helm-2to3.git
$ cd helm-2to3
$ make
$ helm plugin install <your_path>/helm-2to3
```

That last command will use the binary that you built.
