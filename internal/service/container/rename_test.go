// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
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

// Unit tests related to container rename API.
var _ = Describe("Container Rename API ", func() {
	var (
		ctx               context.Context
		mockCtrl          *gomock.Controller
		logger            *mocks_logger.Logger
		cdClient          *mocks_backend.MockContainerdClient
		ncClient          *mocks_backend.MockNerdctlContainerSvc
		con               *mocks_container.MockContainer
		cid               string
		tarExtractor      *mocks_archive.MockTarExtractor
		service           container.Service
		testContainerName string
		opts              ncTypes.ContainerRenameOptions
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

		testContainerName = "testContainerName"
		service = NewService(cdClient, mockNerdctlService{ncClient, nil}, logger, nil, nil, tarExtractor)
		opts = ncTypes.ContainerRenameOptions{
			GOptions: ncTypes.GlobalCommandOptions{},
			Stdout:   nil,
		}
	})
	Context("service", func() {
		It("should not return any error", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), testContainerName).Return(
				[]containerd.Container{}, nil)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			ncClient.EXPECT().RenameContainer(ctx, con, testContainerName, gomock.Any())
			logger.EXPECT().Debugf("no such container: %s", testContainerName)
			logger.EXPECT().Debugf("successfully renamed %s to %s", cid, testContainerName)

			err := service.Rename(ctx, cid, testContainerName, opts)
			Expect(err).Should(BeNil())
		})
		It("should return not found error", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, nil).AnyTimes()
			logger.EXPECT().Debugf("no such container: %s", testContainerName)
			logger.EXPECT().Debugf("no such container: %s", cid)

			err := service.Rename(ctx, cid, testContainerName, opts)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return multiple containers found error", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), testContainerName).Return(
				[]containerd.Container{}, nil)
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf("no such container: %s", testContainerName)
			logger.EXPECT().Debugf("multiple IDs found with provided prefix: %s, total containers found: %d", cid, 2)

			// service should return error
			err := service.Rename(ctx, cid, testContainerName, opts)
			Expect(err).Should(Not(BeNil()))
		})
		It("should return conflict error as container name is taken", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), testContainerName).Return(
				[]containerd.Container{con}, nil)

			expectedErr := errdefs.NewConflict(fmt.Errorf("container with name %s already exists", testContainerName))
			logger.EXPECT().Errorf("Failed to rename container: %s. Error: %v", cid, expectedErr)

			// service should return conflict error.
			err := service.Rename(ctx, cid, testContainerName, opts)
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
		})
		It("should fail due to nerdctl client error", func() {
			// set up the mock to mimic an error occurred  while stopping the container using nerdctl function
			cdClient.EXPECT().SearchContainer(gomock.Any(), testContainerName).Return(
				[]containerd.Container{}, nil)
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			expectedErr := fmt.Errorf("nerdctl error")
			ncClient.EXPECT().RenameContainer(ctx, con, testContainerName, gomock.Any()).Return(expectedErr)
			logger.EXPECT().Debugf("no such container: %s", testContainerName)
			logger.EXPECT().Errorf("Failed to rename container: %s. Error: %v", cid, expectedErr)

			err := service.Rename(ctx, cid, testContainerName, opts)
			Expect(err).Should(Equal(expectedErr))
		})
	})
})
