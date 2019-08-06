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

	utils "github.com/maorfr/helm-plugin-utils/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

        rls "k8s.io/helm/pkg/proto/hapi/release"
)

// GetReleaseVersions returns all rrelease versions from Helm v2 storage for a specified release
func GetReleaseVersions(releaseName, tillerNamespace, label  string) ([]*rls.Release, error) {
	listOptions := utils.ListOptions{
		ReleaseName:     releaseName,
		TillerNamespace: tillerNamespace,
		TillerLabel:     label,
	}
	releases, err := getReleases(listOptions)
	if err != nil {
		return nil, err
	}
	if len(releases) <= 0 {
		return nil, fmt.Errorf("%s has no deployed releases", releaseName)
	}

	return releases, nil

}

// DeleteReleaseVersions deletes all release data from Helm v2 storage for a specified release
func DeleteReleaseVersions(releaseName, tillerNamespace string, versions []int32, dryRun bool) error {
	for _, ver := range versions {
		relVerName := fmt.Sprintf("%s.v%d", releaseName, ver)
                fmt.Printf("[Helm 2] ReleaseVersion \"%s\" will be deleted.\n", fmt.Sprintf("%s.v%d", releaseName, ver))
		if ! dryRun {
                        if err := deleteRelease(relVerName, tillerNamespace); err != nil {
                                return fmt.Errorf("[Helm 2] ReleaseVersion \"%s\" failed to delete with error: %s.\n", fmt.Sprintf("%s.v%d", releaseName, ver), err)
                        }
                        fmt.Printf("[Helm 2] ReleaseVersion \"%s\" deleted.\n", fmt.Sprintf("%s.v%d", releaseName, ver))
		}
        }

	return nil
}

func getReleases(o utils.ListOptions) ([]*rls.Release, error) {
	if o.TillerNamespace == "" {
		o.TillerNamespace = "kube-system"
	}
	if o.TillerLabel == "" {
		o.TillerLabel = "OWNER=TILLER"
	}
	if o.ReleaseName != "" {
		o.TillerLabel += fmt.Sprintf(",NAME=%s", o.ReleaseName)
	}
	clientSet := utils.GetClientSet()
	var releases []*rls.Release
	storage := utils.GetTillerStorage(o.TillerNamespace)
	switch storage {
	case "secrets":
		secrets, err := clientSet.CoreV1().Secrets(o.TillerNamespace).List(metav1.ListOptions{
			LabelSelector: o.TillerLabel,
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
		configMaps, err := clientSet.CoreV1().ConfigMaps(o.TillerNamespace).List(metav1.ListOptions{
			LabelSelector: o.TillerLabel,
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

	if len(releases) > 1 {
		releases = reverse(releases)
	}

	return releases, nil
}

func getRelease(itemReleaseData string) *rls.Release {
	data, _ := utils.DecodeRelease(itemReleaseData)
	return data
}

func deleteRelease(releaseVersionName, tillerNamespace string)  error {
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
        clientSet := utils.GetClientSet()
        storage := utils.GetTillerStorage(tillerNamespace)
        switch storage {
        case "secrets":
                return clientSet.CoreV1().Secrets(tillerNamespace).Delete(releaseVersionName, &metav1.DeleteOptions{})
        case "configmaps":
                return clientSet.CoreV1().ConfigMaps(tillerNamespace).Delete(releaseVersionName, &metav1.DeleteOptions{})
        }
	return nil
}

func reverse(releases []*rls.Release) []*rls.Release {
	for i := 0; i < len(releases)/2; i++ {
		j := len(releases) - i - 1
		releases[i], releases[j] = releases[j], releases[i]
	}
	return releases
}
