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
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
)

type addNodeFlags struct {
	Cloud        string `survey:"cloud"`
	Role         string `survey:"role"`
	Shape        string `survey:"shape"`
	InstanceType string `survey:"instanceType"`
}

var addNodeFlagsData addNodeFlags
var (
	supportedClouds     = []string{"aws", "azure", "gcp", "do"}
	supportedNodeRoles  = []string{"worker", "master"}
	supportedNodeShapes = []string{"small", "medium", "large"}
)

func newNodeAddCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add clusters node",
		Long: `You can create cluster nodes in two ways:
1. Interactive - answer questions one by one with suggested values. 
   To create cluster node in interactive mode just run:
   cast node add
2. Declarative - pass cluster node creation flags manually.

Examples:
  # Add worker node on aws cloud.
  cast node add -c=my-cluster --cloud=aws --role=worker --shape=medium
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleAddNode(cmd, log, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().StringP(flagCluster, "c", "", "cluster name or ID")
	cmd.PersistentFlags().StringVar(&addNodeFlagsData.Cloud, "cloud", "", fmt.Sprintf("node cloud name, possible values: %s)", strings.Join(supportedClouds, ",")))
	cmd.PersistentFlags().StringVar(&addNodeFlagsData.Role, "role", "worker", fmt.Sprintf("node role, possible values: %s)", strings.Join(supportedNodeRoles, ",")))
	cmd.PersistentFlags().StringVar(&addNodeFlagsData.Shape, "shape", "medium", fmt.Sprintf("node shape, possible values: %s)", strings.Join(supportedNodeShapes, ",")))
	cmd.PersistentFlags().StringVar(&addNodeFlagsData.Shape, "instance-type", "", fmt.Sprintf("node instance type"))
	return cmd
}

func handleAddNode(cmd *cobra.Command, log logrus.FieldLogger, api client.Interface) error {
	cluster, err := getClusterFromFlag(cmd, api)
	if err != nil {
		return err
	}
	ctx := cmd.Context()

	var node *sdk.Node
	if cmd.Flags().NFlag() == 1 {
		creds, err := api.ListCloudCredentials(ctx)
		if err != nil {
			return err
		}
		node, err = parseDeclarativeAddNodeForm(cluster, creds)
		if err != nil {
			return err
		}
	} else {
		node = &sdk.Node{
			Cloud:        sdk.CloudType(addNodeFlagsData.Cloud),
			Role:         sdk.NodeType(addNodeFlagsData.Role),
			Shape:        sdk.NodeShape(addNodeFlagsData.Shape),
			InstanceType: addNodeFlagsData.InstanceType,
		}
	}

	if err := api.AddClusterNode(ctx, sdk.ClusterId(cluster.Id), *node); err != nil {
		return err
	}

	log.Info("Cluster node creation is now in progress")

	return nil
}

func parseDeclarativeAddNodeForm(cluster *sdk.KubernetesCluster, creds []sdk.CloudCredentials) (*sdk.Node, error) {
	var clusterClouds []string
	for _, credID := range cluster.CloudCredentialsIDs {
		for _, cred := range creds {
			if cred.Id == credID {
				clusterClouds = append(clusterClouds, cred.Name)
				break
			}
		}
	}

	if len(clusterClouds) == 0 {
		return nil, fmt.Errorf("could not find cluster credentials names, cluster_id=%s", cluster.Id)
	}

	if err := survey.AskOne(&survey.Select{
		Message: "Select cloud:",
		Options: clusterClouds,
		Default: clusterClouds[0],
	}, &addNodeFlagsData.Cloud, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	if err := survey.AskOne(&survey.Select{
		Message: "Select role:",
		Options: supportedNodeRoles,
		Default: supportedNodeRoles[0],
	}, &addNodeFlagsData.Role, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	if err := survey.AskOne(&survey.Select{
		Message: "Select shape:",
		Options: supportedNodeShapes,
		Default: supportedNodeShapes[0],
	}, &addNodeFlagsData.Shape, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	return &sdk.Node{
		Cloud:        sdk.CloudType(addNodeFlagsData.Cloud),
		Role:         sdk.NodeType(addNodeFlagsData.Role),
		Shape:        sdk.NodeShape(addNodeFlagsData.Shape),
		InstanceType: addNodeFlagsData.InstanceType,
	}, nil
}
