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
        "log"
        "os"
	"sync"
        
        "k8s.io/cli-runtime/pkg/genericclioptions"

	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/cli"
        "helm.sh/helm/pkg/kube"
        "helm.sh/helm/pkg/storage"
        "helm.sh/helm/pkg/storage/driver"
)

var (
        settings   cli.EnvSettings
        config     genericclioptions.RESTClientGetter
        configOnce sync.Once
)

// GetActionConfig returns action configuration based on Helm env
func GetActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

        // Initialize the rest of the actionConfig
	err := initActionConfig(actionConfig, namespace)
	if err != nil {
		return nil, err
	}

	return actionConfig, err
}

func initActionConfig(actionConfig *action.Configuration, namespace string) error {
        kc := kube.New(kubeConfig())
        kc.Log = logf

        clientset, err := kc.Factory.KubernetesClientSet()
        if err != nil {
                return err
        }

        var store *storage.Storage
        switch os.Getenv("HELM_DRIVER") {
        case "secret", "secrets", "":
                d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
                d.Log = logf
                store = storage.Init(d)
        case "configmap", "configmaps":
                d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
                d.Log = logf
                store = storage.Init(d)
        case "memory":
                d := driver.NewMemory()
                store = storage.Init(d)
        default:
                return fmt.Errorf("Unknown driver in HELM_DRIVER: " + os.Getenv("HELM_DRIVER"))
        }

        actionConfig.RESTClientGetter = kubeConfig()
        actionConfig.KubeClient = kc
        actionConfig.Releases = store
        actionConfig.Log = logf

	return nil
}

func kubeConfig() genericclioptions.RESTClientGetter {
        configOnce.Do(func() {
                config = kube.GetConfig(settings.KubeConfig, settings.KubeContext, settings.Namespace)
        })
        return config
}

func logf(format string, v ...interface{}) {
        if settings.Debug {
                format = fmt.Sprintf("[debug] %s\n", format)
                log.Output(2, fmt.Sprintf(format, v...))
        }
}
