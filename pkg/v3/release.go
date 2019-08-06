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

package v3
  
import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"

        "helm.sh/helm/pkg/chart"
        "helm.sh/helm/pkg/release"

	v2chrtutil "k8s.io/helm/pkg/chartutil"
	v2chart "k8s.io/helm/pkg/proto/hapi/chart"
	v2rls "k8s.io/helm/pkg/proto/hapi/release"
)

// CreateRelease create a v3 release object from v3 release object
func CreateRelease(v2Rel *v2rls.Release) (*release.Release, error) {
	if v2Rel.Chart == nil || v2Rel.Info == nil {
		return nil, fmt.Errorf("No v2 chart or info metadata")
	}
	chrt, err := mapv2ChartTov3Chart(v2Rel.Chart)
	if err != nil {
		return nil, err
	}
	config, err := mapConfig(v2Rel.Config)
        if err != nil {
                return nil, err
        }
	first, err := mapTimestampToTime(v2Rel.Info.FirstDeployed)
	if err != nil {
                return nil, err
        }
	last, err := mapTimestampToTime(v2Rel.Info.LastDeployed)
	if err != nil {
                return nil, err
	}
	deleted, err := mapTimestampToTime(v2Rel.Info.Deleted)
	if err != nil {
                return nil, err
	}
	status, ok := v2rls.Status_Code_name[int32(v2Rel.Info.Status.Code)]
	if !ok {
		return nil,  fmt.Errorf("Failed to convert status")
	}
	hooks, err:= mapHooks(v2Rel.Hooks)
	if err != nil {
                return nil, err
	}
	lastTestSuiteRun, err := mapTestSuite(v2Rel.Info.Status.LastTestSuiteRun)
	if err != nil {
                return nil, err
        }
        return &release.Release{
                Name:      v2Rel.Name,
                Namespace: v2Rel.Namespace,
                Chart:     chrt,
                Config:    config,
                Info: &release.Info{
                        FirstDeployed:    first,
                        LastDeployed:     last,
			Description:      v2Rel.Info.Description,
			Deleted:          deleted,
                        Status:           release.Status(strings.ToLower(status)),
			Notes:            v2Rel.Info.Status.Notes,
			Resources:        v2Rel.Info.Status.Resources,
			LastTestSuiteRun: lastTestSuiteRun,
                },
		Manifest:v2Rel.Manifest,
		Hooks: hooks,
                Version: int(v2Rel.Version),
        }, nil
}

// StoreRelease stores a release objevct in Helm v3 storage
func StoreRelease(rel *release.Release) error {
	cfg, err := GetActionConfig(rel.Namespace)
	if err != nil {
		return err
	}

	return cfg.Releases.Create(rel)
}

func mapv2ChartTov3Chart(v2Chrt *v2chart.Chart) (*chart.Chart, error) {
	 v3Chrt := new(chart.Chart)
	 v3Chrt.Metadata = mapMetadata(v2Chrt)
	 v3Chrt.Templates = mapTemplates(v2Chrt.Templates)
	 err := mapDependencies(v2Chrt.Dependencies, v3Chrt)
	 if err != nil {
		 return nil, err
         }
	 if v3Chrt.Values, err = mapConfig(v2Chrt.Values); err != nil {
		 return nil, err
         }
	 v3Chrt.Files = mapFiles(v2Chrt.Files)
	 //TODO
	 //v3Chrt.Schema
	 //TODO
	 //v3Chrt.Lock = new(chart.Lock)
	 return v3Chrt, nil
}

func mapMetadata(v2Chrt *v2chart.Chart) (*chart.Metadata) {
	if v2Chrt.Metadata == nil {
		return nil
	}
	metadata := new(chart.Metadata)
	metadata.Name = v2Chrt.Metadata.Name
        metadata.Home = v2Chrt.Metadata.Home
        metadata.Sources = v2Chrt.Metadata.Sources
        metadata.Version = v2Chrt.Metadata.Version
        metadata.Description = v2Chrt.Metadata.Description
        metadata.Keywords = v2Chrt.Metadata.Keywords
	metadata.Maintainers = mapMaintainers(v2Chrt.Metadata.Maintainers)
        metadata.Icon = v2Chrt.Metadata.Icon
        metadata.APIVersion = v2Chrt.Metadata.ApiVersion
        metadata.Condition = v2Chrt.Metadata.Condition
        metadata.Tags = v2Chrt.Metadata.Tags
        metadata.AppVersion = v2Chrt.Metadata.AppVersion
        metadata.Deprecated = v2Chrt.Metadata.Deprecated
        metadata.Annotations = v2Chrt.Metadata.Annotations
        metadata.KubeVersion  = v2Chrt.Metadata.KubeVersion
	//TODO: metadata.Dependencies = 
	//Default to application
	metadata.Type =  "application"
	return metadata
}

func mapMaintainers(v2Maintainers []*v2chart.Maintainer) ([]*chart.Maintainer) {
	if v2Maintainers  == nil {
		return nil
	}
	maintainers := []*chart.Maintainer{}
	for _, val := range v2Maintainers {
		maintainer := new(chart.Maintainer)
		maintainer.Name = val.Name
		maintainer.Email = val.Email
		maintainer.URL = val.Url
		maintainers = append(maintainers, maintainer)
	}
	return maintainers
}

func mapTemplates(v2Templates []*v2chart.Template) ([]*chart.File) {
	if v2Templates  == nil {
		return nil
	}
	files := []*chart.File{}
	 for _, val := range v2Templates {
		 file := new(chart.File)
		 file.Name = val.Name
		 file.Data = val.Data
		 files = append(files, file)
	 }
	 return files
}

