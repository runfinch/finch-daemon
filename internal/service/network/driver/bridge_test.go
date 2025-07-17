// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"

	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/coreos/go-iptables/iptables"
	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
)

var _ = Describe("bridgeDriver HandleCreateOptions", func() {
	var (
		mockController *gomock.Controller
		logger         *mocks_logger.Logger
		driver         *bridgeDriver
		request        types.NetworkCreateRequest
		options        ncTypes.NetworkCreateOptions
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockController)
		driver = &bridgeDriver{logger: logger}
		request = types.NetworkCreateRequest{
			Options: make(map[string]string),
		}
		options = ncTypes.NetworkCreateOptions{
			Options: make(map[string]string),
			Labels:  []string{},
		}
	})

	Context("when processing bridge options", func() {
		It("should handle empty options", func() {
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(BeEmpty())
			Expect(result.Labels).To(BeEmpty())
		})

		It("should process valid host binding IPv4", func() {
			request.Options = map[string]string{
				BridgeHostBindingIpv4Option: "0.0.0.0",
			}
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Times(0)
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(BeEmpty())
		})

		It("should warn on invalid host binding IPv4", func() {
			request.Options = map[string]string{
				BridgeHostBindingIpv4Option: "192.168.1.1",
			}
			logger.EXPECT().Warnf("network option com.docker.network.bridge.host_binding_ipv4 is set to %s, but it must be 0.0.0.0", "192.168.1.1")
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(BeEmpty())
		})

		It("should handle valid ICC option true", func() {
			trueValues := []string{"1", "t", "T", "TRUE", "true", "True"}
			for _, v := range trueValues {
				request.Options = map[string]string{
					BridgeICCOption: v,
				}
				result, err := driver.HandleCreateOptions(request, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Options).To(BeEmpty())
				Expect(driver.disableICC).To(BeFalse(), "Failed for value: "+v)
				Expect(result.Labels).To(BeEmpty(), "Failed for value: "+v)
			}
		})

		It("should handle valid ICC option false", func() {
			falseValues := []string{"0", "f", "F", "FALSE", "false", "False"}
			for _, v := range falseValues {
				request.Options = map[string]string{
					BridgeICCOption: v,
				}
				result, err := driver.HandleCreateOptions(request, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Options).To(BeEmpty())
				Expect(driver.disableICC).To(BeTrue(), "Failed for value: "+v)
				Expect(result.Labels).To(ContainElement(FinchICCLabelIPv4+"=false"), "Failed for value: "+v)
			}
		})

		It("should handle valid ICC option false with IPv6", func() {
			request.Options = map[string]string{
				BridgeICCOption: "false",
			}
			driver.IPv6 = true
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(BeEmpty())
			Expect(driver.disableICC).To(BeTrue())
			Expect(result.Labels).To(ContainElement(FinchICCLabelIPv6 + "=false"))
		})

		It("should warn on invalid ICC option", func() {
			request.Options = map[string]string{
				BridgeICCOption: "invalid",
			}
			logger.EXPECT().Warnf("invalid value for com.docker.network.bridge.enable_icc")
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(BeEmpty())
			Expect(driver.disableICC).To(BeFalse())
			Expect(result.Labels).To(BeEmpty())
		})

		It("should set bridge name", func() {
			request.Options = map[string]string{
				BridgeNameOption: "testbridge",
			}
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(BeEmpty())
			Expect(driver.bridgeName).To(Equal("testbridge"))
		})

		It("should pass through unknown options", func() {
			request.Options = map[string]string{
				"unknown.option": "value",
			}
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(HaveKey("unknown.option"))
			Expect(result.Options["unknown.option"]).To(Equal("value"))
		})

		It("should handle multiple options together", func() {
			request.Options = map[string]string{
				BridgeHostBindingIpv4Option: "0.0.0.0",
				BridgeICCOption:             "false",
				BridgeNameOption:            "testbridge",
				"unknown.option":            "value",
			}
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Times(0)
			result, err := driver.HandleCreateOptions(request, options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Options).To(HaveKey("unknown.option"))
			Expect(driver.disableICC).To(BeTrue())
			Expect(driver.bridgeName).To(Equal("testbridge"))
			Expect(result.Labels).To(ContainElement(FinchICCLabelIPv4 + "=false"))
		})
	})
})

