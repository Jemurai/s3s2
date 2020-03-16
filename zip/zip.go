package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	utils "github.com/tempuslabs/s3s2/utils"
)

// ZipFile zips the provided file.
func ZipFile(InputFn string, OutputFn string, directory string) string {

    log.Infof("Zipping file '%s' to '%s'", InputFn, OutputFn)

	newZipFile, err := os.Create(OutputFn)
	utils.PanicIfError("Unable to create zip file - ", err)
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	zipfile, err := os.Open(InputFn)
	utils.PanicIfError("Unable to open zip file location - ", err)
	defer zipfile.Close()

	// Get the file information
	info, err := zipfile.Stat()
	utils.PanicIfError("Unable to get zip file information - ", err)

	header, err := zip.FileInfoHeader(info)
	utils.PanicIfError("Unable to get zip file header info - ", err)

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = strings.Replace(InputFn, directory, "", -1)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	utils.PanicIfError("Unable to create header info - ", err)

	if _, err = io.Copy(writer, zipfile); err != nil {
		log.Error(err)
	}

	return OutputFn
}

// UnZipFile uncompresses and archive
func UnZipFile(InputFn string, OutputFn string, directory string) string {

	if !strings.HasSuffix(InputFn, ".zip") {
		log.Warnf("Skipping file because it is not a zip file, %s", OutputFn)
		return OutputFn
	}

	zReader, err := zip.OpenReader(InputFn)
    utils.PanicIfError("Unable to open zipreader - ", err)
	defer zReader.Close()

	for _, file := range zReader.Reader.File {

		zippedFile, err := file.Open()
        utils.PanicIfError("Unable to open zipped file - ", err)
		defer zippedFile.Close()

		extractedFilePath := filepath.Join(directory, OutputFn)

		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, file.Mode())
			log.Debugf("Directory Created: '%s'", extractedFilePath)
		} else {

			extractDir := filepath.Dir(extractedFilePath)
			os.MkdirAll(extractDir, os.ModePerm)

			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
            utils.PanicIfError("Unable to open zipreader - ", err)
            log.Debugf("\tFile extracted to: '%s'", extractedFilePath)

			_, err = io.Copy(outputFile, zippedFile)
			utils.PanicIfError("Unable to create zipped file - ", err)

			outputFile.Close()
		}
	}
	return OutputFn
}
