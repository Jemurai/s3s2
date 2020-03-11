
package aws_helpers

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	// local

	session "github.com/aws/aws-sdk-go/aws/session"

	options "github.com/tempuslabs/s3s2_new/options"
	utils "github.com/tempuslabs/s3s2_new/utils"

     // aws
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// UploadFile to S3.
// If the key is present, use it.  If it is not, don't.
// The share command should only allow this to get called
// IFF there is a key or the file has been gpg encrypted
// for the receiver.
func UploadFile(sess *session.Session, folder string, filename string, opts options.Options) error {
	log.Debugf("\tUploading file: '%s'", filename)

	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open file '%q', %v", filename, err)
	}

	basename := f.Name()
	aws_key := utils.OsAgnostic_HandleAwsKey(opts.Org, folder, basename, opts)

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
func DownloadFile(sess *session.Session, string, pullfile string, opts options.Options) (string, error) {
	log.Debugf("\tDownloading file (1): %s", pullfile)

	downloader := s3manager.NewDownloader(sess)

    filename := pullfile
	dirname := filepath.Dir(filename)

	os.MkdirAll(dirname, os.ModePerm)

	file, err := os.Create(utils.ForceBackSlash(filename))
	if err != nil {
		log.Debugf("\tDownloading file (2): %s", filename)
		return "", fmt.Errorf("Unable to open file %q, %v", filename, err)
	}
	defer file.Close()

	log.Debugf("\tDownloading file (3): About to pull %s, from bucket %s", filename, opts.Bucket)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(opts.Bucket),
			Key:    aws.String(filename),
		})

	if err != nil {
		log.Debugf("\tDownloading file (4): %s", filename)
		log.Errorf("Unable to download item '%q', %v", filename, err)
	}
	log.Debugf("\tDownloading file (5): %s", file.Name())
	return file.Name(), nil
}
