// Copyright © 2019 Matt Konda <mkonda@jemurai.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3

import (
	"fmt"
	"os"
	"path/filepath"

	options "github.com/jemurai/s3s2/options"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// UploadFile to S3.
// If the key is present, use it.  If it is not, don't.
// The share command should only allow this to get called
// IFF there is a key or the file has been gpg encrypted
// for the receiver.
func UploadFile(folder string, filename string, options options.Options) error {
	log.Debugf("\tUploading file.")
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(options.Region),
	}))

	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filename, err)
	}

	if options.AwsKey != "" {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:               aws.String(options.Bucket),
			Key:                  aws.String(filepath.Clean(folder + "/" + f.Name())),
			ServerSideEncryption: aws.String("aws:kms"),
			SSEKMSKeyId:          aws.String(options.AwsKey),
			Body:                 f,
		})
		if err != nil {
			return fmt.Errorf("failed to upload file, %v", err)
		}
		log.Debugf("\tFile uploaded to, %s\n", result.Location)
	} else {
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(options.Bucket),
			Key:    aws.String(filepath.Clean(folder + "/" + f.Name())),
			Body:   f,
		})
		if err != nil {
			return fmt.Errorf("failed to upload file, %v", err)
		}
		log.Debugf("\tFile uploaded to, %s\n", result.Location)
	}
	return nil
}

// DownloadFile function to download a file from S3.
func DownloadFile(directory string, pullfile string, options options.Options) (string, error) {
	log.Debugf("\tDownloading file (1): %s", pullfile)
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(options.Region),
	}))

	downloader := s3manager.NewDownloader(sess)

	filename := filepath.Clean(directory + "/" + pullfile)
	dirname := filepath.Dir(filename)
	log.Debugf("\tDownloading file (2): %s", filename)

	os.MkdirAll(dirname, os.ModePerm)
	file, err := os.Create(filename)
	if err != nil {
		log.Debugf("\tDownloading file (3): %s", filename)
		return "", fmt.Errorf("Unable to open file %q, %v", filename, err)
	}
	defer file.Close()

	log.Debugf("\tDownloading file (4): About to pull %s, from bucket %s", pullfile, options.Bucket)
	// TODO:  Add the S3 KMS keys if needed.
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(options.Bucket),
			Key:    aws.String(pullfile),
		})
	if err != nil {
		log.Debugf("\tDownloading file (5): %s", pullfile)
		log.Errorf("Unable to download item %q, %v", pullfile, err)
	}
	log.Debugf("\tDownloading file (6): %s", file.Name())
	return file.Name(), nil
}
