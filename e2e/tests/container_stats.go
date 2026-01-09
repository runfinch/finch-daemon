// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	dockertypes "github.com/docker/docker/api/types/container"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// waitForContainerRunning waits for a container to reach the running state.
func waitForContainerRunning(uClient *http.Client, version string, containerID string, maxRetries int) bool {
	const retryInterval = 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json", containerID)))
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		if res.StatusCode != http.StatusOK {
			time.Sleep(retryInterval)
			continue
		}

		var container types.Container
		err = json.NewDecoder(res.Body).Decode(&container)
		res.Body.Close()
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		if container.State.Status == "running" {
			return true
		}

		time.Sleep(retryInterval)
	}
	return false
}

// ContainerStats tests the `GET containers/{id}/stats` API.
func ContainerStats(opt *option.Option) {
	Describe("get container stats", func() {
		var (
			uClient           *http.Client
			version           string
			wantContainerName string
		)
		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			wantContainerName = fmt.Sprintf("/%s", testContainerName)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should return a 404 error if container does not exist", func() {
			url := client.ConvertToFinchUrl(version, "/containers/container-does-not-exist/stats")
			res, err := uClient.Get(url)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			var errResponse response.Error
			err = json.NewDecoder(res.Body).Decode(&errResponse)
			Expect(err).Should(BeNil())
			Expect(errResponse.Message).ShouldNot(BeEmpty())
		})
		It("should return container stats from container name without streaming", func() {
			cid := command.StdoutStr(
				opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "Infinity",
			)

			isRunning := waitForContainerRunning(uClient, version, cid, 10)
			Expect(isRunning).Should(BeTrue(), "Container should be in running state before checking stats")

			relativeUrl := fmt.Sprintf("/containers/%s/stats?stream=false", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			var statsJSON types.StatsJSON
			err = json.NewDecoder(res.Body).Decode(&statsJSON)
			Expect(err).Should(BeNil())
			expectValidStats(&statsJSON, wantContainerName, cid, 1)
		})
		It("should return container stats from long container ID without streaming", func() {
			cid := command.StdoutStr(
				opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "Infinity",
			)

			isRunning := waitForContainerRunning(uClient, version, cid, 10)
			Expect(isRunning).Should(BeTrue(), "Container should be in running state before checking stats")

			relativeUrl := fmt.Sprintf("/containers/%s/stats?stream=false", cid)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			var statsJSON types.StatsJSON
			err = json.NewDecoder(res.Body).Decode(&statsJSON)
			Expect(err).Should(BeNil())
			expectValidStats(&statsJSON, wantContainerName, cid, 1)
		})
		It("should return container stats from short container ID without streaming", func() {
			cid := command.StdoutStr(
				opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "Infinity",
			)

			isRunning := waitForContainerRunning(uClient, version, cid, 10)
			Expect(isRunning).Should(BeTrue(), "Container should be in running state before checking stats")

			relativeUrl := fmt.Sprintf("/containers/%s/stats?stream=false", cid[:12])
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			var statsJSON types.StatsJSON
			err = json.NewDecoder(res.Body).Decode(&statsJSON)
			Expect(err).Should(BeNil())
			expectValidStats(&statsJSON, wantContainerName, cid, 1)
		})
		It("should stream container stats until the container is removed", func() {
			cid := command.StdoutStr(
				opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "Infinity",
			)

			isRunning := waitForContainerRunning(uClient, version, cid, 10)
			Expect(isRunning).Should(BeTrue(), "Container should be in running state before checking stats")

			go func() {
				time.Sleep(time.Second * 5)
				httpRemoveContainerForce(uClient, version, testContainerName)
			}()

			relativeUrl := fmt.Sprintf("/containers/%s/stats", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			scanner := bufio.NewScanner(res.Body)
			num := 0
			for scanner.Scan() {
				var statsJSON types.StatsJSON
				err = json.Unmarshal(scanner.Bytes(), &statsJSON)
				Expect(err).Should(BeNil())
				isRunning := waitForContainerRunning(uClient, version, cid, 1)
				// confirm need to check the stats were obtained for container running state
				if !isRunning {
					break
				}
				expectValidStats(&statsJSON, wantContainerName, cid, 1)
				num += 1
			}

			Expect(num).Should(BeNumerically(">", 1))
			Expect(num).Should(BeNumerically("<", 10))
		})
		It("should stream stats for a stopped container", func() {
			cid := httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"echo", "hello"})
			httpWaitContainer(uClient, version, testContainerName)

			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/json", cid)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var container types.Container
			err = json.NewDecoder(res.Body).Decode(&container)
			res.Body.Close()
			Expect(err).Should(BeNil())
			Expect(container.State.Status).ShouldNot(Equal("running"))

			go func() {
				time.Sleep(time.Second * 5)
				httpRemoveContainerForce(uClient, version, testContainerName)
			}()

			relativeUrl := fmt.Sprintf("/containers/%s/stats", testContainerName)
			res, err = uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			scanner := bufio.NewScanner(res.Body)
			num := 0
			for scanner.Scan() {
				var statsJSON types.StatsJSON
				err = json.Unmarshal(scanner.Bytes(), &statsJSON)
				Expect(err).Should(BeNil())
				expectEmptyStats(&statsJSON, wantContainerName, cid)
				num += 1
			}

			Expect(num).Should(BeNumerically(">", 1))
			Expect(num).Should(BeNumerically("<", 10))
		})
		It("should stream stats when no network interface is created", func() {
			cid := command.StdoutStr(
				opt,
				"run",
				"-d",
				"--net", "none",
				"--name", testContainerName,
				defaultImage,
				"sleep", "Infinity",
			)

			isRunning := waitForContainerRunning(uClient, version, cid, 10)
			Expect(isRunning).Should(BeTrue(), "Container should be in running state before checking stats")

			go func() {
				time.Sleep(time.Second * 5)
				httpRemoveContainerForce(uClient, version, testContainerName)
			}()

			relativeUrl := fmt.Sprintf("/containers/%s/stats", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			scanner := bufio.NewScanner(res.Body)
			num := 0
			for scanner.Scan() {
				var statsJSON types.StatsJSON
				err = json.Unmarshal(scanner.Bytes(), &statsJSON)
				Expect(err).Should(BeNil())
				isRunning := waitForContainerRunning(uClient, version, cid, 1)
				// confirm need to check the stats were obtained for container running state
				if !isRunning {
					break
				}
				expectValidStats(&statsJSON, wantContainerName, cid, 0)
				num += 1
			}

			Expect(num).Should(BeNumerically(">", 1))
			Expect(num).Should(BeNumerically("<", 10))
		})
		It("should stream stats with multiple network interfaces", func() {
			httpCreateNetwork(uClient, version, "net1")
			httpCreateNetwork(uClient, version, "net2")
			cid := command.StdoutStr(
				opt,
				"run",
				"-d",
				"--net", "net1",
				"--net", "net2",
				"--name", testContainerName,
				defaultImage,
				"sleep", "Infinity",
			)

			isRunning := waitForContainerRunning(uClient, version, cid, 10)
			Expect(isRunning).Should(BeTrue(), "Container should be in running state before checking stats")

			go func() {
				time.Sleep(time.Second * 5)
				httpRemoveContainerForce(uClient, version, testContainerName)
			}()

			relativeUrl := fmt.Sprintf("/containers/%s/stats", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			scanner := bufio.NewScanner(res.Body)
			num := 0
			for scanner.Scan() {
				var statsJSON types.StatsJSON
				err = json.Unmarshal(scanner.Bytes(), &statsJSON)
				Expect(err).Should(BeNil())
				isRunning := waitForContainerRunning(uClient, version, cid, 1)
				// confirm need to check the stats were obtained for container running state
				if !isRunning {
					break
				}
				expectValidStats(&statsJSON, wantContainerName, cid, 2)
				num += 1
			}

			Expect(num).Should(BeNumerically(">", 1))
			Expect(num).Should(BeNumerically("<", 10))
		})
	})
}

