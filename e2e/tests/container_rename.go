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

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// ContainerRename tests the `POST containers/{id}/rename` API.
func ContainerRename(opt *option.Option) {
	Describe("rename a container", func() {
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
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should rename the container", func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			containerShouldBeRunning(opt, testContainerName)
			containerShouldNotExist(opt, testContainerName2)

			relativeUrl := fmt.Sprintf("/containers/%s/rename?name=%s", testContainerName, testContainerName2)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldBeRunning(opt, testContainerName2)
		})
		It("should fail to rename a container to taken name", func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			command.Run(opt, "run", "-d", "--name", testContainerName2, defaultImage, "sleep", "infinity")
			containerShouldBeRunning(opt, testContainerName)
			containerShouldBeRunning(opt, testContainerName2)

			relativeUrl := fmt.Sprintf("/containers/%s/rename?name=%s", testContainerName, testContainerName2)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusConflict))
		})
		It("should fail to rename a container that does not exist", func() {
			containerShouldNotExist(opt, testContainerName)

			relativeUrl := fmt.Sprintf("/containers/%s/rename?name=%s", testContainerName, testContainerName2)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
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
