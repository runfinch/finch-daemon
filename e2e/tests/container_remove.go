// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"
	"github.com/runfinch/finch-daemon/e2e/client"
	"github.com/runfinch/finch-daemon/pkg/api/response"
)

// ContainerRemove tests the `POST containers/{id}/remove` API.
func ContainerRemove(opt *option.Option) {
	Describe("remove a container", func() {
		var (
			uClient *http.Client
			version string
			apiUrl  string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			relativeUrl := fmt.Sprintf("/containers/%s/remove", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should remove the container", func() {
			// start a container that exits as soon as it starts
			command.Run(opt, "run", "--name", testContainerName, defaultImage, "echo", "foo")
			command.Run(opt, "wait", testContainerName)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldNotExist(opt, testContainerName)
		})
		It("should fail to remove a running container", func() {
			// start a container that keeps running
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusConflict))
			containerShouldExist(opt, testContainerName)
		})
		It("should successfully remove a running container with force=true", func() {
			// start a container that keeps running
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")

			relativeUrl := fmt.Sprintf("/containers/%s/remove?force=true", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldNotExist(opt, testContainerName)
		})
		It("should successfully remove a volume associated with it", func() {
			// start a container that keeps running
			command.Run(opt, "run", "-v", "test-vol", "--name", testContainerName, defaultImage)
			command.Run(opt, "wait", testContainerName)
			// get the total number of volumes after creating the new volume
			totalVolumes := len(command.GetAllVolumeNames(opt))
			relativeUrl := fmt.Sprintf("/containers/%s/remove?v=true", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldNotExist(opt, testContainerName)
			vCountAfterRemove := len(command.GetAllVolumeNames(opt))
			Expect(vCountAfterRemove).Should(Equal(totalVolumes - 1))
		})
		It("should fail to remove a container that does not exist", func() {
			// don't create the container and call remove api.
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			var errResponse response.Error
			err = json.NewDecoder(res.Body).Decode(&errResponse)
			Expect(err).Should(BeNil())
			Expect(errResponse.Message).Should(Not(BeEmpty()))
		})

	})
}
