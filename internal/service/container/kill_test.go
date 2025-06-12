// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to the Kill API.
var _ = Describe("Container Kill API", func() {
	var (
		ctx            context.Context
		mockCtrl       *gomock.Controller
		logger         *mocks_logger.Logger
		cdClient       *mocks_backend.MockContainerdClient
		ncContainerSvc *mocks_backend.MockNerdctlContainerSvc
		ncNetworkSvc   *mocks_backend.MockNerdctlNetworkSvc
		svc            *service
		cid            string
		killOptions    ncTypes.ContainerKillOptions
		con            *mocks_container.MockContainer
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncContainerSvc = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		ncNetworkSvc = mocks_backend.NewMockNerdctlNetworkSvc(mockCtrl)

		cid = "test-container-id"
		killOptions = ncTypes.ContainerKillOptions{
			KillSignal: "kill",
		}
		con = mocks_container.NewMockContainer(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()

		svc = &service{
			client:           cdClient,
			nctlContainerSvc: mockNerdctlService{ncContainerSvc, ncNetworkSvc},
			logger:           logger,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("Kill API", func() {
		It("should successfully kill a container", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)
			ncContainerSvc.EXPECT().KillContainer(ctx, cid, killOptions).Return(nil)

			err := svc.Kill(ctx, cid, killOptions)
			Expect(err).Should(BeNil())
		})

		It("should return NotFound error if container is not found", func() {
			mockErr := cerrdefs.ErrNotFound.WithMessage(fmt.Sprintf("no such container: %s", cid))
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(nil, mockErr)
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())

			err := svc.Kill(ctx, cid, killOptions)
			Expect(err.Error()).Should(Equal(errdefs.NewNotFound(fmt.Errorf("no such container: %s", cid)).Error()))
		})

		It("should return a Conflict error if container kill fails with conflict", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)
			// Mock container kill returning a Conflict error
			mockErr := cerrdefs.ErrConflict.WithMessage("conflict while killing container")
			ncContainerSvc.EXPECT().KillContainer(ctx, cid, killOptions).Return(mockErr)

			err := svc.Kill(ctx, cid, killOptions)
			Expect(err.Error()).Should(BeEquivalentTo(errdefs.NewConflict(fmt.Errorf("conflict while killing container")).Error()))
		})

		It("should return a generic error if container kill fails", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)
			mockErr := errors.New("generic error while killing container")
			ncContainerSvc.EXPECT().KillContainer(ctx, cid, killOptions).Return(mockErr)

			err := svc.Kill(ctx, cid, killOptions)
			Expect(err).Should(Equal(mockErr))
		})
	})
})
