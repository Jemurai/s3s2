package utils

import (
	"os"
	"path/filepath"
	"strings"

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


func IsFilePath(key string) bool {
    info, err := os.Stat(key)

    log.Debug(err.Error())

    if os.IsNotExist(err) {
        return false
    } else if strings.Contains(err.Error(), "file name too long") {
        return false
    }

    return !info.IsDir()
}
