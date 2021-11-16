
package gcp_helper

import (
	"context"
	"os"
    "path/filepath"
	"io"
	"io/ioutil"
	
	log "github.com/sirupsen/logrus"
	options "github.com/tempuslabs/s3s2/options"
	utils "github.com/tempuslabs/s3s2/utils"
	"cloud.google.com/go/storage"
)


// Given file, open contents and send to S3
func UploadFile(org string, aws_key string, local_path string, opts options.Options) error {
    ctx := context.Background()
	
	client, err := storage.NewClient(ctx)
	defer client.Close()

	utils.PanicIfError("Unable to get clients - ", err)
	


    file, err := os.Open(local_path)
	defer file.Close()

    utils.PanicIfError("Failed to open file for upload - ", err)

    final_key := utils.ToPosixPath(filepath.Clean(filepath.Join(org, aws_key)))
    log.Debugf("Uploading file '%s' to aws key '%s'", local_path, final_key)

   

	wc := client.Bucket(opts.Bucket).Object(final_key).NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = "text/plain"

	_, err = io.Copy(wc, file)

	if err != nil {
		utils.PanicIfError("Failed to upload file: ", err)
	} else {
		log.Infof("File '%s' uploaded to:  bucket = '%s', key = '%s'", file.Name(), opts.Bucket, final_key)
		return err
	}

	
	return nil

}

// Dedicated function for uploading our lambda trigger file - our way of communicating that s3s2 is done
func UploadLambdaTrigger(org string, folder string, opts options.Options) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	defer client.Close()
	utils.PanicIfError("Unable to get clients - ", err)
	
    file_name := "._lambda_trigger"
	file, err := os.Create(file_name)
	defer file.Close()
	bucket := opts.Bucket
    final_key := utils.ToPosixPath(filepath.Clean(filepath.Join(org, folder, file_name)))
    log.Debugf("Uploading file '%s' to bucket '%s' aws key '%s'", file_name, bucket, final_key)
	wc := client.Bucket(bucket).Object(final_key).NewWriter(ctx)
	defer wc.Close()
	_, err = io.Copy(wc, file)

	utils.PanicIfError("Failed to upload file: ", err)
	log.Infof("File '%s' uploaded to", file_name)
	return err
}


// Given an aws key, download file to local machine
func DownloadFile(bucket string, org string, aws_key string, target_path string) (string, error) {
	
	file, err := os.Create(target_path)
	utils.PanicIfError("Unable to open file - ", err)
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	utils.PanicIfError("Unable to get context client - ", err)

	final_key := filepath.Join(org, aws_key)
	
	rc, err := client.Bucket(bucket).Object(final_key).NewReader(ctx)
	utils.PanicIfError("Unable to get client - ", err)

	log.Infof("Downloading from key '%s' to file '%s'", final_key, target_path)

	data, err := ioutil.ReadAll(rc)
	utils.PanicIfError("Unable to download file - ", err)
	defer rc.Close()

	_, err = file.Write(data)
	utils.PanicIfError("Error downloading file to local a file - ", err)
	
	defer file.Close()

	return file.Name(), err
}
