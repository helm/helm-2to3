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

	"github.com/spf13/cobra"
)

var (
	settings *EnvSettings
)

func NewRootCmd(out io.Writer, args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "2to3",
		Short:        "Migrate and Cleanup Helm v2 configuration and releases in-place to Helm v3",
		Long:         "Migrate and Cleanup Helm v2 configuration and releases in-place to Helm v3",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("no arguments accepted")
			}
			return nil
		},
	}

	flags := cmd.PersistentFlags()
	flags.Parse(args)
	settings = new(EnvSettings)

	cmd.AddCommand(
		newCleanupCmd(out),
		newConvertCmd(out),
		newMoveConfigCmd(out),
	)

	return cmd
}
