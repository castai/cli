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
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
	"github.com/castai/cli/pkg/command"
)

func newNodeListCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clusters nodes",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleListNodes(cmd, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().StringP(flagCluster, "c", "", "cluster name or ID")
	command.AddJSONOutput(cmd)
	return cmd
}

func handleListNodes(cmd *cobra.Command, api client.Interface) error {
	clusterID, err := getClusterIDFromFlag(cmd, api)
	if err != nil {
		return err
	}
	res, err := api.ListClusterNodes(cmd.Context(), sdk.ClusterId(clusterID))
	if err != nil {
		return err
	}

	if command.OutputJSON() {
		command.PrintOutput(res)
		return nil
	}

	printNodesListTable(cmd.OutOrStdout(), res)

	return nil
}

func printNodesListTable(out io.Writer, items []sdk.Node) {
	t := table.NewWriter()
	t.SetStyle(command.DefaultTableStyle)
	t.SetOutputMirror(out)
	t.AppendHeader(table.Row{"ID", "Name", "Role", "Shape", "Status", "Cloud", "Public_IP", "Private_IP"})
	for _, item := range items {
		if item.Network == nil {
			item.Network = &sdk.NodeNetwork{}
		}
		if item.State == nil {
			var empty string
			item.State = &sdk.NodeState{Phase: &empty}
		}
		t.AppendRow(table.Row{
			nodeValueString(item.Id),
			nodeValueString(item.Name),
			item.Role,
			item.Shape,
			nodeValueString(item.State.Phase),
			item.Cloud,
			item.Network.PublicIp,
			item.Network.PrivateIp,
		})
	}
	t.Render()
}

func nodeValueString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
