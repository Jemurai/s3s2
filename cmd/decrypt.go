
package cmd

import (
	"os"
	"strings"
	"sync"
	"time"
	"path/filepath"
	"golang.org/x/crypto/openpgp/packet"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	session "github.com/aws/aws-sdk-go/aws/session"

	log "github.com/sirupsen/logrus"

    // local
	archive "github.com/tempuslabs/s3s2_new/archive"
	encrypt "github.com/tempuslabs/s3s2_new/encrypt"
	manifest "github.com/tempuslabs/s3s2_new/manifest"
	options "github.com/tempuslabs/s3s2_new/options"
	aws_helpers "github.com/tempuslabs/s3s2_new/aws_helpers"
	utils "github.com/tempuslabs/s3s2_new/utils"
)

var opts options.Options

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Retrieve files that are stored securely in S3 and decrypt them",
	Long:  `Retrieve files that are stored securely in S3 and decrypt them`,

	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		opts := buildDecryptOptions()
		checkDecryptOptions(opts)

        // top level clients
        sess := utils.GetAwsSession(opts)
	    _pubKey := encrypt.GetPubKey(opts)
	    _privKey := encrypt.GetPrivKey(opts)

		if strings.HasSuffix(opts.File, "manifest.json") {

			log.Debugf("Manifest file: %s, %s", opts.Destination, opts.File)

			fn, err := aws_helpers.DownloadFile(sess, opts.Destination, opts.File, opts)
			if err != nil {
				log.Error(err)
			}

			m := manifest.ReadManifest(fn)
			org := m.Organization
			folder := m.Folder

            // Iterates over the files inside the manifest, infers their full path, downloads, then decrypts
            sem := make(chan struct{}, 12)
			var wg sync.WaitGroup
			for i := 0; i < len(m.Files); i++ {
				if !strings.HasSuffix(m.Files[i].Name, "manifest.json") {
					wg.Add(1)

                    // files uploaded from windows appear as '\\nested_dir\\filename.txt'
                    cleaned_fn := utils.ForceBackSlash(m.Files[i].Name)
                    log.Debugf("Cleaned filename: %s", cleaned_fn)

					f := utils.OsAgnostic_HandleAwsKey(org, folder, cleaned_fn + ".zip.gpg", opts)

					go func(f string, opts options.Options) {
					    sem <- struct{}{}
			            defer func() { <-sem }()
						defer wg.Done()
						decryptFile(sess, _pubKey, _privKey, f, opts)
					}(f, opts)
				}
			}
			wg.Wait()
			utils.CleanupDirectory(filepath.Join(opts.Destination, m.Folder))

		} else {
			decryptFile(sess, _pubKey, _privKey, utils.ForceBackSlash(opts.File), opts)
		}
		timing(start, "Elapsed time: %f")
	},
}

func decryptFile(sess *session.Session, _pubkey *packet.PublicKey, _privkey *packet.PrivateKey, file string, opts options.Options) {
	log.Debugf("Processing %s", file)
	start := time.Now()

	fn, err := aws_helpers.DownloadFile(sess, opts.Destination, file, opts)
	if err != nil {
		log.Fatal(err)
	}

	stat, _ := os.Stat(fn)

	log.Debugf("Stat of file: %v", stat.Size())
	downloadTime := timing(start, "\tDownload time (sec): %f")

	encryptTime := time.Now()

    if strings.HasSuffix(file, ".gpg") {
		log.Debugf("Would be decrypting here... %s", fn)
		encrypt.Decrypt(_pubkey, _privkey, fn, opts)
		fn = strings.TrimSuffix(fn, ".gpg")
		encryptTime = timing(downloadTime, "\tDecrypt time (sec): %f")
	}

	log.Debugf("\tDecompressing file: %s", fn)
	fn = archive.UnZipFile(fn, opts.Destination)

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
	awsProfile := viper.GetString("awsprofile")

	privKey := viper.GetString("my-private-key")
	pubKey := viper.GetString("my-public-key")
	ssmPrivKey := viper.GetString("ssm-private-key")
	ssmPubKey := viper.GetString("ssm-public-key")

	options := options.Options{
		Bucket:      bucket,
		File:        file,
		Destination: destination,
		Region:      region,
		PrivKey:     privKey,
		PubKey:      pubKey,
		SSMPrivKey:  ssmPrivKey,
		SSMPubKey:   ssmPubKey,
		AwsProfile:  awsProfile,
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

	decryptCmd.PersistentFlags().String("awsprofile", "", "AWS profile to use when establishing sessions with AWS's SDK.")
	decryptCmd.PersistentFlags().String("my-private-key", "", "The receiver's private key.  A local file path.")
	decryptCmd.PersistentFlags().String("my-public-key", "", "The receiver's public key.  A local file path.")
    decryptCmd.PersistentFlags().String("ssm-private-key", "", "The receiver's private key.  A parameter name in SSM.")
	decryptCmd.PersistentFlags().String("ssm-public-key", "", "The receiver's public key.  A parameter name in SSM.")

	viper.BindPFlag("file", decryptCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("destination", decryptCmd.PersistentFlags().Lookup("destination"))
	viper.BindPFlag("awsprofile", decryptCmd.PersistentFlags().Lookup("awsprofile"))
	viper.BindPFlag("my-private-key", decryptCmd.PersistentFlags().Lookup("my-private-key"))
	viper.BindPFlag("my-public-key", decryptCmd.PersistentFlags().Lookup("my-public-key"))
    viper.BindPFlag("ssm-private-key", decryptCmd.PersistentFlags().Lookup("ssm-private-key"))
	viper.BindPFlag("ssm-public-key", decryptCmd.PersistentFlags().Lookup("ssm-public-key"))

	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
}
