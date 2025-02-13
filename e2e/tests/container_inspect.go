// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
