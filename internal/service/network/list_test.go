// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/containerd/nerdctl/pkg/netutil"
	"github.com/containernetworking/cni/libcni"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/network"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

var _ = Describe("Network List API ", func() {
	var (
		ctx         context.Context
		mockCtrl    *gomock.Controller
		cdClient    *mocks_backend.MockContainerdClient
		ncNetClient *mocks_backend.MockNerdctlNetworkSvc
		logger      *mocks_logger.Logger
		service     network.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncNetClient = mocks_backend.NewMockNerdctlNetworkSvc(mockCtrl)
		logger = mocks_logger.NewLogger(mockCtrl)
		service = NewService(cdClient, ncNetClient, logger)
	})
	Context("service", func() {
		It("should return 0 networks when nothing is found", func() {
			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return(nil, nil)

			resp, err := service.List(ctx)
			Expect(err).Should(BeNil())
			Expect(len(resp)).Should(Equal(0))
		})
		It("should pass through errors from FilterNetworks", func() {
			expErr := "filter network error"
			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return(nil, fmt.Errorf("%s", expErr))

			_, err := service.List(ctx)
			Expect(err.Error()).Should(Equal(expErr))
		})
		It("should return the found networks", func() {
			expNetName := "testnet"
			expNetID := "abcdefg"

			expNetList := make([]*netutil.NetworkConfig, 1)
			expNetList[0] = &netutil.NetworkConfig{
				NetworkConfigList: &libcni.NetworkConfigList{Name: expNetName},
				NerdctlID:         &expNetID,
			}
			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return(expNetList, nil)

			resp, err := service.List(ctx)
			Expect(err).Should(BeNil())
			Expect(len(resp)).Should(Equal(1))
			Expect(resp[0].Name).Should(Equal(expNetName))
			Expect(resp[0].ID).Should(Equal(expNetID))
		})
	})
})
