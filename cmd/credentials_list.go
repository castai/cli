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

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/client/sdk"
	"github.com/castai/cast-cli/pkg/command"
)

func newCredentialsListCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud credentials",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleCredentialsList(cmd, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	command.AddJSONOutput(cmd)
	return cmd
}

func handleCredentialsList(cmd *cobra.Command, api client.Interface) error {
	res, err := api.ListCloudCredentials(cmd.Context())
	if err != nil {
		return err
	}

	if command.OutputJSON() {
		command.PrintOutput(res)
		return nil
	}

	printCredentialsListTable(cmd.OutOrStdout(), res)
	return nil
}

func printCredentialsListTable(out io.Writer, items []sdk.CloudCredentials) {
	t := table.NewWriter()
	t.SetStyle(command.DefaultTableStyle)
	t.SetOutputMirror(out)
	t.AppendHeader(table.Row{"ID", "Name", "Cloud", "Clusters"})
	for _, item := range items {
		t.AppendRow(table.Row{
			item.Id,
			item.Name,
			item.Cloud,
			getCredentialsUsedByCount(item),
		})
	}
	t.Render()
}

func getCredentialsUsedByCount(item sdk.CloudCredentials) int {
	if item.UsedBy == nil {
		return 0
	}
	return len(*item.UsedBy)
}
