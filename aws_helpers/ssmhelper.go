
package aws_helpers

import (
    log "github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Fetches value associated with provided keyname from SSM store
func GetParameterValue(ssm_service *ssm.SSM, keyname string) string {

    withDecryption := true

	param, err := ssm_service.GetParameter(&ssm.GetParameterInput{
		Name:           &keyname,
		WithDecryption: &withDecryption,
	})

    if err != nil {
		log.Fatal(err)
	}

    return *param.Parameter.Value
}

