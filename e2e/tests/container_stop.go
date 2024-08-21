// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
	"github.com/runfinch/finch-daemon/pkg/api/response"
)

// ContainerStop tests the `POST containers/{id}/stop` API.
func ContainerStop(opt *option.Option) {
	Describe("stop a container", func() {
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
			relativeUrl := fmt.Sprintf("/containers/%s/stop", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should stop the container", func() {
			// start a container that keeps running
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			containerShouldBeRunning(opt, testContainerName)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldNotBeRunning(opt, testContainerName)
		})
		It("should fail to stop a stopped container", func() {
			// start a container that exits as soon as starts
			command.Run(opt, "run", "--name", testContainerName, defaultImage, "echo", "foo")
			command.Run(opt, "wait", testContainerName)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotModified))
		})
		It("should fail to stop a container that does not exist", func() {
			// don't create the container and call stop api.
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			var errResponse response.Error
			err = json.NewDecoder(res.Body).Decode(&errResponse)
			Expect(err).Should(BeNil())
			Expect(errResponse.Message).Should(Not(BeEmpty()))
		})
		It("should stop the container", func() {
			// start a container that keeps running
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			containerShouldBeRunning(opt, testContainerName)

			// stop the container with a timeout of 10 seconds
			now := time.Now()
			relativeUrl := fmt.Sprintf("/containers/%s/stop?t=10", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			res, err := uClient.Post(apiUrl, "application/json", nil)
			later := time.Now()
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			elapsed := later.Sub(now)
			Expect(elapsed.Seconds()).Should(BeNumerically(">", 9.0))
		})
	})
}
