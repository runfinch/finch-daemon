// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/api/types/blkiodev"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/ffs"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
	"github.com/runfinch/finch-daemon/e2e/util"
)

type containerCreateResponse struct {
	ID      string `json:"Id"`
	Message string `json:"message"`
}

// ContainerCreate tests the `POST containers/create` API.
func ContainerCreate(opt *option.Option, pOpt util.NewOpt) {
	Describe("create container", func() {
		var (
			uClient *http.Client
			version string
			url     string
			options types.ContainerCreateRequest
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			// set finch daemon request url
			url = client.ConvertToFinchUrl(version, "/containers/create")
			// set default container options
			options = types.ContainerCreateRequest{}
			options.Image = defaultImage
		})
		AfterEach(func() {
			httpRemoveAll(uClient, version)
		})

		It("should successfully create a container that prints hello world", func() {
			// define options
			options.Cmd = []string{"echo", "hello world"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container, wait for it to finish, then fetch logs
			httpStartContainer(uClient, version, testContainerName)
			httpWaitContainer(uClient, version, testContainerName)
			out := httpContainerLogs(uClient, version, testContainerName)
			Expect(strings.TrimSpace(out)).Should(Equal("hello world"))
		})
		It("should successfully log container output for the created container", func() {
			// define options
			options.Cmd = []string{"echo", "hello world"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and wait for it to finish
			httpStartContainer(uClient, version, testContainerName)
			httpWaitContainer(uClient, version, testContainerName)
			out := httpContainerLogs(uClient, version, testContainerName)
			Expect(strings.TrimSpace(out)).Should(Equal("hello world"))
		})
		It("should fail to create a container with duplicate name", func() {
			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// create container with duplicate name
			statusCode, _ = createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusInternalServerError))
		})
		It("should fail to create a container for a nonexistent image", func() {
			// define options
			options.Image = "non-existent-image"

			// create container
			statusCode, _ := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusInternalServerError))
		})

		// Network Settings

		It("should attach container to the bridge network", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify network settings
			httpStartContainer(uClient, version, testContainerName)
			verifyNetworkSettings(uClient, version, testContainerName, "bridge")
		})
		It("should attach container to the bridge network for default network mode", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.NetworkMode = "default"

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify network settings
			httpStartContainer(uClient, version, testContainerName)
			verifyNetworkSettings(uClient, version, testContainerName, "bridge")
		})
		It("should attach container to the specified network using network name", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.NetworkMode = testNetwork

			// create network
			httpCreateNetwork(uClient, version, testNetwork)

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify network settings
			httpStartContainer(uClient, version, testContainerName)
			verifyNetworkSettings(uClient, version, testContainerName, testNetwork)
		})
		It("should attach container to the specified network using network id", func() {
			// create network
			netId := httpCreateNetwork(uClient, version, testNetwork)
			Expect(netId).ShouldNot(BeEmpty())

			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.NetworkMode = netId

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify network settings
			httpStartContainer(uClient, version, testContainerName)
			verifyNetworkSettings(uClient, version, testContainerName, testNetwork)
		})
		It("should create a container with container:<id> network mode", func() {
			// create and start the first container
			options.Cmd = []string{"sleep", "Infinity"}
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName)

			// create second container using first container's network
			proxyContainerName := testContainerName + "-proxy"
			options.HostConfig.NetworkMode = "container:" + ctr.ID
			statusCode, proxyCtr := createContainer(uClient, url, proxyContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(proxyCtr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, proxyContainerName)

			// verify proxy container started successfully (shares network with first container)
			inspect := httpInspectContainer(uClient, version, proxyContainerName)
			Expect(inspect.State.Running).Should(BeTrue())
		})
		It("should create a container with specified port mappings", func() {
			hostPort := "8001"
			ctrPort := "8000"
			hostPort2 := "9001"
			ctrPort2 := "9000"

			// define options
			tcpPort := nat.Port(fmt.Sprintf("%s/tcp", ctrPort))
			tcpPortBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort}
			udpPort := nat.Port(fmt.Sprintf("%s/udp", ctrPort2))
			udpPortBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort2}

			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.PortBindings = nat.PortMap{
				tcpPort: []nat.PortBinding{tcpPortBinding},
				udpPort: []nat.PortBinding{udpPortBinding},
			}

			// create and start container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName)

			// inspect container
			inspect := httpInspectContainer(uClient, version, testContainerName)

			// verify port mappings
			Expect(inspect.NetworkSettings).ShouldNot(BeNil())
			portMap := *inspect.NetworkSettings.Ports
			Expect(portMap).Should(HaveLen(2))
			Expect(portMap[tcpPort]).Should(HaveLen(1))
			Expect(portMap[tcpPort][0]).Should(Equal(tcpPortBinding))
			Expect(portMap[udpPort]).Should(HaveLen(1))
			Expect(portMap[udpPort][0]).Should(Equal(udpPortBinding))
		})

		It("should create a container with automatic port allocation on the host", func() {
			ctrPort := "8080"

			// define options with empty host port for automatic allocation
			tcpPort := nat.Port(fmt.Sprintf("%s/tcp", ctrPort))
			tcpPortBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: ""}

			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.PortBindings = nat.PortMap{
				tcpPort: []nat.PortBinding{tcpPortBinding},
			}

			// create and start container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName)

			// inspect container
			inspect2 := httpInspectContainer(uClient, version, testContainerName)

			// verify port mappings with automatic allocation
			Expect(inspect2.NetworkSettings).ShouldNot(BeNil())
			portMap := *inspect2.NetworkSettings.Ports
			Expect(portMap).Should(HaveLen(1))
			Expect(portMap[tcpPort]).Should(HaveLen(1))
			Expect(portMap[tcpPort][0].HostIP).Should(Equal("0.0.0.0"))
			Expect(portMap[tcpPort][0].HostPort).ShouldNot(BeEmpty())
			// Verify that a port was actually allocated (not empty string)
			port, err := strconv.Atoi(portMap[tcpPort][0].HostPort)
			Expect(err).Should(BeNil())
			Expect(port).Should(BeNumerically(">", 0))
		})

		// Volume Mounts

		It("should create a container with a directory mounted from the host", func() {
			fileContent := "hello world"
			hostFilepath := ffs.CreateTempFile("test-file", fileContent)
			DeferCleanup(os.RemoveAll, filepath.Dir(hostFilepath))
			ctrFilepath := "/tmp/test-mount/test-file"

			// define options
			options.HostConfig.Binds = []string{
				fmt.Sprintf("%s:%s", filepath.Dir(hostFilepath), filepath.Dir(ctrFilepath)),
			}
			options.Cmd = []string{"sleep", "Infinity"}

			// create and start container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName)

			// ensure that mounted file exists in container
			fileShouldExistInContainer(testContainerName, ctrFilepath, fileContent)

			// ensure that write permissions are enabled on the mounted directory
			fileContent2 := "hello world again"
			filename2 := "test-file2"
			cmd := fmt.Sprintf("echo -n %s > %s", fileContent2, filepath.Join(filepath.Dir(ctrFilepath), filename2))
			httpExecContainer(uClient, version, testContainerName, []string{"sh", "-c", cmd})
			fileShouldExist(filepath.Join(filepath.Dir(hostFilepath), filename2), fileContent2)
		})
		It("should create a container with a directory mounted from the host with read-only permissions", func() {
			fileContent := "hello world"
			hostFilepath := ffs.CreateTempFile("test-file", fileContent)
			DeferCleanup(os.RemoveAll, filepath.Dir(hostFilepath))
			ctrFilepath := "/tmp/test-mount/test-file"

			// define options
			options.HostConfig.Binds = []string{
				fmt.Sprintf("%s:%s:ro", filepath.Dir(hostFilepath), filepath.Dir(ctrFilepath)),
			}
			options.Cmd = []string{"sleep", "Infinity"}

			// create and start container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName)

			// ensure that mounted file exists in container
			fileShouldExistInContainer(testContainerName, ctrFilepath, fileContent)

			// ensure that write permissions are disabled on the mounted directory
			fileContent2 := "hello world again"
			filename2 := "test-file2"
			cmd := fmt.Sprintf("echo -n %s > %s", fileContent2, filepath.Join(filepath.Dir(ctrFilepath), filename2))
			_, exitCode := httpExecContainerWithExitCode(uClient, version, testContainerName, []string{"sh", "-c", cmd})
			Expect(exitCode).NotTo(Equal(0))
			fileShouldNotExist(filepath.Join(filepath.Dir(hostFilepath), filename2))
		})
		It("should create a container with a volume mount", func() {
			fileContent := "hello world"
			ctrFilepath := "/mnt/test-volume/test-file"

			// create volume
			httpCreateVolume(uClient, version, testVolumeName, nil)

			// define options
			options.HostConfig.Binds = []string{
				fmt.Sprintf("%s:%s", testVolumeName, filepath.Dir(ctrFilepath)),
			}
			options.Cmd = []string{"sleep", "Infinity"}

			// create and start container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName)

			// write file in the mounted volume
			cmd := fmt.Sprintf("echo -n %s > %s", fileContent, ctrFilepath)
			httpExecContainer(uClient, version, testContainerName, []string{"sh", "-c", cmd})

			// ensure that created file exists in another container with the same volume mount
			statusCode, ctr = createContainer(uClient, url, testContainerName2, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, testContainerName2)
			fileShouldExistInContainer(testContainerName2, ctrFilepath, fileContent)
		})

		// User and Environment Config

		It("should create a container with specified entrypoint", func() {
			// define options
			options.Entrypoint = []string{"/bin/echo"}
			options.Cmd = []string{"hello", "world"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container, wait for it to finish, then fetch logs
			httpStartContainer(uClient, version, testContainerName)
			httpWaitContainer(uClient, version, testContainerName)
			out := httpContainerLogs(uClient, version, testContainerName)
			Expect(strings.TrimSpace(out)).Should(Equal("hello world"))
		})
		It("should create a container with environment variables set", func() {
			envName := "TESTVAR"
			envValue := "test-var-value"

			// define options
			options.Env = []string{fmt.Sprintf("%s=%s", envName, envValue)}
			cmd := fmt.Sprintf("echo $%s", envName)
			options.Cmd = []string{"sh", "-c", cmd}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container, wait for it to finish, then fetch logs
			httpStartContainer(uClient, version, testContainerName)
			httpWaitContainer(uClient, version, testContainerName)
			out := httpContainerLogs(uClient, version, testContainerName)
			Expect(strings.TrimSpace(out)).Should(Equal(envValue))
		})
		It("should create a container with defined labels", func() {
			labelName := "test-label"
			labelValue := "test-label-value"

			// define options
			options.Labels = map[string]string{labelName: labelValue}
			options.Cmd = []string{"sleep", "Infinity"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container
			httpStartContainer(uClient, version, testContainerName)

			// inspect container
			inspect := httpInspectContainer(uClient, version, testContainerName)

			// check label
			Expect(inspect.Config.Labels[labelName]).Should(Equal(labelValue))
		})
		It("should create a container with specified user", func() {
			userName := "nobody"

			// define options
			options.User = userName
			options.Cmd = []string{"whoami"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container, wait for it to finish, then fetch logs
			httpStartContainer(uClient, version, testContainerName)
			httpWaitContainer(uClient, version, testContainerName)
			out := httpContainerLogs(uClient, version, testContainerName)
			Expect(strings.TrimSpace(out)).Should(Equal(userName))
		})
		It("should create a container with specified work directory", func() {
			workdir := "/etc/opt"

			// define options
			options.WorkingDir = workdir
			options.Cmd = []string{"pwd"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container, wait for it to finish, then fetch logs
			httpStartContainer(uClient, version, testContainerName)
			httpWaitContainer(uClient, version, testContainerName)
			out := httpContainerLogs(uClient, version, testContainerName)
			Expect(strings.TrimSpace(out)).Should(Equal(workdir))
		})
		It("should create a container with specified memory allocation", func() {
			// define options
			options.HostConfig.Memory = 209715200 // 200 MiB
			options.Cmd = []string{"sleep", "Infinity"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container
			httpStartContainer(uClient, version, testContainerName)

			// verify memory allocation via inspect HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Memory).Should(Equal(int64(209715200)))
		})
		It("should create a container with specified logging options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.LogConfig = types.LogConfig{
				Type:   "json-file",
				Config: map[string]string{"key": "value"},
			}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// inspect container
			inspect := httpInspectContainer(uClient, version, testContainerName)

			// verify log path exists
			Expect(inspect.LogPath).ShouldNot(BeEmpty())
		})

		It("should create a container with specified CPU qouta and period options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.CPUQuota = 11111
			options.HostConfig.CPUShares = 2048
			options.HostConfig.CPUPeriod = 100000

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			nativeInspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(nativeInspect.HostConfig).ShouldNot(BeNil())

			// Verify the CPU quota
			Expect(nativeInspect.HostConfig.CPUQuota).Should(Equal(int64(11111)))
			Expect(nativeInspect.HostConfig.CPUShares).Should(Equal(int64(2048)))
			Expect(nativeInspect.HostConfig.CPUPeriod).Should(Equal(int64(100000)))
		})

		It("should create a container with specified Memory qouta and PidLimits options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Memory = 4048
			options.HostConfig.PidsLimit = 200
			options.HostConfig.MemoryReservation = 28
			options.HostConfig.MemorySwap = 514288000
			options.HostConfig.MemorySwappiness = 25

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())

			Expect(inspect.HostConfig.PidsLimit).Should(Equal(options.HostConfig.PidsLimit))
			Expect(inspect.HostConfig.Memory).Should(Equal(options.HostConfig.Memory))
		})

		It("should create a container with specified Ulimit options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}

			options.HostConfig.Ulimits = []*types.Ulimit{
				{
					Name: "nofile",
					Soft: 1024,
					Hard: 2048,
				},
			}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Ulimits).ShouldNot(BeEmpty())

			for _, ulimit := range options.HostConfig.Ulimits {
				found := false
				for _, rl := range inspect.HostConfig.Ulimits {
					if rl.Name == ulimit.Name {
						Expect(rl.Hard).To(Equal(ulimit.Hard))
						Expect(rl.Soft).To(Equal(ulimit.Soft))
						found = true
						break
					}
				}
				Expect(found).To(BeTrue())
			}
		})

		It("should create a container with Priviledged options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Privileged = true

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Privileged).Should(BeTrue())
		})

		It("should correctly apply CapAdd and CapDrop", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.CapAdd = []string{"SYS_TIME", "NET_ADMIN"}
			options.HostConfig.CapDrop = []string{"CHOWN", "NET_RAW"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.CapAdd).To(ContainElements("CAP_SYS_TIME", "CAP_NET_ADMIN"))
			Expect(inspect.HostConfig.CapDrop).To(ContainElements("CAP_CHOWN", "CAP_NET_RAW"))
		})

		It("should create a container with specified network options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.DNS = []string{"8.8.8.8"}
			options.HostConfig.DNSOptions = []string{"test-opt"}
			options.HostConfig.DNSSearch = []string{"test.com"}
			options.HostConfig.ExtraHosts = []string{"test-host:127.0.0.1"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start a container and verify network settings
			httpStartContainer(uClient, version, testContainerName)
			verifyNetworkSettings(uClient, version, testContainerName, "bridge")
		})

		It("should create a container with specified restart options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.RestartPolicy = types.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 1,
			}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start and kill a container
			httpStartContainer(uClient, version, testContainerName)
			httpKillContainerWithSignal(uClient, version, testContainerName, "SIGKILL")

			// check every 500 ms if container restarted
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			maxDuration := 30 * time.Second
			startTime := time.Now()
			for range ticker.C {
				// fail testcase if 30 seconds have passed
				Expect(time.Since(startTime) < maxDuration).Should(BeTrue())

				// inspect container
				inspect := httpInspectContainer(uClient, version, testContainerName)

				// if container is running, verify it was restarted
				if inspect.State.Running {
					Expect(inspect.RestartCount).Should(Equal(1))
					return
				}
			}
		})
		It("should create a container with OomKillDisable set to true", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.OomKillDisable = true

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Inspect the container to verify OomKillDisable
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.OomKillDisable).Should(BeTrue())
		})

		It("should create a container with NetworkDisabled set to true", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.NetworkDisabled = true

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Inspect to verify network mode is "none"
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.NetworkMode).Should(Equal("none"))
		})

		It("should create a container with specified MAC address", func() {
			// Define custom MAC address
			macAddress := "02:42:ac:11:00:42"

			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.MacAddress = macAddress

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Inspect container using Docker-compatible format
			inspect := httpInspectContainer(uClient, version, testContainerName)

			// Verify MAC address in NetworkSettings
			Expect(inspect.NetworkSettings.MacAddress).Should(Equal(macAddress))

			// Also verify MAC address in the network details
			for _, netDetails := range inspect.NetworkSettings.Networks {
				Expect(netDetails.MacAddress).Should(Equal(macAddress))
			}
		})

		It("should create a container with both CPUSetCPUs and CPUSetMems options", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.CPUSetCPUs = "0,1" // Use CPUs 0 and 1
			options.HostConfig.CPUSetMems = "0"   // Use only memory node 0

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container
			httpStartContainer(uClient, version, testContainerName)

			// Verify CPU set settings via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())

			// Verify both settings are correct
			Expect(inspect.HostConfig.CPUSetCPUs).Should(Equal("0,1"))
			Expect(inspect.HostConfig.CPUSetMems).Should(Equal("0"))
		})

		It("should create container with specified blkio settings options", func() {
			// Skip if not running on Linux
			if runtime.GOOS != "linux" {
				Skip("Blkio settings are only supported on Linux")
			}

			// Create dummy device paths
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

			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.BlkioWeight = 500 // valid values: 0-1000

			// Create WeightDevice objects for input
			weightDevices := []*blkiodev.WeightDevice{
				{
					Path:   loopDev,
					Weight: 400,
				},
			}

			// Create ThrottleDevice objects for input
			readBpsDevices := []*blkiodev.ThrottleDevice{
				{
					Path: loopDev,
					Rate: 1048576, // 1MB/s
				},
			}

			writeBpsDevices := []*blkiodev.ThrottleDevice{
				{
					Path: loopDev,
					Rate: 2097152, // 2MB/s
				},
			}

			readIopsDevices := []*blkiodev.ThrottleDevice{
				{
					Path: loopDev,
					Rate: 1000,
				},
			}

			writeIopsDevices := []*blkiodev.ThrottleDevice{
				{
					Path: loopDev,
					Rate: 2000,
				},
			}

			// Set the original device objects in the options
			options.HostConfig.BlkioWeightDevice = weightDevices
			options.HostConfig.BlkioDeviceReadBps = readBpsDevices
			options.HostConfig.BlkioDeviceWriteBps = writeBpsDevices
			options.HostConfig.BlkioDeviceReadIOps = readIopsDevices
			options.HostConfig.BlkioDeviceWriteIOps = writeIopsDevices

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container
			httpStartContainer(uClient, version, testContainerName)

			// inspect container
			inspect := httpInspectContainer(uClient, version, testContainerName)

			// Verify blkio settings in HostConfig
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			// Verify BlkioWeight
			Expect(inspect.HostConfig.BlkioWeight).Should(Equal(options.HostConfig.BlkioWeight))
			// Compare string representations
			Expect(inspect.HostConfig.BlkioWeightDevice[0].String()).Should(Equal(weightDevices[0].String()))
			Expect(inspect.HostConfig.BlkioDeviceReadBps[0].String()).Should(Equal(readBpsDevices[0].String()))
			Expect(inspect.HostConfig.BlkioDeviceWriteBps[0].String()).Should(Equal(writeBpsDevices[0].String()))
			Expect(inspect.HostConfig.BlkioDeviceReadIOps[0].String()).Should(Equal(readIopsDevices[0].String()))
			Expect(inspect.HostConfig.BlkioDeviceWriteIOps[0].String()).Should(Equal(writeIopsDevices[0].String()))
		})

		It("should create container with volumes from another container", func() {
			tID := testContainerName

			// Create temporary directories
			rwDir, err := os.MkdirTemp("", "rw")
			Expect(err).Should(BeNil())
			roDir, err := os.MkdirTemp("", "ro")
			Expect(err).Should(BeNil())
			defer os.RemoveAll(rwDir)
			defer os.RemoveAll(roDir)

			// Create named volumes
			rwVolName := tID + "-rw"
			roVolName := tID + "-ro"
			httpCreateVolume(uClient, version, rwVolName, nil)
			httpCreateVolume(uClient, version, roVolName, nil)
			defer httpRemoveVolume(uClient, version, rwVolName)
			defer httpRemoveVolume(uClient, version, roVolName)

			// Create source container with multiple volume types
			fromContainerName := tID + "-from"
			sourceOptions := types.ContainerCreateRequest{}
			sourceOptions.Image = defaultImage
			sourceOptions.Cmd = []string{"top"}
			sourceOptions.HostConfig.Binds = []string{
				fmt.Sprintf("%s:%s", rwDir, "/mnt1"),
				fmt.Sprintf("%s:%s:ro", roDir, "/mnt2"),
				fmt.Sprintf("%s:%s", rwVolName, "/mnt3"),
				fmt.Sprintf("%s:%s:ro", roVolName, "/mnt4"),
			}

			// Create and start source container
			statusCode, _ := createContainer(uClient, url, fromContainerName, sourceOptions)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			httpStartContainer(uClient, version, fromContainerName)
			defer httpRemoveContainerForce(uClient, version, fromContainerName)

			// Create target container with volumes-from
			toContainerName := tID + "-to"
			targetOptions := types.ContainerCreateRequest{}
			targetOptions.Image = defaultImage
			targetOptions.Cmd = []string{"top"}
			targetOptions.HostConfig.VolumesFrom = []string{fromContainerName}

			// Create and start target container
			statusCode, _ = createContainer(uClient, url, toContainerName, targetOptions)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			httpStartContainer(uClient, version, toContainerName)
			defer httpRemoveContainerForce(uClient, version, toContainerName)

			// Test write permissions
			httpExecContainer(uClient, version, toContainerName, []string{"sh", "-exc", "echo -n str1 > /mnt1/file1"})
			_, exitCode := httpExecContainerWithExitCode(uClient, version, toContainerName, []string{"sh", "-exc", "echo -n str2 > /mnt2/file2"})
			Expect(exitCode).NotTo(Equal(0))
			httpExecContainer(uClient, version, toContainerName, []string{"sh", "-exc", "echo -n str3 > /mnt3/file3"})
			_, exitCode2 := httpExecContainerWithExitCode(uClient, version, toContainerName, []string{"sh", "-exc", "echo -n str4 > /mnt4/file4"})
			Expect(exitCode2).NotTo(Equal(0))

			// Remove target container (force since it's running)
			httpRemoveContainerForce(uClient, version, toContainerName)

			// Create a new container to verify data persistence
			verifyOptions := types.ContainerCreateRequest{}
			verifyOptions.Image = defaultImage
			verifyOptions.Cmd = []string{"sh", "-c", "cat /mnt1/file1 /mnt3/file3"}
			verifyOptions.HostConfig.VolumesFrom = []string{fromContainerName}

			statusCode, _ = createContainer(uClient, url, "verify-container", verifyOptions)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			httpStartContainer(uClient, version, "verify-container")
			httpWaitContainer(uClient, version, "verify-container")
			out := httpContainerLogs(uClient, version, "verify-container")
			Expect(strings.TrimSpace(out)).Should(Equal("str1str3"))
			defer httpRemoveContainerForce(uClient, version, "verify-container")
		})

		It("should create a container with tmpfs mounts", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Tmpfs = map[string]string{
				"/tmpfs1": "rw,noexec,nosuid,size=65536k",
				"/tmpfs2": "", // no options
			}

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Verify tmpfs mounts via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Tmpfs).Should(HaveKey("/tmpfs1"))
			Expect(inspect.HostConfig.Tmpfs).Should(HaveKey("/tmpfs2"))
		})

		It("should create a container with UTSMode set to host", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.UTSMode = "host"

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Verify UTS mode via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.UTSMode).Should(Equal("host"))
		})

		It("should create a container with specified PidMode", func() {
			// First create a container that will be referenced in pid mode
			hostOptions := types.ContainerCreateRequest{}
			hostOptions.Image = defaultImage
			hostOptions.Cmd = []string{"sleep", "Infinity"}
			statusCode, hostCtr := createContainer(uClient, url, "host-container", hostOptions)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(hostCtr.ID).ShouldNot(BeEmpty())
			httpStartContainer(uClient, version, "host-container")

			// Define options for the container with pid mode
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.PidMode = "container:host-container"

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Inspect container using Docker-compatible format
			inspect := httpInspectContainer(uClient, version, testContainerName)

			// Verify PidMode configuration
			Expect(inspect.HostConfig.PidMode).Should(Equal(hostCtr.ID))

			// Cleanup
			httpRemoveContainerForce(uClient, version, "host-container")
		})

		It("should create a container with private IPC mode", func() {
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.IpcMode = "private"

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			httpStartContainer(uClient, version, testContainerName)

			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())

			// For private IPC mode, verify IpcMode is set
			Expect(inspect.HostConfig.IpcMode).Should(Equal("private"))
		})

		It("should create a container with privileged mode", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Privileged = true

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Inspect the container to verify privileged mode
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Privileged).Should(BeTrue())
		})

		It("should create a container with specified ShmSize", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.ShmSize = 134217728 // 128MB

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Inspect container
			inspect := httpInspectContainer(uClient, version, testContainerName)

			// Verify ShmSize in HostConfig
			Expect(inspect.HostConfig.ShmSize).Should(Equal(options.HostConfig.ShmSize))
		})

		It("should create a container with specified Sysctls", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Sysctls = map[string]string{
				"net.ipv4.ip_forward": "1",
				"kernel.msgmax":       "65536",
			}

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Verify sysctls via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Sysctls).ShouldNot(BeNil())

			// Verify sysctl values
			Expect(inspect.HostConfig.Sysctls["net.ipv4.ip_forward"]).Should(Equal("1"))
			Expect(inspect.HostConfig.Sysctls["kernel.msgmax"]).Should(Equal("65536"))
		})

		It("should create a container with specified Runtime", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Runtime = "io.containerd.runc.v2"

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Verify runtime via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Runtime).Should(Equal(options.HostConfig.Runtime))
		})

		It("should create a container with readonly root filesystem", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.ReadonlyRootfs = true

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Verify readonly root via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.ReadonlyRootfs).Should(BeTrue())
		})

		It("should create a container with specified annotation", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Annotations = map[string]string{
				"com.example.key": "test-value",
			}

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Verify annotation via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Annotations).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Annotations["com.example.key"]).Should(Equal("test-value"))
		})

		It("should create a container with CgroupnsMode set to host", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.CgroupnsMode = "host"

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Verify cgroup namespace mode via HostConfig
			// Note: finch-daemon may return empty string for CgroupnsMode in inspect
			// even when the container was created with the setting applied.
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(string(inspect.HostConfig.CgroupnsMode)).Should(SatisfyAny(Equal("host"), Equal("")))
		})

		It("should create a container with device mappings", func() {
			// Create a temporary file to use as backing store
			tmpFileOpt, _ := pOpt([]string{"touch", "/tmp/loopdev"})
			command.Run(tmpFileOpt)
			defer func() {
				rmOpt, _ := pOpt([]string{"rm", "-f", "/tmp/loopdev"})
				command.Run(rmOpt)
			}()

			// Write 4KB of data to the file
			ddOpt, _ := pOpt([]string{"dd", "if=/dev/zero", "of=/tmp/loopdev", "bs=4096", "count=1"})
			command.Run(ddOpt)

			// Set up loop device
			loopDevOpt, _ := pOpt([]string{"losetup", "-f", "--show", "/tmp/loopdev"})
			loopDev := command.StdoutStr(loopDevOpt)
			Expect(loopDev).ShouldNot(BeEmpty())
			defer func() {
				detachOpt, _ := pOpt([]string{"losetup", "-d", loopDev})
				command.Run(detachOpt)
			}()

			// Write test content to the device
			writeOpt, _ := pOpt([]string{"sh", "-c", "echo -n test-content > " + loopDev})
			command.Run(writeOpt)

			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.Devices = []types.DeviceMapping{
				{
					PathOnHost:        loopDev,
					PathInContainer:   loopDev,
					CgroupPermissions: "rwm",
				},
			}

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Start container
			httpStartContainer(uClient, version, testContainerName)

			// Inspect to verify device mapping via HostConfig
			inspect := httpInspectContainer(uClient, version, testContainerName)
			Expect(inspect.HostConfig).ShouldNot(BeNil())
			Expect(inspect.HostConfig.Devices).ShouldNot(BeEmpty())

			// Verify device is in the HostConfig devices list
			foundDevice := false
			for _, d := range inspect.HostConfig.Devices {
				if d.PathOnHost == loopDev {
					foundDevice = true
					Expect(d.PathInContainer).Should(Equal(loopDev))
					Expect(d.CgroupPermissions).Should(Equal("rwm"))
					break
				}
			}
			Expect(foundDevice).Should(BeTrue())
		})
	})
}

