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

package archive

import (
	"github.com/jemurai/s3s2/options"
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ZipFile archives the provided list of files into a Zip file.
// This is functional but not currently used in S3S2 in favor of
// the Zst archive format which is faster and better compression.
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
		log.Debugf("this is the files name from zreader " + file.Name)
		extractedFilePath := filepath.Join(
			destination,
			file.Name,
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
