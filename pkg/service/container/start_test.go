// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/pkg/api/handlers/container"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/pkg/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/pkg/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/pkg/mocks/mocks_logger"
)

// Unit tests related to container start API.
var _ = Describe("Container Start API ", func() {
	var (
		ctx          context.Context
		mockCtrl     *gomock.Controller
		logger       *mocks_logger.Logger
		cdClient     *mocks_backend.MockContainerdClient
		ncClient     *mocks_backend.MockNerdctlContainerSvc
		con          *mocks_container.MockContainer
		cid          string
		tarExtractor *mocks_archive.MockTarExtractor
		service      container.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)
		service = NewService(cdClient, mockNerdctlService{ncClient, nil}, logger, nil, nil, tarExtractor)
	})
	Context("service", func() {
		It("should not return any error", func() {
			// set up the mock to return a container that is in running state
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			ncClient.EXPECT().StartContainer(ctx, con).Return(nil)
			logger.EXPECT().Debugf("starting container: %s", cid)
			logger.EXPECT().Debugf("successfully started: %s", cid)

			err := service.Start(ctx, cid)
			Expect(err).Should(BeNil())
		})
		It("should return not found error", func() {
			// set up the mock to mimic no container found for the provided container id
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf("no such container: %s", gomock.Any())

			// service should return NotFound error
			err := service.Start(ctx, cid)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return multiple containers found error", func() {
			// set up the mock to mimic two containers found for the provided container id
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf("multiple IDs found with provided prefix: %s, total containers found: %d", cid, 2)

			// service should return error
			err := service.Start(ctx, cid)
			Expect(err).Should(Not(BeNil()))
		})
		It("should return not modified error as container is running already", func() {
			// set up the mock to return a container that is already in running state
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			// service should return not modified error.
			err := service.Start(ctx, cid)
			Expect(errdefs.IsNotModified(err)).Should(BeTrue())
		})
		It("should return not modified error as container is paused", func() {
			// set up the mock to return a container that is not running
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Paused)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			// service should return not modified error.
			err := service.Start(ctx, cid)
			Expect(err).Should(Not(BeNil()))
		})
		It("should fail due to nerdctl client error", func() {
			// set up the mock to mimic an error occurred while starting the container using nerdctl function.
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Created)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			expectedErr := fmt.Errorf("nerdctl error")
			ncClient.EXPECT().StartContainer(ctx, con).Return(expectedErr)
			logger.EXPECT().Errorf("Failed to start container: %s. Error: %v", cid, expectedErr)
			logger.EXPECT().Debugf("starting container: %s", cid)

			// service should return not modified error.
			err := service.Start(ctx, cid)
			Expect(err).Should(Equal(expectedErr))
		})
	})
})
