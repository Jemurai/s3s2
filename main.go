
package main

import (
	"os"
	"github.com/tempuslabs/s3s2/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Executing S3S2...")
	cmd.Execute()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})
	log.Infof("Completed S3S2.")
}

