// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	ncContainer "github.com/containerd/nerdctl/v2/pkg/cmd/container"
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

// Unit tests related to container remove API.
var _ = Describe("Container Remove API ", func() {
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
		It("should successfully remove the container", func() {
			// set up the mock to return a container with no error while searching for containers
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			// set up the mock to verify the remove container is called and proper error msg was logged
			ncClient.EXPECT().RemoveContainer(ctx, con, false, false)
			logger.EXPECT().Debugf("removing container: %s", cid)

			// service should not return any error
			err := service.Remove(ctx, cid, false, false)
			Expect(err).Should(BeNil())
		})
		It("should return internal error", func() {
			// set up the mock to mimic there was an error while searching for the container
			mockErr := fmt.Errorf("some error occurred during container search")
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, mockErr)
			logger.EXPECT().Errorf("failed to search container: %s. error: %s", cid, mockErr.Error())

			// service should return with an error
			err := service.Remove(ctx, cid, false, false)
			Expect(err.Error()).Should(Equal(mockErr.Error()))
		})
		It("should return no container found", func() {
			// set up the mock to mimic there is no container with the provided container id
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf("no such container: %s", cid)

			// service should return NotFound error
			err := service.Remove(ctx, cid, false, false)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return multiple containers found error", func() {
			// set up the mock to return multiple containers that matches the container id prefix provided by the user
			firstCon := mocks_container.NewMockContainer(mockCtrl)
			firstCon.EXPECT().ID().Return(cid + "_1").AnyTimes()
			secondCon := mocks_container.NewMockContainer(mockCtrl)
			secondCon.EXPECT().ID().Return(cid + "_2").AnyTimes()
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{firstCon, secondCon}, nil)
			logger.EXPECT().Debugf("multiple IDs found with provided prefix: %s, total containers found: %d",
				cid, 2)

			// service should return error
			err := service.Remove(ctx, cid, false, false)
			Expect(err).Should(Not(BeNil()))
		})
		It("should return conflict error as container is running", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			// set up the mock to mimic the container is running
			ncClient.EXPECT().RemoveContainer(gomock.Any(), con, gomock.Any(), gomock.Any()).
				Return(ncContainer.NewStatusError(cid, containerd.Running))
			logger.EXPECT().Debugf("removing container: %s", cid)
			logger.EXPECT().Debugf("Container is in running or pausing state. Failed to remove container: %s", cid)

			// service should return conflict error
			err := service.Remove(ctx, cid, false, false)
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
		})

		It("should return error due to failure in deleting a container", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			// set up the mock to mimic the container delete failed in containerd
			mockErr := fmt.Errorf("some random error to delete")
			ncClient.EXPECT().RemoveContainer(gomock.Any(), con, gomock.Any(), gomock.Any()).
				Return(mockErr)
			logger.EXPECT().Debugf("removing container: %s", cid)
			logger.EXPECT().Errorf("Failed to remove container: %s. Error: %s", con.ID(), mockErr.Error())
			err := service.Remove(ctx, cid, false, false)
			// service should return error
			Expect(err).Should(Not(BeNil()))
		})
	})
})
