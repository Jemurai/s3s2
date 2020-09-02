
package cmd

import (
	"sync"
	"time"
	"fmt"
	"os"
	"strings"
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
		viper.BindPFlag("region", cmd.Flags().Lookup("region"))
		viper.BindPFlag("parallelism", cmd.Flags().Lookup("parallelism"))
		viper.BindPFlag("aws-profile", cmd.Flags().Lookup("aws-profile"))
		viper.BindPFlag("ssm-public-key", cmd.Flags().Lookup("ssm-public-key"))
		cmd.MarkFlagRequired("directory")
		cmd.MarkFlagRequired("org")
		cmd.MarkFlagRequired("region")
	},
	Run: func(cmd *cobra.Command, args []string) {

		opts := buildShareOptions(cmd)
		checkShareOptions(opts)

		start := time.Now()
		fnuuid := start.Format("20060102150405") // golang uses numeric constants for timestamp formatting

		file_structs, file_structs_metadata, err := file.GetFileStructsFromDir(opts.Directory, opts)
		utils.PanicIfError("Error reading directory", err)

		if len(file_structs) == 0 && len(file_structs_metadata) == 0 {
		    panic("No files from input directory were read. This means the directory is empty or only contains invalid files.")
		}

		file_struct_batches := append([][]file.File{file_structs_metadata}, file.ChunkArray(file_structs, opts.BatchSize)...)

	    var work_folder string
        if opts.ScratchDirectory != "" {
            work_folder = filepath.Join(opts.ScratchDirectory, opts.Org)
        } else {
            work_folder = opts.Directory
        }

        sess := utils.GetAwsSession(opts)
	    _pubKey := encrypt.GetPubKey(sess, opts)

	    sem := make(chan int, opts.Parallelism)

		change_s3_folders_at_size := 100000 + len(file_structs_metadata)
        current_s3_folder_size := 0
        current_s3_batch := 0

        var batch_folder string
		var all_uploaded_files_so_far []file.File
		var m manifest.Manifest
		var wg sync.WaitGroup

		batch_folder = fmt.Sprintf("%s_s3s2_%s_%d", opts.Prefix, fnuuid, current_s3_batch)

        // for each batch
		for i_batch, batch := range file_struct_batches {

		    log.Infof("Processing batch '%d'...", i_batch)

		    // refresh session every batch
		    sess = utils.GetAwsSession(opts)

            // tie off this current s3 directory allowing us to decrypt in batches of this size
            // this is used to create digestable folders for decrypt
		    if current_s3_folder_size + len(batch) > change_s3_folders_at_size {

                // fire lambda for the batch we are tieing off
		        if opts.LambdaTrigger == true {
		            aws_helpers.UploadLambdaTrigger(sess, opts.Org, batch_folder, opts)
		        }

                // reset / increment variables
		        current_s3_folder_size = 0
		        current_s3_batch += 1
		        batch_folder = fmt.Sprintf("%s_s3s2_%s_%d", opts.Prefix, fnuuid, current_s3_batch)

                // ensure the new s3 folder also has the metadata files
		        for _, mdf := range file_structs_metadata {
		            processFile(sess, _pubKey, batch_folder, work_folder, mdf, opts)
		            current_s3_folder_size += 1
		        }

		        all_uploaded_files_so_far = file_structs_metadata

            }

            wg.Add(len(batch))

            // for each file in batch
            for _, fs := range batch {
                go func(wg *sync.WaitGroup, sess *session.Session, _pubkey *packet.PublicKey, folder string, fs file.File, opts options.Options) {
                    sem <- 1
                    defer func() { <-sem }()
                    defer wg.Done()
                    processFile(sess, _pubKey, batch_folder, work_folder, fs, opts)
                }(&wg, sess, _pubKey, batch_folder, fs, opts)
            }

            wg.Wait()

            current_s3_folder_size += len(batch)
		    all_uploaded_files_so_far = append(all_uploaded_files_so_far, batch...)

            // upon batch completion
            m, err = manifest.BuildManifest(all_uploaded_files_so_far, batch_folder, opts)
            utils.PanicIfError("Error building Manifest", err)

            // create manifest in top-level directory - overwrite any existing manifest to include latest batch
            manifest_aws_key := filepath.Join(batch_folder, m.Name)
            manifest_local := filepath.Join(opts.Directory, m.Name)
            err = aws_helpers.UploadFile(sess, opts.Org, manifest_aws_key, manifest_local, opts)
            utils.PanicIfError("Error uploading Manifest", err)

            // archive the files we processed in this batch, dont archive metadata files until entire process is done
            if opts.ArchiveDirectory != "" && i_batch != 0 {
                log.Infof("Archiving files in batch '%d'", i_batch)
                file.ArchiveFileStructs(batch, opts.Directory, opts.ArchiveDirectory)
            }

            log.Infof("Successfully processed batch '%d'", i_batch)

        }
        // archive metafiles now
        if opts.ArchiveDirectory != "" {
            file.ArchiveFileStructs(file_structs_metadata, opts.Directory, opts.ArchiveDirectory)
        }

        utils.RemoveContents(opts.Directory)

        if opts.ScratchDirectory != "" {
            os.Remove(work_folder)
        }

		utils.Timing(start, "Elapsed time: %f")
        if opts.LambdaTrigger == true {
            aws_helpers.UploadLambdaTrigger(sess, opts.Org, batch_folder, opts)
        }
    },
}

