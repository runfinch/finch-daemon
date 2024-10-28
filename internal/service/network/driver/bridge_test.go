// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"

	"github.com/coreos/go-iptables/iptables"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
)

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
