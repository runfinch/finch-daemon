// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

func NetworkRemove(opt *option.Option) {
	Describe("remove a network", func() {
		var (
			uClient *http.Client
			version string
			apiUrl  string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			relativeUrl := fmt.Sprintf("/networks/%s", testNetwork)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})
		It("should remove the network by name", func() {
			httpCreateNetwork(uClient, version, testNetwork)
			req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
			Expect(err).Should(BeNil())
			res, err := uClient.Do(req)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
		})
		It("should remove the network by id", func() {
			networkId := command.StdoutStr(opt, "network", "create", testNetwork)
			relativeUrl := fmt.Sprintf("/networks/%s", networkId)
			apiUrl = client.ConvertToFinchUrl(version, relativeUrl)
			req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
			Expect(err).Should(BeNil())
			res, err := uClient.Do(req)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
		})
		It("should not remove a network in use", func() {
			httpCreateNetwork(uClient, version, testNetwork)
			httpRunContainerWithOptions(uClient, version, testContainerName, types.ContainerCreateRequest{
				ContainerConfig: types.ContainerConfig{
					Image: defaultImage,
					Cmd:   []string{"sleep", "infinity"},
				},
				HostConfig: types.ContainerHostConfig{
					NetworkMode: testNetwork,
				},
			})
			req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
			Expect(err).Should(BeNil())
			res, err := uClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			defer res.Body.Close()
			body, err := httputil.DumpResponse(res, true)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(body).Should(ContainSubstring("\"test-network\\\" is in use by container"))
			Expect(res.StatusCode).Should(Equal(http.StatusForbidden))
		})
		It("should return an error when network is not found", func() {
			httpCreateNetwork(uClient, version, "notfound")
			req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
			Expect(err).ShouldNot(HaveOccurred())
			res, err := uClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
		})
	})
}
