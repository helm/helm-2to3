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

		RunE: run,
	}

	flags := cmd.Flags()
	settings.AddFlags(flags)

	flags.BoolVar(&deletev2Releases, "delete-v2-releases", false, "v2 releases are deleted after migration. By default, the v2 releases are retained")

	return cmd

}

func run(cmd *cobra.Command, args []string) error {
	releaseName := args[0]
	if settings.releaseStorage != "configmaps" && settings.releaseStorage != "secrets" {
		return errors.New("release-storage flag needs to be 'configmaps' or 'secrets'")
	}
	return Convert(releaseName)
}

// Convert converts Helm 2 release into Helm 3 release. It maps the Helm v2 release versions
// of the release into Helm v3 equivalent and stores the release versions. The underlying Kubernetes resources
// are untouched. Note: The namespaces of each release version need to exist in the Kubernetes  cluster.
// The Helm 2 release is retained by default, unless the '--delete-v2-releases' flag is set.
func Convert(releaseName string) error {
	if settings.dryRun {
		fmt.Printf("NOTE: This is in dry-run mode, the following actions will not be executed.\n")
		fmt.Printf("Run without --dry-run to take the actions described below:\n\n")
	}

	fmt.Printf("Release \"%s\" will be converted from Helm 2 to Helm 3.\n", releaseName)

	fmt.Printf("[Helm 3] Release \"%s\" will be created.\n", releaseName)

	retrieveOptions := v2.RetrieveOptions{
		ReleaseName:      releaseName,
		TillerNamespace:  settings.tillerNamespace,
		TillerLabel:      settings.label,
		TillerOutCluster: settings.tillerOutCluster,
		StorageType:      settings.releaseStorage,
	}
	v2Releases, err := v2.GetReleaseVersions(retrieveOptions)
	if err != nil {
		return err
	}

	versions := []int32{}
	for i := len(v2Releases) - 1; i >= 0; i-- {
		v2Release := v2Releases[i]
		relVerName := v2.GetReleaseVersionName(releaseName, v2Release.Version)
		fmt.Printf("[Helm 3] ReleaseVersion \"%s\" will be created.\n", relVerName)
		if !settings.dryRun {
			if err := createV3ReleaseVersion(v2Release); err != nil {
				return err
			}
			fmt.Printf("[Helm 3] ReleaseVersion \"%s\" created.\n", relVerName)
		}
		versions = append(versions, v2Release.Version)
	}
	if !settings.dryRun {
		fmt.Printf("[Helm 3] Release \"%s\" created.\n", releaseName)
	}

	if deletev2Releases {
		fmt.Printf("[Helm 2] Release \"%s\" will be deleted.\n", releaseName)
		deleteOptions := v2.DeleteOptions{
			DryRun:   settings.dryRun,
			Versions: versions,
		}
		if err := v2.DeleteReleaseVersions(retrieveOptions, deleteOptions); err != nil {
			return err
		}
		if !settings.dryRun {
			fmt.Printf("[Helm 2] Release \"%s\" deleted.\n", releaseName)

			fmt.Printf("Release \"%s\" was converted successfully from Helm 2 to Helm 3.\n", releaseName)
		}
	} else {
		if !settings.dryRun {
			fmt.Printf("Release \"%s\" was converted successfully from Helm 2 to Helm 3. Note: the v2 releases still remain and should be removed to avoid conflicts with the migrated v3 releases.\n", releaseName)
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
