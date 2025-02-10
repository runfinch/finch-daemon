// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/exec"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Exec Start API ", func() {
	var (
		ctx         context.Context
		mockCtrl    *gomock.Controller
		logger      *mocks_logger.Logger
		cdClient    *mocks_backend.MockContainerdClient
		con         *mocks_container.MockContainer
		task        *mocks_container.MockTask
		proc        *mocks_container.MockProcess
		rw          io.ReadWriteCloser
		success     bool
		successResp func()
		statusC     chan containerd.ExitStatus
		s           exec.Service
		startOpts   *types.ExecStartOptions
	)
	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		proc = mocks_container.NewMockProcess(mockCtrl)
		s = NewService(cdClient, logger)
		rw = &CloseableBuffer{
			bytes.NewBuffer([]byte{}),
		}

		success = false
		successResp = func() {
			success = true
		}
		statusC = make(chan containerd.ExitStatus)
	})
	Context("service", func() {
		Context("detach", func() {
			BeforeEach(func() {
				startOpts = &types.ExecStartOptions{
					ExecStartCheck: &types.ExecStartCheck{
						Detach:      true,
						Tty:         false,
						ConsoleSize: &[2]uint{123, 321},
					},
					ConID:           "123",
					ExecID:          "exec-123",
					Stdin:           rw,
					Stdout:          rw,
					Stderr:          rw,
					SuccessResponse: successResp,
				}
			})
			It("should not return error on success", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Running,
				}, nil)
				proc.EXPECT().Wait(ctx).Return(statusC, nil)
				proc.EXPECT().Start(ctx).Return(nil)

				err := s.Start(ctx, startOpts)
				Expect(err).Should(BeNil())
				Expect(success).Should(BeTrue())
			})
			It("should return a NotFound error if the container is not found", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return(nil, cerrdefs.ErrNotFound)
				logger.EXPECT().Errorf("failed to search container: %s. error: %v", "123", gomock.Any())

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			})
			It("should return a Conflict error if the task is not found", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(nil, cerrdefs.ErrNotFound)

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(errdefs.IsConflict(err)).Should(BeTrue())
			})
			It("should return a Conflict error if the task status cannot be found", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{}, cerrdefs.ErrNotFound)

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(errdefs.IsConflict(err)).Should(BeTrue())
			})
			It("should return a Conflict error if the task status is not running", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Stopped,
				}, nil)

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(errdefs.IsConflict(err)).Should(BeTrue())
			})
			It("should return a NotFound error if the process is not found", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(nil, cerrdefs.ErrNotFound)

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			})
			It("should pass through errors from loadExecInstance", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{}, errors.New("getContainer error"))
				logger.EXPECT().Errorf("failed to search container: %s. error: %v", "123", gomock.Any())

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(errdefs.IsNotFound(err)).Should(BeFalse())
				Expect(err.Error()).Should(Equal("getContainer error"))
			})
			It("should pass through errors from task.Status", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{}, errors.New("status error"))

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(Equal("status error"))
			})
			It("should pass through errors from proc.Wait", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Running,
				}, nil)
				proc.EXPECT().Wait(ctx).Return(nil, errors.New("wait error"))

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(Equal("wait error"))
			})
			It("should pass through errors from proc.Start", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Running,
				}, nil)
				proc.EXPECT().Wait(ctx).Return(statusC, nil)
				proc.EXPECT().Start(ctx).Return(errors.New("start error"))

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(Equal("start error"))
			})
		})
		Context("attach", func() {
			var now time.Time
			BeforeEach(func() {
				startOpts = &types.ExecStartOptions{
					ExecStartCheck: &types.ExecStartCheck{
						Detach:      false,
						Tty:         true,
						ConsoleSize: &[2]uint{123, 321},
					},
					ConID:           "123",
					ExecID:          "exec-123",
					Stdin:           rw,
					Stdout:          rw,
					Stderr:          rw,
					SuccessResponse: successResp,
				}
				now = time.Now()
			})
			It("should not throw any errors on success", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", gomock.Any()).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Running,
				}, nil)
				proc.EXPECT().Wait(ctx).Return(statusC, nil)
				proc.EXPECT().Resize(ctx, uint32(321), uint32(123)).Return(nil)
				proc.EXPECT().Start(ctx).Return(nil)

				go func() {
					statusC <- *containerd.NewExitStatus(uint32(0), now, nil)
				}()

				err := s.Start(ctx, startOpts)
				Expect(err).Should(BeNil())
				Expect(success).Should(BeTrue())
			})
			It("should log errors from proc.Resize", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", gomock.Any()).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Running,
				}, nil)
				proc.EXPECT().Wait(ctx).Return(statusC, nil)
				proc.EXPECT().Resize(ctx, uint32(321), uint32(123)).Return(errors.New("resize error"))
				logger.EXPECT().Errorf("could not resize console: %v", errors.New("resize error"))
				proc.EXPECT().Start(ctx).Return(nil)

				go func() {
					statusC <- *containerd.NewExitStatus(uint32(0), now, nil)
				}()

				err := s.Start(ctx, startOpts)
				Expect(err).Should(BeNil())
				Expect(success).Should(BeTrue())
			})
			It("should return errors from the process", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", gomock.Any()).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Running,
				}, nil)
				proc.EXPECT().Wait(ctx).Return(statusC, nil)
				proc.EXPECT().Resize(ctx, uint32(321), uint32(123)).Return(nil)
				proc.EXPECT().Start(ctx).Return(nil)

				go func() {
					statusC <- *containerd.NewExitStatus(uint32(0), now, errors.New("process error"))
				}()

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(Equal("process error"))
				Expect(success).Should(BeTrue())
			})
			It("should throw an error on a non-zero exit code", func() {
				cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
				con.EXPECT().Task(ctx, nil).Return(task, nil)
				task.EXPECT().LoadProcess(ctx, "exec-123", gomock.Any()).Return(proc, nil)
				task.EXPECT().Status(ctx).Return(containerd.Status{
					Status: containerd.Running,
				}, nil)
				proc.EXPECT().Wait(ctx).Return(statusC, nil)
				proc.EXPECT().Resize(ctx, uint32(321), uint32(123)).Return(nil)
				proc.EXPECT().Start(ctx).Return(nil)

				go func() {
					statusC <- *containerd.NewExitStatus(uint32(1), now, nil)
				}()

				err := s.Start(ctx, startOpts)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(Equal("exec failed with exit code 1"))
				Expect(success).Should(BeTrue())
			})
		})
	})
})

type CloseableBuffer struct {
	*bytes.Buffer
}

func (buf *CloseableBuffer) Close() error {
	// NOP
	return nil
}
