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
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"helm-2to3/pkg/v2"
	"helm-2to3/pkg/v3"

	v2rel "k8s.io/helm/pkg/proto/hapi/release"
)

var (
	deletev2Releases bool
)

type ConvertOptions struct {
	DeleteRelease    bool
	DryRun           bool
	ReleaseName      string
	StorageType      string
	TillerLabel      string
	TillerNamespace  string
	TillerOutCluster bool
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

	flags.BoolVar(&deletev2Releases, "delete-v2-releases", false, "v2 releases are deleted after migration. By default, the v2 releases are retained")

	return cmd

}

func runConvert(cmd *cobra.Command, args []string) error {
	releaseName := args[0]
	if settings.releaseStorage != "configmaps" && settings.releaseStorage != "secrets" {
		return errors.New("release-storage flag needs to be 'configmaps' or 'secrets'")
	}
	convertOptions := ConvertOptions{
		DeleteRelease:    deletev2Releases,
		DryRun:           settings.dryRun,
		ReleaseName:      releaseName,
		StorageType:      settings.releaseStorage,
		TillerLabel:      settings.label,
		TillerNamespace:  settings.tillerNamespace,
		TillerOutCluster: settings.tillerOutCluster,
	}
	return Convert(convertOptions)
}

// Convert converts Helm 2 release into Helm 3 release. It maps the Helm v2 release versions
// of the release into Helm v3 equivalent and stores the release versions. The underlying Kubernetes resources
// are untouched. Note: The namespaces of each release version need to exist in the Kubernetes  cluster.
// The Helm 2 release is retained by default, unless the '--delete-v2-releases' flag is set.
func Convert(convertOptions ConvertOptions) error {
	if convertOptions.DryRun {
		fmt.Printf("NOTE: This is in dry-run mode, the following actions will not be executed.\n")
		fmt.Printf("Run without --dry-run to take the actions described below:\n\n")
	}

	fmt.Printf("Release \"%s\" will be converted from Helm v2 to Helm v3.\n", convertOptions.ReleaseName)

	fmt.Printf("[Helm 3] Release \"%s\" will be created.\n", convertOptions.ReleaseName)

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

	versions := []int32{}
	for i := len(v2Releases) - 1; i >= 0; i-- {
		v2Release := v2Releases[i]
		relVerName := v2.GetReleaseVersionName(convertOptions.ReleaseName, v2Release.Version)
		fmt.Printf("[Helm 3] ReleaseVersion \"%s\" will be created.\n", relVerName)
		if !convertOptions.DryRun {
			if err := createV3ReleaseVersion(v2Release); err != nil {
				return err
			}
			fmt.Printf("[Helm 3] ReleaseVersion \"%s\" created.\n", relVerName)
		}
		versions = append(versions, v2Release.Version)
	}
	if !convertOptions.DryRun {
		fmt.Printf("[Helm 3] Release \"%s\" created.\n", convertOptions.ReleaseName)
	}

	if convertOptions.DeleteRelease {
		fmt.Printf("[Helm 2] Release \"%s\" will be deleted.\n", convertOptions.ReleaseName)
		deleteOptions := v2.DeleteOptions{
			DryRun:   convertOptions.DryRun,
			Versions: versions,
		}
		if err := v2.DeleteReleaseVersions(retrieveOptions, deleteOptions); err != nil {
			return err
		}
		if !convertOptions.DryRun {
			fmt.Printf("[Helm 2] Release \"%s\" deleted.\n", convertOptions.ReleaseName)

			fmt.Printf("Release \"%s\" was converted successfully from Helm v2 to Helm v3.\n", convertOptions.ReleaseName)
		}
	} else {
		if !convertOptions.DryRun {
			fmt.Printf("Release \"%s\" was converted successfully from Helm v2 to Helm v3.\n", convertOptions.ReleaseName)
			fmt.Printf("Note: The v2 release information still remains and should be removed to avoid conflicts with the migrated v3 release.\n")
			fmt.Printf("v2 release information should only be removed using `helm 2to3` cleanup and when all releases have been migrated over.\n")
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
