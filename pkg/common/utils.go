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

package common

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	v2 "github.com/helm/helm-2to3/pkg/v2"
	v3 "github.com/helm/helm-2to3/pkg/v3"
)

const sep = string(filepath.Separator)

// Copyv2HomeTov3 copies the v2 home directory to the v3 home directory .
// Note that this is not a direct 1-1 copy
func Copyv2HomeTov3(dryRun bool) error {
	v2HomeDir := v2.HomeDir()
	log.Printf("[Helm 2] Home directory: %s\n", v2HomeDir)
	v3ConfigDir := v3.ConfigDir()
	log.Printf("[Helm 3] Config directory: %s\n", v3ConfigDir)
	v3DataDir := v3.DataDir()
	log.Printf("[Helm 3] Data directory: %s\n", v3DataDir)

	// Create Helm v3 config directory if needed
	log.Printf("[Helm 3] Create config folder \"%s\" .\n", v3ConfigDir)
	var err error
	if !dryRun {
		err = ensureDir(v3ConfigDir)
		if err != nil {
			return fmt.Errorf("[Helm 3] Failed to create config folder \"%s\" due to the following error: %s", v3ConfigDir, err)
		}
		log.Printf("[Helm 3] Config folder \"%s\" created.\n", v3ConfigDir)
	}

	// Move repo config
	v2RepoConfig := v2HomeDir + sep + "repository" + sep + "repositories.yaml"
	v3RepoConfig := v3ConfigDir + sep + "repositories.yaml"
	log.Printf("[Helm 2] repositories file \"%s\" will copy to [Helm 3] config folder \"%s\" .\n", v2RepoConfig, v3RepoConfig)
	if !dryRun {
		err = copyFile(v2RepoConfig, v3RepoConfig)
		if err != nil {
			return fmt.Errorf("Failed to copy [Helm 2] repository file \"%s\" due to the following error: %s", v2RepoConfig, err)
		}
		log.Printf("[Helm 2] repositories file \"%s\" copied successfully to [Helm 3] config folder \"%s\" .\n", v2RepoConfig, v3RepoConfig)
	}

	// Not moving local repo as it is safer to recreate: e.g. v2HomeDir/repository/local

	// Not moving the cache as it is safer to recreate when needed
	// v2HomeDir/cache and v2HomeDir/repository/cache

	// Create Helm v3 data directory if needed
	log.Printf("[Helm 3] Create data folder \"%s\" .\n", v3DataDir)
	if !dryRun {
		err = ensureDir(v3DataDir)
		if err != nil {
			return fmt.Errorf("[Helm 3] Failed to create data folder \"%s\" due to the following error: %s", v3DataDir, err)
		}
		log.Printf("[Helm 3] data folder \"%s\" created.\n", v3DataDir)
	}

	// Move plugins
	v2Plugins := v2HomeDir + sep + "plugins"
	v3Plugins := v3DataDir + sep + "plugins"
	log.Printf("[Helm 2] plugins \"%s\" will copy to [Helm 3] data folder \"%s\" .\n", v2Plugins, v3Plugins)
	if !dryRun {
		err = copyDir(v2Plugins, v3Plugins)
		if err != nil {
			return fmt.Errorf("Failed to copy [Helm 2] plugins directory \"%s\" due to the following error: %s", v2Plugins, err)
		}
		log.Printf("[Helm 2] plugins \"%s\" copied successfully to [Helm 3] data folder \"%s\" .\n", v2Plugins, v3Plugins)
	}

	// Move starters
	v2Starters := v2HomeDir + sep + "starters"
	v3Starters := v3DataDir + sep + "starters"
	log.Printf("[Helm 2] starters \"%s\" will copy to [Helm 3] data folder \"%s\" .\n", v2Starters, v3Starters)
	if !dryRun {
		err = copyDir(v2Starters, v3Starters)
		if err != nil {
			return fmt.Errorf("Failed to copy [Helm 2] starters \"%s\" due to the following error: %s", v2Starters, err)
		}
		log.Printf("[Helm 2] starters \"%s\" copied successfully to [Helm 3] data folder \"%s\" .\n", v2Starters, v3Starters)
	}

	return nil
}

// AskConfirmation provides a prompt for user to confirm continuation with operation
func AskConfirmation(operation, specificMsg string) (bool, error) {
	fmt.Printf("[%s/confirm] Are you sure you want to %s? [y/N]: ", operation, specificMsg)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return false, errors.Wrap(err, "couldn't read from standard input")
	}
	answer := scanner.Text()
	if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
		return true, nil
	}
	return false, nil
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
		srcFileName := srcDirName + sep + obj.Name()
		destFileName := destDirName + sep + obj.Name()
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
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func copySymLink(fileInfo os.FileInfo, srcDirName, destDirName string) error {
	originFileName, err := os.Readlink(srcDirName + sep + fileInfo.Name())
	if err != nil {
		return err
	}
	newSymLinkName := destDirName + sep + fileInfo.Name()
	err = os.Symlink(originFileName, newSymLinkName)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
