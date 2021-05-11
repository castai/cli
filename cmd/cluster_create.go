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
	"net"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/client/sdk"
	"github.com/castai/cli/pkg/config"
)

type clusterCreateFlags struct {
	Name          string   `survey:"name"`
	Region        string   `survey:"region"`
	Credentials   []string `survey:"credentials"`
	Configuration string   `survey:"configuration"`
	VPN           string   `survey:"vpn"`
	Nodes         []string
	Wait          bool
	AWSVPCCidr    string
	GCPVPCCidr    string
	AzureVPCCidr  string
	DOVPCCidr     string
}

var clusterCreateFlagsData clusterCreateFlags

const (
	vpnTypeWireGuardCrossLocationMesh = "wireguard_cross_location_mesh"
	vpnTypeWireGuardFullMesh          = "wireguard_full_mesh"
	vpnTypeCloudProvider              = "cloud_provider"

	clusterConfigurationBasic = "basic"
	clusterConfigurationHA    = "ha"
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
    --wait 

  # Create HA cluster with 3 clouds from custom nodes definitions.
  cast cluster create \
    --name=my-demo-cluster \
    --region=eu-central \
    --credentials=aws --credentials=gcp --credentials=do \
    --node=aws,master,medium \
    --node=gcp,master,medium \
    --node=do,master,medium \
    --node=aws,worker,small \
    --node=gcp,worker,medium \
    --node=do,worker,large \
    --vpn=wireguard_full_mesh \
    --wait
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
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.Configuration, "configuration", clusterConfigurationBasic, "quick cluster nodes configuration, available values: basic,ha")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.VPN, "vpn", "", "virtual private network type between clouds, available values: cloud_provider, wireguard_cross_location_mesh, wireguard_full_mesh")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.AWSVPCCidr, "aws-vpc-cidr", "", "optional custom AWS VPC IPv4 CIDR, eg. --aws-vpc-cidr=10.10.0.0/16")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.GCPVPCCidr, "gcp-vpc-cidr", "", "optional custom GCP VPC IPv4 CIDR, eg. --gcp-vpc-cidr=10.0.0.0/16")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.AzureVPCCidr, "azure-vpc-cidr", "", "optional custom AZURE VPC IPv4 CIDR, eg. --azure-vpc-cidr=10.20.0.0/16")
	cmd.PersistentFlags().StringVar(&clusterCreateFlagsData.DOVPCCidr, "do-vpc-cidr", "", "optional custom DO IPv4 CIDR, eg. --do-vpc-cidr=10.100.0.0/16")
	cmd.PersistentFlags().BoolVar(&clusterCreateFlagsData.Wait, "wait", false, "wait until operation finishes, eg. --wait=true")
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

	if !clusterCreateFlagsData.Wait {
		log.Infof("Cluster creation is now in progress. Check status by running 'cast cluster get %s'", cluster.Name)
		return nil
	}

	log.Info("Cluster creation is now in progress. It is safe to close this terminal.")
	if err := waitClusterCreatedWithProgress(cmd.Context(), log, api, cluster.Id); err != nil {
		return err
	}
	log.Infof("Cluster is ready. Check status by running 'cast cluster get %s'", cluster.Name)
	return nil
}

