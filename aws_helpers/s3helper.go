
package aws_helpers

import (
	"os"

    "path/filepath"
	log "github.com/sirupsen/logrus"

	// local

	session "github.com/aws/aws-sdk-go/aws/session"
	options "github.com/tempuslabs/s3s2/options"
	utils "github.com/tempuslabs/s3s2/utils"

     // aws
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Upload
func UploadFile(sess *session.Session, org string, aws_key string, local_path string, opts options.Options) error {
    uploader := s3manager.NewUploader(sess)

    file, err := os.Open(local_path)
    utils.PanicIfError("Failed to open file for upload - ", err)

    final_key := utils.ToPosixPath(filepath.Join(org, aws_key))
    log.Debugf("Uploading file '%s' to aws key '%s'", local_path, final_key)

	if opts.AwsKey != "" {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:               aws.String(opts.Bucket),
			Key:                  aws.String(final_key),
			ServerSideEncryption: aws.String("aws:kms"),
			SSEKMSKeyId:          aws.String(opts.AwsKey),
			Body:                 file,
		})
		utils.PanicIfError("Failed to upload file: ", err)
		log.Infof("File '%s' uploaded to: '%s'", file.Name(), result.Location)
		file.Close()
		return err

	} else {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(opts.Bucket),
			Key:    aws.String(final_key),
			Body:   file,
		})
		utils.PanicIfError("Failed to upload file: ", err)
		log.Infof("File '%s' uploaded to: '%s'", file.Name(), result.Location)
		file.Close()
		return err
	}
}

// Download
func DownloadFile(sess *session.Session, bucket string, org string, aws_key string, target_path string) (string, error) {
	file, err := os.Create(target_path)
	utils.PanicIfError("Unable to open file - ", err)

	final_key := filepath.Join(org, aws_key)

	log.Infof("Downloading from key '%s' to file '%s'", final_key, target_path)

	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(final_key),
		})

    if err != nil {
        log.Errorf("Error downloading file '%s'", final_key)
    }

	defer file.Close()

	return file.Name(), err
}
