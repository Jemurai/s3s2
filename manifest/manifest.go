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

package manifest

import (
	"log"
	"os"
	"path/filepath"
)

// FileDescription is meta info about a file we will want to
// include in the Manifest.
type FileDescription struct {
	Name string
	Size int64
	//	Type string
	//	Hash string
	//	Date int64
}

// Manifest is a description of files.
type Manifest struct {
	Name  string
	Files []FileDescription
}

// BuildManifest builds a manifest from a directory.
func BuildManifest(directory string) Manifest {
	var files []FileDescription
	err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// TODO:  Add detail to file description.
			files = append(files, FileDescription{path, info.Size()})
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	// TODO:  Add more info to Manifest itself.
	manifest := Manifest{"Manifest", files}
	writeManifest(manifest, directory)
	return manifest
}

func writeManifest(manifest Manifest, directory string) error {
	// TODO:  Write JSON encoded manifest file.

	return nil
}
