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

// Copyv2HomeTov3 copies the v2 home directory to the v3 home directory .
// Note that this is not a direct 1-1 copy
func Copyv2HomeTov3(dryRun bool) error {
	v2HomeDir := v2.HomeDir()
	log.Printf("[Helm 2] Home directory: %s\n", v2HomeDir)
	v3ConfigDir := v3.ConfigDir()
	log.Printf("[Helm 3] Config directory: %s\n", v3ConfigDir)
	v3DataDir := v3.DataDir()
	log.Printf("[Helm 3] Data directory: %s\n", v3DataDir)
	v3CacheDir := v3.CacheDir()
	log.Printf("[Helm 3] Cache directory: %s\n", v3CacheDir)

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
	v2RepoConfig := filepath.Join(v2HomeDir, "repository", "repositories.yaml")
	v3RepoConfig := filepath.Join(v3ConfigDir, "repositories.yaml")
	log.Printf("[Helm 2] repositories file \"%s\" will copy to [Helm 3] config folder \"%s\" .\n", v2RepoConfig, v3RepoConfig)
	if !dryRun {
		err = copyFile(v2RepoConfig, v3RepoConfig)
		if err != nil {
			return fmt.Errorf("Failed to copy [Helm 2] repository file \"%s\" due to the following error: %s", v2RepoConfig, err)
		}
		log.Printf("[Helm 2] repositories file \"%s\" copied successfully to [Helm 3] config folder \"%s\" .\n", v2RepoConfig, v3RepoConfig)
	}

	// Not moving local repo and its cache, as it is safer to recreate: e.g. v2HomeDir/repository/local v2HomeDir/repository/cache

	// Create Helm v3 cache directory if needed
	log.Printf("[Helm 3] Create cache folder \"%s\" .\n", v3CacheDir)
	if !dryRun {
		err = ensureDir(v3CacheDir)
		if err != nil {
			return fmt.Errorf("[Helm 3] Failed to create cache folder \"%s\" due to the following error: %s", v3CacheDir, err)
		}
		log.Printf("[Helm 3] cache folder \"%s\" created.\n", v3CacheDir)
	}

	// Create Helm v3 data directory if needed
	log.Printf("[Helm 3] Create data folder \"%s\" .\n", v3DataDir)
	if !dryRun {
		err = ensureDir(v3DataDir)
		if err != nil {
			return fmt.Errorf("[Helm 3] Failed to create data folder \"%s\" due to the following error: %s", v3DataDir, err)
		}
		log.Printf("[Helm 3] data folder \"%s\" created.\n", v3DataDir)
	}

	// Handle plugins
	v2Plugins := filepath.Join(v2HomeDir, "cache", "plugins")
	plugins, _ := pathExists(v2Plugins)
	if plugins {
		// Move plugins
		v2Plugins := filepath.Join(v2HomeDir, "cache", "plugins")
		v3Plugins := filepath.Join(v3CacheDir, "plugins")
		log.Printf("[Helm 2] plugins \"%s\" will copy to [Helm 3] cache folder \"%s\" .\n", v2Plugins, v3Plugins)
		if !dryRun {
			err = copyDir(v2Plugins, v3Plugins)
			if err != nil {
				return fmt.Errorf("Failed to copy [Helm 2] plugins directory \"%s\" due to the following error: %s", v2Plugins, err)
			}
			log.Printf("[Helm 2] plugins \"%s\" copied successfully to [Helm 3] cache folder \"%s\" .\n", v2Plugins, v3Plugins)
		}

		// Recreate the  plugin symbolic links for v3 path
		v2Links := filepath.Join(v2HomeDir, "plugins")
		log.Printf("[Helm 2] plugin symbolic links \"%s\" will copy to [Helm 3] data folder \"%s\" .\n", v2Links, v3DataDir)
		if !dryRun {
			err = reCreatePluginSymLinks(v2Links, v3DataDir, v3CacheDir)
			if err != nil {
				return fmt.Errorf("Failed to copy [Helm 2] plugin links \"%s\" due to the following error: %s", v2Links, err)
			}
			log.Printf("[Helm 2] plugin links \"%s\" copied successfully to [Helm 3] data folder \"%s\" .\n", v2Links, v3DataDir)
		}
	}

	// Move starters
	v2Starters := filepath.Join(v2HomeDir, "starters")
	v3Starters := filepath.Join(v3DataDir, "starters")
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
	st, err := os.Stat(srcFileName)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(destFileName, input, st.Mode())
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
	if err != nil {
		return fmt.Errorf("Failed to copy directory  due to the following error: %s", err)
	}
	for _, obj := range objects {
		srcFileName := filepath.Join(srcDirName, obj.Name())
		destFileName := filepath.Join(destDirName, obj.Name())
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

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func copySymLink(fileInfo os.FileInfo, srcDirName, destDirName string) error {
	originFileName, err := os.Readlink(filepath.Join(srcDirName, fileInfo.Name()))
	if err != nil {
		return err
	}
	newSymLinkName := filepath.Join(destDirName, fileInfo.Name())
	err = os.Symlink(originFileName, newSymLinkName)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func reCreatePluginSymLinks(srcDirName, v3DataDir, v3CacheDir string) error {
	v3PluginDataDir := filepath.Join(v3DataDir, "plugins")
	err := ensureDir(v3PluginDataDir)
	if err != nil {
		return fmt.Errorf("Failed to create folder \"%s\" due to the following error: %s", v3PluginDataDir, err)
	}
	directory, _ := os.Open(srcDirName)
	objects, err := directory.Readdir(-1)
	if err != nil {
		return fmt.Errorf("Failed to re-create symlinks due to the following error: %s", err)
	}
	for _, obj := range objects {
		srcFileName := filepath.Join(srcDirName, obj.Name())
		if !obj.IsDir() {
			fileInfo, err := os.Lstat(srcFileName)
			if err != nil {
				return fmt.Errorf("Failed to check file \"%s\" stats  due to the following error: %s", srcFileName, err)
			}
			if fileInfo.Mode()&os.ModeSymlink != 0 {
				symLinkName := obj.Name()
				newFullSymLinkName := filepath.Join(v3PluginDataDir, symLinkName)
				origFullFileName, err := os.Readlink(filepath.Join(srcDirName, fileInfo.Name()))
				if err != nil {
					return fmt.Errorf("Failed to re-create symlink for \"%s\" due to the following error: %s", symLinkName, err)
				}
				newFullFileName := filepath.Join(v3CacheDir, "plugins", filepath.Base(origFullFileName))
				err = os.Symlink(newFullFileName, newFullSymLinkName)
				if err != nil && !os.IsExist(err) {
					return fmt.Errorf("Failed to re-create symlink for \"%s\" due to the following error: %s", newFullSymLinkName, err)
				}
			}
		}
	}
	return nil
}
