// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/jemurai/s3s2/encrypt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var keydir string
var keyprefix string

// genkeyCmd represents the genkey command
var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate new gpg keys.",
	Long:  `Generate new gpg keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Generating new keys with name %s in: %s", keyprefix, keydir)
		encrypt.GenerateKeys(keydir, keyprefix, 4096)
	},
}

func init() {
	rootCmd.AddCommand(genkeyCmd)

	genkeyCmd.PersistentFlags().StringVar(&keydir, "keydir", "", "The directory to write the key files to.")
	genkeyCmd.PersistentFlags().StringVar(&keyprefix, "keyprefix", "", "The directory to write the key files to.")

	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}
