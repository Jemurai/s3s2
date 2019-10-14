package utils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// CleanupFile deletes a file
func CleanupFile(fn string) {
	var err = os.Remove(fn)
	if err != nil {
		log.Warnf("\tIssue deleting file: %s", fn)
	} else {
		log.Debugf("\tCleaned up: %s", fn)
	}
}

// CleanupDirectory deletes a file
func CleanupDirectory(fn string) {
	var err = os.RemoveAll(fn)
	if err != nil {
		log.Warnf("\tIssue deleting file: %s", fn)
	} else {
		log.Debugf("\tCleaned up: %s", fn)
	}
}


func OsAgnostic_HandleAwsKey(org string, folder string, fn string) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(org, folder, fn)))
}

