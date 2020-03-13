package utils

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	session "github.com/aws/aws-sdk-go/aws/session"

	options "github.com/tempuslabs/s3s2_new/options"
	log "github.com/sirupsen/logrus"
)

// Helper function to log a debug message of the elapsed time since input time
func Timing(start time.Time, message string) time.Time {
	current := time.Now()
	elapsed := current.Sub(start)
	log.Debugf(message, elapsed.Seconds())
	return current
}

// Helper function to log an error if exists
func LogIfError(msg string, err error) {
    if err != nil {
        log.Error(msg, err)
    }
}

// CleanupFile deletes a file
func CleanupFile(fs string) error {
	err := os.Remove(fs)
	LogIfError("Issue deleting file - ", err)
	return err
}

// CleanupDirectory deletes a file
func CleanupDirectory(fn string) {
    if fn != "/" {
        var err = os.RemoveAll(fn)
        LogIfError("Issue deleting file - ", err)
	}
}

// Will remove duplicate os.seperators from input string
// Will NOT convert forward slashes to back slashes
func ToSlashClean(s string) string {
    return filepath.ToSlash(filepath.Clean(s))
}

// Logic to force paths with forward slashes to backslashes. Main solution for Linux handling files uploaded via Windows
func ToPosixPath(s string) string {
    return strings.Replace(ToSlashClean(s), "\\", "/", -1)
}

func GetRelativePath(path string, relative_to string) string {
    rel, err := filepath.Rel(relative_to, path)

    if err != nil {
        log.Warnf("Unable to get relative path for : '%s'", path)
        return path
    } else {
        return ToSlashClean(rel)
        }
    }

// Builds filepath using blackslashes, regardless of operating system
// is used to make aws-compatible object keys
func OsAgnostic_HandleAwsKey(org string, folder string, fn string, opts options.Options) string {
    rel_path := GetRelativePath(fn, opts.Directory)
	return ToSlashClean(filepath.Join(org, folder, rel_path))
}

func getAwsConfig(opts options.Options) aws.Config {
    conf := aws.Config{Region: aws.String(opts.Region),}
    return conf
    }


// Easily add new command line arguments to influence the creation of AWS sessions
func GetAwsSession(opts options.Options) *session.Session {
    var sess *session.Session
    var err error

    if opts.AwsProfile != "" {
        sess, err = session.NewSessionWithOptions(session.Options{
        Profile: opts.AwsProfile,
        Config: getAwsConfig(opts),
        SharedConfigState: session.SharedConfigEnable,
        })
    } else {
        sess, err = session.NewSessionWithOptions(session.Options{
        Config: getAwsConfig(opts),
        })
    }

    if err != nil {
        panic(fmt.Sprintf("Unable to make AWS session: '%e'", err))
    } else {
        log.Warnf("Using AWS profile '%s'", opts.AwsProfile)
    }

    return sess
}
