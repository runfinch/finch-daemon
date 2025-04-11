// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/mock/gomock"
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

var _ = Describe("Container Unpause API", func() {
	var (
		ctx            context.Context
		mockCtrl       *gomock.Controller
		logger         *mocks_logger.Logger
		cdClient       *mocks_backend.MockContainerdClient
		ncContainerSvc *mocks_backend.MockNerdctlContainerSvc
		ncNetworkSvc   *mocks_backend.MockNerdctlNetworkSvc
		svc            *service
		cid            string
		con            *mocks_container.MockContainer
		unpauseOptions ncTypes.ContainerUnpauseOptions
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncContainerSvc = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		ncNetworkSvc = mocks_backend.NewMockNerdctlNetworkSvc(mockCtrl)

		cid = "test-container-id"
		unpauseOptions = ncTypes.ContainerUnpauseOptions{}
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

	Context("Unpause API", func() {
		It("should successfully unpause a paused container", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Paused)
			ncContainerSvc.EXPECT().UnpauseContainer(ctx, cid, unpauseOptions).Return(nil)

			err := svc.Unpause(ctx, cid, unpauseOptions)
			Expect(err).Should(BeNil())
		})

		It("should return NotFound error if container is not found", func() {
			mockErr := cerrdefs.ErrNotFound.WithMessage(fmt.Sprintf("no such container: %s", cid))
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(nil, mockErr)
			logger.EXPECT().Errorf("failed to search container: %s. error: %s", cid, mockErr.Error())

			err := svc.Unpause(ctx, cid, unpauseOptions)
			Expect(err.Error()).Should(Equal(errdefs.NewNotFound(fmt.Errorf("no such container: %s", cid)).Error()))
		})

		It("should return a Conflict error if container is already running", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)

			err := svc.Unpause(ctx, cid, unpauseOptions)
			Expect(err.Error()).Should(Equal(errdefs.NewConflict(fmt.Errorf("Container %s is not paused", cid)).Error()))
		})

		It("should return a Conflict error if container is not paused", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)

			err := svc.Unpause(ctx, cid, unpauseOptions)
			Expect(err.Error()).Should(Equal(errdefs.NewConflict(fmt.Errorf("Container %s is not paused", cid)).Error()))
		})

		It("should return a generic error if unpause operation fails", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Paused)
			mockErr := errors.New("generic error while unpausing container")
			ncContainerSvc.EXPECT().UnpauseContainer(ctx, cid, unpauseOptions).Return(mockErr)

			err := svc.Unpause(ctx, cid, unpauseOptions)
			Expect(err).Should(Equal(mockErr))
		})
	})
})
