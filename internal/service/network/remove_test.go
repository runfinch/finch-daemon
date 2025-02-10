// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/containernetworking/cni/libcni"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/network"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Network Remove API ", func() {
	var (
		ctx               context.Context
		mockCtrl          *gomock.Controller
		cdClient          *mocks_backend.MockContainerdClient
		ncNetClient       *mocks_backend.MockNerdctlNetworkSvc
		logger            *mocks_logger.Logger
		service           network.Service
		networkId         string
		networkName       string
		file              string
		mockNetworkConfig *netutil.NetworkConfig
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncNetClient = mocks_backend.NewMockNerdctlNetworkSvc(mockCtrl)
		logger = mocks_logger.NewLogger(mockCtrl)
		service = NewService(cdClient, ncNetClient, logger)
		// initialize dummy values
		networkId = "123"
		networkName = "test"
		file = "someFile"
		mockNetworkConfig = &netutil.NetworkConfig{
			NetworkConfigList: &libcni.NetworkConfigList{
				Name: networkName,
			},
			NerdctlID: &networkId,
			File:      file,
		}
	})
	Context("service", func() {
		It("should not return any error", func() {
			logger.EXPECT().Infof("network delete: network Id %s", networkId)
			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{mockNetworkConfig}, nil)
			ncNetClient.EXPECT().UsedNetworkInfo(gomock.Any()).Return(make(map[string][]string), nil)
			ncNetClient.EXPECT().RemoveNetwork(mockNetworkConfig).Return(nil)
			err := service.Remove(ctx, networkId)
			Expect(err).Should(BeNil())
		})
		It("should return forbidden error when network is used by a container", func() {
			logger.EXPECT().Infof("network delete: network Id %s", networkId)
			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{mockNetworkConfig}, nil)
			// Map literal to represent network "test" used by container "container"
			u := map[string][]string{
				"test": {"container"},
			}
			ncNetClient.EXPECT().UsedNetworkInfo(gomock.Any()).Return(u, nil)
			err := service.Remove(ctx, networkId)
			Expect(err.Error()).Should(Equal(fmt.Errorf("network %q is in use by container %q", networkId, u[networkName]).Error()))
			Expect(errdefs.IsForbiddenError(err)).Should(Equal(true))
		})
		It("should return forbidden error when attempting to remove predefined networks", func() {
			logger.EXPECT().Infof("network delete: network Id %s", networkId)
			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{mockNetworkConfig}, nil)
			ncNetClient.EXPECT().UsedNetworkInfo(gomock.Any()).Return(make(map[string][]string), nil)
			mockNetworkConfig.File = ""
			err := service.Remove(ctx, networkId)
			Expect(err.Error()).Should(Equal(fmt.Errorf("%s is a pre-defined network and cannot be removed", networkId).Error()))
			Expect(errdefs.IsForbiddenError(err)).Should(Equal(true))
		})
	})
})
