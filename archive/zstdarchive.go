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
	zstd "github.com/valyala/gozstd"

	"os"
)

// ZstdFiles archives the provided list of files into a Zstd file.
func ZstdFiles(filename string, files []string) error {

	newZstdFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZstdFile.Close()

	zWriter := zstd.NewWriter(newZstdFile)
	defer zWriter.Release()

	// Add files to zip
	for _, file := range files {

		zfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zfile.Close()

	}
	return nil
}
