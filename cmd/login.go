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
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/config"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to CAST AI",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		token := cmd.Flag("token").Value.String()
		apiUrl := cmd.Flag("api-url").Value.String()
		if err := handleLogin(token, apiUrl); err != nil {
			log.Fatalf("🤭 login failed: %v\n", err)
			return
		}
	},
}

func handleLogin(token string, apiUrl string) error {
	// Store valid access token to file.
	if err := config.StoreCredentials(token, apiUrl); err != nil {
		return err
	}

	// Check that access token is valid.
	apiClient, err := client.New()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := apiClient.ListAuthTokensWithResponse(ctx)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("could not verify access token, got status code %d", resp.StatusCode())
	}

	return nil
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().String("token", "", "CAST AI API access token")
	loginCmd.MarkPersistentFlagRequired("token")
}
