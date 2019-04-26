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
	"io"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	zstd "github.com/valyala/gozstd"

	"os"
)

// ZstdFile archives the provided file.
func ZstdFile(filename string) string {

	file, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	zfilename := filename + ".zst"
	newZstdFile, err := os.Create(zfilename)
	if err != nil {
		log.Error(err)
	}
	defer newZstdFile.Close()

	zWriter := zstd.NewWriter(newZstdFile)
	defer zWriter.Close()

	if count, err := io.Copy(zWriter, file); err != nil {
		log.Error(err)
	} else {
		log.Debugf("\tWrote bytes %d", count)
	}
	return zfilename
}

// UnZstdFile uncompresses and archive
func UnZstdFile(filename string) string {

	if !strings.HasSuffix(filename, ".zst") {
		return filename
	}

	zfile, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	defer zfile.Close()

	zReader := zstd.NewReader(zfile)
	defer zReader.Release()

	fn := strings.TrimSuffix(filename, filepath.Ext(filename))

	newFile, err := os.Create(fn)
	if err != nil {
		log.Error(err)
	}
	defer newFile.Close()

	if count, err := io.Copy(newFile, zReader); err != nil {
		log.Error(err)
	} else {
		log.Debugf("\tWrote bytes %d", count)
	}
	return fn
}
