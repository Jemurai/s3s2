
package cmd

import (
	"strings"
	"sync"
	"time"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
    "golang.org/x/crypto/openpgp/packet"

    log "github.com/sirupsen/logrus"
	uuid "github.com/satori/go.uuid"

    // local
	archive "github.com/tempuslabs/s3s2/archive"
	encrypt "github.com/tempuslabs/s3s2/encrypt"
	manifest "github.com/tempuslabs/s3s2/manifest"
	options "github.com/tempuslabs/s3s2/options"
	aws_helpers "github.com/tempuslabs/s3s2/aws_helpers"
	utils "github.com/tempuslabs/s3s2/utils"
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

	    _pubKey := encrypt.GetPubKey(opts)

		fnuuid, _ := uuid.NewV4()
		batch_folder := opts.Prefix + "_s3s2_" + fnuuid.String()

		m := manifest.BuildManifest(batch_folder, opts)
		fmt_manifest_path := filepath.Join(opts.Directory, m.Name)

		if err := aws_helpers.UploadFile(batch_folder, fmt_manifest_path, opts); err != nil {
			log.Error(err)
		}

		utils.CleanupFile(fmt_manifest_path)
        sem := make(chan struct{}, 12)
		var wg sync.WaitGroup
		for i := 0; i < len(m.Files); i++ {

			local_path := filepath.Join(opts.Directory, m.Files[i].Name)
            wg.Add(1)
			go func(folder string, fn string, opts options.Options, wg *sync.WaitGroup) {
			    sem <- struct{}{}
			    defer func() { <-sem }()
				defer wg.Done()
				processFile(_pubKey, folder, fn, opts)
			}(batch_folder, local_path, opts, &wg)
		}
		wg.Wait()
		timing(start, "Elapsed time: %f")

	},
}

func processFile(_pubkey *packet.PublicKey, folder string, fn string, opts options.Options) {
	log.Debugf("Processing '%s'", fn)
	start := time.Now()

	fn = archive.ZipFile(fn, opts)
	archiveTime := timing(start, "\tArchive time (sec): %f")
	log.Debugf("\tCompressing file: '%s'", fn)

	encrypt.Encrypt(_pubkey, fn, opts)

	fn = fn + ".gpg"

	encryptTime := timing(archiveTime, "\tEncrypt time (sec): %f")

	err := aws_helpers.UploadFile(folder, fn, opts)
	if err != nil {
		log.Fatal(err)
	}

    log.Debug("Cleaning...")
	utils.CleanupFile(fn)
	if strings.HasSuffix(fn, ".gpg") {
		zipName := strings.TrimSuffix(fn, ".gpg")
		utils.CleanupFile(zipName)
	}

	timing(encryptTime, "\tUpload time (sec): %f")
	log.Debugf("\tProcessed '%s'", fn)
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

	directory := filepath.Clean(viper.GetString("directory"))

	awsKey := viper.GetString("awskey")
	bucket := viper.GetString("bucket")
	region := viper.GetString("region")

	org := viper.GetString("org")
	prefix := viper.GetString("prefix")

	pubKey := filepath.Clean(viper.GetString("receiver-public-key"))
	hash := viper.GetBool("hash")

	options := options.Options{
		Directory: directory,
		AwsKey:    awsKey,
		Bucket:    bucket,
		Region:    region,
		Org:       org,
		Prefix:     prefix,
		PubKey:    pubKey,
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
