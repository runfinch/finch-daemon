// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// ImageRemove tests the delete image API - `DELETE /images/{id}`
func ImageRemove(opt *option.Option) {
	Describe("remove an image", func() {
		var (
			uClient *http.Client
			version string
			apiUrl  string
			req     *http.Request
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			relativeUrl := fmt.Sprintf("/images/%s", defaultImage)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			var err error
			req, err = http.NewRequest("DELETE", apiUrl, nil)
			Expect(err).Should(BeNil())
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		Context("by name", func() {
			BeforeEach(func() {
				relativeUrl := fmt.Sprintf("/images/%s", defaultImage)
				apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
				var err error
				req, err = http.NewRequest("DELETE", apiUrl, nil)
				Expect(err).ShouldNot(HaveOccurred())
			})
			It("should remove the image", func() {
				pullImage(opt, defaultImage)
				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
				imageShouldNotExist(opt, defaultImage)
			})
			It("should fail to remove the image of a running container", func() {
				// start a container that keeps running
				command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusConflict))
				imageShouldExist(opt, defaultImage)
			})
			It("should fail to remove the image used in a stopped container", func() {
				// start a container that exits as soon as starts
				command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage)
				command.Run(opt, "wait", testContainerName)
				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusConflict))
				imageShouldExist(opt, defaultImage)
			})
			It("should successfully remove an image used in a stopped container with force=true", func() {
				// start a container that exits as soon as starts
				command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage)
				command.Run(opt, "wait", testContainerName)
				req, err := http.NewRequest("DELETE", apiUrl+"?force=true", nil)
				Expect(err).ShouldNot(HaveOccurred())
				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
				imageShouldNotExist(opt, defaultImage)
			})
			It("should fail to remove as image does not exist", func() {
				// don't pull the image and try to delete
				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})
		Context("by id", func() {
			BeforeEach(func() {
				imageID := pullImage(opt, defaultImage)
				relativeUrl := fmt.Sprintf("/images/%s", imageID)
				apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
				var err error
				req, err = http.NewRequest("DELETE", apiUrl, nil)
				Expect(err).ShouldNot(HaveOccurred())
			})
			It("should successfully remove an image", func() {
				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
				imageShouldNotExist(opt, defaultImage)
			})
			It("should fail to remove if multiple image with same id", func() {
				//create a new tag will create a reference with same id
				command.Run(opt, "image", "tag", defaultImage, "custom-image:latest")
				imageShouldExist(opt, "custom-image:latest")
				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusConflict))
				imageShouldExist(opt, defaultImage)
			})
			It("should successfully remove multiple images with same id using force=true", func() {
				req, err := http.NewRequest("DELETE", apiUrl+"?force=true", nil)
				Expect(err).ShouldNot(HaveOccurred())
				//create a new tag will create a reference with same id
				command.Run(opt, "image", "tag", defaultImage, "custom-image:latest")
				imageShouldExist(opt, "custom-image:latest")

				res, err := uClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
				imageShouldNotExist(opt, defaultImage)
				imageShouldNotExist(opt, "custom-image:latest")

			})
			//TODO: need to add a e2e test to make sure proper untagged and deleted value is generated for image remove api.

		})
	})
}
