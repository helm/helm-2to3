module github.com/helm/helm-2to3

go 1.16

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/maorfr/helm-plugin-utils v0.0.0-20200827170302-51b70049c73f
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	helm.sh/helm/v3 v3.6.3
	k8s.io/apimachinery v0.21.0
	k8s.io/helm v2.17.0+incompatible
)

replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