// expectValidStats ensures that the data contained in the stats object is valid.
func expectValidStats(st *types.StatsJSON, name, id string, numNetworks int) {
	// verify container name and ID
	Expect(st.Name).Should(Equal(name))
	Expect(st.ID).Should(Equal(id))

	// check that the time difference between last read and current read
	// is approximately 1 second
	t := time.Time{}
	if st.PreRead != t {
		Expect(st.Read).Should(BeTemporally("~", st.PreRead.Add(time.Second), time.Millisecond*100))
	}

	Expect(st.PidsStats.Current).ShouldNot(BeZero())
	Expect(st.CPUStats.OnlineCPUs).ShouldNot(BeZero())
	Expect(st.CPUStats.SystemUsage).ShouldNot(BeZero())

	if numNetworks == 0 {
		Expect(st.Networks).Should(BeNil())
	} else {
		Expect(st.Networks).ShouldNot(BeNil())
		Expect(len(st.Networks)).Should(Equal(numNetworks))
	}
}

// expectEmptyStats ensures that the data contained in the stats object is empty
// which is the case with containers that are not running.
func expectEmptyStats(st *types.StatsJSON, name, id string) {
	Expect(st.Name).Should(Equal(name))
	Expect(st.ID).Should(Equal(id))
	Expect(st.Read).Should(Equal(time.Time{}))
	Expect(st.PreRead).Should(Equal(time.Time{}))
	Expect(st.PidsStats).Should(Equal(dockertypes.PidsStats{}))
	Expect(st.BlkioStats).Should(Equal(dockertypes.BlkioStats{}))
	Expect(st.CPUStats).Should(Equal(types.CPUStats{}))
	Expect(st.PreCPUStats).Should(Equal(types.CPUStats{}))
	Expect(st.MemoryStats).Should(Equal(dockertypes.MemoryStats{}))
}
