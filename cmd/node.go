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

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/client/sdk"
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

func parseNodeIDFromCMDArgs(cmd *cobra.Command, api client.Interface, clusterID string) (string, error) {
	if len(cmd.Flags().Args()) == 0 {
		return "", errors.New("node ID or name is required")
	}
	nodeIDOrName := cmd.Flags().Args()[0]
	return parseNodeIDFromValue(cmd.Context(), api, nodeIDOrName, clusterID)
}

func parseNodeIDFromValue(ctx context.Context, api client.Interface, nodeIDOrName, clusterID string) (string, error) {
	uuidID, err := uuid.Parse(nodeIDOrName)
	if err == nil {
		return uuidID.String(), nil
	}
	nodes, err := api.ListClusterNodes(ctx, sdk.ClusterId(clusterID))
	if err != nil {
		return "", err
	}
	for _, node := range nodes {
		if node.Name != nil && node.Id != nil && strings.ToLower(*node.Name) == strings.ToLower(nodeIDOrName) {
			return *node.Id, nil
		}
	}
	return "", fmt.Errorf("nodeID for %s not found", nodeIDOrName)
}
