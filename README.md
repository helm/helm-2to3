# Helm 2to3 Plugin

This is a Helm plugin which migrates Helm v2 releases in-place to Helm v3

## Usage

migrate Helm v2 releases in-place to Helm v3

```
$ helm 2to3 convert [flags] RELEASE
```

### Flags:

```
   --dry-run               simulate a convert
  -h, --help               help for convert
  -l, --label string       label to select tiller resources by (default "OWNER=TILLER")
  -t, --tiller-ns string   namespace of Tiller (default "kube-system")
```

## Install

**TODO**

### Developer (From Source) Install

If you would like to handle the build yourself, this is how recommend doing it.

First, set up your environment:

- You need to have [Go](http://golang.org) installed. Make sure to set `$GOPATH`

Clone this repo into a directory, for example your `$GOPATH`. You can use `go get -d github.com/hickeyma/helm-2to3`
for that.

```
$ cd $GOPATH/src/github.com/hickeyma/helm-2to3
$ make build
$ helm plugin install $GOPATH/src/github.com/hickeyma/helm-2to3
```

That last command will use the binary that you built.
