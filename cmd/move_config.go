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

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/helmpath"
)

func newMoveConfigCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move config",
		Short: "migrate Helm v2 repositories in-place to Helm v3",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("config argument has to be specified")
			}
			return nil
		},
		RunE: runMove,
	}

	flags := cmd.Flags()
	flags.BoolVar(&dryRun, "dry-run", false, "simulate a move config")

	return cmd

}

func runMove(cmd *cobra.Command, args []string) error {
	moveArgName := args[0]

	if moveArgName != "config" {
		return errors.New("config argument has to be specified")
	}

	// set home dirs for helm v2 and v3
	// TODO: add support for optional env vars for non default path of home dirs
	v2HomeDir, err := homedir.Dir()
	if err != nil {
		return err
	}

	v3HomeDir := helmpath.ConfigPath()

	return Move(v2HomeDir, v3HomeDir)
}

func Move(v2HomeDir, v3HomeDir string) error {
	fmt.Printf("Helm v2 repositories in folder \"%s/.helm\" will be copied to Helm v3 home folder \"%s\" .\n", v2HomeDir, v3HomeDir)

	return nil
}
