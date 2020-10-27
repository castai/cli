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
	"net/http"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/internal/client"
	"github.com/castai/cast-cli/pkg/command"
	"github.com/castai/cast-cli/pkg/sdk"
)

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleListClusters(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
	command.AddJSONOutput(clusterListCmd)
}

func handleListClusters() error {
	apiClient, err := client.New()
	if err != nil {
		return err
	}

	ctx, cancel := client.DefaultContext()
	defer cancel()

	resp, err := apiClient.ListKubernetesClustersWithResponse(ctx, &sdk.ListKubernetesClustersParams{})
	if err := client.CheckResponse(resp, err, http.StatusOK); err != nil {
		return err
	}

	if command.OutputJSON() {
		command.PrintOutput(resp.JSON200.Items)
		return nil
	}

	printClustersListTable(resp.JSON200.Items)
	return nil
}

func printClustersListTable(clusters []sdk.KubernetesCluster) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Status", "Clouds", "Region"})
	for _, cluster := range clusters {
		t.AppendRow(table.Row{
			cluster.Id,
			cluster.Name,
			cluster.Status,
			strings.Join(getClusterCloudsNames(cluster), " "),
			cluster.Region.DisplayName,
		})
		t.AppendSeparator()
	}
	t.Render()
}

func getClusterCloudsNames(cluster sdk.KubernetesCluster) []string {
	var res []string
	clouds := map[string]struct{}{}
	for _, node := range cluster.Nodes {
		clouds[string(node.Cloud)] = struct{}{}
	}
	for c := range clouds {
		res = append(res, c)
	}
	return res
}