func mapDependencies(v2Dependencies []*v2chart.Chart, chart *chart.Chart) error {
	if v2Dependencies  == nil {
		return nil
	}
	for _, val := range v2Dependencies {
		dependency, err := mapv2ChartTov3Chart(val)
		if err != nil {
			return err
		}
		chart.AddDependency(dependency)
	}
	return nil
}

func mapConfig(v2Config *v2chart.Config) (map[string]interface{}, error) {
	if v2Config  == nil {
		return nil, nil
	}
	values, err := v2chrtutil.ReadValues([]byte(v2Config.Raw))
        if err != nil {
		return nil, err
        }
	return values, nil
}

func mapFiles(v2Files []*any.Any) ([]*chart.File) {
	if mapFiles  == nil {
		return nil
	}
	files := []*chart.File{}
	for _, f := range v2Files {
		file := new(chart.File)
		file.Name = f.TypeUrl
                file.Data = f.Value
                files = append(files, file)
	 }
	 return files
}

func mapHooks(v2Hooks []*v2rls.Hook) ([]*release.Hook, error) {
	if v2Hooks  == nil {
		return nil, nil
	}
	hooks := []*release.Hook{}
	for _, val := range v2Hooks {
		hook := new(release.Hook)
		hook.Name = val.Name
		hook.Kind = val.Kind
                hook.Path = val.Path 
                hook.Manifest =  val.Manifest
		events, err := mapHookEvents(val.Events)
		if err != nil {
			return nil, err
		}
		hook.Events = events
		hook.Weight = int(val.Weight)
		lastRun, err:= mapTimestampToTime(val.LastRun)
                if err != nil {
                       return nil, err
                }
		hook.LastRun = lastRun
		policies, err := mapHookDeletePolicies(val.DeletePolicies)
                if err != nil {
                        return nil, err
                }
                hook.DeletePolicies = policies
                hooks = append(hooks, hook)
	}
	return hooks, nil
}

func mapHookEvents(v2HookEvents [] v2rls.Hook_Event) ([]release.HookEvent, error) {
	if v2HookEvents  == nil {
		return nil, nil
	}
	hookEvents := []release.HookEvent{}
	for _, val := range v2HookEvents {
	        v2EventStr, ok := v2rls.Hook_Event_name[int32(val)]
                if !ok {
                        return nil,  fmt.Errorf("Failed to convert hook event")
                }
	        event := release.HookEvent(strings.ToLower(v2EventStr))
		hookEvents = append(hookEvents, event)
	}
	return hookEvents, nil
}

func mapHookDeletePolicies(v2HookDelPolicies [] v2rls.Hook_DeletePolicy) ([]release.HookDeletePolicy, error) {
	if v2HookDelPolicies  == nil {
		return nil, nil
	}
	hookDelPolicies := []release.HookDeletePolicy{}
	for _, val := range v2HookDelPolicies {
	        v2PolicyStr, ok := v2rls.Hook_DeletePolicy_name[int32(val)]
                if !ok {
                        return nil,  fmt.Errorf("Failed to convert hook delete policy")
                }
	        policy := release.HookDeletePolicy(strings.ToLower(v2PolicyStr))
		hookDelPolicies = append(hookDelPolicies, policy)
	}
	return hookDelPolicies, nil
}

func mapTestSuite(v2LastTestRun *v2rls.TestSuite) (*release.TestSuite, error) {
        if v2LastTestRun == nil {
		return nil, nil
	}
	testSuite := new(release.TestSuite)
	startAt, err := mapTimestampToTime(v2LastTestRun.StartedAt)
        if err != nil {
               return nil, err
        }
        testSuite.StartedAt = startAt
	completeAt, err:= mapTimestampToTime(v2LastTestRun.CompletedAt)
        if err != nil {
               return nil, err
        }
        testSuite.CompletedAt = completeAt
	testRuns, err := mapTestRuns(v2LastTestRun.Results)
	if err != nil {
               return nil, err
        }
        testSuite.Results = testRuns
	return testSuite, nil
}

func mapTestRuns(v2Runs []*v2rls.TestRun) ([]*release.TestRun, error) {
	if v2Runs == nil {
		return nil, nil
	}
	testRuns := []*release.TestRun{}
	for _, val := range v2Runs {
		run := new(release.TestRun)
		run.Name = val.Name
		run.Info = val.Info
		startAt, err:= mapTimestampToTime(val.StartedAt)
                if err != nil {
                       return nil, err
                }
                run.StartedAt = startAt
                completeAt, err:= mapTimestampToTime(val.CompletedAt)
                if err != nil {
                       return nil, err
                }
                run.CompletedAt = completeAt
	        v2RunStatusStr, ok := v2rls.TestRun_Status_name[int32(val.Status)]
                if !ok {
                        return nil,  fmt.Errorf("Failed to convert test run status")
                }
	        run.Status = release.TestRunStatus(strings.ToLower(v2RunStatusStr))
		testRuns = append(testRuns, run)
	}
	return testRuns, nil
}

func mapTimestampToTime(ts *timestamp.Timestamp) (time.Time, error){
	var mappedTime time.Time
	var err error
        if ts != nil {
		mappedTime, err = ptypes.Timestamp(ts)
                if err != nil {
                        return mappedTime, err
                }
        }
	return mappedTime, nil
}