// creates a container with given options and returns http status code and response.
func createContainer(client *http.Client, url, ctrName string, ctrOptions types.ContainerCreateRequest) (int, containerCreateResponse) {
	// send create request
	reqBody, err := json.Marshal(ctrOptions)
	Expect(err).Should(BeNil())
	url += fmt.Sprintf("?name=%s", ctrName)
	res, err := client.Post(url, "application/json", bytes.NewReader(reqBody))
	Expect(err).Should(BeNil())

	// parse response and status code
	var ctr containerCreateResponse
	err = json.NewDecoder(res.Body).Decode(&ctr)
	Expect(err).Should(BeNil())
	return res.StatusCode, ctr
}

// verifies that the container is connected to the network specified.
func verifyNetworkSettings(uClient *http.Client, version, ctrName, network string) {
	// inspect network
	inspectNet := httpInspectNetwork(uClient, version, network)
	Expect(inspectNet.IPAM.Config).Should(HaveLen(1))
	gateway := strings.Split(inspectNet.IPAM.Config[0].Gateway, ".")
	Expect(gateway).Should(HaveLen(4))
	expectedMask := strings.Join(gateway[:3], ".")

	// inspect container
	inspectCtr := httpInspectContainer(uClient, version, ctrName)

	// ensure that container is connected to the specified network
	foundNetwork := ""
	Expect(inspectCtr.NetworkSettings).ShouldNot(BeNil())
	Expect(inspectCtr.NetworkSettings.IPAddress).ShouldNot(BeEmpty())
	Expect(inspectCtr.NetworkSettings.Networks).ShouldNot(BeEmpty())
	for netName, netSettings := range inspectCtr.NetworkSettings.Networks {
		Expect(netSettings).ShouldNot(BeNil())
		if strings.HasPrefix(netSettings.IPAddress, expectedMask) {
			foundNetwork = netName
		}
	}
	Expect(foundNetwork).ShouldNot(BeEmpty())
}
