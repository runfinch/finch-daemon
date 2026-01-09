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

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
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
			command.RemoveAll(opt)
		})

		It("should successfully create a container that prints hello world", func() {
			// define options
			options.Cmd = []string{"echo", "hello world"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify output
			out := command.StdoutStr(opt, "start", "-a", testContainerName)
			Expect(out).Should(Equal("hello world"))
		})
		It("should successfully log container output for the created container", func() {
			// define options
			options.Cmd = []string{"echo", "hello world"}

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify output
			httpStartContainer(uClient, version, testContainerName)
			out := command.StdoutStr(opt, "logs", testContainerName)
			Expect(out).Should(Equal("hello world"))
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
			verifyNetworkSettings(opt, testContainerName, "bridge")
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
			verifyNetworkSettings(opt, testContainerName, "bridge")
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
			verifyNetworkSettings(opt, testContainerName, testNetwork)
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
			verifyNetworkSettings(opt, testContainerName, testNetwork)
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
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// verify port mappings
			Expect(inspect[0].NetworkSettings).ShouldNot(BeNil())
			portMap := *inspect[0].NetworkSettings.Ports
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
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// verify port mappings with automatic allocation
			Expect(inspect[0].NetworkSettings).ShouldNot(BeNil())
			portMap := *inspect[0].NetworkSettings.Ports
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
			fileShouldExistInContainer(opt, testContainerName, ctrFilepath, fileContent)

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
			fileShouldExistInContainer(opt, testContainerName, ctrFilepath, fileContent)

			// ensure that write permissions are disabled on the mounted directory
			fileContent2 := "hello world again"
			filename2 := "test-file2"
			cmd := fmt.Sprintf("echo -n %s > %s", fileContent2, filepath.Join(filepath.Dir(ctrFilepath), filename2))
			command.RunWithoutSuccessfulExit(opt, "exec", testContainerName, "sh", "-c", cmd)
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
			fileShouldExistInContainer(opt, testContainerName2, ctrFilepath, fileContent)
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

			// start container and verify entrypoint output
			out := command.StdoutStr(opt, "start", "-a", testContainerName)
			Expect(out).Should(Equal("hello world"))
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

			// start container and verify output
			out := command.StdoutStr(opt, "start", "-a", testContainerName)
			Expect(out).Should(Equal(envValue))
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
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// check label
			Expect(inspect[0].Config.Labels[labelName]).Should(Equal(labelValue))
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

			// start container and verify user
			out := command.StdoutStr(opt, "start", "-a", testContainerName)
			Expect(out).Should(Equal(userName))
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

			// start container and verify current directory
			out := command.StdoutStr(opt, "start", "-a", testContainerName)
			Expect(out).Should(Equal(workdir))
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

			// verify memory allocation from stats command
			resp := command.StdoutStr(opt, "stats", "--no-stream", "--format", "'{{ json .}}'", testContainerName)
			var stats map[string]string
			err := json.Unmarshal([]byte(strings.Trim(resp, "'")), &stats)
			Expect(err).Should(BeNil())
			Expect(stats).Should(HaveKey("MemUsage"))

			memAloc := strings.Split(stats["MemUsage"], " / ")[1]
			Expect(memAloc).Should(Equal("200MiB"))
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
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// verify log path exists
			Expect(inspect[0].LogPath).ShouldNot(BeNil())
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

			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the CPU quota value
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			resources, ok := linux["resources"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			cpu, ok := resources["cpu"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			quota, ok := cpu["quota"].(float64)
			Expect(ok).Should(BeTrue())
			period, ok := cpu["period"].(float64)
			Expect(ok).Should(BeTrue())
			shares, ok := cpu["shares"].(float64)
			Expect(ok).Should(BeTrue())

			// Verify the CPU quota
			Expect(int64(quota)).Should(Equal(int64(11111)))
			Expect(int64(shares)).Should(Equal(int64(2048)))
			Expect(int64(period)).Should(Equal(int64(100000)))
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

			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the CPU quota value
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			resources, ok := linux["resources"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			memory, _ := resources["memory"].(map[string]interface{})

			pids, _ := resources["pids"].(map[string]interface{})

			Expect(int64(pids["limit"].(float64))).Should(Equal(options.HostConfig.PidsLimit))
			Expect(int64(memory["limit"].(float64))).Should(Equal(options.HostConfig.Memory))
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

			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the CPU quota value
			spec, _ := nativeInspect[0]["Spec"].(map[string]interface{})
			rlimits := spec["process"].(map[string]interface{})["rlimits"].([]interface{})
			for _, ulimit := range options.HostConfig.Ulimits {
				found := false
				for _, rlimit := range rlimits {
					r := rlimit.(map[string]interface{})
					if r["type"] == "RLIMIT_NOFILE" {
						Expect(r["hard"]).To(Equal(float64(ulimit.Hard)))
						Expect(r["soft"]).To(Equal(float64(ulimit.Soft)))
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

			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the CPU quota value
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			capabilities := spec["process"].(map[string]interface{})["capabilities"].(map[string]interface{})
			Expect(capabilities["bounding"]).To(ContainElements("CAP_SYS_ADMIN", "CAP_NET_ADMIN", "CAP_SYS_MODULE"))
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

			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the CPU quota value
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			capabilities := spec["process"].(map[string]interface{})["capabilities"].(map[string]interface{})
			Expect(capabilities["bounding"]).To(ContainElements("CAP_SYS_TIME", "CAP_NET_ADMIN"))
			Expect(capabilities["bounding"]).NotTo(ContainElements("CAP_CHOWN", "CAP_NET_RAW"))
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
			verifyNetworkSettings(opt, testContainerName, "bridge")
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
				resp := command.Stdout(opt, "inspect", testContainerName)
				var inspect []*dockercompat.Container
				err := json.Unmarshal(resp, &inspect)
				Expect(err).Should(BeNil())
				Expect(inspect).Should(HaveLen(1))

				// if container is running, verify it was restarted
				if inspect[0].State.Running {
					Expect(inspect[0].RestartCount).Should(Equal(1))
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

			// Inspect the container using native format to verify OomKillDisable
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the linux resources memory section
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			resources, ok := linux["resources"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			memory, ok := resources["memory"].(map[string]interface{})
			Expect(ok).Should(BeTrue())

			// Verify OomKillDisable is set
			oomKillDisable, ok := memory["disableOOMKiller"].(bool)
			Expect(ok).Should(BeTrue())
			Expect(oomKillDisable).Should(BeTrue())
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

			// Inspect using the native format to verify network mode is "none"
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Check that network is set to "none" in nerdctl/networks label
			labels, ok := nativeInspect[0]["Labels"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			networks, ok := labels["nerdctl/networks"].(string)
			Expect(ok).Should(BeTrue())
			Expect(networks).Should(ContainSubstring(`"none"`))
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
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// Verify MAC address in NetworkSettings
			Expect(inspect[0].NetworkSettings.MacAddress).Should(Equal(macAddress))

			// Also verify MAC address in the network details
			for _, netDetails := range inspect[0].NetworkSettings.Networks {
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

			// Get native container configuration
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the CPU settings
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			resources, ok := linux["resources"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			cpu, ok := resources["cpu"].(map[string]interface{})
			Expect(ok).Should(BeTrue())

			// Verify both settings are correct
			cpuSet, ok := cpu["cpus"].(string)
			Expect(ok).Should(BeTrue())
			Expect(cpuSet).Should(Equal("0,1"))

			memSet, ok := cpu["mems"].(string)
			Expect(ok).Should(BeTrue())
			Expect(memSet).Should(Equal("0"))
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
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// Verify blkio settings in LinuxBlkioSettings
			blkioSettings := inspect[0].HostConfig.BlkioSettings
			// Verify BlkioWeight
			Expect(blkioSettings.BlkioWeight).Should(Equal(options.HostConfig.BlkioWeight))
			// Compare string representations
			Expect(blkioSettings.BlkioWeightDevice[0].String()).Should(Equal(weightDevices[0].String()))
			Expect(blkioSettings.BlkioDeviceReadBps[0].String()).Should(Equal(readBpsDevices[0].String()))
			Expect(blkioSettings.BlkioDeviceWriteBps[0].String()).Should(Equal(writeBpsDevices[0].String()))
			Expect(blkioSettings.BlkioDeviceReadIOps[0].String()).Should(Equal(readIopsDevices[0].String()))
			Expect(blkioSettings.BlkioDeviceWriteIOps[0].String()).Should(Equal(writeIopsDevices[0].String()))
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
			defer httpRemoveContainer(uClient, version, fromContainerName)

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
			defer httpRemoveContainer(uClient, version, toContainerName)

			// Test write permissions
			httpExecContainer(uClient, version, toContainerName, []string{"sh", "-exc", "echo -n str1 > /mnt1/file1"})
			command.RunWithoutSuccessfulExit(opt, "exec", toContainerName, "sh", "-exc", "echo -n str2 > /mnt2/file2")
			httpExecContainer(uClient, version, toContainerName, []string{"sh", "-exc", "echo -n str3 > /mnt3/file3"})
			command.RunWithoutSuccessfulExit(opt, "exec", toContainerName, "sh", "-exc", "echo -n str4 > /mnt4/file4")

			// Remove target container
			httpRemoveContainer(uClient, version, toContainerName)

			// Create a new container to verify data persistence
			verifyOptions := types.ContainerCreateRequest{}
			verifyOptions.Image = defaultImage
			verifyOptions.Cmd = []string{"sh", "-c", "cat /mnt1/file1 /mnt3/file3"}
			verifyOptions.HostConfig.VolumesFrom = []string{fromContainerName}

			statusCode, _ = createContainer(uClient, url, "verify-container", verifyOptions)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			out := command.StdoutStr(opt, "start", "-a", "verify-container")
			Expect(out).Should(Equal("str1str3"))
			defer httpRemoveContainer(uClient, version, "verify-container")
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

			// Verify tmpfs mounts using native inspect
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the mounts section
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			mounts, ok := spec["mounts"].([]interface{})
			Expect(ok).Should(BeTrue())

			// Verify tmpfs mounts
			foundMounts := make(map[string]bool)
			for _, mount := range mounts {
				m := mount.(map[string]interface{})
				if m["type"] == "tmpfs" {
					foundMounts[m["destination"].(string)] = true
					if m["destination"] == "/tmpfs1" {
						options := m["options"].([]interface{})
						optionsStr := make([]string, len(options))
						for i, opt := range options {
							optionsStr[i] = opt.(string)
						}
						Expect(optionsStr).Should(ContainElements(
							"rw",
							"noexec",
							"nosuid",
							"size=65536k",
						))
					}
				}
			}
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

			// Inspect using native format to verify UTS namespace configuration
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the namespaces section
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			namespaces, ok := linux["namespaces"].([]interface{})
			Expect(ok).Should(BeTrue())

			// Verify UTS namespace is not present (indicating host namespace is used)
			foundUTSNamespace := false
			for _, ns := range namespaces {
				namespace := ns.(map[string]interface{})
				if namespace["type"] == "uts" {
					foundUTSNamespace = true
					break
				}
			}
			Expect(foundUTSNamespace).Should(BeFalse())
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
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// Verify PidMode configuration
			Expect(inspect[0].HostConfig.PidMode).Should(Equal(hostCtr.ID))

			// Cleanup
			httpRemoveContainer(uClient, version, "host-container")
		})

		It("should create a container with private IPC mode", func() {
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.IpcMode = "private"

			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			httpStartContainer(uClient, version, testContainerName)

			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			namespaces, ok := linux["namespaces"].([]interface{})
			Expect(ok).Should(BeTrue())

			// For private IPC mode, verify IPC namespace is present
			foundIpcNamespace := false
			for _, ns := range namespaces {
				namespace := ns.(map[string]interface{})
				if namespace["type"] == "ipc" {
					foundIpcNamespace = true
					break
				}
			}
			Expect(foundIpcNamespace).Should(BeTrue())
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

			// Inspect the container using native format
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the process capabilities section
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			process, ok := spec["process"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			capabilities, ok := process["capabilities"].(map[string]interface{})
			Expect(ok).Should(BeTrue())

			// Verify privileged capabilities
			// In privileged mode, the container should have extensive capabilities
			expectedCaps := []string{
				"CAP_SYS_ADMIN",
				"CAP_NET_ADMIN",
				"CAP_SYS_MODULE",
			}

			for _, capType := range []string{"bounding", "effective", "permitted"} {
				caps, ok := capabilities[capType].([]interface{})
				Expect(ok).Should(BeTrue())
				capsList := make([]string, len(caps))
				for i, cap := range caps {
					capsList[i] = cap.(string)
				}
				for _, expectedCap := range expectedCaps {
					Expect(capsList).Should(ContainElement(expectedCap))
				}
			}

			// Also verify that devices are allowed in privileged mode
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			resources, ok := linux["resources"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			devices, ok := resources["devices"].([]interface{})
			Expect(ok).Should(BeTrue())

			// In privileged mode, there should be a device rule that allows all devices
			foundAllowAllDevices := false
			for _, device := range devices {
				dev := device.(map[string]interface{})
				if dev["allow"] == true && dev["access"] == "rwm" {
					if _, hasType := dev["type"]; !hasType {
						foundAllowAllDevices = true
						break
					}
				}
			}
			Expect(foundAllowAllDevices).Should(BeTrue())
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

			// Inspect container using Docker-compatible format
			resp := command.Stdout(opt, "inspect", testContainerName)
			var inspect []*dockercompat.Container
			err := json.Unmarshal(resp, &inspect)
			Expect(err).Should(BeNil())
			Expect(inspect).Should(HaveLen(1))

			// Verify ShmSize in HostConfig
			Expect(inspect[0].HostConfig.ShmSize).Should(Equal(options.HostConfig.ShmSize))
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

			// Verify sysctls using native inspect
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the sysctls section
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			sysctls, ok := linux["sysctl"].(map[string]interface{})
			Expect(ok).Should(BeTrue())

			// Verify sysctl values
			Expect(sysctls["net.ipv4.ip_forward"]).Should(Equal("1"))
			Expect(sysctls["kernel.msgmax"]).Should(Equal("65536"))
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

			// Verify runtime using native inspect
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the Runtime section
			runtime, ok := nativeInspect[0]["Runtime"].(map[string]interface{})
			Expect(ok).Should(BeTrue())

			// Verify runtime name
			runtimeName, ok := runtime["Name"].(string)
			Expect(ok).Should(BeTrue())
			Expect(runtimeName).Should(Equal(options.HostConfig.Runtime))
		})

		It("should create a container with readonly root filesystem", func() {
			// Define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.ReadonlyRootfs = true

			// Create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// Additional verification through native inspect
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Verify readonly root in the spec
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			root, ok := spec["root"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			readonly, ok := root["readonly"].(bool)
			Expect(ok).Should(BeTrue())
			Expect(readonly).Should(BeTrue())
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

			// Inspect using native format to verify annotation
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Verify annotation in container spec
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			annotations, ok := spec["annotations"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			Expect(annotations["com.example.key"]).Should(Equal("test-value"))
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

			// Inspect using native format to verify cgroup namespace configuration
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the namespaces section
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			namespaces, ok := linux["namespaces"].([]interface{})
			Expect(ok).Should(BeTrue())

			// For host mode, cgroup namespace should not be present in the namespaces list
			foundCgroupNS := false
			for _, ns := range namespaces {
				namespace := ns.(map[string]interface{})
				if namespace["type"] == "cgroup" {
					foundCgroupNS = true
					break
				}
			}
			Expect(foundCgroupNS).Should(BeFalse())
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

			// Get device info to verify major/minor numbers
			statOpt, _ := pOpt([]string{"stat", "-c", "%t,%T", loopDev})
			devNums := command.StdoutStr(statOpt)
			parts := strings.Split(devNums, ",")
			major, _ := strconv.ParseUint(parts[0], 16, 64)
			minor, _ := strconv.ParseUint(parts[1], 16, 64)

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

			// Inspect using native format
			nativeResp := command.Stdout(opt, "inspect", "--mode=native", testContainerName)
			var nativeInspect []map[string]interface{}
			err2 := json.Unmarshal(nativeResp, &nativeInspect)
			Expect(err2).Should(BeNil())
			Expect(nativeInspect).Should(HaveLen(1))

			// Navigate to the linux section
			spec, ok := nativeInspect[0]["Spec"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			linux, ok := spec["linux"].(map[string]interface{})
			Expect(ok).Should(BeTrue())

			// Verify device in linux.devices
			devices, ok := linux["devices"].([]interface{})
			Expect(ok).Should(BeTrue())

			foundDevice := false
			for _, device := range devices {
				d := device.(map[string]interface{})
				if d["path"] == loopDev {
					foundDevice = true
					Expect(d["type"]).Should(Equal("b")) // block device
					Expect(d["major"].(float64)).Should(Equal(float64(major)))
					Expect(d["minor"].(float64)).Should(Equal(float64(minor)))
					break
				}
			}
			Expect(foundDevice).Should(BeTrue())

			// Verify device permissions in linux.resources.devices
			resources, ok := linux["resources"].(map[string]interface{})
			Expect(ok).Should(BeTrue())
			resourceDevices, ok := resources["devices"].([]interface{})
			Expect(ok).Should(BeTrue())

			// First rule should be deny all
			denyAll := resourceDevices[0].(map[string]interface{})
			Expect(denyAll["allow"]).Should(BeFalse())
			Expect(denyAll["access"]).Should(Equal("rwm"))

			// Should find an allow rule for our device
			foundAllowRule := false
			for _, rule := range resourceDevices {
				r := rule.(map[string]interface{})
				if r["allow"] == true &&
					r["type"] == "b" &&
					r["major"].(float64) == float64(major) &&
					r["minor"].(float64) == float64(minor) {
					foundAllowRule = true
					Expect(r["access"]).Should(Equal("rwm"))
					break
				}
			}
			Expect(foundAllowRule).Should(BeTrue())
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
func verifyNetworkSettings(opt *option.Option, ctrName, network string) {
	// inspect network
	resp := command.Stdout(opt, "network", "inspect", network)
	var inspectNetResp []*dockercompat.Network
	err := json.Unmarshal(resp, &inspectNetResp)
	Expect(err).Should(BeNil())
	Expect(inspectNetResp).Should(HaveLen(1))
	inspectNet := inspectNetResp[0]
	Expect(inspectNet.IPAM.Config).Should(HaveLen(1))
	gateway := strings.Split(inspectNet.IPAM.Config[0].Gateway, ".")
	Expect(gateway).Should(HaveLen(4))
	expectedMask := strings.Join(gateway[:3], ".")

	// inspect container
	resp = command.Stdout(opt, "inspect", ctrName)
	var inspectCtrResp []*dockercompat.Container
	err = json.Unmarshal(resp, &inspectCtrResp)
	Expect(err).Should(BeNil())
	Expect(inspectCtrResp).Should(HaveLen(1))
	inspectCtr := inspectCtrResp[0]

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
