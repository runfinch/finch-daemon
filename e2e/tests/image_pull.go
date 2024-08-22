// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
)

// ImagePull tests the `POST images/create` API.
func ImagePull(opt *option.Option) {
	Describe("pull an image", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			command.RemoveImages(opt)
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should pull the default image successfully", func() {
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s", defaultImage)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(opt, defaultImage)
		})
		It("should do nothing if image already exists", func() {
			command.Run(opt, "pull", defaultImage)
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s", defaultImage)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(opt, defaultImage)
		})
		It("should fail to pull a non-existent image", func() {
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s", nonexistentImageName)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
			waitForResponse(resp)
			imageShouldNotExist(opt, nonexistentImageName)
		})
		It("should pull the alpine image using the specified image tag", func() {
			imageName, imageTag, _ := strings.Cut(olderAlpineImage, ":")
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s&tag=%s", imageName, imageTag)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(opt, olderAlpineImage)
			imageShouldNotExist(opt, alpineImage)
		})
		It("should fail to pull an image with a malformed image name", func() {
			malformedImage := "alpine:image:latest"
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s", malformedImage)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusBadRequest))
			waitForResponse(resp)
		})
		It("should fail to pull an image with a malformed image tag", func() {
			imageName, _, _ := strings.Cut(olderAlpineImage, ":")
			malformedTag := "image:latest"
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s&tag=%s", imageName, malformedTag)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusBadRequest))
			waitForResponse(resp)
			imageShouldNotExist(opt, imageName)
		})
		It("should pull the alpine image with the specified platform", func() {
			platform := "linux/arm64"
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s&platform=%s", alpineImage, platform)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(opt, alpineImage)
		})
		It("should fail to pull an image with invalid platform", func() {
			platform := "invalid"
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s&platform=%s", alpineImage, platform)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusInternalServerError))
			waitForResponse(resp)
			imageShouldNotExist(opt, alpineImage)
		})
	})
}

// waitForResponse waits until the http response is closed with EOF.
func waitForResponse(resp *http.Response) {
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n == 0 && err != nil {
			break
		}
	}
}
