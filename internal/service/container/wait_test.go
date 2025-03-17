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

var _ = Describe("Container Wait API", func() {
	var (
		ctx            context.Context
		mockCtrl       *gomock.Controller
		logger         *mocks_logger.Logger
		cdClient       *mocks_backend.MockContainerdClient
		ncContainerSvc *mocks_backend.MockNerdctlContainerSvc
		svc            *service
		cid            string
		waitOptions    ncTypes.ContainerWaitOptions
		con            *mocks_container.MockContainer
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncContainerSvc = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)

		cid = "test-container-id"
		waitOptions = ncTypes.ContainerWaitOptions{}
		con = mocks_container.NewMockContainer(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()

		svc = &service{
			client:           cdClient,
			nctlContainerSvc: mockNerdctlService{ncContainerSvc, nil},
			logger:           logger,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("Wait API", func() {
		It("should successfully wait for a container", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			logger.EXPECT().Debugf("wait container: %s", cid)
			ncContainerSvc.EXPECT().ContainerWait(ctx, cid, waitOptions).Return(nil)

			err := svc.Wait(ctx, cid, waitOptions)
			Expect(err).Should(BeNil())
		})

		It("should return NotFound error if container is not found", func() {
			mockErr := cerrdefs.ErrNotFound.WithMessage(fmt.Sprintf("no such container: %s", cid))
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(nil, mockErr)
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())

			err := svc.Wait(ctx, cid, waitOptions)
			Expect(err.Error()).Should(Equal(errdefs.NewNotFound(fmt.Errorf("no such container: %s", cid)).Error()))
		})

		It("should return an error if ContainerWait fails", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)
			logger.EXPECT().Debugf("wait container: %s", cid)
			mockErr := errors.New("error waiting for container")
			ncContainerSvc.EXPECT().ContainerWait(ctx, cid, waitOptions).Return(mockErr)

			err := svc.Wait(ctx, cid, waitOptions)
			Expect(err).Should(Equal(mockErr))
		})

	})
})
