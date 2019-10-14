package utils

import (
	"os"

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
