
package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/tempuslabs/s3s2/options"
	utils "github.com/tempuslabs/s3s2/utils"

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
func BuildManifest(batch_folder string, options options.Options) Manifest {

    log.Info("Building manifest...")

    // traverse files in given directory
	var files []FileDescription
	err := filepath.Walk(options.Directory,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if !info.IsDir() && !strings.HasSuffix(path, "manifest.json") {
				sha256hash := hash(path, options)
				fmt_path := utils.GetRelativePath(path, options)
				log.Debugf("Registering file '%s' to manifest as name '%s'...", path, fmt_path)
				files = append(files, FileDescription{fmt_path, info.Size(), info.ModTime(), sha256hash})
			}

			return nil
		})

	if err != nil {
		log.Error(err)
	}

	user, err := user.Current()
	sudoUser := os.Getenv("SUDO_USER") // In case they are sudo'ing, we can know the acting user.
	manifest := Manifest{
		Name:         filepath.Clean("s3s2_manifest.json"),
		Timestamp:    time.Now(),
		Organization: options.Org,
		Username:     user.Name,
		User:         user.Username,
		SudoUser:     sudoUser,
		Folder:       batch_folder,
		Files:        files,
	}

	writeManifest(manifest, options.Directory)
	return manifest
}

// CleanupFile just deletes a file.
func CleanupFile(fn string) {

	var err = os.Remove(fn)

	if err != nil {
		log.Warnf("\tIssue deleting file: '%s'", fn)
	} else {
		log.Debugf("\tCleaned up: '%s'", fn)
	}

}

func hash(file string, options options.Options) string {
	start := time.Now()

	var hash string
	if options.Hash == true {

		hasher := sha256.New()
		s, err := ioutil.ReadFile(file)
		hasher.Write(s)

		if err != nil {
			log.Fatal(err)
		}

		hash = hex.EncodeToString(hasher.Sum(nil))

	} else {
		// Don't actually hash the file.
		hash = "fake-hash"
	}

	current := time.Now()
	elapsed := current.Sub(start)

	log.Debugf("\tTime to hash %s : %v", file, elapsed)

	return hash
}

func writeManifest(manifest Manifest, directory string) error {
	file, _ := json.MarshalIndent(manifest, "", " ")
	filename := filepath.Join(directory, "s3s2_manifest.json")

	log.Debugf("Creating local manifest '%s'", filename)

	ioutil.WriteFile(filename, file, 0644)

	return nil
}
