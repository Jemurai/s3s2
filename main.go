
package main
import (
"github.com/tempuslabs/s3s2/cmd"
log "github.com/sirupsen/logrus"
)

func main() {
    log.Debug("Executing S3S2...")
	cmd.Execute()
}
