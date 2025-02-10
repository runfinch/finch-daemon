// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/containernetworking/cni/libcni"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/network"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Network Inspect API ", func() {
	var (
		ctx                context.Context
		mockCtrl           *gomock.Controller
		cdClient           *mocks_backend.MockContainerdClient
		ncNetClient        *mocks_backend.MockNerdctlNetworkSvc
		logger             *mocks_logger.Logger
		service            network.Service
		networkId          string
		networkName        string
		mockNetworkConfig  *netutil.NetworkConfig
		mockNetworkInspect *dockercompat.Network
		expNetworkResp     *types.NetworkInspectResponse
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
		mockNetworkConfig = &netutil.NetworkConfig{
			NetworkConfigList: &libcni.NetworkConfigList{
				Name: networkName,
			},
			NerdctlID: &networkId,
		}
		mockNetworkInspect = &dockercompat.Network{
			ID:     networkId,
			Name:   networkName,
			Labels: map[string]string{"testLabel": "testValue"},
			IPAM: dockercompat.IPAM{
				Config: []dockercompat.IPAMConfig{
					{Subnet: "10.5.2.0/24", Gateway: "10.5.2.1"},
				},
			},
		}
		expNetworkResp = &types.NetworkInspectResponse{
			ID:     networkId,
			Name:   networkName,
			Labels: mockNetworkInspect.Labels,
			IPAM:   mockNetworkInspect.IPAM,
		}
	})
	Context("service", func() {
		It("should not return any error", func() {
			logger.EXPECT().Infof("network inspect: network Id %s", networkId)

			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{mockNetworkConfig}, nil)
			ncNetClient.EXPECT().InspectNetwork(gomock.Any(), mockNetworkConfig).Return(mockNetworkInspect, nil)

			resp, err := service.Inspect(ctx, networkId)
			Expect(err).Should(BeNil())
			Expect(resp).Should(Equal(expNetworkResp))
		})
		It("should pass through errors from FilterNetworks", func() {
			logger.EXPECT().Infof("network inspect: network Id %s", networkId)

			mockErr := fmt.Errorf("error from FilterNetworks")
			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, mockErr)
			logger.EXPECT().Errorf("failed to search network: %s. error: %s", networkId, mockErr.Error())
			logger.EXPECT().Debugf("Failed to get network: %s", mockErr)

			resp, err := service.Inspect(ctx, networkId)
			Expect(err).Should(Equal(mockErr))
			Expect(resp).Should(BeNil())
		})
		It("should return a notFound error when no network is found", func() {
			logger.EXPECT().Infof("network inspect: network Id %s", networkId)

			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)
			logger.EXPECT().Debugf("no such network %s", networkId)
			logger.EXPECT().Debugf("Failed to get network: %s", gomock.Any())

			resp, err := service.Inspect(ctx, networkId)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(resp).Should(BeNil())
		})
		It("should return an error when multiple networks are found", func() {
			logger.EXPECT().Infof("network inspect: network Id %s", networkId)

			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{mockNetworkConfig, mockNetworkConfig}, nil)
			logger.EXPECT().Debugf("multiple IDs found with provided prefix: %s, total networks found: %d",
				networkId, 2)
			logger.EXPECT().Debugf("Failed to get network: %s", gomock.Any())

			resp, err := service.Inspect(ctx, networkId)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal(fmt.Sprintf("multiple networks found with ID: %s", networkId)))
			Expect(resp).Should(BeNil())
		})
		It("should return an error when InspectNetwork fails", func() {
			inspectErr := errors.New("network inspect error")
			logger.EXPECT().Infof("network inspect: network Id %s", networkId)

			ncNetClient.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{mockNetworkConfig}, nil)
			ncNetClient.EXPECT().InspectNetwork(gomock.Any(), mockNetworkConfig).Return(nil, inspectErr)
			logger.EXPECT().Debugf("Failed to inspect network: %s", inspectErr)

			resp, err := service.Inspect(ctx, networkId)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal(inspectErr.Error()))
			Expect(resp).Should(BeNil())
		})
	})
})
