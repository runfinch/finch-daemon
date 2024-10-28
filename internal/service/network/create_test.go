// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/nerdctl/pkg/netutil"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/network"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/internal/service/network/driver"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

var _ = Describe("Network Service Create Network Implementation", func() {
	const (
		networkName = "test-network"
		networkID   = "f2ce5cdfcb34238294c247a218b764347f78e55b0f61d00c6364df0ffe3a1de9"
	)

	var (
		ctx              context.Context
		mockController   *gomock.Controller
		cdClient         *mocks_backend.MockContainerdClient
		ncNetClient      *mocks_backend.MockNerdctlNetworkSvc
		logger           *mocks_logger.Logger
		service          network.Service
		mockBridgeDriver *mocks_network.DriverHandler
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockController = gomock.NewController(GinkgoT())
		cdClient = mocks_backend.NewMockContainerdClient(mockController)
		ncNetClient = mocks_backend.NewMockNerdctlNetworkSvc(mockController)
		logger = mocks_logger.NewLogger(mockController)
		service = NewService(cdClient, ncNetClient, logger)
		mockBridgeDriver = mocks_network.NewDriverHandler(mockController)
	})

	When("a create network call is successful", func() {
		It("should return the network ID", func() {
			request := types.NewCreateNetworkRequest(networkName)

			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			nid := networkID
			ncNetClient.EXPECT().CreateNetwork(gomock.Any()).Return(&netutil.NetworkConfig{
				NerdctlID: &nid,
			}, nil)

			response, err := service.Create(ctx, *request)
			Expect(response.ID).Should(Equal(networkID))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("a network already exists", func() {
		When("a request collides with an already existing user defined network", func() {
			It("should return the network ID and a warning that the network exists already", func() {
				request := types.NewCreateNetworkRequest(networkName)

				nid := networkID
				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{{NerdctlID: &nid}}, nil)

				response, err := service.Create(ctx, *request)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.ID).Should(Equal(networkID))
				Expect(response.Warning).Should(ContainSubstring("already exists"))
			})
		})

		When("a request collides with an already existing default network", func() {
			It("should return a warning that the network exists already", func() {
				request := types.NewCreateNetworkRequest(networkName)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{{NerdctlID: nil}}, nil)

				response, err := service.Create(ctx, *request)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.ID).Should(BeEmpty())
				Expect(response.Warning).Should(ContainSubstring("already exists"))
			})
		})
	})

	When("a network plugin is not supported", func() {
		It("should return an error the driver was not found", func() {
			request := types.NewCreateNetworkRequest(networkName)

			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			ncNetClient.EXPECT().CreateNetwork(gomock.Any()).Return(nil, errUnsupportedCNIDriver)

			response, err := service.Create(ctx, *request)
			Expect(response.ID).Should(BeEmpty())
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(MatchError(ContainSubstring("not found")))
		})
	})

	Context("returns from nerdctl which should not happen", func() {
		When("nerdctl successfully creates the network but returns nil network", func() {
			It("should return an error that the network ID was not found", func() {
				request := types.NewCreateNetworkRequest(networkName)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).Return(nil, nil)

				response, err := service.Create(ctx, *request)
				Expect(response.ID).Should(BeEmpty())
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(ContainSubstring("not found")))
			})
		})

		When("nerdctl successfully creates the network but does not return a network ID", func() {
			It("should return an error that the network ID was not found", func() {
				request := types.NewCreateNetworkRequest(networkName)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).Return(&netutil.NetworkConfig{}, nil)

				response, err := service.Create(ctx, *request)
				Expect(response.ID).Should(BeEmpty())
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(ContainSubstring("not found")))
			})
		})
	})

	When("a create network error occurs", func() {
		It("should return the error", func() {
			request := types.NewCreateNetworkRequest(networkName)

			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			errFromNerd := errors.New("create network failed")
			ncNetClient.EXPECT().CreateNetwork(gomock.Any()).Return(nil, errFromNerd)

			response, err := service.Create(ctx, *request)
			Expect(response.ID).Should(BeEmpty())
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(errFromNerd))
		})
	})

	Context("Nerdctl default configuration", func() {
		const (
			defaultExpectedDriver     = "bridge"
			defaultExpectedIPAMDriver = "default"

			overrideExpectedDriver     = "baby"
			overrideExpectedIPAMDriver = "baby-ipam"
		)

		When("a request is missing nerdctl required configuration", func() {
			It("should apply the default configuration", func() {
				request := types.NewCreateNetworkRequest(networkName)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				nid := networkID
				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
					Expect(actual.Driver).Should(Equal(defaultExpectedDriver))
					Expect(actual.IPAMDriver).Should(Equal(defaultExpectedIPAMDriver))
					return &netutil.NetworkConfig{NerdctlID: &nid}, nil
				})

				service.Create(ctx, *request)
			})
		})

		When("a request provides nerdctl required configuration", func() {
			It("should override the default configuration", func() {
				request := types.NewCreateNetworkRequest(
					networkName,
					types.WithDriver(overrideExpectedDriver),
					types.WithIPAM(types.IPAM{Driver: overrideExpectedIPAMDriver}),
				)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				nid := networkID
				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
					Expect(actual.Driver).Should(Equal(overrideExpectedDriver))
					Expect(actual.IPAMDriver).Should(Equal(overrideExpectedIPAMDriver))
					return &netutil.NetworkConfig{NerdctlID: &nid}, nil
				})

				service.Create(ctx, *request)
			})
		})
	})

	Context("IPAM configuration", func() {
		const (
			expectedIPRange = "172.20.10.0/24"
			expectedGateway = "172.20.10.11"
		)
		expectedSubnets := []string{"172.20.0.0/16"}
		When("multiple IPAM configuration entries are specified", func() {
			It("should use the first IPAM object", func() {
				request := types.NewCreateNetworkRequest(
					networkName,
					types.WithIPAM(types.IPAM{
						Driver: "default",
						Config: []map[string]string{
							{
								"Subnet":  expectedSubnets[0],
								"IPRange": expectedIPRange,
								"Gateway": expectedGateway,
							},
							{
								"Subnet":  "2001:db8:abcd::/64",
								"Gateway": "2001:db8:abcd::1011",
							},
						},
					}),
				)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				nid := networkID
				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
					Expect(actual.Subnets).Should(Equal(expectedSubnets))
					Expect(actual.IPRange).Should(Equal(expectedIPRange))
					Expect(actual.Gateway).Should(Equal(expectedGateway))
					return &netutil.NetworkConfig{NerdctlID: &nid}, nil
				})

				service.Create(ctx, *request)
			})
		})

		Context("partial IPAM configuration", func() {
			When("only subnet is specified", func() {
				It("should use the configuration that is available", func() {
					request := types.NewCreateNetworkRequest(
						networkName,
						types.WithIPAM(types.IPAM{
							Driver: "default",
							Config: []map[string]string{
								{
									"Subnet": expectedSubnets[0],
								},
							},
						}),
					)

					ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
					logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

					nid := networkID
					ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
						Expect(actual.Subnets).Should(Equal(expectedSubnets))
						Expect(actual.IPRange).Should(BeEmpty())
						Expect(actual.Gateway).Should(BeEmpty())
						return &netutil.NetworkConfig{NerdctlID: &nid}, nil
					})

					service.Create(ctx, *request)
				})
			})

			When("only IP range is specified", func() {
				It("should use the configuration that is available", func() {
					request := types.NewCreateNetworkRequest(
						networkName,
						types.WithIPAM(types.IPAM{
							Driver: "default",
							Config: []map[string]string{
								{
									"IPRange": expectedIPRange,
								},
							},
						}),
					)

					ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
					logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

					nid := networkID
					ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
						Expect(actual.Subnets).Should(BeEmpty())
						Expect(actual.IPRange).Should(Equal(expectedIPRange))
						Expect(actual.Gateway).Should(BeEmpty())
						return &netutil.NetworkConfig{NerdctlID: &nid}, nil
					})

					service.Create(ctx, *request)
				})
			})

			When("only gateway is specified", func() {
				It("should use the configuration that is available", func() {
					request := types.NewCreateNetworkRequest(
						networkName,
						types.WithIPAM(types.IPAM{
							Driver: "default",
							Config: []map[string]string{
								{
									"Gateway": expectedGateway,
								},
							},
						}),
					)

					ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
					logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

					nid := networkID
					ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
						Expect(actual.Subnets).Should(BeEmpty())
						Expect(actual.IPRange).Should(BeEmpty())
						Expect(actual.Gateway).Should(Equal(expectedGateway))
						return &netutil.NetworkConfig{NerdctlID: &nid}, nil
					})

					service.Create(ctx, *request)
				})
			})
		})
	})

	Context("ICC configuration", func() {
		When("com.docker.network.bridge.enable_icc is set to true in the request", func() {
			It("should not change default behavior", func() {
				request := types.NewCreateNetworkRequest(
					networkName,
					types.WithOptions(map[string]string{
						driver.BridgeICCOption: "true",
					}),
				)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				nid := networkID
				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
					// Check if the label exists
					checkLabel := driver.BridgeICCOption + "=true"
					labelExists := false
					for _, label := range actual.Labels {
						if label == checkLabel {
							labelExists = true
							break
						}
					}

					Expect(labelExists).To(BeFalse(), fmt.Sprintf("Label '%s' should not exist in Labels", checkLabel))

					return &netutil.NetworkConfig{NerdctlID: &nid}, nil
				})

				response, err := service.Create(ctx, *request)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.ID).Should(Equal(networkID))
			})
		})

		When("ICC is not specified in the request", func() {
			It("should use the default setting in the create options and not set any icc labels", func() {
				request := types.NewCreateNetworkRequest(networkName)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				nid := networkID
				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
					// Check if the label exists
					expectedLabel := driver.BridgeICCOption + "=true"
					labelExists := false
					for _, label := range actual.Labels {
						if label == expectedLabel {
							labelExists = true
							break
						}
					}

					Expect(labelExists).To(BeFalse(), fmt.Sprintf("Label '%s' should not exist in Labels", expectedLabel))

					return &netutil.NetworkConfig{NerdctlID: &nid}, nil
				})

				response, err := service.Create(ctx, *request)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.ID).Should(Equal(networkID))
			})
		})

		When("com.docker.network.bridge.enable_icc is set to false in the request", func() {
			It("should set ICC to false in the create options", func() {
				request := types.NewCreateNetworkRequest(
					networkName,
					types.WithOptions(map[string]string{
						driver.BridgeICCOption: "false",
					}),
				)

				ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				nid := networkID
				ncNetClient.EXPECT().CreateNetwork(gomock.Any()).DoAndReturn(func(actual netutil.CreateOptions) (*netutil.NetworkConfig, error) {
					// Check if the label exists
					expectedLabel := driver.FinchICCLabelIPv4 + "=false"
					labelExists := false
					for _, label := range actual.Labels {
						if label == expectedLabel {
							labelExists = true
							break
						}
					}

					Expect(labelExists).To(BeTrue(), fmt.Sprintf("Label '%s' should exist in Labels", expectedLabel))

					return &netutil.NetworkConfig{NerdctlID: &nid}, nil
				})

				originalDriverFunc := driver.NewBridgeDriver
				driver.NewBridgeDriver = func(netClient backend.NerdctlNetworkSvc, logger flog.Logger, isIPv6 bool) (driver.DriverHandler, error) {
					return mockBridgeDriver, nil
				}
				defer func() { driver.NewBridgeDriver = originalDriverFunc }()

				// Set up expectations for mockBridgeDriver
				mockBridgeDriver.EXPECT().HandleCreateOptions(gomock.Any(), gomock.Any()).DoAndReturn(
					func(request types.NetworkCreateRequest, options netutil.CreateOptions) (netutil.CreateOptions, error) {
						// Mock the behavior for BridgeICCOption set to false
						// Remove the option from the options map

						delete(options.Options, driver.BridgeICCOption)
						options.Labels = append(options.Labels, driver.FinchICCLabelIPv4+"=false")
						return options, nil
					}).AnyTimes()

				mockBridgeDriver.EXPECT().HandlePostCreate(gomock.Any()).Return("", nil).AnyTimes()

				response, err := service.Create(ctx, *request)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.ID).Should(Equal(networkID))
			})
		})
	})
})
