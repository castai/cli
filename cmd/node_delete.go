/*
Copyright © 2021 CAST AI

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
	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
)

var flagDeleteClusterNodeConfirm bool

func newNodeDeleteCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <node_name_or_id>",
		Short: "Delete clusters node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleDeleteNode(cmd, log, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().StringP(flagCluster, "c", "", "cluster name or ID")
	cmd.PersistentFlags().BoolVarP(&flagDeleteClusterNodeConfirm, "yes", "y", false, "confirm cluster node deletion")
	return cmd
}

func handleDeleteNode(cmd *cobra.Command, log logrus.FieldLogger, api client.Interface) error {
	clusterID, err := getClusterIDFromFlag(cmd, api)
	if err != nil {
		return err
	}

	node, err := getNode(cmd, api, clusterID)
	if err != nil {
		return err
	}

	if !flagDeleteClusterNodeConfirm {
		if err := survey.AskOne(&survey.Confirm{
			Message: "Are you sure?",
		}, &flagDeleteClusterNodeConfirm); err != nil {
			return err
		}
	}

	if !flagDeleteClusterNodeConfirm {
		log.Info("Cluster node delete canceled")
		return nil
	}

	if err := api.DeleteClusterNode(cmd.Context(), sdk.ClusterId(clusterID), *node.Id); err != nil {
		return err
	}

	log.Info("Cluster node deletion is now in progress")

	return nil
}
