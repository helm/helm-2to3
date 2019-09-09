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
		Short: "migrate Helm v2 configuration in-place to Helm v3",
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

	return move()
}

func v2HomeDir() string {
	if homeDir, exists := os.LookupEnv("HELM_V2_HOME"); exists {
		return homeDir
	}

	homeDir, _ := homedir.Dir()
	defaultDir := homeDir + "/.helm"
	fmt.Printf("[Helm 2] Home directory: %s\n", defaultDir)
	return defaultDir
}

func v3ConfigDir() string {
	if homeDir, exists := os.LookupEnv("HELM_V3_CONFIG"); exists {
		return homeDir
	}

	defaultDir := helmpath.ConfigPath()
	fmt.Printf("[Helm 3] Config directory: %s\n", defaultDir)
	return defaultDir
}

func v3DataDir() string {
	if homeDir, exists := os.LookupEnv("HELM_V3_DATA"); exists {
		return homeDir
	}

	defaultDir := helmpath.DataPath()
	fmt.Printf("[Helm 3] Data directory: %s\n", defaultDir)
	return defaultDir
}

func move() error {
	v2HomeDir := v2HomeDir()
	v3ConfigDir := v3ConfigDir()
	v3DataDir := v3DataDir()

	// Create Helm v3 config directory if needed
	fmt.Printf("i[Helm 3] Create config folder \"%s\" .\n", v3ConfigDir)
	err := ensureDir(v3ConfigDir)
	if err != nil {
		return fmt.Errorf("[Helm 3] Failed to create config folder \"%s\" due to the following error: %s", v3ConfigDir, err)
	}
	fmt.Printf("[Helm 3] Config folder \"%s\" created.\n", v3ConfigDir)

	// Move repo config
	v2RepoConfig := v2HomeDir + "/repository/repositories.yaml"
	v3RepoConfig := v3ConfigDir + "/repositories.yaml"
	fmt.Printf("[Helm 2] repositories file \"%s\" will copy to [Helm 3] config folder \"%s\" .\n", v2RepoConfig, v3RepoConfig)
	err = copyFile(v2RepoConfig, v3RepoConfig)
	if err != nil {
		return fmt.Errorf("Failed to copy [Helm 2] repository file \"%s\" due to the following error: %s", v2RepoConfig, err)
	}
	fmt.Printf("[Helm 2] repositories file \"%s\" copied successfully to [Helm 3] config folder \"%s\" .\n", v2RepoConfig, v3RepoConfig)

	// Bot moving local repo as it is no longer5 supported in v3: v2HomeDir/repository/local

	// Not moving the cache as it is safer to recreate when needed
	// v2HomeDir/cache and v2HomeDir/repository/cache

	// Create Helm v3 data directory if needed
	fmt.Printf("[Helm 3] Create data folder \"%s\" .\n", v3DataDir)
	err = ensureDir(v3DataDir)
	if err != nil {
		return fmt.Errorf("[Helm 3] Failed to create data folder \"%s\" due to the following error: %s", v3DataDir, err)
	}
	fmt.Printf("[Helm 3] data folder \"%s\" created.\n", v3DataDir)

	// Move plugins
	v2Plugins := v2HomeDir + "/plugins"
	v3Plugins := v3DataDir + "/plugins"
	fmt.Printf("[Helm 2] plugins \"%s\" will copy to [Helm 3] data folder \"%s\" .\n", v2Plugins, v3Plugins)
	err = copyDir(v2Plugins, v3Plugins)
	if err != nil {
		return fmt.Errorf("Failed to copy [Helm 2] plugins directory \"%s\" due to the following error: %s", v2Plugins, err)
	}
	fmt.Printf("[Helm 2] plugins \"%s\" copied successfully to [Helm 3] data folder \"%s\" .\n", v2Plugins, v3Plugins)

	// Move starters
	v2Starters := v2HomeDir + "/starters"
	v3Starters := v3DataDir + "/starters"
	fmt.Printf("[Helm 2] starters \"%s\" will copy to [Helm 3] data folder \"%s\" .\n", v2Starters, v3Starters)
	err = copyDir(v2Starters, v3Starters)
	if err != nil {
		return fmt.Errorf("Failed to copy [Helm 2] starters \"%s\" due to the following error: %s", v2Starters, err)
	}
	fmt.Printf("[Helm 2] starters \"%s\" copied successfully to [Helm 3] data folder \"%s\" .\n", v2Starters, v3Starters)

	return nil
}

func copyFile(srcFileName, destFileName string) error {
	input, err := ioutil.ReadFile(srcFileName)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(destFileName, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func copyDir(srcDirName, destDirName string) error {
	err := ensureDir(destDirName)
	if err != nil {
		return fmt.Errorf("Failed to create folder \"%s\" due to the following error: %s", destDirName, err)
	}

	directory, _ := os.Open(srcDirName)
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		srcFileName := srcDirName + "/" + obj.Name()
		destFileName := destDirName + "/" + obj.Name()
		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(srcFileName, destFileName)
			if err != nil {
				return fmt.Errorf("Failed to copy folder \"%s\" to folder \"%s\" due to the following error: %s", srcFileName, destFileName, err)
			}
		} else {
			fileInfo, err := os.Lstat(srcFileName)
			if err != nil {
				return fmt.Errorf("Failed to check file \"%s\" stats  due to the following error: %s", srcFileName, err)
			}
			if fileInfo.Mode()&os.ModeSymlink != 0 {
				err = copySymLink(obj, srcDirName, destDirName)
				if err != nil {
					return fmt.Errorf("Failed to create symlink for  \"%s\" due to the following error: %s", obj.Name(), err)
				}
			} else {
				err = copyFile(srcFileName, destFileName)
				if err != nil {
					return fmt.Errorf("Failed to copy file  \"%s\" to \"%s\" due to the following error: %s", srcFileName, destFileName, err)
				}
			}
		}
	}

	return nil
}

func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil || !os.IsExist(err) {
		return err
	}
	return nil
}

func copySymLink(fileInfo os.FileInfo, srcDirName, destDirName string) error {
	originFileName, err := os.Readlink(srcDirName + "/" + fileInfo.Name())
	if err != nil {
		return err
	}
	newSymLinkName := destDirName + "/" + fileInfo.Name()
	err = os.Symlink(originFileName, newSymLinkName)
	if err != nil {
		return err
	}
	return nil
}
