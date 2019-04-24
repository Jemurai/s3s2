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

	"github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	zstd "github.com/valyala/gozstd"

	"os"
)

// ZstdFile archives the provided file.
func ZstdFile(filename string) string {
	zfilename := filename + ".zst"
	newZstdFile, err := os.Create(zfilename)
	if err != nil {
		log.Error(err)
	}
	defer newZstdFile.Close()

	zWriter := zstd.NewWriter(newZstdFile)
	defer zWriter.Release()

	file, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	if _, err = io.Copy(zWriter, file); err != nil {
		log.Error(err)
	}

	return zfilename
}
