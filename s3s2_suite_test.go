package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestS3s2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3s2 Suite")
}
