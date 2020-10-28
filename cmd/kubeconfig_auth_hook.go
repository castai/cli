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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/castai/cast-cli/internal/client"
	"github.com/castai/cast-cli/pkg/sdk"
)

var kubeconfigAuthHookCmd = &cobra.Command{
	Use:   "kubeconfig-auth-hook <cluster_id>",
	Short: "Hook to perform transparent auth and firewall",
	Run: func(cmd *cobra.Command, args []string) {
		clusterID := requireClusterID(cmd, args)
		if err := handleKubeconfigAuthHook(clusterID.String()); err != nil {
			log.Fatal(err)
		}
	},
}

type execCredentialsStatus struct {
	ClientCertificateData string `json:"clientCertificateData"`
	ClientKeyData         string `json:"clientKeyData"`
}

type execCredentials struct {
	ApiVersion string                `json:"apiVersion"`
	King       string                `json:"king"`
	Status     execCredentialsStatus `json:"status"`
}

// TODO: This logic is not for production. It fetches kubeconfig and adds firewall on each call.
func handleKubeconfigAuthHook(clusterID string) error {
	apiClient, err := client.New()
	if err != nil {
		return err
	}

	ctx, cancel := client.DefaultContext()
	defer cancel()
	resp, err := apiClient.GetClusterKubeconfigWithResponse(ctx, sdk.ClusterId(clusterID))
	if err := client.CheckResponse(resp, err, http.StatusOK); err != nil {
		return err
	}

	rawConfig, err := getRawConfig(resp.Body)
	if err != nil {
		return err
	}

	if len(rawConfig.AuthInfos) == 0 {
		return errors.New("no AuthInfos in kubeconfig")
	}
	var userAuth *api.AuthInfo
	for _, v := range rawConfig.AuthInfos {
		userAuth = v
	}

	creds := execCredentials{
		ApiVersion: "client.authentication.k8s.io/v1beta1",
		King:       "ExecCredential",
		Status: execCredentialsStatus{
			ClientCertificateData: string(userAuth.ClientCertificateData),
			ClientKeyData:         string(userAuth.ClientKeyData),
		},
	}

	if err := handleFirewallAllow(clusterID, ""); err != nil {
		return err
	}

	bytes, err := json.Marshal(creds)
	fmt.Fprint(os.Stdout, string(bytes))
	return nil
}

func init() {
	rootCmd.AddCommand(kubeconfigAuthHookCmd)
}
