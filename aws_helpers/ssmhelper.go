
package aws_helpers

import (
	"github.com/aws/aws-sdk-go/service/ssm"
	utils "github.com/tempuslabs/s3s2_new/utils"
)

// Fetches value associated with provided keyname from SSM store
func GetParameterValue(ssm_service *ssm.SSM, keyname string) string {
    withDecryption := true

	param, err := ssm_service.GetParameter(&ssm.GetParameterInput{
		Name:           &keyname,
		WithDecryption: &withDecryption,
	})

	utils.PanicIfError("Error getting SSM parameter - ", err)

    return *param.Parameter.Value
}
