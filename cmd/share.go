// Copyright © 2019 Matt Konda <mkonda@jemurai.com>
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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	manifest "github.com/jemurai/s3s2/manifest"
)

// ShareContext is the context we extract from the command.
type ShareContext struct {
	Directory string
	Bucket    string
	Key       string
}

// shareCmd represents the share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share a file",
	Long: `Share a file to S3.
	
Behind the scenes, s3s2 checks to ensure the file is 
either GPG encrypted or passes S3 headers indicating
that it will be encrypted.`,
	Run: func(cmd *cobra.Command, args []string) {
		context := buildContext(cmd)
		m := manifest.BuildManifest(context.Directory)

		for i := 0; i < len(m.Files); i++ {
			//fmt.Println(m.Files[i].Name)
		}
		b, _ := json.Marshal(m)
		fmt.Println(b)

		// TODO: Archive
		// TODO: Encrypt with GPG
		// TODO: Encryption config for S3
		// TODO: Upload file to S3
	},
}

// buildContext sets up the ShareContext we're going to use
// to keep track of our state while we go.
func buildContext(cmd *cobra.Command) ShareContext {
	directory, _ := cmd.PersistentFlags().GetString("directory")
	bucket, _ := cmd.PersistentFlags().GetString("bucket")
	key, _ := cmd.PersistentFlags().GetString("key")
	context := ShareContext{directory, bucket, key}
	return context
}

func init() {
	rootCmd.AddCommand(shareCmd)
	shareCmd.PersistentFlags().String("bucket", "", "The bucket to share the file to.")
	shareCmd.MarkFlagRequired("bucket")
	shareCmd.PersistentFlags().String("directory", "", "The directory to zip, encrypt and share.")
	shareCmd.MarkFlagRequired("directory")
	shareCmd.PersistentFlags().String("key", "", "The receiver's public key.  A link or a local file path.")
}
