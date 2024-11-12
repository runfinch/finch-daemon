// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/netutil"
	"github.com/containernetworking/cni/libcni"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to container create API.
var _ = Describe("Container Create API ", func() {
	var (
		ctx            context.Context
		mockCtrl       *gomock.Controller
		logger         *mocks_logger.Logger
		cdClient       *mocks_backend.MockContainerdClient
		ncContainerSvc *mocks_backend.MockNerdctlContainerSvc
		ncNetworkSvc   *mocks_backend.MockNerdctlNetworkSvc
		ncExe          string
		image          string
		cmd            []string
		createOpt      types.ContainerCreateOptions
		createOptExp   types.ContainerCreateOptions
		netOpt         types.NetworkOptions
		netManager     *mocks_container.MockNetworkOptionsManager
		con            *mocks_container.MockContainer
		cid            string
		svc            *service
		tarExtractor   *mocks_archive.MockTarExtractor
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncContainerSvc = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		ncNetworkSvc = mocks_backend.NewMockNerdctlNetworkSvc(mockCtrl)
		ncExe = "/usr/local/bin/nerdctl"
		image = "test-image"
		cmd = []string{"echo", "hello world"}
		createOpt = types.ContainerCreateOptions{}
		createOptExp = types.ContainerCreateOptions{NerdctlCmd: ncExe, NerdctlArgs: []string{}}
		netOpt = types.NetworkOptions{}
		netManager = mocks_container.NewMockNetworkOptionsManager(mockCtrl)
		cid = "test-container-id"
		con = mocks_container.NewMockContainer(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)

		svc = &service{
			client:           cdClient,
			nctlContainerSvc: mockNerdctlService{ncContainerSvc, ncNetworkSvc},
			logger:           logger,
			tarExtractor:     tarExtractor,
		}
	})
	Context("service", func() {
		It("should successfully create a container", func() {
			ncContainerSvc.EXPECT().GetNerdctlExe().Return(ncExe, nil)
			ncContainerSvc.EXPECT().NewNetworkingOptionsManager(netOpt).Return(netManager, nil)

			// container create arguments
			args := []string{image}
			args = append(args, cmd...)
			ncContainerSvc.EXPECT().CreateContainer(ctx, args, netManager, createOptExp).Return(
				con, nil, nil)

			// service should not return any error and the returned cid should match expected
			cidResult, err := svc.Create(ctx, image, cmd, createOpt, netOpt)
			Expect(cidResult).Should(Equal(cid))
			Expect(err).Should(BeNil())
		})
		It("should return internal error for network options create failure", func() {
			mockErr := errors.New("error while creating networking options")
			ncContainerSvc.EXPECT().GetNerdctlExe().Return(ncExe, nil)
			ncContainerSvc.EXPECT().NewNetworkingOptionsManager(gomock.Any()).Return(nil, mockErr)

			// service should return with an error
			cidResult, err := svc.Create(ctx, image, nil, createOpt, netOpt)
			Expect(cidResult).Should(BeEmpty())
			Expect(err.Error()).Should(Equal(mockErr.Error()))
		})
		It("should return internal error for container create failure", func() {
			ncContainerSvc.EXPECT().GetNerdctlExe().Return(ncExe, nil)
			ncContainerSvc.EXPECT().NewNetworkingOptionsManager(netOpt).Return(netManager, nil)

			// container create arguments
			args := []string{image}
			args = append(args, cmd...)

			mockErr := errors.New("error while creating a container")
			ncContainerSvc.EXPECT().CreateContainer(ctx, args, netManager, createOptExp).Return(
				nil, nil, mockErr)

			// service should return with an error
			cidResult, err := svc.Create(ctx, image, cmd, createOpt, netOpt)
			Expect(cidResult).Should(BeEmpty())
			Expect(err.Error()).Should(Equal(mockErr.Error()))
		})
		It("should call garbage collector upon container create failure", func() {
			ncContainerSvc.EXPECT().GetNerdctlExe().Return(ncExe, nil)
			ncContainerSvc.EXPECT().NewNetworkingOptionsManager(netOpt).Return(netManager, nil)

			// container create arguments
			args := []string{image}
			args = append(args, cmd...)

			// define mock garbage cleanup method
			gcFlag := false
			mockGc := func() {
				gcFlag = true
			}

			mockErr := errors.New("error while creating a container")
			ncContainerSvc.EXPECT().CreateContainer(ctx, args, netManager, createOptExp).Return(
				nil, mockGc, mockErr)

			// service should call garbage collector and return with an error
			cidResult, err := svc.Create(ctx, image, cmd, createOpt, netOpt)
			Expect(cidResult).Should(BeEmpty())
			Expect(gcFlag).Should(BeTrue())
			Expect(err.Error()).Should(Equal(mockErr.Error()))
		})
		It("should return not-found error if image was not found", func() {
			ncContainerSvc.EXPECT().GetNerdctlExe().Return(ncExe, nil)
			ncContainerSvc.EXPECT().NewNetworkingOptionsManager(netOpt).Return(netManager, nil)

			// container create arguments
			args := []string{image}
			args = append(args, cmd...)

			ncContainerSvc.EXPECT().CreateContainer(ctx, args, netManager, createOptExp).Return(
				nil, nil, cerrdefs.ErrNotFound)

			// service should return with an error
			cidResult, err := svc.Create(ctx, image, cmd, createOpt, netOpt)
			Expect(cidResult).Should(BeEmpty())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return invalid-format error if the inputs are invalid", func() {
			ncContainerSvc.EXPECT().GetNerdctlExe().Return(ncExe, nil)
			ncContainerSvc.EXPECT().NewNetworkingOptionsManager(netOpt).Return(netManager, nil)

			// container create arguments
			args := []string{image}
			args = append(args, cmd...)

			ncContainerSvc.EXPECT().CreateContainer(ctx, args, netManager, createOptExp).Return(
				nil, nil, cerrdefs.ErrInvalidArgument)

			// service should return with an error
			cidResult, err := svc.Create(ctx, image, cmd, createOpt, netOpt)
			Expect(cidResult).Should(BeEmpty())
			Expect(errdefs.IsInvalidFormat(err)).Should(BeTrue())
		})
		It("should return conflict error if container name already exists", func() {
			ncContainerSvc.EXPECT().GetNerdctlExe().Return(ncExe, nil)
			ncContainerSvc.EXPECT().NewNetworkingOptionsManager(netOpt).Return(netManager, nil)

			// container create arguments
			args := []string{image}
			args = append(args, cmd...)

			ncContainerSvc.EXPECT().CreateContainer(ctx, args, netManager, createOptExp).Return(
				nil, nil, cerrdefs.ErrAlreadyExists)

			// service should return with an error
			cidResult, err := svc.Create(ctx, image, cmd, createOpt, netOpt)
			Expect(cidResult).Should(BeEmpty())
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
		})
		It("should return an error if nerdctl binary was not found", func() {
			mockErr := errors.New("could not find nerdctl binary")
			ncContainerSvc.EXPECT().GetNerdctlExe().Return("", mockErr)

			// service should return with an error
			cidResult, err := svc.Create(ctx, image, nil, createOpt, netOpt)
			Expect(cidResult).Should(BeEmpty())
			Expect(err.Error()).Should(ContainSubstring(mockErr.Error()))
		})
	})
	Context("translate network IDs", func() {
		It("should translate network ids to network names for specified networks", func() {
			// network options
			netIds := []string{"network-id1", "network-id2"}
			netNames := []string{"network1", "network2"}
			netOpt := types.NetworkOptions{
				NetworkSlice: netIds,
			}

			// FilterNetworks returns a single network config for each network id
			ncNetworkSvc.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{
				{NetworkConfigList: &libcni.NetworkConfigList{Name: netNames[0]}},
			}, nil)
			ncNetworkSvc.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{
				{NetworkConfigList: &libcni.NetworkConfigList{Name: netNames[1]}},
			}, nil)

			// network ids should be translated to corresponding names without error
			err := svc.translateNetworkIds(&netOpt)
			Expect(err).Should(BeNil())
			Expect(netOpt.NetworkSlice).Should(Equal(netNames))
		})
		It("should ignore host, none, and bridge networks for network id translation", func() {
			// network options
			netIds := []string{"network-id1", "bridge", "host", "network-id2", "none"}
			netNames := []string{"network1", "bridge", "host", "network2", "none"}
			netOpt := types.NetworkOptions{
				NetworkSlice: netIds,
			}

			// FilterNetworks returns a single network config for each network id
			// but should not be called for bridge, host, and none networks
			ncNetworkSvc.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{
				{NetworkConfigList: &libcni.NetworkConfigList{Name: netNames[0]}},
			}, nil)
			ncNetworkSvc.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{
				{NetworkConfigList: &libcni.NetworkConfigList{Name: netNames[3]}},
			}, nil)

			// network ids should be translated to corresponding names without error
			err := svc.translateNetworkIds(&netOpt)
			Expect(err).Should(BeNil())
			Expect(netOpt.NetworkSlice).Should(Equal(netNames))
		})
		It("should return an error if filter networks failed", func() {
			mockErr := errors.New("filter networks failure")

			// network options
			netOpt := types.NetworkOptions{
				NetworkSlice: []string{"test-network-id"},
			}

			// FilterNetworks returns an error
			ncNetworkSvc.EXPECT().FilterNetworks(gomock.Any()).Return(nil, mockErr)

			// function should propagate the error from FilterNetworks
			err := svc.translateNetworkIds(&netOpt)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(ContainSubstring(mockErr.Error()))
		})
		It("should return an error if network was not found", func() {
			// network options
			netOpt := types.NetworkOptions{
				NetworkSlice: []string{"test-network-id"},
			}

			// FilterNetworks returns 0 networks
			ncNetworkSvc.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{}, nil)

			// function should return a not found error
			err := svc.translateNetworkIds(&netOpt)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return an error if multiple networks are found for the same id", func() {
			// network options
			netOpt := types.NetworkOptions{
				NetworkSlice: []string{"test-network-id"},
			}

			// FilterNetworks returns 2 networks
			ncNetworkSvc.EXPECT().FilterNetworks(gomock.Any()).Return([]*netutil.NetworkConfig{
				{NetworkConfigList: &libcni.NetworkConfigList{Name: "network1"}},
				{NetworkConfigList: &libcni.NetworkConfigList{Name: "network2"}},
			}, nil)

			// function should return an error
			err := svc.translateNetworkIds(&netOpt)
			Expect(err).ShouldNot(BeNil())
		})
	})
})
