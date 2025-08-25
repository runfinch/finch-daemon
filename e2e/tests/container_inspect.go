// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// ContainerInspect tests the `GET containers/{id}/json` API.
func ContainerInspect(opt *option.Option) {
	Describe("inspect container", func() {
		var (
			uClient           *http.Client
			version           string
			containerId       string
			containerName     string
			wantContainerName string
		)
		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			containerName = testContainerName
			wantContainerName = fmt.Sprintf("/%s", containerName)
			containerId = command.StdoutStr(opt, "run", "-d", "--name", containerName, defaultImage, "sleep", "infinity")
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should inspect the container by ID", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json", containerId)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got types.Container
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(got.ID).Should(Equal(containerId))
			Expect(got.Name).Should(Equal(wantContainerName))
			Expect(got.State.Status).Should(Equal("running"))
		})

		It("should inspect the container by name", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json", containerName)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got types.Container
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(got.ID).Should(Equal(containerId))
			Expect(got.Name).Should(Equal(wantContainerName))
			Expect(got.State.Status).Should(Equal("running"))
		})

		It("should return size information when size parameter is true", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json?size=1", containerId)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got types.Container
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(got.SizeRw).ShouldNot(BeNil())
			Expect(got.SizeRootFs).ShouldNot(BeNil())
		})

		It("should return hostconfig in inspect response", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json", containerId)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got types.Container
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(got.HostConfig).ShouldNot(BeNil())
			Expect(got.HostConfig.PidMode).Should(BeEmpty())
			Expect(got.HostConfig.IpcMode).Should(BeEmpty())
			Expect(got.HostConfig.ReadonlyRootfs).Should(BeFalse())
		})

		It("should contain port mappings in hostconfig if container is created with one", func() {
			hostPort := "8001"
			ctrPort := "8000"
			hostPort2 := "9001"
			ctrPort2 := "9000"

			tcpPort := nat.Port(fmt.Sprintf("%s/tcp", ctrPort))
			tcpPortBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort}
			udpPort := nat.Port(fmt.Sprintf("%s/udp", ctrPort2))
			udpPortBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort2}

			createOptions := types.ContainerCreateRequest{}
			createOptions.Image = defaultImage
			createOptions.Cmd = []string{"sleep", "Infinity"}
			createOptions.HostConfig.PortBindings = nat.PortMap{
				tcpPort: []nat.PortBinding{tcpPortBinding},
				udpPort: []nat.PortBinding{udpPortBinding},
			}

			// create and start container with port mapping
			containerCreateUrl := client.ConvertToFinchUrl(version, "/containers/create")
			portMapContainerName := testContainerName + "-portmap"
			statusCode, ctr := createContainer(uClient, containerCreateUrl, portMapContainerName, createOptions)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			command.Run(opt, "start", portMapContainerName)

			// inspect container
			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json", ctr.ID)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got types.Container
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())

			// verify port mappings
			Expect(got.HostConfig).ShouldNot(BeNil())
			Expect(got.HostConfig.PortBindings).ShouldNot(BeNil())
			Expect(got.HostConfig.PortBindings).Should(HaveLen(2))
			Expect(got.HostConfig.PortBindings[tcpPort]).Should(HaveLen(1))
			Expect(got.HostConfig.PortBindings[tcpPort]).Should(HaveLen(1))
			Expect(got.HostConfig.PortBindings[tcpPort][0]).Should(Equal(tcpPortBinding))
			Expect(got.HostConfig.PortBindings[udpPort]).Should(HaveLen(1))
			Expect(got.HostConfig.PortBindings[udpPort][0]).Should(Equal(udpPortBinding))
		})

		It("should return 404 error when container does not exist", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/nonexistent/json"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			Expect(body).Should(MatchJSON(`{"message": "no such container: nonexistent"}`))
		})
	})
}
