/*
Copyright © 2020 CAST AI

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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/sdk"
)

var (
	flagKubeconfigPath string
)

func newClusterGetKubeconfigCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-kubeconfig <cluster_name_or_id>",
		Short: "Get cluster kubeconfig",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleClusterGetKubeconfig(cmd, log, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	defaultKubeConfigDir := getDefaultKubeconfigPath()
	cmd.PersistentFlags().StringVar(&flagKubeconfigPath, "path", defaultKubeConfigDir, "(optional) absolute path to the kubeconfig file")
	if defaultKubeConfigDir == "" {
		cmd.MarkPersistentFlagRequired("path")
	}
	return cmd
}

func getDefaultKubeconfigPath() string {
	home := homedir.HomeDir()
	if home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

func handleClusterGetKubeconfig(cmd *cobra.Command, log logrus.FieldLogger, api client.Interface) (err error) {
	clusterID, err := getClusterID(cmd, api)
	if err != nil {
		return err
	}

	resp, err := api.GetClusterKubeconfig(cmd.Context(), sdk.ClusterId(clusterID))
	if err != nil {
		return err
	}
	newConfig, err := getRawConfig(resp)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			log.Infof("Kubeconfig saved to %s", flagKubeconfigPath)
		}
	}()

	// If there is no already created kubconfig in given path create it and exit.
	if _, err := os.Stat(flagKubeconfigPath); os.IsNotExist(err) {
		if err := clientcmd.WriteToFile(newConfig, flagKubeconfigPath); err != nil {
			return err
		}
		return nil
	}

	// Merge with existing kubeconfig and set default context to new config's context.
	currentConfigBytes, err := ioutil.ReadFile(flagKubeconfigPath)
	if err != nil {
		return err
	}
	currentConfig, err := getRawConfig(currentConfigBytes)
	if err != nil {
		return err
	}

	currentConfig = mergeConfigs(currentConfig, newConfig)
	if err := clientcmd.WriteToFile(currentConfig, flagKubeconfigPath); err != nil {
		return err
	}

	return nil
}

func mergeConfigs(current, other api.Config) api.Config {
	for k, v := range other.Clusters {
		current.Clusters[k] = v
	}
	for k, v := range other.Contexts {
		current.Contexts[k] = v
	}
	for k, v := range other.AuthInfos {
		current.AuthInfos[k] = v
	}
	current.CurrentContext = other.CurrentContext
	return current
}

func getRawConfig(bytes []byte) (api.Config, error) {
	config, err := clientcmd.NewClientConfigFromBytes(bytes)
	if err != nil {
		return api.Config{}, err
	}
	return config.RawConfig()
}
