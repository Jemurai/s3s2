
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

	"github.com/tempuslabs/s3s2_new/options"
	utils "github.com/tempuslabs/s3s2_new/utils"

)

// Manifest is a description of files.
type Manifest struct {
	Name         string
	Timestamp    time.Time
	Organization string
	Username     string
	User         string
	SudoUser     string
	Folder       string
	Files        []utils.File
}

func GetManifestPath(m *Manifest, opts) string {
  return filepath.Join(opts.Directory, m.Name)
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

func BuildManifest(file_structs []utils.File, batch_folder string, options options.Options) Manifest {

    log.Info("Building manifest...")

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
		Files:        file_structs,
	}

	writeManifest(manifest, options.Directory)

	return manifest
}

func hash(file string, options options.Options) string {

	var hash string
	if options.Hash == true {
		start := time.Now()

		hasher := sha256.New()
		s, err := ioutil.ReadFile(file)

		if err != nil {
			log.Fatal(err)
		}

		hasher.Write(s)
		hash = hex.EncodeToString(hasher.Sum(nil))

	    timing(start, "Time to hash: %f")

	} else {
		// Don't actually hash the file.
		hash = ""
	}

	return hash
}

func writeManifest(manifest Manifest, directory string) error {
	file, _ := json.MarshalIndent(manifest, "", " ")
	filename := filepath.Join(directory, "s3s2_manifest.json")

	log.Debugf("Creating local manifest '%s'", filename)

	ioutil.WriteFile(filename, file, 0644)

	return nil
}
