// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
)

// ImageExport tests the `GET /images/{name}/get` API.
func ImageExport(opt *option.Option) {
	Describe("export an image", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			httpRemoveAllImages(uClient, version)
		})
		AfterEach(func() {
			httpRemoveAll(uClient, version)
		})

		It("should export an image successfully", func() {
			httpPullImage(uClient, version, defaultImage)
			relativeUrl := fmt.Sprintf("/images/%s/get", defaultImage)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Get(url)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).Should(Equal("application/x-tar"))

			// Verify response body is not empty (tar archive)
			body, err := io.ReadAll(resp.Body)
			Expect(err).Should(BeNil())
			Expect(len(body)).Should(BeNumerically(">", 0))
			resp.Body.Close()
		})

		It("should return 404 for non-existent image", func() {
			relativeUrl := fmt.Sprintf("/images/%s/get", nonexistentImageName)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Get(url)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
			resp.Body.Close()
		})
	})
}
