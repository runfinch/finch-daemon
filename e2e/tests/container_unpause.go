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

func ContainerUnpause(opt *option.Option) {
	Describe("unpause a container", func() {
		var (
			uClient    *http.Client
			version    string
			unpauseUrl string
			pauseUrl   string
		)

		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			relativeUrl_pause := fmt.Sprintf("/containers/%s/pause", testContainerName)
			relativeUrl_unpause := fmt.Sprintf("/containers/%s/unpause", testContainerName)
			unpauseUrl = client.ConvertToFinchUrl(version, relativeUrl_unpause)
			pauseUrl = client.ConvertToFinchUrl(version, relativeUrl_pause)
		})

		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should unpause a paused container", func() {
			// Start a container that keeps running
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")

			// Pause this running container
			res, err := uClient.Post(pauseUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))

			// Verify container is paused
			output := command.StdoutStr(opt, "inspect", "--format", "{{.State.Status}}", testContainerName)
			Expect(output).Should(Equal("paused"))

			// Unpause the paused container
			res, err = uClient.Post(unpauseUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))

			// Verify container is running again
			output = command.StdoutStr(opt, "inspect", "--format", "{{.State.Status}}", testContainerName)
			Expect(output).Should(Equal("running"))
		})

		It("should fail to unpause a non-existent container", func() {
			res, err := uClient.Post(unpauseUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))

			var body response.Error
			err = json.NewDecoder(res.Body).Decode(&body)
			Expect(err).Should(BeNil())
		})

		It("should fail to unpause a running container", func() {
			// Start a container that keeps running
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")

			// Try to unpause the running container
			res, err := uClient.Post(unpauseUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusConflict))

			var body response.Error
			err = json.NewDecoder(res.Body).Decode(&body)
			Expect(err).Should(BeNil())
			Expect(body.Message).Should(Equal(fmt.Sprintf("Container %s is not paused", testContainerName)))

			// Verify container is still running
			output := command.StdoutStr(opt, "inspect", "--format", "{{.State.Status}}", testContainerName)
			Expect(output).Should(Equal("running"))
		})

		It("should fail to unpause a stopped container", func() {
			// Create and start a container
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")

			// Verify container is running
			output := command.StdoutStr(opt, "inspect", "--format", "{{.State.Status}}", testContainerName)
			Expect(output).Should(Equal("running"))

			// Stop the container with a timeout to ensure it stops
			command.Run(opt, "stop", "-t", "1", testContainerName)

			// Try to unpause the stopped container
			res, err := uClient.Post(unpauseUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusConflict))

			var body response.Error
			err = json.NewDecoder(res.Body).Decode(&body)
			Expect(err).Should(BeNil())
			Expect(body.Message).Should(Equal(fmt.Sprintf("Container %s is not paused", testContainerName)))
		})
	})
}
