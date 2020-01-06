package archive

import (
	"github.com/tempuslabs/s3s2/options"
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ZipFile zips the provided file.
func ZipFile(filename string, options options.Options) string {
	zfilename := filename + ".zip"

	log.Debugf("The file name is " + zfilename)

	newZipFile, err := os.Create(zfilename)
	if err != nil {
		log.Error(err)
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	zipfile, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	defer zipfile.Close()

	// Get the file information
	info, err := zipfile.Stat()
	if err != nil {
		log.Error(err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		log.Error(err)
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = strings.Replace(filename, options.Directory, "", -1)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Error(err)
	}
	if _, err = io.Copy(writer, zipfile); err != nil {
		log.Error(err)
	}
	return zfilename
}

// UnZipFile uncompresses and archive
func UnZipFile(filename string, destination string) string {
	log.Debugf("Unzipping file %s", filename)
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

		// this value can differ from the input 'filename' argument, dir\\filename vs dir/filename
		log.Debugf("this is the files name from zreader " + file.Name)

		extractedFilePath := filepath.Join(
			destination,
			"decrypted",
			returnFn,
		)

		log.Debugf("\tExtracted path: %s", extractedFilePath)
		if file.FileInfo().IsDir() {
			log.Println("Directory Created:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			log.Println("\tFile extracted:", file.Name)

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
			returnFn = outputFile.Name()
			outputFile.Close()
		}
	}
	log.Debugf("\tUnzip returning file name %s", returnFn)
	return returnFn
}
