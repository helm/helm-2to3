module github.com/helm/helm-2to3

go 1.16

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/maorfr/helm-plugin-utils v0.0.0-20200827170302-51b70049c73f
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/ugorji/go v1.1.4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	helm.sh/helm/v3 v3.5.1
	k8s.io/apimachinery v0.20.2
	k8s.io/helm v2.17.0+incompatible
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	gotest.tools => gotest.tools/v3 v3.0.2
)
