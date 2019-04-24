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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jemurai/s3s2/options"
)

// FileDescription is meta info about a file we will want to
// include in the Manifest.
type FileDescription struct {
	Name     string
	Size     int64
	Modified time.Time
	Hash     string
}

// Manifest is a description of files.
type Manifest struct {
	Name         string
	Timestamp    time.Time
	Organization string
	Username     string
	User         string
	SudoUser     string
	Files        []FileDescription
}

// BuildManifest builds a manifest from a directory.
func BuildManifest(options options.Options) Manifest {
	var files []FileDescription
	err := filepath.Walk(options.Directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				sha256hash := hash(path)
				files = append(files, FileDescription{path, info.Size(), info.ModTime(), sha256hash})
			}
			return nil
		})
	if err != nil {
		log.Error(err)
	}

	user, err := user.Current()
	sudoUser := os.Getenv("SUDO_USER") // In case they are sudo'ing, we can know the acting user.
	manifest := Manifest{options.Directory + "/s3s2_manifest.json", time.Now(), options.Org, user.Name, user.Username, sudoUser, files}
	writeManifest(manifest, options.Directory)
	return manifest
}

func hash(file string) string {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(file)
	hasher.Write(s)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func writeManifest(manifest Manifest, directory string) error {
	file, _ := json.MarshalIndent(manifest, "", " ")
	filename := directory + "/s3s2_manifest.json"
	log.Debug(filename)
	ioutil.WriteFile(filename, file, 0644)
	return nil
}
