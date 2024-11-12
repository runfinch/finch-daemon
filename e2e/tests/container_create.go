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
	"strings"
	"time"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/ffs"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

type containerCreateResponse struct {
	ID      string `json:"Id"`
	Message string `json:"message"`
}

// ContainerCreate tests the `POST containers/create` API.
func ContainerCreate(opt *option.Option) {
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
			command.Run(opt, "start", testContainerName)
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
			command.Run(opt, "start", testContainerName)
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
			command.Run(opt, "start", testContainerName)
			verifyNetworkSettings(opt, testContainerName, "bridge")
		})
		It("should attach container to the specified network using network name", func() {
			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.NetworkMode = testNetwork

			// create network
			command.Run(opt, "network", "create", testNetwork)

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify network settings
			command.Run(opt, "start", testContainerName)
			verifyNetworkSettings(opt, testContainerName, testNetwork)
		})
		It("should attach container to the specified network using network id", func() {
			// create network
			netId := command.StdoutStr(opt, "network", "create", testNetwork)
			Expect(netId).ShouldNot(BeEmpty())

			// define options
			options.Cmd = []string{"sleep", "Infinity"}
			options.HostConfig.NetworkMode = netId

			// create container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			// start container and verify network settings
			command.Run(opt, "start", testContainerName)
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
			command.Run(opt, "start", testContainerName)

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
			command.Run(opt, "start", testContainerName)

			// ensure that mounted file exists in container
			fileShouldExistInContainer(opt, testContainerName, ctrFilepath, fileContent)

			// ensure that write permissions are enabled on the mounted directory
			fileContent2 := "hello world again"
			filename2 := "test-file2"
			cmd := fmt.Sprintf("echo -n %s > %s", fileContent2, filepath.Join(filepath.Dir(ctrFilepath), filename2))
			command.Run(opt, "exec", testContainerName, "sh", "-c", cmd)
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
			command.Run(opt, "start", testContainerName)

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
			command.Run(opt, "volume", "create", testVolumeName)

			// define options
			options.HostConfig.Binds = []string{
				fmt.Sprintf("%s:%s", testVolumeName, filepath.Dir(ctrFilepath)),
			}
			options.Cmd = []string{"sleep", "Infinity"}

			// create and start container
			statusCode, ctr := createContainer(uClient, url, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			command.Run(opt, "start", testContainerName)

			// write file in the mounted volume
			cmd := fmt.Sprintf("echo -n %s > %s", fileContent, ctrFilepath)
			command.Run(opt, "exec", testContainerName, "sh", "-c", cmd)

			// ensure that created file exists in another container with the same volume mount
			statusCode, ctr = createContainer(uClient, url, testContainerName2, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())
			command.Run(opt, "start", testContainerName2)
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
			command.Run(opt, "start", testContainerName)

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
			command.Run(opt, "start", testContainerName)

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
			command.Run(opt, "start", testContainerName)
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
			command.Run(opt, "start", testContainerName)
			command.Run(opt, "kill", "--signal=SIGKILL", testContainerName)

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
