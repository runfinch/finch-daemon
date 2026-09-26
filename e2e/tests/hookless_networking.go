// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"io"
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

// HooklessNetworking tests that CNI networking works when set up inline by
// finch-daemon (via ocihook.Run) instead of via OCI hooks forked by runc.
// All containers are created and started via the HTTP API so they go through
// customStart, which strips hooks and calls setupNetworking directly.
func HooklessNetworking(opt *option.Option) {
	Describe("hookless CNI networking", func() {
		var (
			uClient   *http.Client
			version   string
			createUrl string
		)
		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			createUrl = client.ConvertToFinchUrl(version, "/containers/create")
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should assign an IP address on the bridge network", func() {
			// Create and start a container via HTTP API.
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sleep", "infinity"}

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			startContainer(uClient, version, testContainerName)

			// Inspect the container and verify it has an IP address.
			inspectUrl := fmt.Sprintf("/containers/%s/json", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, inspectUrl))
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			var ctrInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&ctrInspect)
			Expect(err).Should(BeNil())
			Expect(ctrInspect.NetworkSettings).ShouldNot(BeNil())
			Expect(ctrInspect.NetworkSettings.IPAddress).ShouldNot(BeEmpty())
		})

		It("should allow outbound connectivity", func() {
			// wget with a short timeout — exit 0 confirms outbound works.
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"wget", "-q", "--spider", "--timeout=5", "http://example.com"}

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			startContainer(uClient, version, testContainerName)

			// Wait for the container to finish (wget exits quickly).
			waitUrl := fmt.Sprintf("/containers/%s/wait", testContainerName)
			res, err := uClient.Post(client.ConvertToFinchUrl(version, waitUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			var waitResp struct {
				StatusCode int `json:"StatusCode"`
			}
			err = json.NewDecoder(res.Body).Decode(&waitResp)
			Expect(err).Should(BeNil())
			Expect(waitResp.StatusCode).Should(Equal(0), "wget should exit 0 — outbound connectivity failed")
		})

		It("should publish a port and make it reachable from the host", func() {
			options := types.ContainerCreateRequest{}
			options.Image = nginxImage
			options.ExposedPorts = nat.PortSet{
				"80/tcp": struct{}{},
			}
			options.HostConfig.PortBindings = nat.PortMap{
				"80/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "18080"}},
			}

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			startContainer(uClient, version, testContainerName)
			time.Sleep(2 * time.Second) // give nginx time to start

			// Verify the published port is reachable.
			resp, err := http.Get("http://127.0.0.1:18080/") //nolint:noctx // test helper
			Expect(err).Should(BeNil(), "published port 18080 not reachable")
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			resp.Body.Close()
		})

		It("should not assign networking with --network none", func() {
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sleep", "infinity"}
			options.HostConfig.NetworkMode = "none"

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			startContainer(uClient, version, testContainerName)

			// Inspect — should have no IP address.
			inspectUrl := fmt.Sprintf("/containers/%s/json", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, inspectUrl))
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			var ctrInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&ctrInspect)
			Expect(err).Should(BeNil())
			Expect(ctrInspect.NetworkSettings.IPAddress).Should(BeEmpty())
		})

		It("should clean up CNI state after container stop", func() {
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sleep", "infinity"}

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			startContainer(uClient, version, testContainerName)
			time.Sleep(1 * time.Second)

			// Stop the container — postStop watcher should run CNI teardown.
			stopUrl := fmt.Sprintf("/containers/%s/stop", testContainerName)
			res, err := uClient.Post(client.ConvertToFinchUrl(version, stopUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			res.Body.Close()
			time.Sleep(2 * time.Second)

			// Verify indirectly: remove the stopped container, start a new one with
			// the same name — if CNI state leaked, the new container would fail to
			// get an IP or would get a duplicate.
			options2 := types.ContainerCreateRequest{}
			options2.Image = defaultImage
			options2.Cmd = []string{"sleep", "infinity"}

			// Remove the stopped container first.
			removeUrl := fmt.Sprintf("/containers/%s?force=true", testContainerName)
			req, err := http.NewRequest(http.MethodDelete, client.ConvertToFinchUrl(version, removeUrl), nil)
			Expect(err).Should(BeNil())
			res, err = uClient.Do(req)
			Expect(err).Should(BeNil())
			res.Body.Close()

			// Create and start a new container with the same name.
			statusCode2, ctr2 := createContainer(uClient, createUrl, testContainerName, options2)
			Expect(statusCode2).Should(Equal(http.StatusCreated))
			Expect(ctr2.ID).ShouldNot(BeEmpty())

			startContainer(uClient, version, testContainerName)

			// Verify the new container got an IP (proves CNI state was cleaned up).
			inspectUrl := fmt.Sprintf("/containers/%s/json", testContainerName)
			res, err = uClient.Get(client.ConvertToFinchUrl(version, inspectUrl))
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			var ctrInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&ctrInspect)
			Expect(err).Should(BeNil())
			Expect(ctrInspect.NetworkSettings.IPAddress).ShouldNot(BeEmpty())
		})

		It("should assign an IP on a user-defined network", func() {
			// Create a custom network via nerdctl CLI.
			command.Run(opt, "network", "create", testNetwork)

			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sleep", "infinity"}
			options.HostConfig.NetworkMode = testNetwork

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			startContainer(uClient, version, testContainerName)

			// Inspect and verify IP is on the custom network.
			inspectUrl := fmt.Sprintf("/containers/%s/json", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, inspectUrl))
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			var ctrInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&ctrInspect)
			Expect(err).Should(BeNil())
			Expect(ctrInspect.NetworkSettings).ShouldNot(BeNil())
			Expect(ctrInspect.NetworkSettings.Networks).Should(HaveKey(testNetwork))
		})

		It("should allow two containers on the same bridge to communicate", func() {
			// Start a "server" container.
			serverOpts := types.ContainerCreateRequest{}
			serverOpts.Image = defaultImage
			serverOpts.Cmd = []string{"sleep", "infinity"}

			statusCode, serverCtr := createContainer(uClient, createUrl, testContainerName, serverOpts)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(serverCtr.ID).ShouldNot(BeEmpty())
			startContainer(uClient, version, testContainerName)

			// Get the server container's IP.
			inspectUrl := fmt.Sprintf("/containers/%s/json", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, inspectUrl))
			Expect(err).Should(BeNil())
			var serverInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&serverInspect)
			Expect(err).Should(BeNil())
			res.Body.Close()
			serverIP := serverInspect.NetworkSettings.IPAddress
			Expect(serverIP).ShouldNot(BeEmpty())

			// Start a "client" container that pings the server by IP.
			clientOpts := types.ContainerCreateRequest{}
			clientOpts.Image = defaultImage
			clientOpts.Cmd = []string{"ping", "-c", "3", "-W", "5", serverIP}

			statusCode, clientCtr := createContainer(uClient, createUrl, testContainerName2, clientOpts)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(clientCtr.ID).ShouldNot(BeEmpty())
			startContainer(uClient, version, testContainerName2)

			// Wait for the ping container to finish.
			waitUrl := fmt.Sprintf("/containers/%s/wait", testContainerName2)
			res, err = uClient.Post(client.ConvertToFinchUrl(version, waitUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			var waitResp struct {
				StatusCode int `json:"StatusCode"`
			}
			err = json.NewDecoder(res.Body).Decode(&waitResp)
			Expect(err).Should(BeNil())
			Expect(waitResp.StatusCode).Should(Equal(0), "ping should exit 0 — inter-container connectivity failed")
		})

		It("should preserve networking across daemon restart", Serial, func() {
			daemonExe := getFinchDaemonExe()
			socketPath := getSocketPath()

			// Create and start a networked container via HTTP API.
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sleep", "infinity"}

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			startContainer(uClient, version, testContainerName)

			// Record the IP before restart.
			inspectUrl := fmt.Sprintf("/containers/%s/json", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, inspectUrl))
			Expect(err).Should(BeNil())
			var preInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&preInspect)
			Expect(err).Should(BeNil())
			res.Body.Close()
			originalIP := preInspect.NetworkSettings.IPAddress
			Expect(originalIP).ShouldNot(BeEmpty())

			// Kill and restart the daemon.
			killDaemon()
			startDaemon(daemonExe, socketPath)
			waitForDaemon(uClient, version)

			// Container should still be running with the same IP.
			res, err = uClient.Get(client.ConvertToFinchUrl(version, inspectUrl))
			Expect(err).Should(BeNil())
			var postInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&postInspect)
			Expect(err).Should(BeNil())
			res.Body.Close()
			Expect(postInspect.State.Running).Should(BeTrue(), "container should still be running after daemon restart")
			Expect(postInspect.NetworkSettings.IPAddress).Should(Equal(originalIP))

			// Stop the container — exercises the reattached postStop watcher for CNI teardown.
			stopUrl := fmt.Sprintf("/containers/%s/stop", testContainerName)
			res, err = uClient.Post(client.ConvertToFinchUrl(version, stopUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			res.Body.Close()
			time.Sleep(2 * time.Second)

			// Remove the stopped container.
			removeUrl := fmt.Sprintf("/containers/%s?force=true", testContainerName)
			req, err := http.NewRequest(http.MethodDelete, client.ConvertToFinchUrl(version, removeUrl), nil)
			Expect(err).Should(BeNil())
			res, err = uClient.Do(req)
			Expect(err).Should(BeNil())
			res.Body.Close()

			// Create a new container — if CNI state leaked, this would fail or get a stale IP.
			options2 := types.ContainerCreateRequest{}
			options2.Image = defaultImage
			options2.Cmd = []string{"sleep", "infinity"}

			statusCode2, ctr2 := createContainer(uClient, createUrl, testContainerName, options2)
			Expect(statusCode2).Should(Equal(http.StatusCreated))
			Expect(ctr2.ID).ShouldNot(BeEmpty())
			startContainer(uClient, version, testContainerName)

			inspectUrl2 := fmt.Sprintf("/containers/%s/json", testContainerName)
			res, err = uClient.Get(client.ConvertToFinchUrl(version, inspectUrl2))
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			var finalInspect types.Container
			err = json.NewDecoder(res.Body).Decode(&finalInspect)
			Expect(err).Should(BeNil())
			Expect(finalInspect.NetworkSettings.IPAddress).ShouldNot(BeEmpty(),
				"new container should get an IP — CNI teardown after reattach should have cleaned up")
		})
	})
}

// startContainer starts a container via the HTTP API and asserts success.
func startContainer(uClient *http.Client, version, name string) {
	startUrl := fmt.Sprintf("/containers/%s/start", name)
	res, err := uClient.Post(client.ConvertToFinchUrl(version, startUrl), "application/json", nil)
	Expect(err).Should(BeNil())
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	Expect(res.StatusCode).Should(SatisfyAny(
		Equal(http.StatusNoContent),
		Equal(http.StatusNotModified)),
		fmt.Sprintf("start failed: %s", string(body)))
}
