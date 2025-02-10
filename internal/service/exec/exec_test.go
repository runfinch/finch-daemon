// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"errors"
	"testing"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// TestExecService is the entry point of exec service package's unit tests using ginkgo.
func TestExecService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Exec APIs Service")
}

var _ = Describe("Exec API service common ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		con      *mocks_container.MockContainer
		task     *mocks_container.MockTask
		proc     *mocks_container.MockProcess
		cid      string
		execId   string
		s        service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		cid = "123"
		execId = "exec-123"
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		proc = mocks_container.NewMockProcess(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()
		s = service{
			client: cdClient,
			logger: logger,
		}
	})
	Context("getContainer", func() {
		It("should return the container object if it was found", func() {
			// search method returns one container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(Equal(con))
			Expect(err).Should(BeNil())
		})
		It("should return an error if search container method fails", func() {
			// search method returns no container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				nil, errors.New("search container error"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
		It("should return NotFound error if no container was found", func() {
			// search method returns no container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return an error if multiple containers were found", func() {
			// search method returns two containers
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
	})
	Context("loadExecInstance", func() {
		It("should return the exec instance if found", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().LoadProcess(ctx, execId, nil).Return(proc, nil)

			Expect(s.loadExecInstance(ctx, cid, execId, nil)).Should(Equal(&execInstance{
				Container: con,
				Task:      task,
				Process:   proc,
			}))
		})
		It("should use the provided attach to attach to the process", func() {
			attach := cio.NewAttach()

			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			// we can't check equality because cio.Attach is a function type, which can't be compared. non-nil should
			// be sufficient, though, as the function either uses the provided attach or nil
			task.EXPECT().LoadProcess(ctx, execId, gomock.Not(nil)).Return(proc, nil)

			Expect(s.loadExecInstance(ctx, cid, execId, attach)).Should(Equal(&execInstance{
				Container: con,
				Task:      task,
				Process:   proc,
			}))
		})
		It("should return a NotFound error if the container is not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			result, err := s.loadExecInstance(ctx, cid, execId, nil)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(HavePrefix("container not found:"))
			Expect(result).Should(BeNil())
		})
		It("should pass through other errors from getContainer", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{}, errors.New("getContainer error"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())

			result, err := s.loadExecInstance(ctx, cid, execId, nil)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("getContainer error"))
			Expect(result).Should(BeNil())
		})
		It("should return a NotFound error if the task is not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, cerrdefs.ErrNotFound)

			result, err := s.loadExecInstance(ctx, cid, execId, nil)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(HavePrefix("task not found:"))
			Expect(result).Should(BeNil())
		})
		It("should pass through other errors from con.Task", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, errors.New("task error"))

			result, err := s.loadExecInstance(ctx, cid, execId, nil)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("task error"))
			Expect(result).Should(BeNil())
		})
		It("should return a NotFound error if the process is not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().LoadProcess(ctx, execId, nil).Return(nil, cerrdefs.ErrNotFound)

			result, err := s.loadExecInstance(ctx, cid, execId, nil)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(HavePrefix("process not found:"))
			Expect(result).Should(BeNil())
		})
		It("should pass through other errors from task.Process", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().LoadProcess(ctx, execId, nil).Return(nil, errors.New("process error"))

			result, err := s.loadExecInstance(ctx, cid, execId, nil)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("process error"))
			Expect(result).Should(BeNil())
		})
	})
})
