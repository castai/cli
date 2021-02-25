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
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
	"github.com/castai/cli/pkg/command"
)

func newRegionListCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List regions",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleListRegions(cmd, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	command.AddJSONOutput(cmd)
	return cmd
}

func handleListRegions(cmd *cobra.Command, api client.Interface) error {
	res, err := api.ListRegions(cmd.Context())
	if err != nil {
		return err
	}
	if command.OutputJSON() {
		command.PrintOutput(res)
		return nil
	}

	printRegionsListTable(cmd.OutOrStdout(), res)
	return nil
}

func printRegionsListTable(out io.Writer, items []sdk.CastRegion) {
	t := table.NewWriter()
	t.SetStyle(command.DefaultTableStyle)
	t.SetOutputMirror(out)
	t.AppendHeader(table.Row{"Name", "DisplayName", "Clouds"})
	for _, item := range items {
		t.AppendRow(table.Row{
			item.Name,
			item.DisplayName,
			strings.Join(getRegionClouds(item), " "),
		})
	}
	t.Render()
}

func getRegionClouds(item sdk.CastRegion) []string {
	res := make([]string, len(item.Clouds))
	for i, cloud := range item.Clouds {
		res[i] = cloud.Name
	}
	return res
}