func processFile(sess *session.Session, _pubkey *packet.PublicKey, aws_folder string, work_folder string, fs file.File, opts options.Options) {
	log.Debugf("Processing file '%s'", fs.Name)
	start := time.Now()

	fn_source := fs.GetSourceName(opts.Directory)
	fn_zip := fs.GetZipName(work_folder)
	fn_encrypt := fs.GetEncryptedName(work_folder)
	fn_aws_key := fs.GetEncryptedName(aws_folder)

	zip.ZipFile(fn_source, fn_zip, work_folder)
	encrypt.EncryptFile(_pubkey, fn_zip, fn_encrypt, opts)

	err := aws_helpers.UploadFile(sess, opts.Org, fn_aws_key, fn_encrypt, opts)

	if err != nil {
	    utils.PanicIfError("Error uploading file - ", err)
	} else {
	    utils.Timing(start, fmt.Sprintf("\tProcessed file '%s' in ", fs.Name) + "%f seconds")
	}

	// remove the zipped and encrypted files
    utils.CleanupFile(fn_zip)
	utils.CleanupFile(fn_encrypt)

    // these file names are often /internal_dir/basename
    // this line is a non-performant way for each file to be responsible for cleaning up the directory they were in
	if opts.ScratchDirectory != "" {
        nested_dir_crypt, _ := filepath.Split(fn_encrypt)
        source_dir_empty, _ := utils.IsDirEmpty(nested_dir_crypt)

        if source_dir_empty == true {
            os.Remove(nested_dir_crypt)
        }
    }

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
	ssmPubKey := filepath.Clean(viper.GetString("ssm-public-key"))

	archive_directory := viper.GetString("archive-directory")
	scratch_directory := viper.GetString("scratch-directory")
	aws_profile := viper.GetString("aws-profile")
	parallelism := viper.GetInt("parallelism")
	batchSize := viper.GetInt("batch-size")

	var metaDataFiles []string
	if viper.GetString("metadata-files") != "" {
	    metaDataFiles = strings.Split(viper.GetString("metadata-files"), ",")
	    }

	lambdaTrigger := viper.GetBool("lambda-trigger")

	options := options.Options{
		Directory       : directory,
		AwsKey          : awsKey,
		Bucket          : bucket,
		Region          : region,
		Org             : org,
		Prefix          : prefix,
		PubKey          : pubKey,
		SSMPubKey       : ssmPubKey,
		ScratchDirectory: scratch_directory,
		ArchiveDirectory: archive_directory,
		AwsProfile      : aws_profile,
		Parallelism     : parallelism,
		BatchSize       : batchSize,
		MetaDataFiles   : metaDataFiles,
		LambdaTrigger   : lambdaTrigger,
	}

	debug := viper.GetBool("debug")
	if debug != true {
		log.SetLevel(log.InfoLevel)
	}
	log.Debugf("Captured options: %s", options)

	return options
}
// Any assertions that need to be made regarding input arguments
func checkShareOptions(options options.Options) {
    log.Debug("Checking input arguments...")

	if options.AwsKey == "" && options.PubKey == "" && options.SSMPubKey == "" {
		panic("Need to supply either AWS Key for S3 level encryption or a public key for GPG encryption or both!. Insufficient key material to perform safe encryption.")
	}

	if options.Org == "" {
	    panic("A Org must be provided.")
	}

	if options.Bucket == "" {
		panic("A bucket must be provided.")
	}

	if options.Directory == "" {
		panic("Need to supply a destination for the files to decrypt.  Should be a local path.")
	}

    if options.Directory == "/" {
        panic("Input directory cannot be root!")
    }

	if !strings.Contains(strings.ToLower(options.Prefix), "clinical") && !strings.Contains(strings.ToLower(options.Prefix), "documents") {
	    panic("Prefix command line argument must contain 'clinical' or 'documents' to abide by our lambda trigger!")
	}
}

