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
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/client/sdk"
)

func newClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cluster",
		Short: "Manage clusters",
	}
}

func getClusterID(cmd *cobra.Command, api client.Interface) (string, error) {
	if len(cmd.Flags().Args()) == 0 {
		return "", errors.New("cluster ID or name is required")
	}
	clusterIDOrName := cmd.Flags().Args()[0]
	uuidID, err := uuid.Parse(clusterIDOrName)
	if err == nil {
		return uuidID.String(), nil
	}
	clusters, err := api.ListKubernetesClusters(cmd.Context(), &sdk.ListKubernetesClustersParams{})
	if err != nil {
		return "", err
	}
	for _, cluster := range clusters {
		if strings.ToLower(cluster.Name) == strings.ToLower(clusterIDOrName) {
			return cluster.Id, nil
		}
	}
	return "", fmt.Errorf("clusterID for %s not found", clusterIDOrName)
}
