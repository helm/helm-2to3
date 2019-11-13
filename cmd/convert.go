/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"io"
	"log"

	"github.com/spf13/cobra"
	v2rel "k8s.io/helm/pkg/proto/hapi/release"

	v2 "github.com/helm/helm-2to3/pkg/v2"
	v3 "github.com/helm/helm-2to3/pkg/v3"
)

var (
	deletev2Releases   bool
	maxReleaseVersions int
)

type ConvertOptions struct {
	DeleteRelease      bool
	DryRun             bool
	MaxReleaseVersions int
	ReleaseName        string
	StorageType        string
	TillerLabel        string
	TillerNamespace    string
	TillerOutCluster   bool
}

func newConvertCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert [flags] RELEASE",
		Short: "migrate Helm v2 release in-place to Helm v3",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("name of release to be converted has to be defined")
			}
			return nil
		},

		RunE: runConvert,
	}

	flags := cmd.Flags()
	settings.AddFlags(flags)

	flags.BoolVar(&deletev2Releases, "delete-v2-releases", false, "v2 release versions are deleted after migration. By default, the v2 release versions are retained")
	flags.IntVar(&maxReleaseVersions, "release-versions-max", 10, "limit the maximum number of versions converted per release. Use 0 for no limit")

	return cmd

}

func runConvert(cmd *cobra.Command, args []string) error {
	releaseName := args[0]
	if settings.releaseStorage != "configmaps" && settings.releaseStorage != "secrets" {
		return errors.New("release-storage flag needs to be 'configmaps' or 'secrets'")
	}
	convertOptions := ConvertOptions{
		DeleteRelease:      deletev2Releases,
		DryRun:             settings.dryRun,
		MaxReleaseVersions: maxReleaseVersions,
		ReleaseName:        releaseName,
		StorageType:        settings.releaseStorage,
		TillerLabel:        settings.label,
		TillerNamespace:    settings.tillerNamespace,
		TillerOutCluster:   settings.tillerOutCluster,
	}
	return Convert(convertOptions)
}

// Convert converts Helm 2 release into Helm 3 release. It maps the Helm v2 release versions
// of the release into Helm v3 equivalent and stores the release versions. The underlying Kubernetes resources
// are untouched. Note: The namespaces of each release version need to exist in the Kubernetes  cluster.
// The Helm 2 release is retained by default, unless the '--delete-v2-releases' flag is set.
func Convert(convertOptions ConvertOptions) error {
	if convertOptions.DryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
	}

	log.Printf("Release \"%s\" will be converted from Helm v2 to Helm v3.\n", convertOptions.ReleaseName)

	log.Printf("[Helm 3] Release \"%s\" will be created.\n", convertOptions.ReleaseName)

	retrieveOptions := v2.RetrieveOptions{
		ReleaseName:      convertOptions.ReleaseName,
		TillerNamespace:  convertOptions.TillerNamespace,
		TillerLabel:      convertOptions.TillerLabel,
		TillerOutCluster: convertOptions.TillerOutCluster,
		StorageType:      convertOptions.StorageType,
	}
	v2Releases, err := v2.GetReleaseVersions(retrieveOptions)
	if err != nil {
		return err
	}

	// Limit release versions to migrate.
	// Limit is based on newest versions.
	v2RelVerLen := len(v2Releases)
	startIndex := 0
	if convertOptions.MaxReleaseVersions > 0 && convertOptions.MaxReleaseVersions < v2RelVerLen {
		log.Println()
		log.Printf("NOTE: The max release versions \"%d\" is less than the actual release versions \"%d\".", convertOptions.MaxReleaseVersions, v2RelVerLen)
		log.Printf("This means only \"%d\" of the latest release versions will be converted.", convertOptions.MaxReleaseVersions)
		if convertOptions.DeleteRelease {
			log.Println("This also means some versions will remain in Helm v2 storage that will no longer be visible to Helm v2 commands like 'helm list'. Plugin 'cleanup' command will remove them from storage.")
		}
		log.Println()
		startIndex = v2RelVerLen - convertOptions.MaxReleaseVersions
	}

	versions := []int32{}
	for i := startIndex; i < v2RelVerLen; i++ {
		v2Release := v2Releases[i]
		relVerName := v2.GetReleaseVersionName(convertOptions.ReleaseName, v2Release.Version)
		log.Printf("[Helm 3] ReleaseVersion \"%s\" will be created.\n", relVerName)
		if !convertOptions.DryRun {
			if err := createV3ReleaseVersion(v2Release); err != nil {
				return err
			}
			log.Printf("[Helm 3] ReleaseVersion \"%s\" created.\n", relVerName)
		}
		versions = append(versions, v2Release.Version)
	}
	if !convertOptions.DryRun {
		log.Printf("[Helm 3] Release \"%s\" created.\n", convertOptions.ReleaseName)
	}

	if convertOptions.DeleteRelease {
		log.Printf("[Helm 2] Release \"%s\" will be deleted.\n", convertOptions.ReleaseName)
		deleteOptions := v2.DeleteOptions{
			DryRun:   convertOptions.DryRun,
			Versions: versions,
		}
		if err := v2.DeleteReleaseVersions(retrieveOptions, deleteOptions); err != nil {
			return err
		}
		if !convertOptions.DryRun {
			log.Printf("[Helm 2] Release \"%s\" deleted.\n", convertOptions.ReleaseName)

			log.Printf("Release \"%s\" was converted successfully from Helm v2 to Helm v3.\n", convertOptions.ReleaseName)
		}
	} else {
		if !convertOptions.DryRun {
			log.Printf("Release \"%s\" was converted successfully from Helm v2 to Helm v3.\n", convertOptions.ReleaseName)
			log.Println("Note: The v2 release information still remains and should be removed to avoid conflicts with the migrated v3 release.")
			log.Println("v2 release information should only be removed using `helm 2to3` cleanup and when all releases have been migrated over.")
		}
	}

	return nil
}

func createV3ReleaseVersion(v2Release *v2rel.Release) error {
	v3Release, err := v3.CreateRelease(v2Release)
	if err != nil {
		return err
	}
	return v3.StoreRelease(v3Release)
}
