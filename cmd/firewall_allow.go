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
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/sdk"
)

var (
	flagCidr string
)

var firewallAddCmd = &cobra.Command{
	Use:   "allow <cluster_id>",
	Short: "Allow firewall rule to access API Server",
	Long:  `Example: cast firewall allow <cluster_id> --cidr 0.0.0.0/32`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterID := requireClusterID(cmd, args)
		if err := handleFirewallAllow(clusterID.String(), flagCidr); err != nil {
			log.Fatal(err)
		}
	},
}

func handleFirewallAllow(clusterID, cidr string) error {
	apiClient, err := client.New()
	if err != nil {
		return err
	}

	if cidr == "" {
		pubIP, err := getPublicIP()
		if err != nil {
			return err
		}
		cidr = fmt.Sprintf("%s/32", pubIP)
	}

	ctx, cancel := client.DefaultContext()
	defer cancel()
	body := sdk.CreateOrUpdateFirewallJSONRequestBody{
		Cidr:      cidr,
		ClusterId: clusterID,
	}
	resp, err := apiClient.CreateOrUpdateFirewallWithResponse(ctx, body)
	if err := client.CheckResponse(resp, err, http.StatusOK); err != nil {
		return err
	}
	return nil
}

func init() {
	firewallCmd.AddCommand(firewallAddCmd)
	firewallAddCmd.PersistentFlags().StringVar(&flagCidr, "cidr", "", "--cidr 0.0.0.0/32")
}
