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
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/helm/helm-2to3/pkg/common"
	v2 "github.com/helm/helm-2to3/pkg/v2"
)

var (
	configCleanup  bool
	releaseCleanup bool
	tillerCleanup  bool
)

type CleanupOptions struct {
	ConfigCleanup    bool
	DryRun           bool
	ReleaseCleanup   bool
	StorageType      string
	TillerCleanup    bool
	TillerLabel      string
	TillerNamespace  string
	TillerOutCluster bool
}

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

	flags.BoolVar(&configCleanup, "config-cleanup", false, "if set, configuration cleanup performed")
	flags.BoolVar(&releaseCleanup, "release-cleanup", false, "if set, release data cleanup performed")
	flags.BoolVar(&tillerCleanup, "tiller-cleanup", false, "if set, Tiller cleanup performed")

	return cmd
}

func runCleanup(cmd *cobra.Command, args []string) error {
	cleanupOptions := CleanupOptions{
		ConfigCleanup:    configCleanup,
		DryRun:           settings.dryRun,
		ReleaseCleanup:   releaseCleanup,
		StorageType:      settings.releaseStorage,
		TillerCleanup:    tillerCleanup,
		TillerLabel:      settings.label,
		TillerNamespace:  settings.tillerNamespace,
		TillerOutCluster: settings.tillerOutCluster,
	}
	return Cleanup(cleanupOptions)
}

// Cleanup will delete all release data for in specified namespace and owner label. It will remove
// the Tiller server deployed as per namespace and owner label. It is also delete the Helm gv2 home directory
// which contains the Helm configuration. Helm v2 will be unusable after this operation.
func Cleanup(cleanupOptions CleanupOptions) error {
	var message strings.Builder

	if !cleanupOptions.ConfigCleanup && !cleanupOptions.ReleaseCleanup && !cleanupOptions.TillerCleanup {
		cleanupOptions.ConfigCleanup = true
		cleanupOptions.ReleaseCleanup = true
		cleanupOptions.TillerCleanup = true
	}

	if cleanupOptions.DryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
	}

	fmt.Fprint(&message, "WARNING: ")
	if cleanupOptions.ConfigCleanup {
		fmt.Fprint(&message, "\"Helm v2 Configuration\" ")
	}
	if cleanupOptions.ReleaseCleanup {
		fmt.Fprint(&message, "\"Release Data\" ")
	}
	if cleanupOptions.TillerCleanup {
		fmt.Fprint(&message, "\"Release Data\" ")
	}
	fmt.Fprintln(&message, "will be removed. ")
	if cleanupOptions.ReleaseCleanup {
		fmt.Fprintln(&message, "This will clean up all releases managed by Helm v2. It will not be possible to restore them if you haven't made a backup of the releases.")
	}
	fmt.Fprintln(&message, "Helm v2 may not be usable afterwards.")

	fmt.Println(message.String())

	doCleanup, err := common.AskConfirmation("Cleanup", "cleanup Helm v2 data")
	if err != nil {
		return err
	}
	if !doCleanup {
		log.Println("Cleanup will not proceed as the user didn't answer (Y|y) in order to continue.")
		return nil
	}

	log.Printf("\nHelm v2 data will be cleaned up.\n")

	if cleanupOptions.ReleaseCleanup {
		log.Println("[Helm 2] Releases will be deleted.")
		retrieveOptions := v2.RetrieveOptions{
			ReleaseName:      "",
			TillerNamespace:  cleanupOptions.TillerNamespace,
			TillerLabel:      cleanupOptions.TillerLabel,
			TillerOutCluster: cleanupOptions.TillerOutCluster,
			StorageType:      cleanupOptions.StorageType,
		}
		err = v2.DeleteAllReleaseVersions(retrieveOptions, cleanupOptions.DryRun)
		if err != nil {
			return err
		}
		if !cleanupOptions.DryRun {
			log.Println("[Helm 2] Releases deleted.")
		}
	}

	if !cleanupOptions.TillerOutCluster && cleanupOptions.TillerCleanup {
		log.Printf("[Helm 2] Tiller in \"%s\" namespace will be removed.\n", cleanupOptions.TillerNamespace)
		err = v2.RemoveTiller(cleanupOptions.TillerNamespace, cleanupOptions.DryRun)
		if err != nil {
			return err
		}
		if !cleanupOptions.DryRun {
			log.Printf("[Helm 2] Tiller in \"%s\" namespace was removed.\n", cleanupOptions.TillerNamespace)
		}
	}

	if cleanupOptions.ConfigCleanup {
		err = v2.RemoveHomeFolder(cleanupOptions.DryRun)
		if err != nil {
			return err
		}
	}

	if !cleanupOptions.DryRun {
		log.Println("Helm v2 data was cleaned up successfully.")
	}
	return nil
}
