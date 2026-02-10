// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/go-cni"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"go.uber.org/mock/gomock"

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

			// make Labels() return an error to skip updateContainerMetadata
			con.EXPECT().Labels(ctx).Return(nil, errors.New("mock error"))

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
	Context("updateContainerMetadata", func() {
		It("should successfully update container metadata", func() {
			createOpt := types.ContainerCreateOptions{
				NerdctlCmd: ncExe,
				GOptions: types.GlobalCommandOptions{
					DataRoot: "/tmp/test",
					Address:  "/run/containerd/containerd.sock",
				},
			}
			netOpt := types.NetworkOptions{
				PortMappings: []cni.PortMapping{
					{ContainerPort: 80, HostPort: 8080, Protocol: "tcp"},
				},
			}

			// Mock container expectations
			con.EXPECT().Labels(ctx).Return(map[string]string{}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{Annotations: map[string]string{}}, nil)
			con.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).Return(nil)

			err := updateContainerMetadata(ctx, createOpt, netOpt, con)
			Expect(err).Should(BeNil())
		})

		It("should return error when Labels() fails", func() {
			createOpt := types.ContainerCreateOptions{}
			netOpt := types.NetworkOptions{}
			mockErr := errors.New("failed to get labels")

			con.EXPECT().Labels(ctx).Return(nil, mockErr)

			err := updateContainerMetadata(ctx, createOpt, netOpt, con)
			Expect(err).Should(Equal(mockErr))
		})

		It("should return error when Spec() fails", func() {
			createOpt := types.ContainerCreateOptions{}
			netOpt := types.NetworkOptions{}
			mockErr := errors.New("failed to get spec")

			con.EXPECT().Labels(ctx).Return(map[string]string{}, nil)
			con.EXPECT().Spec(ctx).Return(nil, mockErr)

			err := updateContainerMetadata(ctx, createOpt, netOpt, con)
			Expect(err).Should(Equal(mockErr))
		})
	})
})
