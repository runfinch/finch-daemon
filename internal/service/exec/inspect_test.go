// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"errors"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/exec"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_cio"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Exec Inspect API ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		con      *mocks_container.MockContainer
		task     *mocks_container.MockTask
		proc     *mocks_container.MockProcess
		procIO   *mocks_cio.MockIO
		s        exec.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		proc = mocks_container.NewMockProcess(mockCtrl)
		procIO = mocks_cio.NewMockIO(mockCtrl)
		s = NewService(cdClient, logger)
	})
	Context("service", func() {
		It("should not return an error on success", func() {
			now := time.Now()
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
			proc.EXPECT().Status(ctx).Return(containerd.Status{
				Status:     containerd.Running,
				ExitStatus: 0,
				ExitTime:   now,
			}, nil)
			proc.EXPECT().ID().Return("exec-123")
			con.EXPECT().ID().Return("123")
			proc.EXPECT().Pid().Return(uint32(123))
			proc.EXPECT().IO().Times(5).Return(procIO)
			procIO.EXPECT().Config().Times(4).Return(cio.Config{
				Terminal: false,
				Stdin:    "",
				Stdout:   "",
				Stderr:   "",
			})

			exitCode := 0
			Expect(s.Inspect(ctx, "123", "exec-123")).Should(Equal(&types.ExecInspect{
				ID:       "exec-123",
				Running:  true,
				ExitCode: &exitCode,
				ProcessConfig: &types.ExecProcessConfig{
					Tty: false,
				},
				OpenStdin:   false,
				OpenStderr:  false,
				OpenStdout:  false,
				CanRemove:   true,
				ContainerID: "123",
				DetachKeys:  []byte(""),
				Pid:         123,
			}))
		})
		It("should return a NotFound error if loadExecInstance returns NotFound", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return(nil, cerrdefs.ErrNotFound)
			logger.EXPECT().Errorf("failed to search container: %s. error: %v", "123", gomock.Any())

			inspectResult, err := s.Inspect(ctx, "123", "exec-123")
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(inspectResult).Should(BeNil())
		})
		It("should pass through non-NotFound errors from loadExecInstance", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return(nil, errors.New("getContainer error"))
			logger.EXPECT().Errorf("failed to search container: %s. error: %v", "123", gomock.Any())

			inspectResult, err := s.Inspect(ctx, "123", "exec-123")
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("getContainer error"))
			Expect(inspectResult).Should(BeNil())
		})
		It("should log a warning if proc.Status returns an error", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
			proc.EXPECT().Status(ctx).Return(containerd.Status{}, errors.New("status error"))
			proc.EXPECT().ID().Return("exec-123").Times(2)
			logger.EXPECT().Warnf("error getting process status for proc %s: %v", "exec-123", gomock.Any())
			con.EXPECT().ID().Return("123")
			proc.EXPECT().Pid().Return(uint32(123))
			proc.EXPECT().IO().Times(5).Return(procIO)
			procIO.EXPECT().Config().Times(4).Return(cio.Config{
				Terminal: false,
				Stdin:    "",
				Stdout:   "",
				Stderr:   "",
			})

			exitCode := 0
			Expect(s.Inspect(ctx, "123", "exec-123")).Should(Equal(&types.ExecInspect{
				ID:       "exec-123",
				Running:  false,
				ExitCode: &exitCode,
				ProcessConfig: &types.ExecProcessConfig{
					Tty: false,
				},
				OpenStdin:   false,
				OpenStderr:  false,
				OpenStdout:  false,
				CanRemove:   false,
				ContainerID: "123",
				DetachKeys:  []byte(""),
				Pid:         123,
			}))
		})
	})
})
