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

	utils "github.com/helm/helm-2to3/pkg/utils"
)

func newMoveConfigCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move config",
		Short: "migrate Helm v2 configuration in-place to Helm v3",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("config argument has to be specified")
			}
			return nil
		},
		RunE: runMove,
	}

	flags := cmd.Flags()
	settings.AddBaseFlags(flags)
	flags.BoolVar(&skipConfirmation, "skip-confirmation", false, "if set, skips confirmation message before performing move")
	return cmd
}

func runMove(cmd *cobra.Command, args []string) error {
	moveArgName := args[0]

	if moveArgName != "config" {
		return errors.New("config argument has to be specified")
	}

	return Move(settings.DryRun)
}

// Moves/copies v2 configuration to v2 configuration. It copies repository config,
// plugins and starters. It does not copy cache.
func Move(dryRun bool) error {
	var err error
	var doCleanup bool
	if dryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
	}

	log.Println("WARNING: Helm v3 configuration may be overwritten during this operation.")
	log.Println()
	if skipConfirmation {

		log.Println("Skipping confirmation before performing move.")
		doCleanup = true
	} else {
		doCleanup, err = utils.AskConfirmation("Move Config", "move the v2 configuration")
		if err != nil {
			return err
		}
	}
	if !doCleanup {
		log.Println("Move will not proceed as the user didn't answer (Y|y) in order to continue.")
		return nil
	}

	log.Println("\nHelm v2 configuration will be moved to Helm v3 configuration.")
	err = utils.Copyv2HomeTov3(dryRun)
	if err != nil {
		return err
	}
	if !dryRun {
		log.Println("Helm v2 configuration was moved successfully to Helm v3 configuration.")
	}
	return nil
}
