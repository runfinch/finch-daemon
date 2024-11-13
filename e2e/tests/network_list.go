// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// NetworkList tests calling the get networks api.
func NetworkList(opt *option.Option) {
	Describe("lists the networks", func() {
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
		It("should return bridge network by default", func() {
			relativeUrl := client.ConvertToFinchUrl(version, "/networks")

			res, err := uClient.Get(relativeUrl)
			Expect(err).Should(BeNil())

			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			ls := new([]*types.NetworkInspectResponse)
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			err = json.Unmarshal(body, ls)
			Expect(err).Should(BeNil())
			Expect((*ls)[0].Name).Should(Equal("bridge"))
		})
		It("should return a list with a new network", func() {
			expName := "test-net"
			command.Run(opt, "network", "create", expName)
			relativeUrl := client.ConvertToFinchUrl(version, "/networks")

			res, err := uClient.Get(relativeUrl)
			Expect(err).Should(BeNil())

			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			ls := new([]*types.NetworkInspectResponse)
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			err = json.Unmarshal(body, ls)
			Expect(err).Should(BeNil())
			Expect((*ls)[0].Name).Should(Equal(expName))
		})
	})
}
