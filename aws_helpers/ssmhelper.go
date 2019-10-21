
package aws_helpers

import (

	options "github.com/tempuslabs/s3s2/options"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)


func GetParameterValue(keyname string, options options.Options) string {

    withDecryption := true

	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(options.Region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		panic(err)
    }

    ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(options.Region))

	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &keyname,
		WithDecryption: &withDecryption,
	})

	value := *param.Parameter.Value

    return value
}

