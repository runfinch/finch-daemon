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

	"github.com/runfinch/finch-daemon/e2e/client"
	"github.com/runfinch/finch-daemon/pkg/api/response"
)

// ContainerStart tests the `POST containers/{id}/start` API.
func ContainerStart(opt *option.Option) {
	Describe("start a container", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			command.Run(opt, "create", "--name", testContainerName, defaultImage, "echo", "foo")
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should start the container", func() {
			relativeUrl := fmt.Sprintf("/containers/%s/start", testContainerName)
			res, err := uClient.Post(client.ConvertToFinchUrl(version, relativeUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
		})
		It("should fail to start the container", func() {
			// start a container that does not exist
			relativeUrl := client.ConvertToFinchUrl(version, "/containers/container-does-not-exist/start")
			res, err := uClient.Post(relativeUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			var errResponse response.Error
			err = json.NewDecoder(res.Body).Decode(&errResponse)
			Expect(err).Should(BeNil())
			Expect(errResponse.Message).Should(Not(BeEmpty()))
		})
	})
}
