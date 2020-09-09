
package cmd

import (
	"os"
	"fmt"
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
	zip "github.com/tempuslabs/s3s2/zip"
	encrypt "github.com/tempuslabs/s3s2/encrypt"
	manifest "github.com/tempuslabs/s3s2/manifest"
	options "github.com/tempuslabs/s3s2/options"
	aws_helpers "github.com/tempuslabs/s3s2/aws_helpers"
	utils "github.com/tempuslabs/s3s2/utils"
	file "github.com/tempuslabs/s3s2/file"
)


var opts options.Options

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Retrieve files that are stored securely in S3 and decrypt them",
	Long:  `Retrieve files that are stored securely in S3 and decrypt them`,
	// bug in Viper prevents shared flag names across different commands
	// placing these in the prerun is the workaround
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("directory", cmd.Flags().Lookup("directory"))
		viper.BindPFlag("region", cmd.Flags().Lookup("region"))
		viper.BindPFlag("parallelism", cmd.Flags().Lookup("parallelism"))
		viper.BindPFlag("aws-profile", cmd.Flags().Lookup("aws-profile"))
		viper.BindPFlag("ssm-public-key", cmd.Flags().Lookup("ssm-public-key"))
		cmd.MarkFlagRequired("directory")
		cmd.MarkFlagRequired("region")
	},
	Run: func(cmd *cobra.Command, args []string) {

		opts := buildDecryptOptions()
		checkDecryptOptions(opts)

        // top level clients
        sess := utils.GetAwsSession(opts)
	    _pubKey := encrypt.GetPubKey(sess, opts)
	    _privKey := encrypt.GetPrivKey(sess, opts)

	    os.MkdirAll(opts.Directory, os.ModePerm)

	    // if downloading via manifest
		if strings.HasSuffix(opts.File, "manifest.json") {

		    log.Info("Detected manifest file...")

		    target_manifest_path := filepath.Join(opts.Directory, filepath.Base(opts.File))
			fn, err := aws_helpers.DownloadFile(sess, opts.Bucket, opts.Org, opts.File, target_manifest_path)
			utils.PanicIfError("Unable to download file - ", err)

			m := manifest.ReadManifest(fn)
			batch_folder := m.Folder
			file_structs := m.Files

            var wg sync.WaitGroup
            sem := make(chan int, opts.Parallelism)

            for _, fs := range file_structs {
                wg.Add(1)
                go func(wg *sync.WaitGroup, sess *session.Session, _pubkey *packet.PublicKey, _privKey *packet.PrivateKey, folder string, fs file.File, opts options.Options) {
                    sem <- 1
                    defer func() { <-sem }()
                    defer wg.Done()
                    // if block is for cases where AWS session expires, so we re-create session and attempt file again
                    if decryptFile(sess, _pubKey, _privKey, m, fs, opts) != nil {
                        sess = utils.GetAwsSession(opts)
                        err := decryptFile(sess, _pubKey, _privKey, m, fs, opts)
                        if err != nil {}
                            log.Warn("Error during decrypt-file session expiration if block!")
                            log.Errorf("Error: '%v'", err)
                            panic(err)
                    }
                }(&wg, sess, _pubKey, _privKey, batch_folder, fs, opts)
            }
            wg.Wait()
        }
	},
}

func decryptFile(sess *session.Session, _pubkey *packet.PublicKey, _privkey *packet.PrivateKey, m manifest.Manifest, fs file.File, opts options.Options) error {
	start := time.Now()
	log.Debugf("Starting decryption on file '%s'", fs.Name)

	// enforce posix path
	fs.Name = utils.ToPosixPath(fs.Name)

	aws_key := fs.GetEncryptedName(m.Folder)
	target_path := fs.GetEncryptedName(opts.Directory)

	fn_zip := fs.GetZipName(opts.Directory)
	fn_decrypt := fs.GetSourceName("decrypted")

	nested_dir := filepath.Dir(target_path)
	os.MkdirAll(nested_dir, os.ModePerm)

	_, err := aws_helpers.DownloadFile(sess, opts.Bucket, m.Organization, aws_key, target_path)
	utils.PanicIfError("Unable to download file - ", err)

    encrypt.DecryptFile(_pubkey, _privkey, target_path, fn_zip, opts)
	zip.UnZipFile(fn_zip, fn_decrypt, opts.Directory)

    utils.Timing(start, fmt.Sprintf("\tProcessed file '%s' in ", fs.Name) + "%f seconds")

	return err
}

