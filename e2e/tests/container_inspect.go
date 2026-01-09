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
	"github.com/runfinch/finch-daemon/e2e/util"
)

// ContainerInspect tests the `GET containers/{id}/json` API.
func ContainerInspect(opt *option.Option, pOpt util.NewOpt) {
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

		It("should return hostconfig with proper values", func() {
			// setup cpu/mem options
			createOptions := types.ContainerCreateRequest{}
			createOptions.Image = defaultImage
			createOptions.Cmd = []string{"sleep", "Infinity"}
			createOptions.HostConfig.ShmSize = 134217728
			createOptions.HostConfig.CPUSetCPUs = "0,1"
			createOptions.HostConfig.CPUSetMems = "0"
			createOptions.HostConfig.CPUShares = 2048
			createOptions.HostConfig.CPUPeriod = 100000
			createOptions.HostConfig.Memory = 134217728
			createOptions.HostConfig.MemorySwap = 514288000

			// setup port mappings
			hostPort := "8001"
			ctrPort := "8000"
			hostPort2 := "9001"
			ctrPort2 := "9000"
			tcpPort := nat.Port(fmt.Sprintf("%s/tcp", ctrPort))
			tcpPortBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort}
			udpPort := nat.Port(fmt.Sprintf("%s/udp", ctrPort2))
			udpPortBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort2}
			createOptions.HostConfig.PortBindings = nat.PortMap{
				tcpPort: []nat.PortBinding{tcpPortBinding},
				udpPort: []nat.PortBinding{udpPortBinding},
			}

			// setup devices
			tmpFileOpt, _ := pOpt([]string{"touch", "/tmp/loopdev"})
			command.Run(tmpFileOpt)
			defer func() {
				rmOpt, _ := pOpt([]string{"rm", "-f", "/tmp/loopdev"})
				command.Run(rmOpt)
			}()
			ddOpt, _ := pOpt([]string{"dd", "if=/dev/zero", "of=/tmp/loopdev", "bs=4096", "count=1"})
			command.Run(ddOpt)
			loopDevOpt, _ := pOpt([]string{"losetup", "-f", "--show", "/tmp/loopdev"})
			loopDev := command.StdoutStr(loopDevOpt)
			Expect(loopDev).ShouldNot(BeEmpty())
			defer func() {
				detachOpt, _ := pOpt([]string{"losetup", "-d", loopDev})
				command.Run(detachOpt)
			}()
			createOptions.HostConfig.Devices = []types.DeviceMapping{
				{
					PathOnHost:        loopDev,
					PathInContainer:   loopDev,
					CgroupPermissions: "rwm",
				},
			}

			// create and start container with options
			containerCreateUrl := client.ConvertToFinchUrl(version, "/containers/create")
			statusCode, ctr := createContainer(uClient, containerCreateUrl, testContainerName2, createOptions)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName2)

			// inspect container
			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json", ctr.ID)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got types.Container
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())

			// verify hostconfig
			Expect(got.HostConfig).ShouldNot(BeNil())
			Expect(got.HostConfig.ContainerIDFile).Should(BeEmpty())
			Expect(got.HostConfig.PidMode).Should(BeEmpty())
			Expect(got.HostConfig.IpcMode).Should(Equal("private"))
			Expect(got.HostConfig.ReadonlyRootfs).Should(BeFalse())
			Expect(got.HostConfig.Devices).Should(HaveLen(1))
			Expect(got.HostConfig.Devices[0]).Should(Equal(createOptions.HostConfig.Devices[0]))
			Expect(got.HostConfig.ShmSize).Should(Equal(createOptions.HostConfig.ShmSize))
			Expect(got.HostConfig.CPUSetCPUs).Should(Equal(createOptions.HostConfig.CPUSetCPUs))
			Expect(got.HostConfig.CPUSetMems).Should(Equal(createOptions.HostConfig.CPUSetMems))
			Expect(got.HostConfig.CPUShares).Should(Equal(createOptions.HostConfig.CPUShares))
			Expect(got.HostConfig.CPUPeriod).Should(Equal(createOptions.HostConfig.CPUPeriod))
			Expect(got.HostConfig.Memory).Should(Equal(createOptions.HostConfig.Memory))
			Expect(got.HostConfig.MemorySwap).Should(Equal(createOptions.HostConfig.MemorySwap))
			Expect(got.HostConfig.PortBindings).Should(HaveLen(2))
			Expect(got.HostConfig.PortBindings[tcpPort]).Should(HaveLen(1))
			Expect(got.HostConfig.PortBindings[tcpPort]).Should(HaveLen(1))
			Expect(got.HostConfig.PortBindings[tcpPort][0]).Should(Equal(tcpPortBinding))
			Expect(got.HostConfig.PortBindings[udpPort]).Should(HaveLen(1))
			Expect(got.HostConfig.PortBindings[udpPort][0]).Should(Equal(udpPortBinding))
			// Sysctls can be empty or contain "net.ipv4.ip_unprivileged_port_start" depending on the environment.
			// See - https://github.com/containerd/nerdctl/blob/53e7b272af14b075a5e8d7b95a5c2d862a1620f8/cmd/nerdctl/container/container_inspect_linux_test.go#L366
			Expect(got.HostConfig.Sysctls).Should(Or(HaveLen(0), HaveLen(1)))
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