func waitClusterCreatedWithProgress(ctx context.Context, log logrus.FieldLogger, api client.Interface, clusterID string) (err error) {
	written := make(map[string]struct{})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			// Check if cluster is ready.
			cluster, err := api.GetCluster(ctx, sdk.ClusterId(clusterID))
			if err != nil {
				log.Warn(err)
				continue
			}
			if cluster.Status == "ready" {
				return nil
			}

			// Print feedback events.
			events, err := api.GetClusterFeedbackEvents(ctx, sdk.ClusterId(clusterID))
			if err != nil {
				log.Warn(err)
				continue
			}
			sort.Slice(events, func(i, j int) bool {
				return events[i].CreatedAt.Before(events[j].CreatedAt)
			})
			for _, e := range events {
				if _, ok := written[e.Id]; !ok {
					logFn := log.Infof
					if e.Severity == "error" {
						logFn = log.Errorf
					}
					logFn("%s %s", e.CreatedAt.Format(time.RFC3339), e.Message)
					written[e.Id] = struct{}{}
				}
			}
		}
	}
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
	}

	err := survey.Ask(qs, &clusterCreateFlagsData)
	if err != nil {
		return nil, err
	}

	if len(clusterCreateFlagsData.Credentials) > 1 {
		if err := survey.AskOne(&survey.Select{
			Message: "Select virtual private network:",
			Options: filterVPNOptions(lists, clusterCreateFlagsData.Credentials).displayNames(),
		}, &clusterCreateFlagsData.VPN, survey.WithValidator(survey.Required)); err != nil {
			return nil, err
		}
	}

	if err := survey.AskOne(&survey.Confirm{
		Message: "Wait for cluster creation?",
		Default: true,
	}, &clusterCreateFlagsData.Wait); err != nil {
		return nil, err
	}

	req, err := toCreateClusterRequest(lists, clusterCreateFlagsData)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func filterVPNOptions(lists *clusterCreationSelectLists, selectedClouds []string) selectOptionList {
	var containsDO bool
	for _, cloudDisplayName := range selectedClouds {
		if cloud, ok := lists.credentials.find(cloudDisplayName); ok && cloud.name == "do" {
			containsDO = true
			break
		}
	}

	if !containsDO {
		return lists.vpns
	}

	var res selectOptionList
	for _, option := range lists.vpns {
		if option.name != vpnTypeCloudProvider {
			res = append(res, option)
		}
	}
	return res
}

func toCreateClusterRequest(lists *clusterCreationSelectLists, flags clusterCreateFlags) (*sdk.CreateNewClusterJSONRequestBody, error) {
	if len(flags.Credentials) == 0 {
		return nil, errors.New("cloud credentials are required, eg. --credentials=gcp-creds-1")
	}
	selectedCloudCredentialIDs := make([]string, 0, len(flags.Credentials))
	selectedClouds := make([]string, 0, len(flags.Credentials))
	for _, credential := range flags.Credentials {
		v, ok := lists.credentials.find(credential)
		if !ok {
			return nil, fmt.Errorf("cloud credentials value '%s' is not valid, available values: %s", credential, strings.Join(lists.credentials.names(), ", "))
		}
		selectedClouds = append(selectedClouds, v.extra["cloud"])
		selectedCloudCredentialIDs = append(selectedCloudCredentialIDs, v.extra["id"])
	}

	region, ok := lists.regions.find(flags.Region)
	if !ok {
		return nil, fmt.Errorf("region value '%s' is not valid, available values: %s", flags.Region, strings.Join(lists.regions.names(), ", "))
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
			return nil, fmt.Errorf("configuration value '%s' is not valid, available values: %s", flags.Configuration, strings.Join(lists.clusterConfigurations.names(), ", "))
		}
		nodes = toAPINodesFromConfiguration(selectedClouds, configuration.name)
	}

	networkSpec, err := toNetwork(lists, flags, len(selectedCloudCredentialIDs))
	if err != nil {
		return nil, err
	}

	return &sdk.CreateNewClusterJSONRequestBody{
		Name:                flags.Name,
		CloudCredentialsIDs: selectedCloudCredentialIDs,
		Region:              region.name,
		Network:             networkSpec,
		Addons:              defaultAddons(),
		Nodes:               nodes,
	}, nil
}

