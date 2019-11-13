module github.com/helm/helm-2to3

go 1.13

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/maorfr/helm-plugin-utils v0.0.0-20181205064038-588190cb5e3b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	helm.sh/helm/v3 v3.0.0
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/helm v2.16.1+incompatible
)

replace github.com/docker/docker => github.com/docker/docker v0.0.0-20190731150326-928381b2215c
