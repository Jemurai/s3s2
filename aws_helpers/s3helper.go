
package aws_helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	options "github.com/tempuslabs/s3s2/options"
	log "github.com/sirupsen/logrus"
	utils "github.com/tempuslabs/s3s2/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// UploadFile to S3.
// If the key is present, use it.  If it is not, don't.
// The share command should only allow this to get called
// IFF there is a key or the file has been gpg encrypted
// for the receiver.
func UploadFile(folder string, filename string, opts options.Options) error {
	log.Debugf("\tUploading file: '%s'", filename)

    sess := utils.GetAwsSession(opts)

	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open file '%q', %v", filename, err)
	}

	basename := filepath.Base(f.Name())
	aws_key := utils.OsAgnostic_HandleAwsKey(opts.Org, folder, basename)

	if opts.AwsKey != "" {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:               aws.String(opts.Bucket),
			Key:                  aws.String(aws_key),
			ServerSideEncryption: aws.String("aws:kms"),
			SSEKMSKeyId:          aws.String(opts.AwsKey),
			Body:                 f,
		})
		if err != nil {
			return fmt.Errorf("Failed to upload file: '%v'", err)
		}
		log.Infof("\tFile '%s' uploaded to: '%s'", filename, result.Location)
	} else {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(opts.Bucket),
			Key:    aws.String(aws_key),
			Body:   f,
		})
		if err != nil {
			return fmt.Errorf("Failed to upload file: %v", err)
		}
		log.Infof("\tFile '%s' uploaded to: '%s'", filename, result.Location)
	}
	return nil
}

// DownloadFile function to download a file from S3.
func DownloadFile(directory string, pullfile string, opts options.Options) (string, error) {
	log.Debugf("\tDownloading file (1): %s", pullfile)

	sess := utils.GetAwsSession(opts)
	downloader := s3manager.NewDownloader(sess)

	fmt_pullfile := strings.Replace(pullfile, "\\", "", -1)
	fmt_directory := strings.Replace(directory, "\\", "", -1)

	filename := filepath.Clean(fmt_directory + "/" + fmt_pullfile)
	dirname := filepath.Dir(filename)
	log.Debugf("\tDownloading file (2): %s", filename)

	os.MkdirAll(dirname, os.ModePerm)
	file, err := os.Create(filename)
	if err != nil {
		log.Debugf("\tDownloading file (3): %s", filename)
		return "", fmt.Errorf("Unable to open file %q, %v", filename, err)
	}
	defer file.Close()

	log.Debugf("\tDownloading file (4): About to pull %s, from bucket %s", fmt_pullfile, opts.Bucket)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(opts.Bucket),
			Key:    aws.String(fmt_pullfile),
		})

	if err != nil {
		log.Debugf("\tDownloading file (5): %s", fmt_pullfile)
		log.Errorf("Unable to download item '%q', %v", fmt_pullfile, err)
	}
	log.Debugf("\tDownloading file (6): %s", file.Name())
	return file.Name(), nil
}
