package utils

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	session "github.com/aws/aws-sdk-go/aws/session"

	options "github.com/tempuslabs/s3s2/options"
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
func PanicIfError(msg string, err error) {
    if err != nil {
        panic(msg)
    }
}

// CleanupFile deletes a file
func CleanupFile(fs string) error {
	err := os.Remove(fs)
	PanicIfError("Issue deleting file - ", err)
	return err
}

// CleanupDirectory deletes a file
func CleanupDirectory(fn string) {
    if fn != "/" {
        var err = os.RemoveAll(fn)
        PanicIfError("Issue deleting file - ", err)
	}
}

// Will remove duplicate os.seperators from input string
// Will NOT convert forward slashes to back slashes
// Serves as general cleansing function
func ToSlashClean(s string) string {
    return filepath.ToSlash(filepath.Clean(s))
}

// Logic to force paths with forward slashes to backslashes. Main solution for Linux handling files uploaded via Windows
func ToPosixPath(s string) string {
    return ToSlashClean(strings.Replace(s, "\\", "/", -1))
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

// Influence creation of AWS config
func getAwsConfig(opts options.Options) aws.Config {
    conf := aws.Config{Region: aws.String(opts.Region)}
    return conf
    }

// Easily add new command line arguments to influence the creation of AWS sessions
func GetAwsSession(opts options.Options) *session.Session {
    var sess *session.Session

    // intended on share when ran on partner server using credential files
    if opts.AwsProfile != "" {
        sess = session.Must(session.NewSessionWithOptions(session.Options{
        Profile: opts.AwsProfile,
        Config: getAwsConfig(opts),
        SharedConfigState: session.SharedConfigEnable,
        }))
    // intended on decrypt when ran on ec2 instance using sts
    } else {
        sess = session.Must(session.NewSessionWithOptions(session.Options{
        Config: getAwsConfig(opts),
        AssumeRoleDuration: 12 * time.Hour,
        }))
    }

    return sess
}
