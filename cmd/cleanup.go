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
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"helm-2to3/pkg/common"
	"helm-2to3/pkg/v2"
)

/*var (
	settings *EnvSettings
)*/

func newCleanupCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "cleanup Helm v2 configuration, release data and Tiller deployment",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: runCleanup,
	}

	flags := cmd.Flags()
	settings.AddFlags(flags)

	return cmd
}

func runCleanup(cmd *cobra.Command, args []string) error {
	return Cleanup()
}

// Cleanup will delete all release data for in specified namespace and owner label. It will remove
// the Tiller server deployed as per namespace and owner label. It is also delete the Helm gv2 home directory
// which contains the Helm configuration. Helm v2 will be unusable after this operation.
func Cleanup() error {
	if settings.dryRun {
		fmt.Printf("NOTE: This is in dry-run mode, the following actions will not be executed.\n")
		fmt.Printf("Run without --dry-run to take the actions described below:\n\n")
	}

	fmt.Printf("WARNING: Helm v2 Configuration, Release Data and Tiller Deployment will be removed.\n")
	fmt.Printf("This will clean up all releases managed by Helm v2. It will not be possible to restore them if you haven't made a backup of the releases.\n")
	fmt.Printf("Helm v2 will not be usable afterwards.\n\n")

	doCleanup, err := common.AskConfirmation("Cleanup", "cleanup Helm v2 data")
	if err != nil {
		return err
	}
	if !doCleanup {
		fmt.Printf("Cleanup will not proceed as the user didn't answer (Y|y) in order to continue.\n")
		return nil
	}

	fmt.Printf("\nHelm v2 data will be cleaned up.\n")

	fmt.Printf("[Helm 2] Releases will be deleted.\n")
	retrieveOptions := v2.RetrieveOptions{
		ReleaseName:      "",
		TillerNamespace:  settings.tillerNamespace,
		TillerLabel:      settings.label,
		TillerOutCluster: settings.tillerOutCluster,
		StorageType:      settings.releaseStorage,
	}
	err = v2.DeleteAllReleaseVersions(retrieveOptions, settings.dryRun)
	if err != nil {
		return err
	}
	if !settings.dryRun {
		fmt.Printf("[Helm 2] Releases deleted.\n")
	}

	if !settings.tillerOutCluster {
		fmt.Printf("[Helm 2] Tiller in \"%s\" namespace will be removed.\n", settings.tillerNamespace)
		err = v2.RemoveTiller(settings.tillerNamespace, settings.dryRun)
		if err != nil {
			return err
		}
		if !settings.dryRun {
			fmt.Printf("[Helm 2] Tiller in \"%s\" namespace was removed.\n", settings.tillerNamespace)
		}
	}

	err = v2.RemoveHomeFolder(settings.dryRun)
	if err != nil {
		return err
	}

	if !settings.dryRun {
		fmt.Printf("Helm v2 data was cleaned up successfully.\n")
	}
	return nil
}
