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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/config"
	"github.com/castai/cli/pkg/ipify"
	"github.com/castai/cli/pkg/ssh"
)

func NewRootCmd(log logrus.FieldLogger, cfg *config.Config, api client.Interface, terminal ssh.Terminal, ipify ipify.Client) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cast",
		Short: "CAST AI Command Line Interface",
		Long:  ``,
	}
	// Configure.
	rootCmd.AddCommand(newConfigureCmd(log))
	// Credentials.
	credentialsCmd := newCredentialsCmd()
	credentialsCmd.AddCommand(newCredentialsListCmd(log, api))
	rootCmd.AddCommand(credentialsCmd)
	// Cluster.
	clusterCmd := newClusterCmd()
	clusterCmd.AddCommand(newClusterListCmd(log, api))
	clusterCmd.AddCommand(newClusterGetCmd(log, api))
	clusterCmd.AddCommand(newClusterCreateCmd(log, cfg, api))
	clusterCmd.AddCommand(newClusterGetKubeconfigCmd(log, api))
	clusterCmd.AddCommand(newClusterDeleteCmd(log, api))
	clusterCmd.AddCommand(newClusterReconcileCmd(log, api))
	rootCmd.AddCommand(clusterCmd)
	// Cluster nodes.
	nodeCmd := newNodeCmd()
	nodeCmd.AddCommand(newNodeListCmd(log, api))
	nodeCmd.AddCommand(newNodeSSHCmd(log, api, terminal, ipify))
	nodeCmd.AddCommand(newNodeAddCmd(log, api))
	nodeCmd.AddCommand(newNodeDeleteCmd(log, api))
	rootCmd.AddCommand(nodeCmd)
	// Completion.
	rootCmd.AddCommand(newCompletionCmd())
	// Region.
	regionCmd := newRegionCmd()
	regionCmd.AddCommand(newRegionListCmd(log, api))
	rootCmd.AddCommand(regionCmd)
	// Version.
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}
