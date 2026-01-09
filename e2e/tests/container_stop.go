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

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/e2e/client"
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
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"sleep", "infinity"})
			containerShouldBeRunning(opt, testContainerName)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldNotBeRunning(opt, testContainerName)
		})
		It("should fail to stop a stopped container", func() {
			// start a container that exits as soon as starts
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"echo", "foo"})
			httpWaitContainer(uClient, version, testContainerName)

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
		It("should stop the container with timeout", func() {
			// start a container that keeps running
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"sleep", "infinity"})
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
		It("should stop the container with SIGINT signal", func() {
			// Start a container that only logs the signal it receives
			httpRunContainer(uClient, version, testContainerName, defaultImage,
				[]string{"sh", "-c", `trap 'echo "Received signal: SIGINT"' SIGINT; while true; do sleep 1; done`})
			containerShouldBeRunning(opt, testContainerName)

			// Stop the container with SIGINT signal
			relativeUrl := fmt.Sprintf("/containers/%s/stop?signal=SIGINT", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))

			// Verify container is stopped by the API
			containerShouldNotBeRunning(opt, testContainerName)

			logs := httpContainerLogs(uClient, version, testContainerName)
			Expect(logs).Should(ContainSubstring("Received signal: SIGINT"))
		})
	})
}
