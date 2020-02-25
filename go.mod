module github.com/helm/helm-2to3

go 1.13

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/maorfr/helm-plugin-utils v0.0.0-20200216074820-36d2fcf6ae86
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	helm.sh/helm/v3 v3.1.0
	k8s.io/apimachinery v0.17.2
	k8s.io/helm v2.16.3+incompatible
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
)
