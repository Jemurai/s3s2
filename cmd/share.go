
package cmd

import (
	"sync"
	"time"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
    "golang.org/x/crypto/openpgp/packet"

    session "github.com/aws/aws-sdk-go/aws/session"

    log "github.com/sirupsen/logrus"

    // local
	zip "github.com/tempuslabs/s3s2/zip"
	encrypt "github.com/tempuslabs/s3s2/encrypt"
	manifest "github.com/tempuslabs/s3s2/manifest"
	options "github.com/tempuslabs/s3s2/options"
	aws_helpers "github.com/tempuslabs/s3s2/aws_helpers"
	file "github.com/tempuslabs/s3s2/file"
	utils "github.com/tempuslabs/s3s2/utils"

)

// shareCmd represents the share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Encrypt and upload to S3",
	Long: `Given a directory, encrypt all non-private-file contents and upload to S3.
    Behind the scenes, S3S2 checks to ensure the file is
    either GPG encrypted or passes S3 headers indicating
    that it will be encrypted.`,
	// bug in Viper prevents shared flag names across different commands
	// placing these in the prerun is the workaround
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("directory", cmd.Flags().Lookup("directory"))
		viper.BindPFlag("org", cmd.Flags().Lookup("org"))
		viper.BindPFlag("region", cmd.Flags().Lookup("region"))
		viper.BindPFlag("parallelism", cmd.Flags().Lookup("parallelism"))
		cmd.MarkFlagRequired("directory")
		cmd.MarkFlagRequired("org")
		cmd.MarkFlagRequired("region")
	},
	Run: func(cmd *cobra.Command, args []string) {

		opts := buildShareOptions(cmd)
		checkShareOptions(opts)

		start := time.Now()
		fnuuid := start.Format("20060102150405") // golang uses numeric constants for timestamp formatting
		batch_folder := opts.Prefix + "_s3s2_" + fnuuid

        sess := utils.GetAwsSession(opts)
	    _pubKey := encrypt.GetPubKey(sess, opts)

		file_structs, err := file.GetFileStructsFromDir(opts.Directory, opts)
		utils.PanicIfError("Error reading directory", err)

		m, err := manifest.BuildManifest(file_structs, batch_folder, opts)
        utils.PanicIfError("Error building Manifest", err)

        var wg sync.WaitGroup
        sem := make(chan int, opts.Parallelism)

        for _, fs := range m.Files {
            wg.Add(1)
            go func(wg *sync.WaitGroup, sess *session.Session, _pubkey *packet.PublicKey, folder string, fs file.File, opts options.Options) {
                sem <- 1
                defer func() { <-sem }()
                defer wg.Done()
                if processFile(sess, _pubKey, batch_folder, fs, opts) != nil {
                    sess = utils.GetAwsSession(opts)
                    err = processFile(sess, _pubKey, batch_folder, fs, opts)
                }
            }(&wg, sess, _pubKey, batch_folder, fs, opts)
        }
		wg.Wait()

		// create manifest in top-level directory
        manifest_aws_key := filepath.Join(batch_folder, m.Name)
        manifest_local := filepath.Join(opts.Directory, m.Name)

		err = aws_helpers.UploadFile(sess, opts.Org, manifest_aws_key, manifest_local, opts)

		if opts.ArchiveDirectory != "" {
		    utils.PerformArchive(opts.Directory, opts.ArchiveDirectory)
		}

		utils.Timing(start, "Elapsed time: %f")
	},
}

func processFile(sess *session.Session, _pubkey *packet.PublicKey, folder string, fs file.File, opts options.Options) error {
	log.Debugf("Processing '%s'", fs.Name)
	start := time.Now()

	fn_source := fs.GetSourceName(opts.Directory)
	fn_zip :=fs.GetZipName(opts.Directory)
	fn_encrypt := fs.GetEncryptedName(opts.Directory)
	aws_key := fs.GetEncryptedName(folder)

	zip.ZipFile(fn_source, fn_zip, opts.Directory)
	encrypt.EncryptFile(_pubkey, fn_zip, fn_encrypt, opts)

	err := aws_helpers.UploadFile(sess, opts.Org, aws_key, fn_encrypt, opts)

	if err != nil {
	    log.Error("Error uploading file - ", err)
	} else {
	    utils.Timing(start, fmt.Sprintf("\tProcessed file '%s' in ", fs.Name) + "%f seconds")
	}

	// cleanup regardless of the upload succeeding or not, we will retry outside of this function
    utils.CleanupFile(fn_zip)
	utils.CleanupFile(fn_encrypt)

    return err
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
	archive_directory := viper.GetString("archive-directory")
	aws_profile := viper.GetString("aws-profile")
	parallelism := viper.GetInt("parallelism")

	options := options.Options{
		Directory       : directory,
		AwsKey          : awsKey,
		Bucket          : bucket,
		Region          : region,
		Org             : org,
		Prefix          : prefix,
		PubKey          : pubKey,
		ArchiveDirectory: archive_directory,
		AwsProfile      : aws_profile,
		Parallelism     : parallelism,
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
	if options.AwsKey == "" && options.PubKey == "" {
		log.Warn("Need to supply either AWS Key for S3 level encryption or a public key for GPG encryption or both!")
		log.Panic("Insufficient key material to perform safe encryption.")
	} else if options.Bucket == "" {
		log.Warn("Need to supply a bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Directory == "" {
		log.Warn("Need to supply a destination for the files to decrypt.  Should be a local path.")
		log.Panic("Insufficient information to perform decryption.")
	}
}

func init() {
	rootCmd.AddCommand(shareCmd)

	shareCmd.PersistentFlags().String("directory", "", "The directory to zip, encrypt and share.")
	shareCmd.MarkFlagRequired("directory")
	shareCmd.PersistentFlags().String("org", "", "The Org that owns the files.")
	shareCmd.MarkFlagRequired("org")
	shareCmd.PersistentFlags().Int("parallelism", 10, "The maximum number of files to download and decrypt at a time.")

	shareCmd.PersistentFlags().String("prefix", "", "A prefix for the S3 path.")
	shareCmd.PersistentFlags().String("awskey", "", "The agreed upon S3 key to encrypt data with at the bucket.")
	shareCmd.PersistentFlags().String("receiver-public-key", "", "The receiver's public key.  A local file path.")
	shareCmd.PersistentFlags().String("archive-directory", "", "If provided, will move contents of upload directory contents to this location after upload.")
	shareCmd.PersistentFlags().String("aws-profile", "", "AWS Profile to use for the session.")

	viper.BindPFlag("directory", shareCmd.PersistentFlags().Lookup("directory"))
	viper.BindPFlag("org", shareCmd.PersistentFlags().Lookup("org"))
	viper.BindPFlag("prefix", shareCmd.PersistentFlags().Lookup("prefix"))
	viper.BindPFlag("awskey", shareCmd.PersistentFlags().Lookup("awskey"))
	viper.BindPFlag("archive-directory", shareCmd.PersistentFlags().Lookup("archive-directory"))
	viper.BindPFlag("receiver-public-key", shareCmd.PersistentFlags().Lookup("receiver-public-key"))
	viper.BindPFlag("aws-profile", shareCmd.PersistentFlags().Lookup("aws-profile"))
	viper.BindPFlag("parallelism", shareCmd.PersistentFlags().Lookup("parallelism"))

	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}
