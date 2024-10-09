// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
)

var _ = Describe("bridgeDriver DisableICC", func() {
	var (
		mockController *gomock.Controller
		mockIpt        *mocks_network.MockIPTablesWrapper
		driver         *bridgeDriver
		bridgeIface    string
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIpt = mocks_network.NewMockIPTablesWrapper(mockController)
		driver = &bridgeDriver{ipt: mockIpt}
		bridgeIface = "br-mock"
	})

	Context("FINCH-ISOLATE-CHAIN chain does not exist", func() {
		BeforeEach(func() {
			// FINCH-ISOLATE-CHAIN does not exist initially
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(false, nil)

			// NewChain method should be called
			mockIpt.EXPECT().NewChain("filter", "FINCH-ISOLATE-CHAIN").Return(nil)

			// Expect a rule to be added to FORWARD chain that jumps to FINCH-ISOLATE-CHAIN
			mockIpt.EXPECT().Insert("filter", "FORWARD", 1, "-j", "FINCH-ISOLATE-CHAIN").Return(nil)
			// Expect a RETURN rule to be appended in FINCH-ISOLATE-CHAIN
			mockIpt.EXPECT().Append("filter", "FINCH-ISOLATE-CHAIN", "-j", "RETURN").Return(nil)

			// Expect the second Insert call to FINCH-ISOLATE-CHAIN for the DROP rule
			mockIpt.EXPECT().Insert("filter", "FINCH-ISOLATE-CHAIN", 1, "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(nil)
		})

		It("should create and set up the FINCH-ISOLATE-CHAIN", func() {
			err := driver.DisableICC(bridgeIface, true)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("FINCH-ISOLATE-CHAIN exists", func() {
		BeforeEach(func() {
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(true, nil)
		})

		When("insert set to true", func() {
			BeforeEach(func() {
				// Expect the DROP rule to be inserted for packets from and to the same bridge
				mockIpt.EXPECT().Insert("filter", "FINCH-ISOLATE-CHAIN", 1, "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(nil)
			})

			It("should add the DROP rule to the FINCH-ISOLATE-CHAIN", func() {
				err := driver.DisableICC(bridgeIface, true)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("insert set to false", func() {
			BeforeEach(func() {
				// Expect the DROP rule to be removed for packets from and to the same bridge
				mockIpt.EXPECT().DeleteIfExists("filter", "FINCH-ISOLATE-CHAIN", "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(nil)
			})

			It("should remove the DROP rule from the FINCH-ISOLATE-CHAIN", func() {
				err := driver.DisableICC(bridgeIface, false)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Context("when iptables returns an error", func() {
		It("should return an error if ChainExists fails", func() {
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(false, fmt.Errorf("iptables error"))

			err := driver.DisableICC(bridgeIface, true)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to check if FINCH-ISOLATE-CHAIN chain exists"))
		})

		It("should return an error if NewChain fails", func() {
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(false, nil)

			mockIpt.EXPECT().NewChain("filter", "FINCH-ISOLATE-CHAIN").Return(fmt.Errorf("iptables error"))

			err := driver.DisableICC(bridgeIface, true)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create FINCH-ISOLATE-CHAIN chain"))
		})

		It("should return an error if Insert fails while adding DROP rule", func() {
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(true, nil)

			mockIpt.EXPECT().Insert("filter", "FINCH-ISOLATE-CHAIN", 1, "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(fmt.Errorf("iptables error"))

			err := driver.DisableICC(bridgeIface, true)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to add DROP rule"))
		})

		It("should return an error if Delete fails while removing DROP rule", func() {
			mockIpt.EXPECT().ChainExists("filter", "FINCH-ISOLATE-CHAIN").Return(true, nil)

			// Simulate an error while removing the DROP rule
			mockIpt.EXPECT().DeleteIfExists("filter", "FINCH-ISOLATE-CHAIN", "-i", bridgeIface, "-o", bridgeIface, "-j", "DROP").Return(fmt.Errorf("iptables error"))

			err := driver.DisableICC(bridgeIface, false)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to remove DROP rule"))
		})
	})
})
