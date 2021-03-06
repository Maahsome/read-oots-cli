/*
Copyright © 2020 Christopher Maahs <cmaahs@gmail.com>

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

	"github.com/spf13/cobra"
)

// These are in support of goreleaser, which auto-populates main. variables.
var (
	// Version - Receive from main
	Version = "dev"
	// Commit - Receive from main
	Commit = "none"
	// Date - Receive from main
	Date = "unknown"
	// BuiltBy - Receive from main
	BuiltBy = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the application version",
	Long:  `Display the application version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("{\"VersionInfo\": {")
		fmt.Println(fmt.Sprintf("\"SemVer\": \"%s\", \"GitCommit\": \"%s\", \"BuildDate\": \"%s\", \"BuiltBy\": \"%s\"}", Version, Commit, Date, BuiltBy))
		fmt.Println("}")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

}