func init() {
    log.Debug("Initializing share command...")

	rootCmd.AddCommand(shareCmd)

    // core flags
	shareCmd.PersistentFlags().String("directory", "", "The directory to zip, encrypt and share.")
	shareCmd.MarkFlagRequired("directory")
	shareCmd.PersistentFlags().String("org", "", "The Org that owns the files.")
	shareCmd.MarkFlagRequired("org")
	shareCmd.PersistentFlags().String("prefix", "", "A prefix for the S3 path. Currently used to separate clinical and documents files.")

    // technical configuration
	shareCmd.PersistentFlags().Int("parallelism", 10, "The maximum number of files to download and decrypt at a time within a batch.")
	shareCmd.PersistentFlags().Int("batch-size", 10000, "Files are uploaded and archived in batches of this size. Manifest files are updated and uploaded after each factor of batch-size.")
	shareCmd.PersistentFlags().Bool("lambda-trigger", true, "Will send a trigger file to the S3 bucket upon both process completion (when all valid files in the input directory are uploaded) and each internal S3 bucket tie off.")
	shareCmd.PersistentFlags().String("aws-profile", "", "AWS Profile to use for the session.")

    // optional file / file-path configurations
    shareCmd.PersistentFlags().String("scratch-directory", "", "If provided, serves as location where .zip & .gpg files are written to. Is automatically suffixed by org argument. Intended to be leveraged if location will have superior write/read performance. If not provided, .zip and .gpg files are written to the original directory.")
    shareCmd.PersistentFlags().String("archive-directory", "", "If provided, contents of upload directory are moved here after each batch.")
    shareCmd.PersistentFlags().String("metadata-files", "", "If provided, these files are the first to be uploaded and the last to be archived out of the input directory. Comma-separated. I.E. --metadata-files=file1,file2,file3")

    // ssm key options
	shareCmd.PersistentFlags().String("awskey", "", "The agreed upon S3 key to encrypt data with at the bucket.")
	shareCmd.PersistentFlags().String("receiver-public-key", "", "The receiver's public key.  A local file path.")
	shareCmd.PersistentFlags().String("ssm-public-key", "", "The receiver's public key.  A local file path.")

	viper.BindPFlag("directory", shareCmd.PersistentFlags().Lookup("directory"))
	viper.BindPFlag("org", shareCmd.PersistentFlags().Lookup("org"))
	viper.BindPFlag("prefix", shareCmd.PersistentFlags().Lookup("prefix"))
	viper.BindPFlag("parallelism", shareCmd.PersistentFlags().Lookup("parallelism"))
	viper.BindPFlag("batch-size", shareCmd.PersistentFlags().Lookup("batch-size"))
	viper.BindPFlag("lambda-trigger", shareCmd.PersistentFlags().Lookup("lambda-trigger"))
	viper.BindPFlag("scratch-directory", shareCmd.PersistentFlags().Lookup("scratch-directory"))
	viper.BindPFlag("archive-directory", shareCmd.PersistentFlags().Lookup("archive-directory"))
	viper.BindPFlag("metadata-files", shareCmd.PersistentFlags().Lookup("metadata-files"))
	viper.BindPFlag("awskey", shareCmd.PersistentFlags().Lookup("awskey"))
	viper.BindPFlag("receiver-public-key", shareCmd.PersistentFlags().Lookup("receiver-public-key"))
	viper.BindPFlag("ssm-public-key", shareCmd.PersistentFlags().Lookup("ssm-public-key"))
	viper.BindPFlag("aws-profile", shareCmd.PersistentFlags().Lookup("aws-profile"))

	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}
