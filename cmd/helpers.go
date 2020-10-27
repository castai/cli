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
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// usagef prints the Printf formatted message to stderr, prints usage help and
// exits the program
// Note: os.Exit(1) is not recoverable
func usagef(cmd *cobra.Command, msg string, args ...interface{}) {
	txt := fmt.Sprintf(msg, args...)
	fmt.Fprintf(os.Stderr, "Error: %s\n\n", txt)
	cmd.Help()
	os.Exit(1)
}

func requireClusterID(cmd *cobra.Command, args []string) uuid.UUID {
	if len(args) < 1 {
		usagef(cmd, "Missing cluster ID argument")
	}

	clusterID, err := uuid.Parse(args[0])
	if err != nil {
		usagef(cmd, "cluster ID is invalid: %v", err)
	}

	return clusterID
}
