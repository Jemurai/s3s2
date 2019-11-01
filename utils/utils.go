package utils

import (
	"os"
	"path/filepath"

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
	var err = os.RemoveAll(fn)
	if err != nil {
		log.Warnf("\tIssue deleting file: '%s'", fn)
	} else {
		log.Debugf("\tCleaned up: '%s'", fn)
	}
}

// Builds filepath using blackslashes, regardless of operating system
// is used to make aws-compatible object keys
func OsAgnostic_HandleAwsKey(org string, folder string, fn string) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(org, folder, fn)))
}


func getAwsConfig(opts options.Options) aws.Config {
    conf := aws.Config{Region: aws.String(opts.Region),}
    return conf
    }

// Allows for easily adding new command line arguments to
// influene the creation of AWS sessions
func GetAwsSession(opts options.Options) *session.Session {

    sess, err := session.NewSessionWithOptions(session.Options{
    // Specify profile to load for the session's config
    Profile: opts.AwsProfile,
    // Provide SDK Config options, such as Region.
    Config: getAwsConfig(opts),
    // Force enable Shared Config support
    SharedConfigState: session.SharedConfigEnable,
    })

    if err != nil {
        log.Warnf("Unable to make AWS session: '%s'", err)
    } else {
        log.Debugf("Using AWS session with profile: '%s'.", opts.AwsProfile)
    }

    return sess

}
