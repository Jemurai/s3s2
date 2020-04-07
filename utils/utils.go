package utils

import (
	"os"
	"path/filepath"
	"strings"
	"time"
	"io"
	"github.com/aws/aws-sdk-go/aws"

	session "github.com/aws/aws-sdk-go/aws/session"
	options "github.com/tempuslabs/s3s2/options"
	client "github.com/aws/aws-sdk-go/aws/client"
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
        log.Error(msg)
        log.Error(err)
        panic(err)
    }
}

// CleanupFile deletes a file
func CleanupFile(fs string) error {
	err := os.Remove(fs)
	PanicIfError("Issue deleting file - ", err)
	return err
}

func CleanupDirectory(directory string) error {
    err := os.RemoveAll(directory)
    return err
}

func IsDirEmpty(name string) (bool, error) {
    f, err := os.Open(name)
    if err != nil {
        return false, err
    }
    defer f.Close()

    _, err = f.Readdirnames(1) // Or f.Readdir(1)
    if err == io.EOF {
        return true, nil
    }
    return false, err // Either not empty or error, suits both cases
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

// https://gobyexample.com/collection-functions
func index(vs []string, t string) int {
    for i, v := range vs {
        if v == t {
            return i
        }
    }
    return -1
}

// function to mimic Python's """>> if x in []"""
func Include(vs []string, t string) bool {
    return index(vs, t) >= 0
}

// Influence creation of the retry logic used by any aws-config-using tools
func getRetryer() client.DefaultRetryer {
    retryer := client.DefaultRetryer{NumMaxRetries:10}
    return retryer
}

// Influence creation of AWS config
func getAwsConfig(opts options.Options) aws.Config {
    conf := aws.Config{
        Region: aws.String(opts.Region),
        Retryer: getRetryer(),
    }
    return conf
    }

// Easily add new command line arguments to influence the creation of AWS sessions
func GetAwsSession(opts options.Options) *session.Session {
    var sess *session.Session

    // intended on share when ran on partner server using credential files
    if opts.AwsProfile != "" {
        log.Infof("Using AWS Profile '%s'", opts.AwsProfile)
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
