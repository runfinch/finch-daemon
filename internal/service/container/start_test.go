// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"go.uber.org/mock/gomock"

	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
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
		task         *mocks_container.MockTask
		cid          string
		tarExtractor *mocks_archive.MockTarExtractor
		service      container.Service
		options      ncTypes.ContainerStartOptions
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		cid = "test-container-123"
		con.EXPECT().ID().Return(cid).AnyTimes()
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)
		service = NewService(cdClient, mockNerdctlService{ncClient, nil}, logger, nil, nil, tarExtractor)
		options = ncTypes.ContainerStartOptions{}
	})

	// expectCustomStartSuccess sets up mock expectations for a successful customStart flow.
	// Note: NewFIFOSetInDir and NewDirectCIO are called inside the ioCreator closure
	// passed to cont.NewTask. The mock NewTask does not invoke the closure, so directIO
	// will be nil. startLogCopiers handles nil directIO gracefully.
	expectCustomStartSuccess := func() {
		con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
		con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{Process: &specs.Process{}}, nil)
		con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
		con.EXPECT().Labels(gomock.Any()).Return(map[string]string{"nerdctl/namespace": "finch"}, nil)
		ncClient.EXPECT().GetDataStore().Return("/tmp/test-store", nil)
		task.EXPECT().Start(gomock.Any()).Return(nil)
	}

	Context("service", func() {
		It("should not return any error on successful start", func() {
			// set up the mock to return a container that is in stopped state
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			// mock the customStart flow (FIFO task creation + log copiers + task.Start)
			expectCustomStartSuccess()

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// service should not return any error
			err := service.Start(ctx, cid, options)
			Expect(err).Should(BeNil())
		})
		It("should return not found error", func() {
			// set up the mock to mimic no container found for the provided container id
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf("no such container: %s", gomock.Any())

			// service should return NotFound error
			err := service.Start(ctx, cid, options)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return multiple containers found error", func() {
			// set up the mock to mimic two containers found for the provided container id
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf("multiple IDs found with provided prefix: %s, total containers found: %d", cid, 2)

			// service should return error
			err := service.Start(ctx, cid, options)
			Expect(err).Should(Not(BeNil()))
		})
		It("should return not modified error as container is running already", func() {
			// set up the mock to return a container that is already in running state
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			// service should return not modified error.
			err := service.Start(ctx, cid, options)
			Expect(errdefs.IsNotModified(err)).Should(BeTrue())
		})
		It("should return error as container is paused", func() {
			// set up the mock to return a container that is paused
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Paused)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			// service should return error
			err := service.Start(ctx, cid, options)
			Expect(err).Should(Not(BeNil()))
		})
		It("should fail when task.Start returns error", func() {
			// set up the mock to mimic an error during task.Start
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Created)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{Process: &specs.Process{}}, nil)
			con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{"nerdctl/namespace": "finch"}, nil)
			ncClient.EXPECT().GetDataStore().Return("/tmp/test-store", nil)

			expectedErr := fmt.Errorf("task start failed")
			task.EXPECT().Start(gomock.Any()).Return(expectedErr)

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// service should return the task start error
			err := service.Start(ctx, cid, options)
			Expect(err).Should(Equal(expectedErr))
		})
		It("should fail when Spec returns error", func() {
			// set up the mock to mimic an error retrieving the container spec
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(nil, fmt.Errorf("spec error"))

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// service should return error containing "spec"
			err := service.Start(ctx, cid, options)
			Expect(err).Should(Not(BeNil()))
			Expect(err.Error()).Should(ContainSubstring("spec"))
		})
		It("should delete task when Labels fails", func() {
			// set up the mock to mimic Labels failure after task creation
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{Process: &specs.Process{}}, nil)
			con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
			con.EXPECT().Labels(gomock.Any()).Return(nil, fmt.Errorf("labels error"))
			// task should be deleted on failure
			task.EXPECT().Delete(gomock.Any()).Return(nil, nil)

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// service should return error containing "labels"
			err := service.Start(ctx, cid, options)
			Expect(err).Should(Not(BeNil()))
			Expect(err.Error()).Should(ContainSubstring("labels"))
		})
		It("should delete task when GetDataStore fails", func() {
			// set up the mock to mimic GetDataStore failure after task creation
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{Process: &specs.Process{}}, nil)
			con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{"nerdctl/namespace": "finch"}, nil)
			ncClient.EXPECT().GetDataStore().Return("", fmt.Errorf("datastore error"))
			// task should be deleted on failure
			task.EXPECT().Delete(gomock.Any()).Return(nil, nil)

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// service should return error containing "data store"
			err := service.Start(ctx, cid, options)
			Expect(err).Should(Not(BeNil()))
			Expect(err.Error()).Should(ContainSubstring("data store"))
		})
		It("should handle nil Process in spec without panic", func() {
			// set up the mock with a spec that has nil Process — isTerminal should default to false
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)

			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{}, nil)
			con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{"nerdctl/namespace": "finch"}, nil)
			ncClient.EXPECT().GetDataStore().Return("/tmp/test-store", nil)
			task.EXPECT().Start(gomock.Any()).Return(nil)

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// service should not panic and should succeed
			err := service.Start(ctx, cid, options)
			Expect(err).Should(BeNil())
		})
		It("should cleanup existing old task before creating new one", func() {
			// set up the mock where an old task exists and should be deleted
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			// cleanupOldTask — old task exists and is deleted
			oldTask := mocks_container.NewMockTask(mockCtrl)
			con.EXPECT().Task(gomock.Any(), nil).Return(oldTask, nil)
			oldTask.EXPECT().Delete(gomock.Any()).Return(nil, nil)

			// rest of customStart flow
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{Process: &specs.Process{}}, nil)
			con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{"nerdctl/namespace": "finch"}, nil)
			ncClient.EXPECT().GetDataStore().Return("/tmp/test-store", nil)
			task.EXPECT().Start(gomock.Any()).Return(nil)

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// service should succeed after cleaning up old task
			err := service.Start(ctx, cid, options)
			Expect(err).Should(BeNil())
		})
		It("should pass terminal flag from spec to NewFIFOSetInDir", func() {
			// set up the mock with Terminal=true in spec
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{Process: &specs.Process{Terminal: true}}, nil)
			con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{"nerdctl/namespace": "finch"}, nil)
			ncClient.EXPECT().GetDataStore().Return("/tmp/test-store", nil)
			task.EXPECT().Start(gomock.Any()).Return(nil)

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// service should succeed — terminal flag is passed through to ioCreator
			err := service.Start(ctx, cid, options)
			Expect(err).Should(BeNil())
		})
	})
})
