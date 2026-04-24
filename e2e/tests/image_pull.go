// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			httpRemoveAllImages(uClient, version)
		})
		AfterEach(func() {
			httpRemoveAll(uClient, version)
		})

		It("should pull the default image successfully", func() {
			url := buildPullURL(version, defaultImage)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(defaultImage)
		})
		It("should do nothing if image already exists", func() {
			httpPullImage(uClient, version, defaultImage)
			url := buildPullURL(version, defaultImage)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(defaultImage)
		})
		It("should fail to pull a non-existent image", func() {
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s", nonexistentImageName)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
			waitForResponse(resp)
			imageShouldNotExist(nonexistentImageName)
		})
		It("should pull the alpine image using the specified image tag", func() {
			imageName, imageTag, _ := strings.Cut(olderAlpineImage, ":")
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s&tag=%s", imageName, imageTag)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(olderAlpineImage)
			imageShouldNotExist(alpineImage)
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
			imageShouldNotExist(imageName)
		})
		It("should pull the alpine image with the specified platform", func() {
			platform := "linux/arm64"
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s&platform=%s", alpineImage, platform)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			waitForResponse(resp)
			imageShouldExist(alpineImage)
		})
		It("should fail to pull an image with invalid platform", func() {
			platform := "invalid"
			relativeUrl := fmt.Sprintf("/images/create?fromImage=%s&platform=%s", alpineImage, platform)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusInternalServerError))
			waitForResponse(resp)
			imageShouldNotExist(alpineImage)
		})
	})
}

// buildPullURL constructs a /images/create URL with fromImage and tag as separate
// query params, so that images like "localhost:PORT/alpine:latest" are parsed
// correctly by the pull handler.
func buildPullURL(version, imageName string) string {
	repo := imageName
	tag := ""
	if lastColon := strings.LastIndex(imageName, ":"); lastColon > 0 {
		candidate := imageName[lastColon+1:]
		if !strings.Contains(candidate, "/") {
			repo = imageName[:lastColon]
			tag = candidate
		}
	}
	var relativeUrl string
	if tag != "" {
		relativeUrl = fmt.Sprintf("/images/create?fromImage=%s&tag=%s", repo, tag)
	} else {
		relativeUrl = fmt.Sprintf("/images/create?fromImage=%s", imageName)
	}
	return client.ConvertToFinchUrl(version, relativeUrl)
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
