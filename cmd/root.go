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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/castai/cast-cli/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "cast",
	Short: "CAST AI Command Line Interface",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
		if config.GlobalFlags.Debug {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func Execute() {
	rootCmd.PersistentFlags().BoolVar(&config.GlobalFlags.Debug, "debug", false, "Enable debug mode to log api calls")
	rootCmd.PersistentFlags().StringVar(&config.GlobalFlags.ApiURL, "api-url", "https://api.cast.ai/v1", "CAST AI Api URL")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
