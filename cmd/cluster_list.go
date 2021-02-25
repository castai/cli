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
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
	"github.com/castai/cli/pkg/command"
)

var (
	flagIncludeDeletedClusters bool
)

func newClusterListCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all clusters",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleListClusters(cmd, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().BoolVar(&flagIncludeDeletedClusters, "include-deleted", false, "Show deleted clusters too.")
	command.AddJSONOutput(cmd)
	return cmd
}

func handleListClusters(cmd *cobra.Command, api client.Interface) error {
	resp, err := api.ListKubernetesClusters(cmd.Context(), &sdk.ListKubernetesClustersParams{})
	if err != nil {
		return err
	}

	if command.OutputJSON() {
		command.PrintOutput(resp)
		return nil
	}

	printClustersListTable(cmd.OutOrStdout(), resp)
	return nil
}

func printClustersListTable(out io.Writer, items []sdk.KubernetesCluster) {
	t := table.NewWriter()
	t.SetStyle(command.DefaultTableStyle)
	t.SetOutputMirror(out)
	t.AppendHeader(table.Row{"ID", "Name", "Status", "Clouds", "Region"})
	for _, item := range items {
		t.AppendRow(table.Row{
			item.Id,
			item.Name,
			item.Status,
			strings.Join(getClusterCloudsNames(item), " "),
			item.Region.DisplayName,
		})
	}
	t.Render()
}

func getClusterCloudsNames(item sdk.KubernetesCluster) []string {
	var res []string
	clouds := map[string]struct{}{}
	for _, node := range item.Nodes {
		clouds[string(node.Cloud)] = struct{}{}
	}
	for c := range clouds {
		res = append(res, c)
	}
	return res
}
