
package aws_helpers

import (

	options "github.com/tempuslabs/s3s2/options"
	utils "github.com/tempuslabs/s3s2/utils"
    log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Fetches value associated with provided keyname from SSM store
func GetParameterValue(keyname string, opts options.Options) string {

    withDecryption := true

    sess := utils.GetAwsSession(opts)
    ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(opts.Region))

	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &keyname,
		WithDecryption: &withDecryption,
	})

    if err != nil {
		log.Fatal(err)
	}

	value := *param.Parameter.Value

    return value
}

