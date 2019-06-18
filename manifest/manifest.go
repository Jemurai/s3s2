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
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
	Folder       string
	Files        []FileDescription
}

// ReadManifest from a file.
func ReadManifest(file string) Manifest {
	var m Manifest
	rfile, err := os.Open(file)
	if err != nil {
		log.Error(err)
	}
	bytes, err := ioutil.ReadAll(rfile)
	if err != nil {
		log.Error(err)
	}
	json.Unmarshal(bytes, &m)
	return m
}

// BuildManifest builds a manifest from a directory.
// It reads the contents of the directory and captures the file names,
// owners, dates and user into a manifest.json file.
func BuildManifest(folder string, options options.Options) Manifest {
	var files []FileDescription
	err := filepath.Walk(options.Directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && !strings.HasSuffix(path, "manifest.json") {
				sha256hash := hash(path)
				files = append(files, FileDescription{strings.Replace(path, options.Directory, "", -1), info.Size(), info.ModTime(), sha256hash})
			}
			return nil
		})
	if err != nil {
		log.Error(err)
	}

	user, err := user.Current()
	sudoUser := os.Getenv("SUDO_USER") // In case they are sudo'ing, we can know the acting user.
	manifest := Manifest{
		Name:         filepath.Clean("/s3s2_manifest.json"),
		Timestamp:    time.Now(),
		Organization: options.Org,
		Username:     user.Name,
		User:         user.Username,
		SudoUser:     sudoUser,
		Folder:       folder,
		Files:        files,
	}

	writeManifest(manifest, options.Directory)
	return manifest
}

func hash(file string) string {
	start := time.Now()
	/*  This is commented out because it is actually slow.
	hasher := sha256.New()
	s, err := ioutil.ReadFile(file)
	hasher.Write(s)
	if err != nil {
		log.Fatal(err)
	}
	*/
	current := time.Now()
	elapsed := current.Sub(start)
	log.Debugf("\tTime to hash %s : %f", file, elapsed)
	//	return hex.EncodeToString(hasher.Sum(nil))
	return "fake-hash"
}

func writeManifest(manifest Manifest, directory string) error {
	file, _ := json.MarshalIndent(manifest, "", " ")
	filename := directory + "/s3s2_manifest.json"
	log.Debug(filename)
	ioutil.WriteFile(filename, file, 0644)
	return nil
}
