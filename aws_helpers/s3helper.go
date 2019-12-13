
package aws_helpers

import (
	"fmt"
	"os"
	"path/filepath"

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
func UploadFile(uploader *s3manager.Uploader, folder string, filename string, opts options.Options) error {
	log.Debugf("\tUploading file: '%s'", filename)

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
func DownloadFile(downloader *s3manager.Downloader, string, pullfile string, opts options.Options) (string, error) {
	log.Debugf("\tDownloading file (1): %s", pullfile)

	dirname := filepath.Dir(pullfile)
	os.MkdirAll(dirname, os.ModePerm)

	file, err := os.Create(pullfile)
	if err != nil {
		log.Debugf("\tDownloading file (2): %s", pullfile)
		return "", fmt.Errorf("Unable to open file %q, %v", pullfile, err)
	}
	defer file.Close()

	log.Debugf("\tDownloading file (3): About to pull %s, from bucket %s", pullfile, opts.Bucket)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(opts.Bucket),
			Key:    aws.String(pullfile),
		})

	if err != nil {
		log.Debugf("\tDownloading file (4): %s", pullfile)
		log.Errorf("Unable to download item '%q', %v", pullfile, err)
	}
	log.Debugf("\tDownloading file (5): %s", file.Name())
	return file.Name(), nil
}
