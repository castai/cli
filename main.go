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
package main

import (
	"github.com/sirupsen/logrus"

	"github.com/castai/cast-cli/cmd"
	"github.com/castai/cast-cli/pkg/client"
	"github.com/castai/cast-cli/pkg/config"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp:       false,
		DisableLevelTruncation: true,
	})

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Debug {
		log.SetLevel(logrus.DebugLevel)
	}

	api, err := client.New(cfg, log)
	if err != nil {
		log.Fatal(err)
	}

	root := cmd.NewRootCmd(log, api)
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
