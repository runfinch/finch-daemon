// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

func ContainerPause(opt *option.Option) {
	Describe("pause a container", func() {
		var (
			uClient *http.Client
			version string
			apiUrl  string
		)

		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			relativeUrl := fmt.Sprintf("/containers/%s/pause", testContainerName)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
		})

		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should pause a running container", func() {
			// Start a container that keeps running
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"sleep", "infinity"})

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))

			// Verify container is paused
			output := command.StdoutStr(opt, "inspect", "--format", "{{.State.Status}}", testContainerName)
			Expect(output).Should(Equal("paused"))
		})

		It("should fail to pause a non-existent container", func() {
			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))

			var body response.Error
			err = json.NewDecoder(res.Body).Decode(&body)
			Expect(err).Should(BeNil())
		})

		It("should fail to pause a non-running container", func() {
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

		It("should fail to pause an already paused container", func() {
			// Start and pause the container
			httpRunContainer(uClient, version, testContainerName, defaultImage, []string{"sleep", "infinity"})
			httpPauseContainer(uClient, version, testContainerName)

			res, err := uClient.Post(apiUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusConflict))

			var body response.Error
			err = json.NewDecoder(res.Body).Decode(&body)
			Expect(err).Should(BeNil())
		})
	})
}
