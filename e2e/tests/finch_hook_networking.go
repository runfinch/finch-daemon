// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// TODO(remove-nerdctl-full-integration): All command.*(opt, ...) calls in this file must be
// converted to HTTP API calls when merging remove-nerdctl-binary-dep and remove-test-dependency.
// See .kiro/specs/finch-hook-helper-binary/hook-test-gap.md for the migration plan.

package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// FinchhookNetworking tests that CNI hooks ran correctly via finch-hook by verifying
// that a container attached to a bridge network receives an IP address and can reach
// external addresses. These tests confirm the OCI createRuntime/poststop hook pipeline
// is wired to finch-hook rather than nerdctl.
func FinchhookNetworking(opt *option.Option) {
	Describe("finch-hook CNI networking", func() {
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
			command.RemoveAll(opt)
		})

		It("should assign an IP address to a container on the bridge network, confirming CNI hooks ran", func() {
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sleep", "Infinity"}

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			command.Run(opt, "start", testContainerName)

			// IP assignment is the observable proof that the CNI hook fired.
			// If finch-hook failed to run, the container would have no network interface.
			verifyNetworkSettings(opt, testContainerName, "bridge")
		})

		It("should allow a container to reach an external address, confirming CNI masquerade rules are set", func() {
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			// wget with a short timeout so the test fails fast if networking is broken.
			// We only care about the TCP handshake succeeding (exit 0), not the HTTP response.
			options.Cmd = []string{"wget", "-q", "--spider", "--timeout=5", "http://example.com"}

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start -a waits for the container to exit and returns its stdout.
			// A zero exit from wget confirms outbound connectivity via CNI.
			out := command.StdoutStr(opt, "start", "-a", testContainerName)
			_ = out // wget --spider writes nothing to stdout on success
		})

		It("should attach a container to a user-defined bridge network with correct IP settings", func() {
			command.Run(opt, "network", "create", testNetwork)

			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.NetworkMode = testNetwork

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			command.Run(opt, "start", testContainerName)
			verifyNetworkSettings(opt, testContainerName, testNetwork)
		})

		It("should publish a port and make it reachable from the host, confirming nerdctl/ports labels were written", func() {
			// Run a minimal HTTP server inside the container on port 8080.
			// busybox httpd is available in the alpine image.
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{
				"sh", "-c",
				`mkdir -p /www && echo "ok" > /www/index.html && httpd -p 8080 -h /www && sleep 30`,
			}
			options.HostConfig.PortBindings = nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "18080"}},
			}

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			command.Run(opt, "start", testContainerName)

			// Give httpd a moment to start.
			time.Sleep(2 * time.Second)

			// Verify the published port is reachable from the host.
			// This confirms the nerdctl/ports label was written and the OCI hook processed it.
			resp, err := http.Get("http://127.0.0.1:18080/index.html") //nolint:noctx
			Expect(err).Should(BeNil(), fmt.Sprintf("published port 18080 not reachable: container ID %s", ctr.ID))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			_ = resp.Body.Close()
		})
	})
}
