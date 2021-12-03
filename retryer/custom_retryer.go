package retryer

import (
	"strings"
	"github.com/aws/aws-sdk-go/aws/request"
	client "github.com/aws/aws-sdk-go/aws/client"
)

type CustomRetryer struct {
	client.DefaultRetryer
}

// ShouldRetry overrides the SDK's built in DefaultRetryer adding customization
// to retry after read: connection reset error.
func (r CustomRetryer) ShouldRetry(req *request.Request) bool {
	if req.Error!=nil && strings.Contains(req.Error.Error(), "read: connection reset") {
		return true
	}

	// Fallback to SDK's built in retry rules
	return r.DefaultRetryer.ShouldRetry(req)
}