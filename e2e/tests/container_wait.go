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

// ContainerWait tests the `POST containers/{id}/wait` API.
func ContainerWait(opt *option.Option) {
	Describe("wait for a container", func() {
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
			relativeUrl := fmt.Sprintf("/containers/%s/wait", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
		})

		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should wait for the container to exit and return status code", func() {
			// start a container and exit after 3s
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "3")

			// call wait on this container
			relativeUrl := fmt.Sprintf("/containers/%s/wait", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			var waitResponse struct {
				StatusCode int    `json:"StatusCode"`
				Error      string `json:"Error"`
			}
			err = json.NewDecoder(res.Body).Decode(&waitResponse)
			Expect(err).Should(BeNil())
			Expect(waitResponse.StatusCode).Should(Equal(0))
			Expect(waitResponse.Error).Should(BeEmpty())
		})

		It("should fail when container does not exist", func() {
			// don't create the container and call wait api
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))

			var errResponse response.Error
			err = json.NewDecoder(res.Body).Decode(&errResponse)
			Expect(err).Should(BeNil())
			Expect(errResponse.Message).Should(Not(BeEmpty()))
		})

		It("should reject wait condition parameter", func() {
			// start a container
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "5")

			// try to wait with a condition
			relativeUrl := fmt.Sprintf("/containers/%s/wait?condition=not-running", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))

			var errResponse response.Error
			err = json.NewDecoder(res.Body).Decode(&errResponse)
			Expect(err).Should(BeNil())
			Expect(errResponse.Message).Should(ContainSubstring("condition"))
		})
	})
}
