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

package v2

import (
	"fmt"
	"os"
	"strings"

	utils "github.com/maorfr/helm-plugin-utils/pkg"
	"github.com/mitchellh/go-homedir"
)

// RemoveHomeFolder removes the v2 Helm home folder
func RemoveHomeFolder(dryRun bool) error {
	homeDir := HomeDir()
	fmt.Printf("[Helm 2] Home folder \"%s\" will be deleted.\n", homeDir)
	if !dryRun {
		if err := os.RemoveAll(homeDir); err != nil {
			return fmt.Errorf("[Helm 2] Failed to delete \"%s\" due to the following error: %s.\n", homeDir, err)
		}
		fmt.Printf("[Helm 2] Home folder \"%s\" deleted.\n", homeDir)
	}
	return nil

}

// RemoveTiller removes Tiller service in a particular namespace from the cluster
func RemoveTiller(tillerNamespace string, dryRun bool) error {
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	if !dryRun {
		applyCmd := []string{"kubectl", "delete", "--namespace", tillerNamespace, "deploy/tiller-deploy"}
		output := utils.Execute(applyCmd)
		if !strings.Contains(string(output), "\"tiller-deploy\" deleted") {
			return fmt.Errorf("[Helm 2] Failed to remove Tiller service in \"%s\" namespace  due to the following error: %s", tillerNamespace, string(output))
		}
	}
	return nil
}

// HomeDir return the Helm home folder
func HomeDir() string {
	if homeDir, exists := os.LookupEnv("HELM_V2_HOME"); exists {
		return homeDir
	}

	homeDir, _ := homedir.Dir()
	defaultDir := homeDir + "/.helm"
	return defaultDir
}

// GetReleaseVersionName returns release version name
func GetReleaseVersionName(releaseName string, releaseVersion int32) string {
        return fmt.Sprintf("%s.v%d", releaseName, releaseVersion)
}
