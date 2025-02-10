// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to container restart API.
var _ = Describe("Container Restart API ", func() {
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
		timeout      time.Duration
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
		timeout = time.Duration(10)
	})
	Context("service", func() {
		It("should not return any error when Running", func() {
			// set up the mock to return a container that is in running state
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil).AnyTimes()
			//mock the nerdctl client to mock the restart container was successful without any error.
			ncClient.EXPECT().StartContainer(ctx, con).Return(nil)
			ncClient.EXPECT().StopContainer(ctx, con, &timeout).Return(nil)
			gomock.InOrder(
				logger.EXPECT().Debugf("restarting container: %s", cid),
				logger.EXPECT().Debugf("successfully restarted: %s", cid),
			)
			//service should not return any error
			err := service.Restart(ctx, cid, timeout)
			Expect(err).Should(BeNil())
		})
		It("should not return any error when Stopped", func() {
			// set up the mock to return a container that is in running state
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil).AnyTimes()
			//mock the nerdctl client to mock the restart container was successful without any error.
			ncClient.EXPECT().StopContainer(ctx, con, &timeout).Return(errdefs.NewNotModified(fmt.Errorf("err")))
			ncClient.EXPECT().StartContainer(ctx, con).Return(nil)
			gomock.InOrder(
				logger.EXPECT().Debugf("restarting container: %s", cid),
				logger.EXPECT().Debugf("successfully restarted: %s", cid),
			)
			//service should not return any error
			err := service.Restart(ctx, cid, timeout)
			Expect(err).Should(BeNil())
		})
		It("should return not found error", func() {
			// set up the mock to mimic no container found for the provided container id
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf("no such container: %s", gomock.Any())

			// service should return NotFound error
			err := service.Restart(ctx, cid, timeout)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return multiple containers found error", func() {
			// set up the mock to mimic two containers found for the provided container id
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf("multiple IDs found with provided prefix: %s, total containers found: %d", cid, 2)

			// service should return error
			err := service.Restart(ctx, cid, timeout)
			Expect(err).Should(Not(BeNil()))
		})
		It("should fail due to nerdctl client error", func() {
			// set up the mock to mimic an error occurred while starting the container using nerdctl function.
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil).AnyTimes()

			expectedErr := fmt.Errorf("nerdctl error")
			ncClient.EXPECT().StopContainer(ctx, con, &timeout).Return(nil)
			ncClient.EXPECT().StartContainer(ctx, con).Return(expectedErr)
			gomock.InOrder(
				logger.EXPECT().Debugf("restarting container: %s", cid),
			)
			logger.EXPECT().Errorf("Failed to start container: %s. Error: %v", cid, expectedErr)

			// service should return not modified error.
			err := service.Restart(ctx, cid, timeout)
			Expect(err).Should(Equal(expectedErr))
		})
	})
})
