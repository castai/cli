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
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/client/sdk"
	"github.com/castai/cast-cli/pkg/command"
	"github.com/castai/cast-cli/pkg/ssh"
)

const (
	sshPublicKeyName  = "cast_ed25519.pub"
	sshPrivateKeyName = "cast_ed25519"
)

func newNodeSSHCmd(log logrus.FieldLogger, api client.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH into cluster node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleNodeSSH(cmd, log, api); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().StringP(flagCluster, "c", "", "cluster name or ID")
	cmd.MarkPersistentFlagRequired(flagCluster)
	command.AddJSONOutput(cmd)
	return cmd
}

func handleNodeSSH(cmd *cobra.Command, log logrus.FieldLogger, api client.Interface) error {
	clusterIDOrName, err := cmd.Flags().GetString(flagCluster)
	if err != nil {
		return err
	}

	clusterID, err := parseClusterIDFromValue(cmd.Context(), api, clusterIDOrName)
	if err != nil {
		return err
	}

	nodeID, err := parseNodeIDFromCMDArgs(cmd, api, clusterID)
	if err != nil {
		return err
	}

	node, err := api.GetClusterNode(cmd.Context(), sdk.ClusterId(clusterID), nodeID)
	if err != nil {
		return err
	}

	if node.Network == nil || node.Network.PrivateIp == "" || node.Network.PublicIp == "" {
		return errors.New("node is not ready yet")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keysPath := path.Join(home, ".ssh")
	privateKeyPath := filepath.Join(keysPath, sshPrivateKeyName)
	publicKeyPath := filepath.Join(keysPath, sshPublicKeyName)

	keys, err := generateKeys(privateKeyPath, publicKeyPath)
	if err != nil {
		return err
	}

	// Send public key to CAST AI.
	publicIP, err := getPublicIP()
	if err != nil {
		return err
	}
	log.Info("configuring firewall for SSH access")
	accessRuleID, err := api.SetupNodeSSH(cmd.Context(), sdk.ClusterId(clusterID), nodeID, sdk.SetupNodeSshJSONRequestBody{
		PublicKey: base64.StdEncoding.EncodeToString(keys.Public),
		SourceIp:  publicIP,
	})
	if err != nil {
		return err
	}

	// TODO: Fix different users and use 'ubuntu' user on all clouds.
	var user string
	switch node.Cloud {
	case sdk.CloudType_do:
		user = "root"
	case sdk.CloudType_aws:
	case sdk.CloudType_azure:
		user = "ubuntu"
	case sdk.CloudType_gcp:
		user = "TODO"
	}

	log.Info("establishing secure ssh session")
	if err := ssh.Terminal(cmd.Context(), ssh.TerminalConfig{
		PrivateKey: keys.Private,
		User:       user,
		Addr:       fmt.Sprintf("%s:22", node.Network.PublicIp),
	}); err != nil {
		return err
	}

	log.Info("closing firewall access")
	if err := api.CloseNodeSSH(cmd.Context(), sdk.ClusterId(clusterID), nodeID, accessRuleID); err != nil {
		log.Fatal(err)
	}

	return nil
}

func generateKeys(privateKeyPath, publicKeyPath string) (*ssh.Keys, error) {
	_, err := os.Stat(privateKeyPath)

	// If no keys yet generate and return.
	if os.IsNotExist(err) {
		keys, err := ssh.GenerateKeys("cast-cli")
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(privateKeyPath, keys.Private, 0400); err != nil {
			return nil, fmt.Errorf("writing private key: %w", err)
		}
		if err := ioutil.WriteFile(publicKeyPath, keys.Public, 0400); err != nil {
			return nil, fmt.Errorf("writing public key: %w", err)
		}
		return keys, nil
	}

	// Read already generated keys from files
	priv, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	pub, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	return &ssh.Keys{
		Public:  pub,
		Private: priv,
	}, nil
}

func getPublicIP() (string, error) {
	c := http.Client{Timeout: 30 * time.Second}
	resp, err := c.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
