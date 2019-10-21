
package cmd

import (
	"os"
	"strings"
	"sync"
	"time"

	archive "github.com/tempuslabs/s3s2/archive"
	encrypt_ssm "github.com/tempuslabs/s3s2/encrypt_ssm"
	manifest "github.com/tempuslabs/s3s2/manifest"
	options "github.com/tempuslabs/s3s2/options"
	aws_helpers "github.com/tempuslabs/s3s2/aws_helpers"
	utils "github.com/tempuslabs/s3s2/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ops options.Options

// decryptCmd represents the decrypt command
var decryptSSMCmd = &cobra.Command{
	Use:   "decrypt-ssm",
	Short: "Retrieve files that are stored securely in S3 and decrypt them",
	Long:  `Retrieve files that are stored securely in S3 and decrypt them`,

	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		opts := buildDecryptSSMOptions()
		log.Debug(opts)
		checkDecryptSSMOptions(opts)

		if strings.HasSuffix(opts.File, "manifest.json") {
			log.Debugf("manifest file: %s, %s", opts.Destination, opts.File)
			fn, err := aws_helpers.DownloadFile(opts.Destination, opts.File, opts)
			if err != nil {
				log.Error(err)
			}

			m := manifest.ReadManifest(fn)
			org := m.Organization
			folder := m.Folder

			var wg sync.WaitGroup
			for i := 0; i < len(m.Files); i++ {
				if !strings.HasSuffix(m.Files[i].Name, "manifest.json") {
					wg.Add(1)

					f := utils.OsAgnostic_HandleAwsKey(org, folder, m.Files[i].Name + ".zip.gpg")

					go func(f string, opts options.Options) {
						defer wg.Done()
						decryptFileSSM(f, opts)
					}(f, opts)
				}
			}
			wg.Wait()
			utils.CleanupDirectory(opts.Destination + m.Folder)

		} else {
			decryptFileSSM(opts.File, opts)
		}
		timing(start, "Elapsed time: %f")
	},
}

func decryptFileSSM(file string, options options.Options) {
	log.Debugf("Processing %s", file)
	start := time.Now()

	fn, err := aws_helpers.DownloadFile(options.Destination, file, options)

	if err != nil {
		log.Fatal(err)
	}
	stat, _ := os.Stat(fn)
	log.Debugf("Stat of file: %v", stat.Size())

	downloadTime := timing(start, "\tDownload time (sec): %f")

	encryptTime := time.Now()

    // if fetching keys from SSM store
	if options.SSMPubKey != "" && options.SSMPrivKey != "" {
        log.Debugf("Would be decrypting here... %s", fn)
		encrypt_ssm.DecryptSSM(fn, options)
		fn = strings.TrimSuffix(fn, ".gpg")
		encryptTime = timing(downloadTime, "\tDecrypt time (sec): %f")
	}

	log.Debugf("\tDecompressing file: %s", fn)
	fn = archive.UnZipFile(fn, options.Destination)
	// utils.CleanupFile(options.Directory)
	// utils.CleanupFile(fn + ".gpg")

	timing(encryptTime, "\tDecompress time (sec): %f")
	timing(start, "Total time: %f")
	log.Debugf("\tProcessed %s", fn)
}

func buildDecryptSSMOptions() options.Options {
	bucket := viper.GetString("bucket")
	file := viper.GetString("file")
	destination := viper.GetString("destination")
	if !strings.HasSuffix(destination, "/") {
		destination = destination + "/"
	}
	region := viper.GetString("region")
	SSMPrivKey := viper.GetString("ssm-private-key")
	SSMPubKey := viper.GetString("ssm-public-key")

	options := options.Options{
		Bucket:      bucket,
		File:        file,
		Destination: destination,
		Region:      region,
		SSMPrivKey:  SSMPrivKey,
		SSMPubKey:   SSMPubKey,
	}

	debug := viper.GetBool("debug")
	if debug != true {
		log.Debug("Setting Debug in Decrypt")
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("Captured options: ")
	log.Debug(options)
	return options
}

func checkDecryptSSMOptions(options options.Options) {
	if options.File == "" {
		log.Warn("Need to supply a file to decrypt.  Should be the file path within the dbucket but not including the dbucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Bucket == "" {
		log.Warn("Need to supply a bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Destination == "" {
		log.Warn("Need to supply a destination for the files to decrypt.  Should be a local path.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Region == "" {
		log.Warn("Need to supply a region for the S3 bucket.")
		log.Panic("Insufficient information to perform decryption.")
	}
}

func init() {
	rootCmd.AddCommand(decryptSSMCmd)

	decryptSSMCmd.PersistentFlags().String("file", "", "The path to the file to decrypt.  Can be manifest or single file.")
	decryptSSMCmd.MarkFlagRequired("file")
	decryptSSMCmd.PersistentFlags().String("destination", "", "The destination directory to decrypt and unzip.")
	decryptSSMCmd.MarkFlagRequired("destination")
	decryptSSMCmd.PersistentFlags().String("ssm-private-key", "", "The receiver's private key.  A parameter name in SSM.")
	decryptSSMCmd.PersistentFlags().String("ssm-public-key", "", "The receiver's public key.  A parameter name in SSM.")

	viper.BindPFlag("file", decryptSSMCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("destination", decryptSSMCmd.PersistentFlags().Lookup("destination"))
	viper.BindPFlag("ssm-private-key", decryptSSMCmd.PersistentFlags().Lookup("ssm-private-key"))
	viper.BindPFlag("ssm-public-key", decryptSSMCmd.PersistentFlags().Lookup("ssm-public-key"))

	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
}
