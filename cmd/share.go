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
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	archive "github.com/jemurai/s3s2/archive"
	encrypt "github.com/jemurai/s3s2/encrypt"
	manifest "github.com/jemurai/s3s2/manifest"
	options "github.com/jemurai/s3s2/options"
	s3helper "github.com/jemurai/s3s2/s3"
	utils "github.com/jemurai/s3s2/utils"
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
		start := time.Now()
		opts := buildShareOptions(cmd)
		checkShareOptions(opts)
		fnuuid, _ := uuid.NewV4()
		folder := opts.Prefix + "_s3s2_" + fnuuid.String()

		m := manifest.BuildManifest(folder, opts)

		if err := s3helper.UploadFile(folder, opts.Directory+m.Name, opts); err != nil {
			log.Error(err)
		}
		utils.CleanupFile(opts.Directory + m.Name)

		var wg sync.WaitGroup
		for i := 0; i < len(m.Files); i++ {
			wg.Add(1)
			fn := m.Files[i].Name
			go func(folder string, fn string, opts options.Options) {
				defer wg.Done()
				processFile(folder, fn, opts)
			}(folder, fn, opts)
		}
		wg.Wait()
		timing(start, "Elasped time: %f")
	},
}

func processFile(folder string, fn string, options options.Options) {
	log.Debugf("Processing %s", fn)
	start := time.Now()
	fn = archive.ZipFile(options.Directory+fn, options)
	//fn = archive.ZipFile(fn)
	archiveTime := timing(start, "\tArchive time (sec): %f")
	log.Debugf("\tCompressing file: %s", fn)
	if options.PubKey != "" {
		encrypt.Encrypt(fn, options.PubKey)
		fn = fn + ".gpg"
	}
	encryptTime := timing(archiveTime, "\tEncrypt time (sec): %f")
	err := s3helper.UploadFile(folder, fn, options)
	if err != nil {
		log.Fatal(err)
	}

	utils.CleanupFile(fn)
	if strings.HasSuffix(fn, ".gpg") {
		zipName := strings.TrimSuffix(fn, ".gpg")
		utils.CleanupFile(zipName)
	}

	timing(encryptTime, "\tUpload time (sec): %f")
	log.Debugf("\tProcessed %s", fn)
}

func timing(start time.Time, message string) time.Time {
	current := time.Now()
	elapsed := current.Sub(start)
	log.Debugf(message, elapsed.Seconds())
	return current
}

// buildContext sets up the ShareContext we're going to use
// to keep track of our state while we go.
func buildShareOptions(cmd *cobra.Command) options.Options {
	directory := viper.GetString("directory")
	bucket := viper.GetString("bucket")
	region := viper.GetString("region")
	pubKey := viper.GetString("receiver-public-key")
	awsKey := viper.GetString("awskey")
	org := viper.GetString("org")
	prefix := viper.GetString("prefix")
	hash := viper.GetBool("hash")

	options := options.Options{
		Directory: directory,
		Bucket:    bucket,
		Region:    region,
		PubKey:    pubKey,
		AwsKey:    awsKey,
		Org:       org,
		Prefix:    prefix,
		Hash:      hash,
	}

	debug := viper.GetBool("debug")
	if debug != true {
		log.SetLevel(log.InfoLevel)
	}
	log.Debug("Captured options: ")
	log.Debug(options)

	return options
}

func checkShareOptions(options options.Options) {
	if options.AwsKey != "" || options.PubKey != "" {
		// OK, that's good.  Looks like we have a key.
	} else {
		log.Warn("Need to supply either AWS Key for S3 level encryption or a public key for GPG encryption or both!")
		log.Panic("Insufficient key material to perform safe encryption.")
	}
}

func init() {
	rootCmd.AddCommand(shareCmd)

	shareCmd.PersistentFlags().String("directory", "", "The directory to zip, encrypt and share.")
	shareCmd.MarkFlagRequired("directory")
	shareCmd.PersistentFlags().String("org", "", "The organization that owns the files.")
	shareCmd.MarkFlagRequired("org")
	shareCmd.PersistentFlags().String("prefix", "", "A prefix for the S3 path.")
	shareCmd.PersistentFlags().String("awskey", "", "The agreed upon S3 key to encrypt data with at the bucket.")
	shareCmd.PersistentFlags().String("receiver-public-key", "", "The receiver's public key.  A local file path.")
	shareCmd.PersistentFlags().Bool("hash", false, "Should the tool calculate hashes (slow)?")

	viper.BindPFlag("directory", shareCmd.PersistentFlags().Lookup("directory"))
	viper.BindPFlag("org", shareCmd.PersistentFlags().Lookup("org"))
	viper.BindPFlag("prefix", shareCmd.PersistentFlags().Lookup("prefix"))
	viper.BindPFlag("awskey", shareCmd.PersistentFlags().Lookup("awskey"))
	viper.BindPFlag("receiver-public-key", shareCmd.PersistentFlags().Lookup("receiver-public-key"))
	viper.BindPFlag("hash", shareCmd.PersistentFlags().Lookup("hash"))

	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

}
