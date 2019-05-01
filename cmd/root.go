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
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var debug bool
var bucket string
var region string
var pubkey string
var privkey string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "s3s2",
	Short: "Easily use S3 in a safe and secure (s2) way.",
	Long: `Easily use S3 in a safe and secure (s2) way.
	
Amazon S3 is an incredibly useful way to share files.
s3s2 is a command line program that helps to ensure that when 
we use S3 we do so safely.  It does this by making the interface
easy and by being opinionated about encrypting data in specific 
ways in S3.

Use s3s2 help to get more information about using s3s2.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.s3s2.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode")
	rootCmd.PersistentFlags().StringVar(&bucket, "bucket", "", "The bucket to work with.")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "The region the bucket is in.")
	rootCmd.PersistentFlags().StringVar(&privkey, "privkey", "", "The receiver's private key.  A local file path.")
	rootCmd.PersistentFlags().StringVar(&pubkey, "pubkey", "", "The receiver's public key.  A local file path.")

	viper.BindPFlag("bucket", rootCmd.PersistentFlags().Lookup("bucket"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("privkey", rootCmd.PersistentFlags().Lookup("privkey"))
	viper.BindPFlag("pubkey", rootCmd.PersistentFlags().Lookup("pubkey"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".s3s2" (with extension!!!).
		viper.AddConfigPath(home)
		viper.SetConfigName(".s3s2")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	} else {
		//Uncomment if problems picking up config file.
		//fmt.Println(err)
	}
}