var _ = Describe("bridgeDriver Disable ICC", func() {
	var (
		mockController *gomock.Controller
		logger         *mocks_logger.Logger
		mockIpt        *mocks_network.MockIPTablesWrapper
		driver         *bridgeDriver
		bridgeIface    string
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockController)
		mockIpt = mocks_network.NewMockIPTablesWrapper(mockController)
		driver = &bridgeDriver{logger: logger}
		bridgeIface = "br-mock"
		newIptablesCommand = func(_ bool) (*iptablesCommand, error) {
			iptCommand := &iptablesCommand{
				protos: make(map[iptables.Protocol]IPTablesWrapper),
			}
			iptCommand.protos[iptables.ProtocolIPv4] = mockIpt
			return iptCommand, nil
		}
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})

	Context("chain does not exist", func() {
		It("should create and set up the FINCH-ISOLATE-CHAIN", func() {
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(false, nil).Times(1)
			mockIpt.EXPECT().NewChain("filter", "FINCH-ISOLATE-CHAIN").Return(nil)
			mockIpt.EXPECT().InsertUnique("filter", "FORWARD", 1, "-j", "FINCH-ISOLATE-CHAIN").Return(nil)
			mockIpt.EXPECT().AppendUnique("filter", "FINCH-ISOLATE-CHAIN", "-j", "RETURN").Return(nil)

			// Expect the second Insert call to FINCH-ISOLATE-CHAIN for the DROP rule
			mockIpt.EXPECT().InsertUnique("filter", "FINCH-ISOLATE-CHAIN", 1, "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(nil)
			err := driver.addICCDropRule(bridgeIface)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("chain setup fails", func() {
		When("we fail to check if chain exists", func() {
			It("should return an error and cleanup", func() {
				mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(false, fmt.Errorf("iptables failed"))

				////expect cleanup to be called
				mockIpt.EXPECT().DeleteChain("filter", "FINCH-ISOLATE-CHAIN").Return(nil)

				err := driver.addICCDropRule(bridgeIface)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to check if iptables FINCH-ISOLATE-CHAIN exists"))
			})
		})
		When("a new chain creation fails", func() {
			It("should return an error and cleanup", func() {
				mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(false, nil)
				mockIpt.EXPECT().NewChain("filter", "FINCH-ISOLATE-CHAIN").Return(fmt.Errorf("iptables failed"))

				//expect cleanup to be called
				mockIpt.EXPECT().DeleteChain("filter", "FINCH-ISOLATE-CHAIN").Return(nil)

				err := driver.addICCDropRule(bridgeIface)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create FINCH-ISOLATE-CHAIN chain"))
			})
		})
	})

	Context("chain already exists", func() {
		BeforeEach(func() {
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(true, nil).AnyTimes()
			mockIpt.EXPECT().InsertUnique("filter", "FORWARD", 1, "-j", "FINCH-ISOLATE-CHAIN").Return(nil).AnyTimes()
			mockIpt.EXPECT().AppendUnique("filter", "FINCH-ISOLATE-CHAIN", "-j", "RETURN").Return(nil).AnyTimes()
		})
		When("addICCDropRule is successful", func() {
			It("should add the DROP rule to the FINCH-ISOLATE-CHAIN", func() {
				// Expect the DROP rule to be inserted for packets from and to the same bridge
				mockIpt.EXPECT().InsertUnique("filter", "FINCH-ISOLATE-CHAIN", 1, "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(nil)
				err := driver.addICCDropRule(bridgeIface)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("removeICCDropRule is successful", func() {
			It("should remove the DROP rule from the FINCH-ISOLATE-CHAIN", func() {
				mockIpt.EXPECT().DeleteIfExists("filter", "FINCH-ISOLATE-CHAIN", "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(nil)
				err := driver.removeICCDropRule(bridgeIface)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
		When("addICCDropRule fails", func() {
			It("should return an error", func() {
				mockIpt.EXPECT().InsertUnique("filter", "FINCH-ISOLATE-CHAIN", 1, "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(fmt.Errorf("iptables error"))

				// //expect cleanup to be called
				mockIpt.EXPECT().DeleteIfExists("filter", "FINCH-ISOLATE-CHAIN", "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(nil)

				err := driver.addICCDropRule(bridgeIface)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to add iptables rule to drop ICC"))
			})
		})

		When("removeICCDropRule fails", func() {
			It("should return an error ", func() {
				mockIpt.EXPECT().DeleteIfExists("filter", "FINCH-ISOLATE-CHAIN", "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(fmt.Errorf("iptables error"))

				err := driver.removeICCDropRule(bridgeIface)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to remove iptables rules to drop ICC"))
			})
		})
	})
})
