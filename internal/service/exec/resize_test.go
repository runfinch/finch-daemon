// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"errors"

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

var _ = Describe("Exec Resize API ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		con      *mocks_container.MockContainer
		task     *mocks_container.MockTask
		proc     *mocks_container.MockProcess
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
		s = NewService(cdClient, logger)
	})
	Context("service", func() {
		It("should not return an error on success", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
			proc.EXPECT().Resize(ctx, uint32(321), uint32(123)).Return(nil)

			err := s.Resize(ctx, &types.ExecResizeOptions{
				ConID:  "123",
				ExecID: "exec-123",
				Height: 123,
				Width:  321,
			})
			Expect(err).Should(BeNil())
		})
		It("should return a not found error if the exec instance is not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return(nil, cerrdefs.ErrNotFound)
			logger.EXPECT().Errorf("failed to search container: %s. error: %v", "123", gomock.Any())

			err := s.Resize(ctx, &types.ExecResizeOptions{
				ConID:  "123",
				ExecID: "exec-123",
				Height: 123,
				Width:  321,
			})
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should pass through any errors from resize", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().LoadProcess(ctx, "exec-123", nil).Return(proc, nil)
			proc.EXPECT().Resize(ctx, uint32(321), uint32(123)).Return(errors.New("resize error"))

			err := s.Resize(ctx, &types.ExecResizeOptions{
				ConID:  "123",
				ExecID: "exec-123",
				Height: 123,
				Width:  321,
			})
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("resize error"))
		})
	})
})
