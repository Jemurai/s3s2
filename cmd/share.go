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
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"

	archive "github.com/jemurai/s3s2/archive"
	encrypt "github.com/jemurai/s3s2/encrypt"
	manifest "github.com/jemurai/s3s2/manifest"
	options "github.com/jemurai/s3s2/options"
	s3helper "github.com/jemurai/s3s2/s3"
)

// shareCmd represents the share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share a file",
	Long: `Share a file to S3.
	
Behind the scenes, s3s2 checks to ensure the file is 
either GPG encrypted or passes S3 headers indicating
that it will be encrypted.`,

	Run: func(cmd *cobra.Command, args []string) {
		options := buildOptions(cmd)
		checkOptions(options)
		m := manifest.BuildManifest(options)

		var filez []string
		for i := 0; i < len(m.Files); i++ {
			filez = append(filez, m.Files[i].Name)
			fmt.Println(m.Files[i].Name, m.Files[i].Hash) // This is just debug.
		}
		fnuuid, _ := uuid.NewV4()
		fn := "s3s2_" + fnuuid.String() + ".zip"
		archive.ZipFiles(fn, filez)

		if options.PubKey != "" {
			encrypt.Encrypt(fn, options.PubKey)
			fn = fn + ".gpg"
		}
		err := s3helper.UploadFile(fn, options)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// buildContext sets up the ShareContext we're going to use
// to keep track of our state while we go.
func buildOptions(cmd *cobra.Command) options.Options {
	directory, _ := cmd.PersistentFlags().GetString("directory")
	bucket, _ := cmd.PersistentFlags().GetString("bucket")
	region, _ := cmd.PersistentFlags().GetString("region")
	pubKey, _ := cmd.PersistentFlags().GetString("pubkey")
	awsKey, _ := cmd.PersistentFlags().GetString("awskey")
	org, _ := cmd.PersistentFlags().GetString("org")
	prefix, _ := cmd.PersistentFlags().GetString("prefix")
	options := options.Options{Directory: directory,
		Bucket: bucket, Region: region,
		PubKey: pubKey, AwsKey: awsKey,
		Org: org, Prefix: prefix}

	return options
}

func checkOptions(options options.Options) {
	if options.AwsKey != "" || options.PubKey != "" {
		// OK, that's good.  Looks like we have a key.
	} else {
		fmt.Println("Need to supply either AWS Key for S3 level encryption or a public key for GPG encryption or both!")
		panic("Insufficient key material to perform safe encryption.")
	}
}

func init() {
	rootCmd.AddCommand(shareCmd)
	shareCmd.PersistentFlags().String("bucket", "", "The bucket to share the file to.")
	shareCmd.MarkFlagRequired("bucket")
	shareCmd.PersistentFlags().String("region", "", "The region the S3 bucket is in. Ex: us-east-1")
	shareCmd.MarkFlagRequired("region")
	shareCmd.PersistentFlags().String("directory", "", "The directory to zip, encrypt and share.")
	shareCmd.MarkFlagRequired("directory")
	shareCmd.PersistentFlags().String("org", "", "The organization that owns the files.")
	shareCmd.MarkFlagRequired("org")
	shareCmd.PersistentFlags().String("prefix", "", "A prefix for the S3 path.")
	shareCmd.PersistentFlags().String("pubkey", "", "The receiver's public key.  A link or a local file path.")
	shareCmd.PersistentFlags().String("awskey", "", "The agreed upon S3 key to encrypt data with at the bucket.")
}
