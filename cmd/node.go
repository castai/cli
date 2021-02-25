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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
)

const (
	flagCluster = "cluster"
)

func newNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "node",
		Short: "Manage clusters nodes",
	}
}

func getNode(cmd *cobra.Command, api client.Interface, clusterID string) (*sdk.Node, error) {
	ctx := cmd.Context()
	// Select node from interactive picker if no args passed.
	if len(cmd.Flags().Args()) == 0 {
		node, err := selectNode(ctx, api, clusterID)
		if err != nil {
			return nil, err
		}
		return node, err
	}

	// Try to search single node by uuid.
	value := cmd.Flags().Args()[0]
	uuidID, err := uuid.Parse(value)
	if err == nil {
		node, err := api.GetClusterNode(cmd.Context(), sdk.ClusterId(clusterID), uuidID.String())
		if err != nil {
			return nil, err
		}
		return node, nil
	}

	// List all nodes and search by node name.
	nodes, err := api.ListClusterNodes(ctx, sdk.ClusterId(clusterID))
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if node.Name != nil && node.Id != nil && strings.ToLower(*node.Name) == strings.ToLower(value) {
			return &node, nil
		}
	}
	return nil, fmt.Errorf("node not found, searched by value=%s", value)
}

func selectNode(ctx context.Context, api client.Interface, clusterID string) (*sdk.Node, error) {
	items, err := api.ListClusterNodes(ctx, sdk.ClusterId(clusterID))
	if err != nil {
		return nil, err
	}
	selectList := make([]string, len(items))
	for i, item := range items {
		selectList[i] = *item.Name
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select node:",
		Options: selectList,
		Default: selectList[0],
	}

	if err := survey.AskOne(prompt, &selected, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	for _, item := range items {
		if *item.Name == selected {
			return &item, nil
		}
	}

	return nil, errors.New("cluster node not found")
}
