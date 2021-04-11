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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
)

func newClusterReconcileCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reconcile <cluster_name_or_id>",
		Short: "Trigger cluster reconcile",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleClusterReconcile(cmd, log, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	return cmd
}

func handleClusterReconcile(cmd *cobra.Command, log logrus.FieldLogger, api client.Interface) error {
	clusterID, err := getClusterIDFromArgs(cmd, api)
	if err != nil {
		return err
	}

	if err := api.TriggerClusterReconcile(cmd.Context(), sdk.ClusterId(clusterID)); err != nil {
		return err
	}

	log.Info("Cluster reconcile triggered successfully")
	return nil
}