func buildDecryptOptions() options.Options {
	bucket := viper.GetString("bucket")
	file := viper.GetString("file")
	directory := viper.GetString("directory")
	org := viper.GetString("org")

	if !strings.HasSuffix(directory, "/") {
		directory = directory + "/"
	}

	region := viper.GetString("region")
	awsProfile := viper.GetString("aws-profile")
	privKey := viper.GetString("my-private-key")
	pubKey := viper.GetString("my-public-key")
	ssmPrivKey := viper.GetString("ssm-private-key")
	ssmPubKey := viper.GetString("ssm-public-key")
	parallelism := viper.GetInt("parallelism")

	options := options.Options{
		Bucket:      bucket,
		File:        file,
		Directory:   directory,
		Org:         org,
		Region:      region,
		PrivKey:     privKey,
		PubKey:      pubKey,
		SSMPrivKey:  ssmPrivKey,
		SSMPubKey:   ssmPubKey,
		AwsProfile:  awsProfile,
		Parallelism: parallelism,
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
		log.Warn("Need to supply a file to decrypt. Should be the file path within the bucket but not including the bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Bucket == "" {
		log.Warn("Need to supply a bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Directory == "" {
		log.Warn("Need to supply a destination for the files to decrypt.  Should be a local path.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Region == "" {
		log.Warn("Need to supply a region for the S3 bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.PubKey == "" && options.SSMPubKey == "" {
	    log.Warn("Need to supply a public encryption key parameter.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.PrivKey == "" && options.SSMPrivKey == "" {
	    log.Warn("Need to supply a private encryption key parameter.")
		log.Panic("Insufficient information to perform decryption.")
	}
}

func init() {
	rootCmd.AddCommand(decryptCmd)

    // core flags
	decryptCmd.PersistentFlags().String("file", "", "The path to the file to decrypt.  Can be manifest or single file.")
	decryptCmd.MarkFlagRequired("file")
	decryptCmd.PersistentFlags().String("directory", "", "The destination directory to decrypt and unzip.")
	decryptCmd.MarkFlagRequired("directory")
	decryptCmd.PersistentFlags().String("region", "", "The AWS region of the target bucket.")
	decryptCmd.MarkFlagRequired("region")

    // technical configuration
	decryptCmd.PersistentFlags().Int("parallelism", 10, "The maximum number of files to download and decrypt at a time.")
	decryptCmd.PersistentFlags().String("awsprofile", "", "AWS profile to use when establishing sessions with AWS's SDK.")

    // ssm keys
	decryptCmd.PersistentFlags().String("my-private-key", "", "The receiver's private key.  A local file path.")
	decryptCmd.PersistentFlags().String("my-public-key", "", "The receiver's public key.  A local file path.")
    decryptCmd.PersistentFlags().String("ssm-private-key", "", "The receiver's private key.  A parameter name in SSM.")
    decryptCmd.PersistentFlags().String("ssm-public-key", "", "The receiver's public key.  A parameter name in SSM.")

	viper.BindPFlag("file", decryptCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("directory", decryptCmd.PersistentFlags().Lookup("directory"))
	viper.BindPFlag("parallelism", decryptCmd.PersistentFlags().Lookup("parallelism"))
	viper.BindPFlag("aws-profile", decryptCmd.PersistentFlags().Lookup("aws-profile"))
	viper.BindPFlag("my-private-key", decryptCmd.PersistentFlags().Lookup("my-private-key"))
	viper.BindPFlag("my-public-key", decryptCmd.PersistentFlags().Lookup("my-public-key"))
    viper.BindPFlag("ssm-private-key", decryptCmd.PersistentFlags().Lookup("ssm-private-key"))
    viper.BindPFlag("ssm-public-key", decryptCmd.PersistentFlags().Lookup("ssm-public-key"))

	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
}
