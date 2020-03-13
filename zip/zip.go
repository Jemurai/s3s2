package zip

import (
	"github.com/tempuslabs/s3s2_new/options"
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	utils "github.com/tempuslabs/s3s2_new/utils"
	file "github.com/tempuslabs/s3s2_new/utils/file"
)

// ZipFile zips the provided file.
func ZipFile(fs file.File, options options.Options) string {

    zip_file_name := fs.GetZipName()

	newZipFile, err := os.Create(zip_file_name)
	utils.LogIfError("Unable to create zip file - ", err)
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	zipfile, err := os.Open(zip_file_name)
	utils.LogIfError("Unable to open zip file location - ", err)
	defer zipfile.Close()

	// Get the file information
	info, err := zipfile.Stat()
	utils.LogIfError("Unable to get zip file information - ", err)

	header, err := zip.FileInfoHeader(info)
	utils.LogIfError("Unable to get zip file header info - ", err)

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = strings.Replace(zip_file_name, options.Directory, "", -1)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	utils.LogIfError("Unable to create header info - ", err)

	if _, err = io.Copy(writer, zipfile); err != nil {
		log.Error(err)
	}

	return zip_file_name
}

// UnZipFile uncompresses and archive
func UnZipFile(filename string, destination string) string {

	returnFn := filename
	if !strings.HasSuffix(filename, ".zip") {
		log.Warnf("Skipping file because it is not a zip file, %s", filename)
		return returnFn
	}

	zReader, err := zip.OpenReader(filename)
	if err != nil {
		log.Error(err)
	}
	defer zReader.Close()
	for _, file := range zReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer zippedFile.Close()

		cleaned_name := utils.ToPosixPath(file.Name)

		extractedFilePath := filepath.Join(
			destination,
			"decrypted",
			cleaned_name,
		)

		log.Debugf("\tExtracted path: %s", extractedFilePath)
		if file.FileInfo().IsDir() {
			log.Println("Directory Created:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			log.Println("\tFile extracted:", cleaned_name)

			extractDir := filepath.Dir(extractedFilePath)
			os.MkdirAll(extractDir, os.ModePerm)
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				log.Fatal(err)
			}

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Fatal(err)
			}
			returnFn = utils.ToPosixPath(cleaned_name)
			outputFile.Close()
		}
	}
	log.Debugf("\tUnzip returning file name %s", returnFn)
	return returnFn
}
