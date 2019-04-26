package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3s2", func() {
	var (
		test string
	)

	BeforeEach(func() {
		test = "hello world"
	})

	Describe("String handling", func() {
		Context("With string", func() {
			It("should not match empty", func() {
				Expect(test).NotTo(Equal(BeEmpty()))
			})
		})

		Context("With string", func() {
			It("should not match nil", func() {
				Expect(test).NotTo(Equal(BeNil()))
			})
		})
	})
})
