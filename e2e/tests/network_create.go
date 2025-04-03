// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bytes"
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
	"github.com/runfinch/finch-daemon/e2e/util"
)

func NetworkCreate(opt *option.Option, pOpt util.NewOpt) {
	Describe("create a network", func() {
		const (
			path        = "/networks/create"
			contentType = "application/json"
		)

		var (
			uclient          *http.Client
			version          = GetDockerApiVersion()
			createNetworkURL = client.ConvertToFinchUrl(version, path)
		)

		BeforeEach(func() {
			uclient = client.NewClient(GetDockerHostUrl())
		})

		createNetwork := func(request types.NetworkCreateRequest) *http.Response {
			json, err := json.Marshal(request)
			Expect(err).ShouldNot(HaveOccurred(), "marshalling request to JSON")

			httpResponse, err := uclient.Post(createNetworkURL, contentType, bytes.NewReader(json))
			Expect(err).ShouldNot(HaveOccurred())

			return httpResponse
		}

		unmarshallHTTPResponse := func(httpResponse *http.Response) *types.NetworkCreateResponse {
			response := &types.NetworkCreateResponse{}

			body, err := io.ReadAll(httpResponse.Body)
			Expect(err).ShouldNot(HaveOccurred(), "reading response body")

			err = json.Unmarshal(body, response)
			Expect(err).ShouldNot(HaveOccurred(), "unmarshalling response from JSON")

			return response
		}

		cleanupNetwork := func(id string) func() {
			return func() {
				command.Run(opt, "network", "remove", id)
			}
		}

		cleanupNetworkWithHTTP := func(network string) func() {
			return func() {
				relativeUrl := fmt.Sprintf("/networks/%s", network)
				apiUrl := client.ConvertToFinchUrl(version, relativeUrl)
				req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
				Expect(err).ShouldNot(HaveOccurred())
				_, err = uclient.Do(req)
				Expect(err).Should(BeNil())
			}
		}

		When("a create network request is received with required configuration", func() {
			It("should return 201 Created and the network ID", func() {
				request := types.NewCreateNetworkRequest(testNetwork)

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				response := unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				DeferCleanup(cleanupNetwork(response.ID))
				Expect(response.Warning).Should(BeEmpty())
			})
		})

		When("a create network request is received with explicit default configuration", func() {
			It("should return 201 Created and the network ID", func() {
				request := types.NewCreateNetworkRequest(testNetwork, withDefaultOptions()...)

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				response := unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				DeferCleanup(cleanupNetwork(response.ID))
				Expect(response.Warning).Should(BeEmpty())
			})
		})

		When("a create network request is received with explicit non-default configuration", func() {
			It("should return 201 Created and the network ID", func() {
				request := types.NewCreateNetworkRequest(testNetwork, withNonDefaultOptions()...)

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				response := unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				DeferCleanup(cleanupNetwork(response.ID))
				Expect(response.Warning).Should(BeEmpty())
			})
		})

		When("a network create request is made with nerdctl unsupported network options", func() {
			It("should return 201 Created and the network ID", func() {
				request := types.NewCreateNetworkRequest(testNetwork, withUnsupportedNetworkOptions()...)

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				response := unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				DeferCleanup(cleanupNetwork(response.ID))
				Expect(response.Warning).Should(BeEmpty())
			})
		})

		When("a network create request is missing the required fields", func() {
			It("should return 500 Internal Server Error", func() {
				// Name is the only required field.
				request := &types.NetworkCreateRequest{}

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusInternalServerError))
			})
		})

		When("consecutive create network requests are made with the same network name", func() {
			It("should return 201 Created, the same network ID, and a warning", func() {
				request := types.NewCreateNetworkRequest(testNetwork)

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				response := unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				Expect(response.Warning).Should(BeEmpty())
				DeferCleanup(cleanupNetwork(response.ID))

				expected := response.ID

				httpResponse = createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				response = unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				Expect(response.ID).Should(Equal(expected))
				Expect(response.Warning).Should(ContainSubstring("already exists"))
			})
		})

		When("a create network request is made with an invalid JSON payload", func() {
			It("should return 400 Bad Request", func() {
				invalidJSON := []byte(fmt.Sprintf(`{Name: %s}`, testNetwork))
				httpResponse, err := uclient.Post(createNetworkURL, contentType, bytes.NewReader(invalidJSON))
				Expect(err).ShouldNot(HaveOccurred(), "sending HTTP request")

				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusBadRequest))
			})
		})

		When("a create network request is made with an unsupported network driver plugin", func() {
			It("should return 404 Not Found", func() {
				request := types.NewCreateNetworkRequest(testNetwork, types.WithDriver("baby"))

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusNotFound))
			})
		})

		When("a network create request is made with network option com.docker.network.bridge.enable_icc set to false", func() {
			It("should return 201 Created and the network ID", func() {
				testBridge := "br-test"
				request := types.NewCreateNetworkRequest(testNetwork, withEnableICCdNetworkOptions("false", testBridge)...)

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				response := unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				Expect(response.Warning).Should(BeEmpty())

				DeferCleanup(cleanupNetworkWithHTTP(testNetwork))

				stdout := command.Stdout(opt, "network", "inspect", testNetwork)
				Expect(stdout).To(ContainSubstring(`"finch.network.bridge.enable_icc.ipv4": "false"`))

				// check iptables rules exists
				iptOpt, _ := pOpt([]string{"iptables"})
				command.Run(iptOpt, "-C", "FINCH-ISOLATE-CHAIN",
					"-i", testBridge, "-o", testBridge, "-j", "DROP")
			})
		})

		When("a network create request is made with network option com.docker.network.bridge.enable_icc set to true", func() {
			It("should create the network without the enable_icc label", func() {
				testBridge := "br-test"
				request := types.NewCreateNetworkRequest(testNetwork, withEnableICCdNetworkOptions("true", testBridge)...)

				httpResponse := createNetwork(*request)
				Expect(httpResponse).Should(HaveHTTPStatus(http.StatusCreated))

				DeferCleanup(cleanupNetworkWithHTTP(testNetwork))

				response := unmarshallHTTPResponse(httpResponse)
				Expect(response.ID).ShouldNot(BeEmpty())
				Expect(response.Warning).Should(BeEmpty())

				stdout := command.Stdout(opt, "network", "inspect", testNetwork)
				Expect(stdout).ShouldNot(ContainSubstring(`"finch.network.bridge.enable_icc.ipv4"`))

				// check iptables rules does not exist
				iptOpt, _ := pOpt([]string{"iptables"})
				command.RunWithoutSuccessfulExit(iptOpt, "-C", "FINCH-ISOLATE-CHAIN",
					"-i", testBridge, "-o", testBridge, "-j", "DROP")
			})
		})
	})
}

