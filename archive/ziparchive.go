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
	"archive/zip"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// ZipFile archives the provided list of files into a Zip file.
// This is functional but not currently used in S3S2 in favor of
// the Zst archive format which is faster and better compression.
func ZipFile(filename string) string {
	zfilename := filename + ".zip"
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
	header.Name = filename

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
