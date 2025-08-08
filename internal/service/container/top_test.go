// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Top API", func() {
	var (
		ctx            context.Context
		mockCtrl       *gomock.Controller
		logger         *mocks_logger.Logger
		cdClient       *mocks_backend.MockContainerdClient
		ncContainerSvc *mocks_backend.MockNerdctlContainerSvc
		ncNetworkSvc   *mocks_backend.MockNerdctlNetworkSvc
		svc            *service
		cid            string
		topOptions     ncTypes.ContainerTopOptions
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
		topOptions = ncTypes.ContainerTopOptions{
			PsArgs: "-ef",
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

	Context("Top API", func() {
		It("should successfully get top processes of a container", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			ncContainerSvc.EXPECT().ContainerTop(ctx, cid, topOptions).Return(nil)

			err := svc.Top(ctx, cid, topOptions)
			Expect(err).Should(BeNil())
		})

		It("should return NotFound error if container is not found", func() {
			mockErr := cerrdefs.ErrNotFound.WithMessage(fmt.Sprintf("no such container: %s", cid))
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(nil, mockErr)
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())

			err := svc.Top(ctx, cid, topOptions)
			Expect(err.Error()).Should(Equal(errdefs.NewNotFound(fmt.Errorf("no such container: %s", cid)).Error()))
		})

		It("should return error from ContainerTop if it fails", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			mockErr := errors.New("failed to get container processes")
			ncContainerSvc.EXPECT().ContainerTop(ctx, cid, topOptions).Return(mockErr)

			err := svc.Top(ctx, cid, topOptions)
			Expect(err).Should(Equal(mockErr))
		})

		It("should return error if SearchContainer fails with non-NotFound error", func() {
			mockErr := errors.New("failed to search container")
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(nil, mockErr)
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())

			err := svc.Top(ctx, cid, topOptions)
			Expect(err).Should(Equal(mockErr))
		})
	})
})
