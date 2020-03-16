
package manifest

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"

    "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"

	file "github.com/tempuslabs/s3s2_new/file"
	utils "github.com/tempuslabs/s3s2_new/utils"
	options "github.com/tempuslabs/s3s2_new/options"

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
	Files        []file.File
}


// ReadManifest from a file.
func ReadManifest(file string) Manifest {
	var m Manifest

	rfile, err := os.Open(file)
	utils.LogIfError("Error opening manifest - ", err)

	bytes, err := ioutil.ReadAll(rfile)
    utils.LogIfError("Error reading manifest - ", err)

	jsoniter.Unmarshal(bytes, &m)

    defer rfile.Close()

	return m
}

func BuildManifest(file_structs []file.File, batch_folder string, options options.Options) (Manifest, error) {

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

	err = writeManifest(manifest, options.Directory)
	utils.LogIfError("Error getting current user - ", err)

	return manifest, err
}


func writeManifest(manifest Manifest, directory string) error {
	file, err := jsoniter.MarshalIndent(manifest, "", " ")
	filename := filepath.Join(directory, "s3s2_manifest.json")

	log.Debugf("Creating local manifest '%s'", filename)

	ioutil.WriteFile(filename, file, 0644)

	log.Debugf("Completed writing manifest '%s'", filename)

	return err
}
