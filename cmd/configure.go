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

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/config"
)

func newConfigureCmd(log logrus.FieldLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Setup initial configuration",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleConfigure(log, cmd); err != nil {
				log.Fatalf("Configuration failed: %v", err)
				return
			}
		},
	}
	return cmd
}

func handleConfigure(log logrus.FieldLogger, cmd *cobra.Command) error {
	qs := []*survey.Question{
		{
			Name: "hostname",
			Prompt: &survey.Input{
				Message: "API hostname:",
				Default: config.DefaultHostname,
				Help:    "API hostname points to CAST AI rest api which is used by CAST AI CLI.",
			},
		},
		{
			Name: "access_token",
			Prompt: &survey.Password{
				Message: "API access token:",
				Help:    "API access token could be created via console UI. See https://docs.cast.ai/api/authentication/ for more details.",
			},
			Validate: survey.Required,
		},
	}
	var answers struct {
		Hostname    string `survey:"hostname"`
		AccessToken string `survey:"access_token"`
	}
	err := survey.Ask(qs, &answers)
	if err != nil {
		return err
	}
	cfg := &config.Config{
		Hostname:    answers.Hostname,
		AccessToken: answers.AccessToken,
		Debug:       false,
	}

	// Check if passed access token and hostname are valid.
	api, err := client.New(cfg, log)
	if err != nil {
		return err
	}
	_, err = api.ListAuthTokens(cmd.Context())
	if err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	if err := configureDefaultRegion(cmd.Context(), cfg, api); err != nil {
		return err
	}

	configPath, err := config.GetPath()
	if err != nil {
		return err
	}
	log.Infof("Configuration saved to %s", configPath)

	return nil
}

func configureDefaultRegion(ctx context.Context, cfg *config.Config, api client.Interface) error {
	regions, err := api.ListRegions(ctx)
	if err != nil {
		return err
	}
	regionsSelectList := make([]string, len(regions))
	for i, region := range regions {
		regionsSelectList[i] = region.DisplayName
	}

	var selectedRegion string
	prompt := &survey.Select{
		Message: "Default region:",
		Options: regionsSelectList,
		Default: regionsSelectList[0],
	}

	if err := survey.AskOne(prompt, &selectedRegion, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	for _, region := range regions {
		if region.DisplayName == selectedRegion {
			cfg.DefaultRegion = region.Name
			if err := config.Save(cfg); err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}