func toNetwork(lists *clusterCreationSelectLists, flags clusterCreateFlags, clouds int) (*sdk.Network, error) {
	res := &sdk.Network{}

	// Set VPN type.
	vpn, ok := lists.vpns.find(flags.VPN)
	if !ok && clouds > 1 {
		return nil, fmt.Errorf("vpn value '%s' is not valid, available values: %s", flags.VPN, strings.Join(lists.vpns.names(), ", "))
	}
	switch vpn.name {
	case vpnTypeWireGuardFullMesh:
		res.Vpn = &sdk.VpnConfig{
			WireGuard: &sdk.WireGuardConfig{Topology: "fullMesh"},
		}
	case vpnTypeWireGuardCrossLocationMesh:
		res.Vpn = &sdk.VpnConfig{
			WireGuard: &sdk.WireGuardConfig{Topology: "crossLocationMesh"},
		}
	case vpnTypeCloudProvider:
		res.Vpn = &sdk.VpnConfig{
			IpSec: &sdk.IpSecConfig{},
		}
	}

	// Set custom VPC CIDR's.
	if cidr := flags.AWSVPCCidr; cidr != "" {
		if err := validateCIDR(cidr); err != nil {
			return nil, err
		}
		res.Aws = &sdk.CloudNetworkConfig{VpcCidr: cidr}
	}
	if cidr := flags.GCPVPCCidr; cidr != "" {
		if err := validateCIDR(cidr); err != nil {
			return nil, err
		}
		res.Gcp = &sdk.CloudNetworkConfig{VpcCidr: cidr}
	}
	if cidr := flags.AzureVPCCidr; cidr != "" {
		if err := validateCIDR(cidr); err != nil {
			return nil, err
		}
		res.Azure = &sdk.CloudNetworkConfig{VpcCidr: cidr}
	}
	if cidr := flags.DOVPCCidr; cidr != "" {
		if err := validateCIDR(cidr); err != nil {
			return nil, err
		}
		res.Do = &sdk.CloudNetworkConfig{VpcCidr: cidr}
	}

	return res, nil
}

func validateCIDR(s string) error {
	_, _, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("invalid cidr: %w", err)
	}
	return nil
}

func toAPINodesFromFlags(nodes []string) ([]sdk.Node, error) {
	res := make([]sdk.Node, len(nodes))
	for i, s := range nodes {
		node, err := parseNode(s)
		if err != nil {
			return nil, err
		}
		res[i] = node
	}
	return res, nil
}

func parseNode(n string) (sdk.Node, error) {
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

func toAPINodesFromConfiguration(selectedClouds []string, clusterConfigurationName string) []sdk.Node {
	var nodes []sdk.Node
	switch clusterConfigurationName {
	case clusterConfigurationBasic:
		// Add master node on first cloud.
		firstCloud := selectedClouds[0]
		nodes = append(nodes, sdk.Node{
			Cloud: sdk.CloudType(firstCloud),
			Role:  "master",
			Shape: "medium",
		})
		// Add worker nodes on each cloud.
		for _, cloud := range selectedClouds {
			nodes = append(nodes, sdk.Node{
				Cloud: sdk.CloudType(cloud),
				Role:  "worker",
				Shape: "small",
			})
		}
	case clusterConfigurationHA:
		// Add master node on each cloud.
		for _, cloud := range selectedClouds {
			nodes = append(nodes, sdk.Node{
				Cloud: sdk.CloudType(cloud),
				Role:  "master",
				Shape: "medium",
			})
		}
		// In case user selected 1 or 2 cloud credentials add additional master on first cloud.
		if len(nodes) < 3 {
			firstCloud := selectedClouds[0]
			for i := len(nodes); i < 3; i++ {
				nodes = append(nodes, sdk.Node{
					Cloud: sdk.CloudType(firstCloud),
					Role:  "master",
					Shape: "medium",
				})
			}
		}
		// Add worker nodes on each cloud.
		for _, cloud := range selectedClouds {
			nodes = append(nodes, sdk.Node{
				Cloud: sdk.CloudType(cloud),
				Role:  "worker",
				Shape: "small",
			})
		}
	}
	return nodes
}

func defaultAddons() *sdk.AddonsConfig {
	// TODO (anjmao): Allow to pass only selected addons during cluster creation.
	return &sdk.AddonsConfig{}
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
