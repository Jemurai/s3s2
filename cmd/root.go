
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
var ssmpubkey string
var ssmprivkey string
var awsprofile string

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
    log.Debug("Initializing root configurations...")
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.s3s2.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode")
	rootCmd.PersistentFlags().StringVar(&bucket, "bucket", "", "The bucket to work with.")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "The region the bucket is in.")

	viper.BindPFlag("bucket", rootCmd.PersistentFlags().Lookup("bucket"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

    log.Debug("Setting log level...")
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})

    log.Debug("Determining config source...")
	if cfgFile != "" {
		// Use config file from the flag.
		log.Debug("Setting config file...")
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		log.Debug("Locating .s3s2 config file...")
		home, err := homedir.Dir()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".s3s2" (with extension!!!).
		viper.AddConfigPath(home)
		viper.SetConfigName(".s3s2")
	}

    log.Debug("Reading environment variables...")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	} else {
		//Uncomment if problems picking up config file.
		//fmt.Println(err)
	}
}