func withDefaultOptions() []types.NetworkCreateOption {
	return []types.NetworkCreateOption{
		types.WithDriver("bridge"),
		types.WithInternal(false),
		types.WithAttachable(false),
		types.WithIngress(false),
		types.WithIPAM(types.IPAM{
			Driver: "default",
		}),
		types.WithEnableIPv6(false),
		types.WithOptions(map[string]string{}),
		types.WithLabels(map[string]string{}),
	}
}

func withNonDefaultOptions() []types.NetworkCreateOption {
	return []types.NetworkCreateOption{
		types.WithDriver("ipvlan"),
		types.WithInternal(true),
		types.WithAttachable(false),
		types.WithIngress(false),
		types.WithIPAM(types.IPAM{
			Driver: "default",
			Config: []map[string]string{
				{
					"Subnet":  "172.20.0.0/16",
					"IPRange": "172.20.10.0/24",
					"Gateway": "172.20.10.11",
				},
				{
					"Subnet":  "2001:db8:abcd::/64",
					"Gateway": "2001:db8:abcd::1011",
				},
			},
			Options: map[string]string{
				"foo": "bar",
			},
		}),
		types.WithEnableIPv6(true),
		types.WithOptions(map[string]string{
			"com.docker.network.driver.mtu": "1000",
		}),
		types.WithLabels(map[string]string{
			"com.example.some-label":       "some-value",
			"com.example.some-other-label": "some-other-value",
		}),
	}
}

func withUnsupportedNetworkOptions() []types.NetworkCreateOption {
	return []types.NetworkCreateOption{
		types.WithIPAM(types.IPAM{
			Driver: "default",
			Config: []map[string]string{
				{
					"Subnet": "240.10.0.0/24",
				},
			},
		}),
		types.WithOptions(map[string]string{
			"com.docker.network.bridge.enable_ip_masquerade": "true",
			"com.docker.network.bridge.host_binding_ipv4":    "0.0.0.0",
			"com.docker.network.bridge.name":                 "EvergreenPointFloatingBridge",
		}),
		types.WithLabels(map[string]string{
			"com.example.some-label":       "some-value",
			"com.example.some-other-label": "some-other-value",
		}),
	}
}

func withEnableICCdNetworkOptions(enableICC string, bridgeName string) []types.NetworkCreateOption {
	return []types.NetworkCreateOption{
		types.WithIPAM(types.IPAM{
			Driver: "default",
			Config: []map[string]string{
				{
					"Subnet": "240.11.0.0/24",
				},
			},
		}),
		types.WithOptions(map[string]string{
			"com.docker.network.bridge.enable_icc": enableICC,
			"com.docker.network.bridge.name":       bridgeName,
		}),
	}
}
