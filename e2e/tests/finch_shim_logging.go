// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// FinchShimLogging tests that the log URI binary:// path is correctly set to finch-shim
// by verifying that GET /containers/{id}/logs returns the expected output after a container
// created via the HTTP API writes to stdout.
func FinchShimLogging() {
	Describe("finch-shim log URI", func() {
		var (
			uClient *http.Client
			version string
			url     string
		)
		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			url = client.ConvertToFinchUrl(version, "/containers/create")
		})
		AfterEach(func() {
			httpRemoveAll(uClient, version)
		})

		It("should return stdout via GET /logs after container created through HTTP API, confirming log URI points to finch-shim", func() {
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"echo", "finch-shim-log-test"}

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start the container and let it run to completion.
			httpStartContainer(uClient, version, testContainerName)
			time.Sleep(1 * time.Second)

			// Retrieve logs via the HTTP API — not via nerdctl CLI.
			// If the log URI was set to finch-shim correctly, containerd will invoke
			// finch-shim in logging mode and the output will be available here.
			logsURL := fmt.Sprintf("/containers/%s/logs?stdout=1&stderr=1&follow=0&tail=0", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, logsURL))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			_ = res.Body.Close()

			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			// The Docker multiplexed stream format prefixes each line with an 8-byte header.
			// body[8:] is the first log line payload.
			Expect(string(body)).Should(ContainSubstring("finch-shim-log-test"))
		})

		It("should return logs for a container that produces multiple lines, confirming streaming works end-to-end", func() {
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sh", "-c", "echo line1; echo line2; echo line3"}

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			httpStartContainer(uClient, version, testContainerName)
			time.Sleep(1 * time.Second)

			logsURL := fmt.Sprintf("/containers/%s/logs?stdout=1&stderr=1&follow=0&tail=0", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, logsURL))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			_ = res.Body.Close()

			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			Expect(string(body)).Should(ContainSubstring("line1"))
			Expect(string(body)).Should(ContainSubstring("line2"))
			Expect(string(body)).Should(ContainSubstring("line3"))
		})
	})
}
