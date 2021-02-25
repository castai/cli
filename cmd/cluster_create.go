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
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cheggaaa/pb/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
	"github.com/castai/cli/pkg/command"
	"github.com/castai/cli/pkg/config"
)

type clusterCreateFlags struct {
	Name          string   `survey:"name"`
	Region        string   `survey:"region"`
	Credentials   []string `survey:"credentials"`
	Configuration string   `survey:"configuration"`
	VPN           string   `survey:"vpn"`
	Progress      bool     `survey:"progress"`
	Nodes         []string
	Wait          bool
}

var clusterCreateFlagsData clusterCreateFlags

const (
	vpnTypeWireGuardCrossLocationMesh = "wireguard_cross_location_mesh"
	vpnTypeWireGuardFullMesh          = "wireguard_full_mesh"
	vpnTypeCloudProvider              = "cloud_provider"

	clusterConfigurationStarter = "starter"
	clusterConfigurationBasic   = "basic"
	clusterConfigurationHA      = "ha"
)

func newClusterCreateCmd(log logrus.FieldLogger, cfg *config.Config, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create kubernetes cluster",
		Long: `
You can create cluster in two ways:
1. Interactive - answer questions one by one with suggested values. 
   To create cluster in interactive mode just run:
   cast cluster create
2. Declarative - pass cluster creation flags manually.

Examples:
  # Create HA cluster with 3 clouds from quick configuration.
  cast cluster create \
    --name=my-demo-cluster \
    --region=eu-central \
    --credentials=aws --credentials=gcp --credentials=do \
    --configuration=ha \
    --vpn=wireguard_cross_location_mesh \
    --wait --progress

  # Create HA cluster with 3 clouds from custom nodes definitions.
  cast cluster create \
    --name=my-demo-cluster \
    --region=eu-central \
    --credentials=aws --credentials=gcp --credentials=do \
    --node=aws,master,medium \
    --node=gcp,master,medium \
    --node=do,master,medium \
    --node=aws,worker,smal \
    --node=gcp,worker,medium \
    --node=do,worker,large \
    --vpn=wireguard_full_mesh \
    --wait --progress
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleCreateCluster(cmd, log, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.Name, "name", "", "cluster name, eg. --name=my-demo-cluster")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.Region, "region", cfg.DefaultRegion, "region in which cluster will be created, eg. --region=eu-central")
	cmd.PersistentFlags().StringSliceVar(&clusterCreateFlagsData.Credentials, "credentials", []string{}, "cloud credentials names, eg. --credentials=aws, --credentials=gcp")
	cmd.PersistentFlags().StringSliceVar(&clusterCreateFlagsData.Nodes, "node", []string{}, "nodes configuration, eg. --node=aws,master,medium --node=gcp,worker,small")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.Configuration, "configuration", clusterConfigurationBasic, "quick cluster nodes configuration, eg. --configuration=ha")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.VPN, "vpn", vpnTypeWireGuardCrossLocationMesh, "virtual private network type between clouds, eg. --vpn=cloud_provider")
	cmd.PersistentFlags().BoolVar(&clusterCreateFlagsData.Wait, "wait", false, "wait until operation finishes, eg. --wait=true")
	cmd.PersistentFlags().BoolVar(&clusterCreateFlagsData.Progress, "progress", false, "show progress bar with estimated time for finish, eg. --progress=true")
	return cmd
}

func handleCreateCluster(cmd *cobra.Command, log logrus.FieldLogger, api client.Interface) error {
	req := &sdk.CreateNewClusterJSONRequestBody{}
	var err error
	if cmd.Flags().NFlag() == 0 {
		req, err = parseInteractiveClusterForm(cmd.Context(), api)
		if err != nil {
			return err
		}
	} else {
		req, err = parseDeclarativeClusterForm(cmd.Context(), api)
		if err != nil {
			return err
		}
	}

	cluster, err := api.CreateNewCluster(cmd.Context(), *req)
	if err != nil {
		return err
	}

	estimatedDuration := estimateClusterCreationDuration(req)

	if !clusterCreateFlagsData.Wait && !clusterCreateFlagsData.Progress {
		log.Infof("Great! Cluster creation is now in progress and should be ready after ~ %s.", estimatedDuration)
		return nil
	}

	if clusterCreateFlagsData.Wait && !clusterCreateFlagsData.Progress {
		err := waitClusterCreated(cmd.Context(), cluster.Id, api)
		if err != nil {
			return err
		}
		log.Info("Great! Cluster is ready.")
		return nil
	}

	if err := waitClusterCreatedWithProgress(cmd.Context(), estimatedDuration, cluster.Id, api); err != nil {
		return err
	}
	log.Info("Great! Cluster is ready.")
	return nil
}

func estimateClusterCreationDuration(req *sdk.CreateNewClusterJSONRequestBody) time.Duration {
	if req.Network.Vpn.WireGuard != nil {
		return 10 * time.Minute
	}

	clouds := map[string]struct{}{}
	for _, node := range req.Nodes {
		clouds[string(node.Cloud)] = struct{}{}
	}
	if len(clouds) == 1 {
		return 10 * time.Minute
	}
	for c := range clouds {
		if c == "azure" {
			return 30 * time.Minute
		}
	}
	return 12 * time.Minute
}

func waitClusterCreated(ctx context.Context, clusterID string, api client.Interface) error {
	for {
		cluster, err := api.GetCluster(ctx, sdk.ClusterId(clusterID))
		if err != nil {
			return err
		}
		if cluster.Status == "ready" {
			return nil
		}

		select {
		case <-time.After(20 * time.Second):
		case <-ctx.Done():
			return nil
		}
	}
}

func waitClusterCreatedWithProgress(ctx context.Context, duration time.Duration, clusterID string, api client.Interface) (err error) {
	waitErr := make(chan error)
	go func() {
		waitErr <- waitClusterCreated(ctx, clusterID, api)
	}()

	command.ShowProgress(ctx, command.ProgressConfig{
		Title:        fmt.Sprintf("Estimated time: %s", duration),
		TotalTimeETA: duration,
		TickInterval: 1 * time.Second,
		StopFunc: func(tick int, bar *pb.ProgressBar) bool {
			select {
			case err = <-waitErr:
				return true
			default:
				return false
			}
		},
	})
	return nil
}

func parseDeclarativeClusterForm(ctx context.Context, api client.Interface) (*sdk.CreateNewClusterJSONRequestBody, error) {
	lists := &clusterCreationSelectLists{}
	if err := lists.load(ctx, api); err != nil {
		return nil, err
	}

	req, err := toCreateClusterRequest(lists, clusterCreateFlagsData)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func parseInteractiveClusterForm(ctx context.Context, api client.Interface) (*sdk.CreateNewClusterJSONRequestBody, error) {
	lists := &clusterCreationSelectLists{}
	if err := lists.load(ctx, api); err != nil {
		return nil, err
	}

	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter cluster name:",
			},
			Validate: survey.Required,
		},
		{
			Name: "credentials",
			Prompt: &survey.MultiSelect{
				Message: "Select cloud credentials:",
				Options: lists.credentials.displayNames(),
			},
			Validate: survey.Required,
		},
		{
			Name: "region",
			Prompt: &survey.Select{
				Message: "Select region:",
				Options: lists.regions.displayNames(),
			},
			Validate: survey.Required,
		},
		{
			Name: "configuration",
			Prompt: &survey.Select{
				Message: "Select initial cluster configuration:",
				Options: lists.clusterConfigurations.displayNames(),
			},
			Validate: survey.Required,
		},
		{
			Name: "vpn",
			Prompt: &survey.Select{
				Message: "Select virtual private network:",
				Options: lists.vpns.displayNames(),
			},
			Validate: survey.Required,
		},
		{
			Name: "progress",
			Prompt: &survey.Confirm{
				Message: "Wait for creation and show progress:",
				Default: true,
			},
		},
	}

	err := survey.Ask(qs, &clusterCreateFlagsData)
	if err != nil {
		return nil, err
	}

	req, err := toCreateClusterRequest(lists, clusterCreateFlagsData)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func toCreateClusterRequest(lists *clusterCreationSelectLists, flags clusterCreateFlags) (*sdk.CreateNewClusterJSONRequestBody, error) {
	if len(flags.Credentials) == 0 {
		return nil, errors.New("cloud credentials are required, eg. --credentials=gcp-creds-1")
	}
	cloudCredentials := selectOptionList{}
	cloudCredentialIDs := make([]string, len(flags.Credentials))
	for i, credential := range flags.Credentials {
		v, ok := lists.credentials.find(credential)
		if !ok {
			return nil, fmt.Errorf("cloud credentials value %s is not valid, available values: %s", credential, strings.Join(lists.credentials.names(), ", "))
		}
		cloudCredentials = append(cloudCredentials, v)
		cloudCredentialIDs[i] = v.extra["id"]
	}

	region, ok := lists.regions.find(flags.Region)
	if !ok {
		return nil, fmt.Errorf("region value %s is not valid, available values: %s", flags.Region, strings.Join(lists.regions.names(), ", "))
	}

	vpn, ok := lists.vpns.find(flags.VPN)
	if !ok {
		return nil, fmt.Errorf("vpn value %s is not valid, available values: %s", flags.VPN, strings.Join(lists.vpns.names(), ", "))
	}

	if flags.Configuration == "" && len(flags.Nodes) == 0 {
		return nil, errors.New("configuration or nodes are required")
	}

	var nodes []sdk.Node
	if len(flags.Nodes) > 0 {
		var err error
		nodes, err = toAPINodesFromFlags(flags.Nodes)
		if err != nil {
			return nil, err
		}
	} else {
		configuration, ok := lists.clusterConfigurations.find(flags.Configuration)
		if !ok {
			return nil, fmt.Errorf("configuration value %s is not valid", flags.VPN)
		}
		nodes = toAPINodesFromConfiguration(cloudCredentials, configuration.name)
	}

	return &sdk.CreateNewClusterJSONRequestBody{
		Name:                flags.Name,
		CloudCredentialsIDs: cloudCredentialIDs,
		Region:              region.name,
		Network:             toNetwork(vpn.name),
		Addons:              defaultAddons(),
		Nodes:               nodes,
	}, nil
}

func toNetwork(name string) *sdk.Network {
	switch name {
	case vpnTypeWireGuardFullMesh:
		return &sdk.Network{Vpn: sdk.VpnConfig{
			WireGuard: &sdk.WireGuardConfig{Topology: "fullMesh"},
		}}
	case vpnTypeWireGuardCrossLocationMesh:
		return &sdk.Network{Vpn: sdk.VpnConfig{
			WireGuard: &sdk.WireGuardConfig{Topology: "crossLocationMesh"},
		}}
	case vpnTypeCloudProvider:
		return &sdk.Network{Vpn: sdk.VpnConfig{
			IpSec: &sdk.IpSecConfig{},
		}}
	}
	return nil
}

func toAPINodesFromFlags(nodes []string) ([]sdk.Node, error) {
	res := make([]sdk.Node, len(nodes))
	for i, s := range nodes {
		node, err := parseNodeFromString(s)
		if err != nil {
			return nil, err
		}
		res[i] = node
	}
	return res, nil
}

func parseNodeFromString(n string) (sdk.Node, error) {
	p := strings.Split(n, ",")
	if len(p) < 2 {
		return sdk.Node{}, fmt.Errorf("unknown node format %q, it should contain cloud, type and shape, eg. --node=aws,master,medium or --node=aws,worker,small", n)
	}

	cloud := strings.TrimSpace(p[0])
	switch cloud {
	case "aws":
	case "gcp":
	case "azure":
	case "do":
	default:
		return sdk.Node{}, fmt.Errorf("unknown node cloud %q, allowed values: aws, gcp, azure, do", cloud)
	}

	role := strings.TrimSpace(p[1])
	switch role {
	case "master":
	case "worker":
	default:
		return sdk.Node{}, fmt.Errorf("unknown node role %q, allowed values: master, worker", role)
	}

	shape := strings.TrimSpace(p[2])
	switch shape {
	case "small":
	case "medium":
	case "large":
	default:
		return sdk.Node{}, fmt.Errorf("unknown node shape %q, allowed values: small, medium, large", shape)
	}

	return sdk.Node{
		Cloud: sdk.CloudType(cloud),
		Role:  sdk.NodeType(role),
		Shape: sdk.NodeShape(shape),
	}, nil
}

func toAPINodesFromConfiguration(cloudCredentials selectOptionList, clusterConfigurationName string) []sdk.Node {
	var nodes []sdk.Node
	switch clusterConfigurationName {
	case clusterConfigurationStarter:
		// Add master node on first cloud.
		firstCloudCredential := cloudCredentials[0]
		nodes = append(nodes, sdk.Node{
			Cloud: sdk.CloudType(firstCloudCredential.extra["cloud"]),
			Role:  "master",
			Shape: "medium",
		})
		// Add worker node on first cloud.
		nodes = append(nodes, sdk.Node{
			Cloud: sdk.CloudType(firstCloudCredential.extra["cloud"]),
			Role:  "worker",
			Shape: "medium",
		})
	case clusterConfigurationBasic:
		// Add master node on first cloud.
		firstCloudCredential := cloudCredentials[0]
		nodes = append(nodes, sdk.Node{
			Cloud: sdk.CloudType(firstCloudCredential.extra["cloud"]),
			Role:  "master",
			Shape: "medium",
		})
		// Add worker nodes on each cloud.
		for _, credential := range cloudCredentials {
			nodes = append(nodes, sdk.Node{
				Cloud: sdk.CloudType(credential.extra["cloud"]),
				Role:  "worker",
				Shape: "small",
			})
		}
	case clusterConfigurationHA:
		// Add master node on each cloud.
		for i := 0; i < len(cloudCredentials); i++ {
			nodes = append(nodes, sdk.Node{
				Cloud: sdk.CloudType(cloudCredentials[i].extra["cloud"]),
				Role:  "master",
				Shape: "medium",
			})
		}
		// In case user selected 1 or 2 cloud credentials add additional master on first cloud.
		if len(nodes) < 3 {
			firstCloudCredential := cloudCredentials[0]
			for i := len(nodes); i < 3; i++ {
				nodes = append(nodes, sdk.Node{
					Cloud: sdk.CloudType(firstCloudCredential.extra["cloud"]),
					Role:  "master",
					Shape: "medium",
				})
			}
		}
		// Add worker nodes on each cloud.
		for _, credential := range cloudCredentials {
			nodes = append(nodes, sdk.Node{
				Cloud: sdk.CloudType(credential.extra["cloud"]),
				Role:  "worker",
				Shape: "small",
			})
		}
	}
	return nodes
}

func defaultAddons() *sdk.AddonsConfig {
	return &sdk.AddonsConfig{
		CertManager: &sdk.CertManagerConfig{Enabled: true},
		Dashboard:   &sdk.DashboardConfig{Enabled: true},
		ElasticLogging: &sdk.ElasticLoggingConfig{
			Config:  nil,
			Enabled: false,
		},
		Grafana:      &sdk.GrafanaConfig{Enabled: true},
		Keda:         &sdk.KedaConfig{Enabled: true},
		NginxIngress: &sdk.NginxIngressConfig{Enabled: true},
	}
}

type clusterCreationSelectLists struct {
	regions               selectOptionList
	credentials           selectOptionList
	clusterConfigurations selectOptionList
	vpns                  selectOptionList
}

func (d *clusterCreationSelectLists) load(ctx context.Context, api client.Interface) error {
	// Setup regions from API.
	regions, err := api.ListRegions(ctx)
	if err != nil {
		return err
	}

	d.regions = make([]selectOption, len(regions))
	for i, item := range regions {
		d.regions[i] = selectOption{
			name:        item.Name,
			displayName: item.DisplayName,
		}
	}

	// Setup credentials from API.
	credentials, err := api.ListCloudCredentials(ctx)
	if err != nil {
		return err
	}
	d.credentials = make([]selectOption, len(credentials))
	for i, item := range credentials {
		d.credentials[i] = selectOption{
			name:        item.Name,
			displayName: fmt.Sprintf("(%s) %s", item.Cloud, item.Name),
			extra:       map[string]string{"cloud": item.Cloud, "id": item.Id},
		}
	}

	// Setup cluster configurations.
	d.clusterConfigurations = selectOptionList{
		{
			name:        "starter",
			displayName: "Starter (1 Control Plane Node, 1 Node)",
		},
		{
			name:        "basic",
			displayName: "Basic (1 Control Plane Node, 1 Node on each cloud)",
		},
		{
			name:        "ha",
			displayName: "Highly Available (3 Control Plane Nodes, 1 Node on each cloud)",
		},
	}

	// Setup vpn options.
	d.vpns = selectOptionList{
		{
			name:        "wireguard_cross_location_mesh",
			displayName: "WireGuard VPN (Cross location mesh)",
		},
		{
			name:        "wireguard_full_mesh",
			displayName: "WireGuard VPN (Full mesh)",
		},
		{
			name:        "cloud_provider",
			displayName: "Cloud provider VPN",
		},
	}

	return nil
}

type selectOptionList []selectOption

func (s selectOptionList) names() []string {
	list := make([]string, len(s))
	for i, option := range s {
		list[i] = option.name
	}
	return list
}

func (s selectOptionList) displayNames() []string {
	list := make([]string, len(s))
	for i, option := range s {
		list[i] = option.displayName
	}
	return list
}

func (s selectOptionList) find(name string) (selectOption, bool) {
	for _, option := range s {
		if option.name == name || option.displayName == name {
			return option, true
		}
	}
	return selectOption{}, false
}

type selectOption struct {
	name        string
	displayName string
	extra       map[string]string
}
