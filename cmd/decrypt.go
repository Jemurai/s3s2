
package cmd

import (
	"os"
	"strings"
	"sync"
	"time"

	archive "github.com/jemurai/s3s2/archive"
	"github.com/jemurai/s3s2/encrypt"
	manifest "github.com/jemurai/s3s2/manifest"
	options "github.com/jemurai/s3s2/options"
	s3helper "github.com/jemurai/s3s2/aws_helpers"
	utils "github.com/jemurai/s3s2/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ops options.Options

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Retrieve files that are stored securely in S3 and decrypt them",
	Long:  `Retrieve files that are stored securely in S3 and decrypt them`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		opts := buildDecryptOptions()
		checkDecryptOptions(opts)

		if strings.HasSuffix(opts.File, "manifest.json") {
			log.Debugf("manifest file: %s, %s", opts.Destination, opts.File)
			fn, err := s3helper.DownloadFile(opts.Destination, opts.File, opts)
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
						decryptFile(f, opts)
					}(f, opts)
				}
			}
			wg.Wait()
			utils.CleanupDirectory(opts.Destination + m.Folder)

		} else {
			decryptFile(opts.File, opts)
		}
		timing(start, "Elapsed time: %f")
	},
}

func decryptFile(file string, options options.Options) {
	log.Debugf("Processing %s", file)
	start := time.Now()

	fn, err := s3helper.DownloadFile(options.Destination, file, options)
	if err != nil {
		log.Fatal(err)
	}
	stat, _ := os.Stat(fn)
	log.Debugf("Stat of file: %v", stat.Size())

	downloadTime := timing(start, "\tDownload time (sec): %f")

	encryptTime := time.Now()
	if options.PrivKey != "" && strings.HasSuffix(file, ".gpg") {
		log.Debugf("Would be decrypting here... %s", fn)
		encrypt.Decrypt(fn, options.PubKey, options.PrivKey)
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

func buildDecryptOptions() options.Options {
	bucket := viper.GetString("bucket")
	file := viper.GetString("file")
	destination := viper.GetString("destination")
	if !strings.HasSuffix(destination, "/") {
		destination = destination + "/"
	}
	region := viper.GetString("region")
	privKey := viper.GetString("my-private-key")
	pubKey := viper.GetString("my-public-key")

	options := options.Options{
		Bucket:      bucket,
		File:        file,
		Destination: destination,
		Region:      region,
		PrivKey:     privKey,
		PubKey:      pubKey,
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

func checkDecryptOptions(options options.Options) {
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
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.PersistentFlags().String("file", "", "The path to the file to decrypt.  Can be manifest or single file.")
	decryptCmd.MarkFlagRequired("file")
	decryptCmd.PersistentFlags().String("destination", "", "The destination directory to decrypt and unzip.")
	decryptCmd.MarkFlagRequired("destination")
	decryptCmd.PersistentFlags().String("my-private-key", "", "The receiver's private key.  A local file path.")
	decryptCmd.PersistentFlags().String("my-public-key", "", "The receiver's public key.  A local file path.")

	viper.BindPFlag("file", decryptCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("destination", decryptCmd.PersistentFlags().Lookup("destination"))
	viper.BindPFlag("my-private-key", decryptCmd.PersistentFlags().Lookup("my-private-key"))
	viper.BindPFlag("my-public-key", decryptCmd.PersistentFlags().Lookup("my-public-key"))

	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
}
