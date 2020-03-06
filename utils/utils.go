package utils

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	session "github.com/aws/aws-sdk-go/aws/session"

	options "github.com/tempuslabs/s3s2/options"
	log "github.com/sirupsen/logrus"
)

// CleanupFile deletes a file
func CleanupFile(fn string) {
	var err = os.Remove(fn)
	if err != nil {
		log.Warnf("\tIssue deleting file: '%s'", fn)
	} else {
		log.Debugf("\tCleaned up: '%s'", fn)
	}
}

// CleanupDirectory deletes a file
func CleanupDirectory(fn string) {
    if fn != "/" {
        var err = os.RemoveAll(fn)
        if err != nil {
            log.Warnf("\tIssue deleting file: '%s'", fn)
            log.Warn(err)
        } else {
            log.Debugf("\tCleaned up: '%s'", fn)
	    }
	}
}

func visit(files *[]string) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Fatal(err)
        }
        *files = append(*files, path)
        return nil
    }
}

func walkDir(opts options.Options) []string {
    var files []string
    err := filepath.Walk(opts.Directory, visit(&files))
    if err != nil {
        panic(err)
    }
    return files
    }

// Copies files from input directory to archive directory
// Only clears input direct if no failures
// Archive Directory is always lateral to input directory
func ArchiveDirectory(opts options.Options) {

    os.Mkdir(path.Join(filepath.Dir(opts.Directory), opts.ArchiveDirectory), os.ModePerm)

    files := walkDir(opts)

    for _, file := range files {
        sourceFileStat, err := os.Stat(file)
        if err != nil {
            panic(err)
        }

        if !sourceFileStat.IsDir() {
            srcFile, err := os.Open(file)
            if err != nil {
                panic(err)
            }
            defer srcFile.Close()

            rel_path := GetRelativePath(file, opts.Directory)
            new_path := path.Join(filepath.Dir(opts.Directory), opts.ArchiveDirectory, filepath.Base(opts.Directory), rel_path)

            os.MkdirAll(filepath.Dir(new_path), os.ModePerm)

            destFile, err := os.Create(new_path) // creates if file doesn't exist
            if err != nil {
                panic(err)
            }
            defer destFile.Close()

            _, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
            if err != nil {
                panic(err)
            }
        }
    }
}

// Force to OS filepath seperator and clean filepaths. * Note * does nothing to filepaths with leading slashes
func ToSlashClean(s string) string {
    return filepath.ToSlash(filepath.Clean(s))
}

// Logic to force paths with forward slashes to backslashes. Main solution for Linux handling files uploaded via Windows
func ForceBackSlash(s string) string {
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


// Allows for easily adding new command line arguments to
// influene the creation of AWS sessions
func GetAwsSession(opts options.Options) *session.Session {
    var sess *session.Session
    var err error

    if opts.AwsProfile != "" {
        sess, err = session.NewSessionWithOptions(session.Options{
        // Specify profile to load for the session's config
        Profile: opts.AwsProfile,
        // Provide SDK Config options, such as Region.
        Config: getAwsConfig(opts),
        // Force enable Shared Config support
        SharedConfigState: session.SharedConfigEnable,
        AssumeRoleDuration: 12 * time.Hour,
        })
    } else {
        sess, err = session.NewSessionWithOptions(session.Options{
        Config: getAwsConfig(opts),
        AssumeRoleDuration: 12 * time.Hour,
        })
    }

    if err != nil {
        log.Warnf("Unable to make AWS session: '%s'", err)
    } else {
        log.Debugf("Using AWS session with profile: '%s'.", opts.AwsProfile)
    }

    return sess

}
