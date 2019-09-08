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
	"io/ioutil"
	"os"

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

	return cmd
}

func runMove(cmd *cobra.Command, args []string) error {
	moveArgName := args[0]

	if moveArgName != "config" {
		return errors.New("config argument has to be specified")
	}

	return move(v2HomeDir(), v3ConfigDir())
}

func v2HomeDir() string {
	if homeDir, exists := os.LookupEnv("HELM_V2_HOME"); exists {
		return homeDir
	}

	homeDir, _ := homedir.Dir()
	defaultDir := homeDir + "/.helm"
	fmt.Println(defaultDir)
	return defaultDir
}

func v3ConfigDir() string {
	if homeDir, exists := os.LookupEnv("HELM_V3_CONFIG"); exists {
		return homeDir
	}

	defaultDir := helmpath.ConfigPath()
	fmt.Println(defaultDir)
	return defaultDir
}

func move(v2HomeDir, v3ConfigDir string) error {
	v2RepoConfig := v2HomeDir + "/repository/repositories.yaml"
	v3RepoConfig := v3ConfigDir + "/repositories.yaml"
	fmt.Printf("Helm v2 repositories file \"%s\" will be copied to Helm v3 config folder \"%s\" .\n", v2RepoConfig, v3RepoConfig)

	// Create v3 repo home dir
	if err := ensureDir(v3ConfigDir); err != nil {
		fmt.Println("Directory creation failed with error: " + err.Error())
		return err
	}

	// set v2 repositories file
	input, err := ioutil.ReadFile(v2RepoConfig)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// copy v3 repositories file
	err = ioutil.WriteFile(v3RepoConfig, input, 0644)
	if err != nil {
		fmt.Println("Error creating", v3RepoConfig)
		fmt.Println(err)
		return err
	}

	return nil
}

func ensureDir(dirName string) error {

	err := os.MkdirAll(dirName, os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}
