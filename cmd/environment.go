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

import "github.com/spf13/pflag"

type EnvSettings struct {
	DryRun           bool
	KubeConfigFile   string
	KubeContext      string
	Label            string
	ReleaseStorage   string
	TillerNamespace  string
	TillerOutCluster bool
}

func New() *EnvSettings {
	envSettings := EnvSettings{}
	return &envSettings
}

// AddBaseFlags binds base flags to the given flagset.
func (s *EnvSettings) AddBaseFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&s.DryRun, "dry-run", false, "simulate a command")
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	s.AddBaseFlags(fs)
	fs.StringVar(&s.KubeConfigFile, "kubeconfig", "", "path to the kubeconfig file")
	fs.StringVar(&s.KubeContext, "kube-context", s.KubeContext, "name of the kubeconfig context to use")
	fs.StringVarP(&s.TillerNamespace, "tiller-ns", "t", "kube-system", "namespace of Tiller")
	fs.StringVarP(&s.Label, "label", "l", "OWNER=TILLER", "label to select Tiller resources by")
	fs.BoolVar(&s.TillerOutCluster, "tiller-out-cluster", false, "when  Tiller is not running in the cluster e.g. Tillerless")
	fs.StringVarP(&s.ReleaseStorage, "release-storage", "s", "secrets", "v2 release storage type/object. It can be 'secrets' or 'configmaps'. This is only used with the 'tiller-out-cluster' flag")

}
