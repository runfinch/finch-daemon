// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

func ContainerKill(opt *option.Option) {
	Describe("kill a container", func() {
		var (
			uClient *http.Client
			version string
			apiUrl  string
		)
		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			relativeUrl := fmt.Sprintf("/containers/%s/kill", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})
		It("should kill the container with default SIGKILL", func() {
			// start a container that keeps running
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"sleep", "infinity"})
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldNotBeRunning(opt, testContainerName)
		})
		It("should fail to kill a non-existent container", func() {
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			var body response.Error
			err = json.NewDecoder(res.Body).Decode(&body)
			Expect(err).Should(BeNil())
		})
		It("should fail to kill a non running container", func() {
			httpCreateContainer(uClient, version, testContainerName, types.ContainerCreateRequest{
				ContainerConfig: types.ContainerConfig{
					Image: defaultImage,
					Cmd:   []string{"sleep", "infinity"},
				},
			})
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusConflict))
			var body response.Error
			err = json.NewDecoder(res.Body).Decode(&body)
			Expect(err).Should(BeNil())
			containerShouldExist(opt, testContainerName)
		})
		It("should kill the container with SIGINT", func() {
			relativeUrl := fmt.Sprintf("/containers/%s/kill?signal=SIGINT", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			// sleep infinity doesnot respond to SIGINT
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"/bin/sh", "-c", "trap 'exit 0' SIGINT; while true; do sleep 1; done"})
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			// This is an async operation as a result we need to wait for the container to exit gracefully before checking the status
			time.Sleep(1 * time.Second)
			containerShouldNotBeRunning(opt, testContainerName)
		})
		It("should not kill the container and throw error on unrecognized signal", func() {
			relativeUrl := fmt.Sprintf("/containers/%s/kill?signal=SIGRAND", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"sleep", "infinity"})
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusInternalServerError))
			containerShouldExist(opt, testContainerName)
			containerShouldBeRunning(opt, testContainerName)
		})
	})
}
