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

// NetworkInspect tests `GET networks/{id}` API.
func NetworkInspect(opt *option.Option) {
	Describe("inspect a network", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})
		It("should inspect network by network name", func() {
			// create network
			netId := command.StdoutStr(opt, "network", "create", testNetwork)

			// call inspect network api
			relativeUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/networks/%s", testNetwork))
			res, err := uClient.Get(relativeUrl)
			Expect(err).Should(BeNil())

			// verify inspect response
			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			var network types.NetworkInspectResponse
			err = json.NewDecoder(res.Body).Decode(&network)
			Expect(err).Should(BeNil())
			Expect(network.Name).Should(Equal(testNetwork))
			Expect(network.ID).Should(Equal(netId))
			Expect(network.IPAM.Config).ShouldNot(BeEmpty())
			// Verify Containers field is present (may be empty for new network)
			Expect(network.Containers).NotTo(BeNil())
		})
		It("should inspect network by long network id", func() {
			// create network
			netId := command.StdoutStr(opt, "network", "create", testNetwork)

			// call inspect network api
			relativeUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/networks/%s", netId))
			res, err := uClient.Get(relativeUrl)
			Expect(err).Should(BeNil())

			// verify inspect response
			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			var network types.NetworkInspectResponse
			err = json.NewDecoder(res.Body).Decode(&network)
			Expect(err).Should(BeNil())
			Expect(network.Name).Should(Equal(testNetwork))
			Expect(network.ID).Should(Equal(netId))
			Expect(network.IPAM.Config).ShouldNot(BeEmpty())
			// Verify Containers field is present (may be empty for new network)
			Expect(network.Containers).NotTo(BeNil())
		})
		It("should inspect network by short network id", func() {
			// create network
			netId := command.StdoutStr(opt, "network", "create", testNetwork)

			// call inspect network api
			relativeUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/networks/%s", netId[:12]))
			res, err := uClient.Get(relativeUrl)
			Expect(err).Should(BeNil())

			// verify inspect response
			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			var network types.NetworkInspectResponse
			err = json.NewDecoder(res.Body).Decode(&network)
			Expect(err).Should(BeNil())
			Expect(network.Name).Should(Equal(testNetwork))
			Expect(network.ID).Should(Equal(netId))
			Expect(network.IPAM.Config).ShouldNot(BeEmpty())
			// Verify Containers field is present (may be empty for new network)
			Expect(network.Containers).NotTo(BeNil())
		})
		It("should inspect network with labels", func() {
			// create network
			netId := command.StdoutStr(opt, "network", "create", "--label", "testLabel=testValue", testNetwork)

			// call inspect network api
			relativeUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/networks/%s", testNetwork))
			res, err := uClient.Get(relativeUrl)
			Expect(err).Should(BeNil())

			// verify inspect response
			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			var network types.NetworkInspectResponse
			err = json.NewDecoder(res.Body).Decode(&network)
			Expect(err).Should(BeNil())
			Expect(network.Name).Should(Equal(testNetwork))
			Expect(network.ID).Should(Equal(netId))
			Expect(network.Labels).Should(Equal(map[string]string{"testLabel": "testValue"}))
			Expect(network.IPAM.Config).ShouldNot(BeEmpty())
			// Verify Containers field is present (may be empty for new network)
			Expect(network.Containers).NotTo(BeNil())
		})
		It("should fail to inspect nonexistent network", func() {
			// call inspect network api
			relativeUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/networks/%s", testNetwork))
			res, err := uClient.Get(relativeUrl)
			Expect(err).Should(BeNil())

			// expect 404 response
			Expect(res).To(HaveHTTPStatus(http.StatusNotFound))
			var message response.Error
			err = json.NewDecoder(res.Body).Decode(&message)
			Expect(err).Should(BeNil())
			Expect(message.Message).Should(Equal(fmt.Sprintf("network %s not found", testNetwork)))
		})
	})
}
