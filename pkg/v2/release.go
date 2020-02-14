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
	"log"
	"sort"

	utils "github.com/maorfr/helm-plugin-utils/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rls "k8s.io/helm/pkg/proto/hapi/release"

	common "github.com/helm/helm-2to3/pkg/common"
)

type RetrieveOptions struct {
	ReleaseName      string
	StorageType      string
	TillerLabel      string
	TillerNamespace  string
	TillerOutCluster bool
}

type DeleteOptions struct {
	DryRun   bool
	Versions []int32
}

// ByReleaseVersion implements sort.Interface based on the rls.Release Version field
type ByReleaseVersion []*rls.Release

func (releases ByReleaseVersion) Len() int           { return len(releases) }
func (releases ByReleaseVersion) Less(i, j int) bool { return releases[i].Version < releases[j].Version }
func (releases ByReleaseVersion) Swap(i, j int)      { releases[i], releases[j] = releases[j], releases[i] }

// GetReleaseVersions returns all release versions from Helm v2 storage for a specified release..
// It is based on Tiller namespace and labels like owner of storage.
func GetReleaseVersions(retOpts RetrieveOptions, kubeConfig common.KubeConfig) ([]*rls.Release, error) {
	releases, err := getReleases(retOpts, kubeConfig)
	if err != nil {
		return nil, err
	}
	if len(releases) <= 0 {
		return nil, fmt.Errorf("%s has no deployed releases\n", retOpts.ReleaseName)
	}

	return releases, nil

}

// DeleteReleaseVersions deletes all release data from Helm v2 storage for a specified release.
// It is based on Tiller namespace and labels like owner of storage.
func DeleteReleaseVersions(retOpts RetrieveOptions, delOpts DeleteOptions, kubeConfig common.KubeConfig) error {
	for _, ver := range delOpts.Versions {
		relVerName := fmt.Sprintf("%s.v%d", retOpts.ReleaseName, ver)
		log.Printf("[Helm 2] ReleaseVersion \"%s\" will be deleted.\n", relVerName)
		if !delOpts.DryRun {
			if err := deleteRelease(retOpts, relVerName, kubeConfig); err != nil {
				return fmt.Errorf("[Helm 2] ReleaseVersion \"%s\" failed to delete with error: %s.\n", relVerName, err)
			}
			log.Printf("[Helm 2] ReleaseVersion \"%s\" deleted.\n", relVerName)
		}
	}

	return nil
}

// DeleteReleaseVersions deletes all release data from Helm v2 storage.
// It is based on Tiller namespace and labels like owner of storage.
func DeleteAllReleaseVersions(retOpts RetrieveOptions, kubeConfig common.KubeConfig, dryRun bool) error {
	if retOpts.TillerNamespace == "" {
		retOpts.TillerNamespace = "kube-system"
	}
	if retOpts.TillerLabel == "" {
		retOpts.TillerLabel = "OWNER=TILLER"
	}
	if retOpts.StorageType == "" {
		retOpts.StorageType = "configmaps"
	}

	// Get all release versions stored for that namespace and owner
	releases, err := getReleases(retOpts, kubeConfig)
	if err != nil {
		return err
	}
	releaseLen := len(releases)
	if releaseLen <= 0 {
		log.Printf("[Helm 2] no deployed releases for namespace: %s, owner: %s\n", retOpts.TillerNamespace, retOpts.TillerLabel)
		return nil
	}

	// Delete each release version from storage
	for i := 0; i < releaseLen; i++ {
		release := releases[i]
		relVerName := GetReleaseVersionName(release.Name, release.Version)
		log.Printf("[Helm 2] ReleaseVersion \"%s\" will be deleted.\n", relVerName)
		if !dryRun {
			if err := deleteRelease(retOpts, relVerName, kubeConfig); err != nil {
				return fmt.Errorf("[Helm 2] ReleaseVersion \"%s\" failed to delete with error: %s.\n", relVerName, err)
			}
			log.Printf("[Helm 2] ReleaseVersion \"%s\" deleted.\n", relVerName)
		}
	}
	return nil
}

func getReleases(retOpts RetrieveOptions, kubeConfig common.KubeConfig) ([]*rls.Release, error) {
	if retOpts.TillerNamespace == "" {
		retOpts.TillerNamespace = "kube-system"
	}
	if retOpts.TillerLabel == "" {
		retOpts.TillerLabel = "OWNER=TILLER"
	}
	if retOpts.ReleaseName != "" {
		retOpts.TillerLabel += fmt.Sprintf(",NAME=%s", retOpts.ReleaseName)
	}
	if retOpts.StorageType == "" {
		retOpts.StorageType = "configmaps"
	}
	storage := getStorageType(retOpts, kubeConfig)
	clientSet := utils.GetClientSetWithKubeConfig(kubeConfig.File, kubeConfig.Context)
	var releases []*rls.Release
	switch storage {
	case "secrets":
		secrets, err := clientSet.CoreV1().Secrets(retOpts.TillerNamespace).List(metav1.ListOptions{
			LabelSelector: retOpts.TillerLabel,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range secrets.Items {
			release := getRelease((string)(item.Data["release"]))
			if release == nil {
				continue
			}
			releases = append(releases, release)
		}
	case "configmaps":
		configMaps, err := clientSet.CoreV1().ConfigMaps(retOpts.TillerNamespace).List(metav1.ListOptions{
			LabelSelector: retOpts.TillerLabel,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range configMaps.Items {
			release := getRelease(item.Data["release"])
			if release == nil {
				continue
			}
			releases = append(releases, release)
		}
	}

	sort.Sort(ByReleaseVersion(releases))

	return releases, nil
}

func getStorageType(retOpts RetrieveOptions, kubeConfig common.KubeConfig) string {
	var storage string
	if !retOpts.TillerOutCluster {
		storage = utils.GetTillerStorageWithKubeConfig(retOpts.TillerNamespace, kubeConfig.File, kubeConfig.Context)
	} else {
		storage = retOpts.StorageType
	}
	return storage
}

func getRelease(itemReleaseData string) *rls.Release {
	data, _ := utils.DecodeRelease(itemReleaseData)
	return data
}

func deleteRelease(retOpts RetrieveOptions, releaseVersionName string, kubeConfig common.KubeConfig) error {
	if retOpts.TillerNamespace == "" {
		retOpts.TillerNamespace = "kube-system"
	}
	if retOpts.StorageType == "" {
		retOpts.StorageType = "configmaps"
	}
	storage := getStorageType(retOpts, kubeConfig)
	clientSet := utils.GetClientSetWithKubeConfig(kubeConfig.File, kubeConfig.Context)
	switch storage {
	case "secrets":
		return clientSet.CoreV1().Secrets(retOpts.TillerNamespace).Delete(releaseVersionName, &metav1.DeleteOptions{})
	case "configmaps":
		return clientSet.CoreV1().ConfigMaps(retOpts.TillerNamespace).Delete(releaseVersionName, &metav1.DeleteOptions{})
	}
	return nil
}
