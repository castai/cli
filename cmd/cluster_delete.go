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

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/sdk"
)

func newClusterDeleteCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <cluster_name_or_id>",
		Short: "Delete cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleDeleteCluster(cmd, log, api); err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd
}

func handleDeleteCluster(cmd *cobra.Command, log logrus.FieldLogger, api client.Interface) error {
	clusterID, err := getClusterID(cmd, api)
	if err != nil {
		return err
	}

	err = api.DeleteCluster(cmd.Context(), sdk.ClusterId(clusterID))
	if err != nil {
		return err
	}

	log.Info("Cluster deletion is now in progress")
	return nil
}