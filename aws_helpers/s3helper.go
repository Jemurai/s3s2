
package aws_helpers

import (
	"os"
	"strings"
    "path/filepath"
	log "github.com/sirupsen/logrus"
	session "github.com/aws/aws-sdk-go/aws/session"
	options "github.com/tempuslabs/s3s2/options"
	utils "github.com/tempuslabs/s3s2/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Given file, open contents and send to S3
func UploadFile(sess *session.Session, org string, aws_key string, local_path string, opts options.Options) error {
    uploader := s3manager.NewUploader(sess)

    file, err := os.Open(local_path)
    utils.PanicIfError("Failed to open file for upload - ", err)

    final_key := utils.ToPosixPath(filepath.Clean(filepath.Join(org, aws_key)))
    log.Debugf("Uploading file '%s' to aws key '%s'", local_path, final_key)

    for {

        if opts.AwsKey != "" {
            result, err := uploader.Upload(&s3manager.UploadInput{
                Bucket:               aws.String(opts.Bucket),
                Key:                  aws.String(final_key),
                ServerSideEncryption: aws.String("aws:kms"),
                SSEKMSKeyId:          aws.String(opts.AwsKey),
                Body:                 file,
            })

            if err != nil {
                utils.PanicIfError("Failed to upload file: ", err)
            } else {
                log.Infof("File '%s' uploaded to: '%s'", file.Name(), result.Location)
                file.Close()
                return err
            }

        } else {
            result, err := uploader.Upload(&s3manager.UploadInput{
                Bucket: aws.String(opts.Bucket),
                Key:    aws.String(final_key),
                Body:   file,
            })

            if err != nil {
                utils.PanicIfError("Failed to upload file: ", err)
            } else {
                log.Infof("File '%s' uploaded to: '%s'", file.Name(), result.Location)
                file.Close()
                return err
            }
        }
    }
}

// Dedicated function for uploading our lambda trigger file - our way of communicating that s3s2 is done
func UploadLambdaTrigger(sess *session.Session, org string, folder string, opts options.Options) error {
    uploader := s3manager.NewUploader(sess)

    file_name := "._lambda_trigger"

    final_key := utils.ToPosixPath(filepath.Clean(filepath.Join(org, folder, file_name)))
    log.Debugf("Uploading file '%s' to aws key '%s'", file_name, final_key)

	if opts.AwsKey != "" {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:               aws.String(opts.Bucket),
			Key:                  aws.String(final_key),
			ServerSideEncryption: aws.String("aws:kms"),
			SSEKMSKeyId:          aws.String(opts.AwsKey),
			Body:                 strings.NewReader(""),
		})
		utils.PanicIfError("Failed to upload file: ", err)
		log.Infof("File '%s' uploaded to: '%s'", file_name, result.Location)
		return err

	} else {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(opts.Bucket),
			Key:    aws.String(final_key),
			Body:   strings.NewReader(""),
		})
		utils.PanicIfError("Failed to upload file: ", err)
		log.Infof("File '%s' uploaded to: '%s'", file_name, result.Location)
		return err
	}
}


// Given an aws key, download file to local machine
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
