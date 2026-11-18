// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"time"

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

// Unit tests related to container restart API.
var _ = Describe("Container Restart API ", func() {
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
		timeout      time.Duration
		options      ncTypes.ContainerRestartOptions
	)
	BeforeEach(func() {
		ctx = context.Background()
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
		timeout = time.Duration(10)
		options = ncTypes.ContainerRestartOptions{
			Timeout: &timeout,
		}
	})
	Context("service", func() {
		// expectCustomStartSuccess sets up mock expectations for a successful customStart flow.
		expectCustomStartSuccess := func() {
			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{Process: &specs.Process{}}, nil)
			con.EXPECT().NewTask(gomock.Any(), gomock.Any()).Return(task, nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{"nerdctl/namespace": "finch"}, nil)
			ncClient.EXPECT().GetDataStore().Return("/tmp/test-store", nil)
			task.EXPECT().Start(gomock.Any()).Return(nil)
		}

		It("should not return any error when Running", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil).AnyTimes()
			ncClient.EXPECT().StopContainer(ctx, con.ID(), gomock.Any()).Return(nil)
			expectCustomStartSuccess()

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			err := service.Restart(ctx, cid, options)
			Expect(err).Should(BeNil())
		})
		It("should not return any error when Stopped", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil).AnyTimes()
			// StopContainer returns NotModified (already stopped) — this is swallowed
			ncClient.EXPECT().StopContainer(ctx, con.ID(), gomock.Any()).Return(errdefs.NewNotModified(fmt.Errorf("already stopped")))
			expectCustomStartSuccess()

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			err := service.Restart(ctx, cid, options)
			Expect(err).Should(BeNil())
		})
		It("should return not found error", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf("no such container: %s", gomock.Any())

			err := service.Restart(ctx, cid, options)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return multiple containers found error", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf("multiple IDs found with provided prefix: %s, total containers found: %d", cid, 2)

			err := service.Restart(ctx, cid, options)
			Expect(err).Should(Not(BeNil()))
		})
		It("should fail when customStart fails", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil).AnyTimes()
			ncClient.EXPECT().StopContainer(ctx, con.ID(), gomock.Any()).Return(nil)

			// customStart fails at Spec
			con.EXPECT().Task(gomock.Any(), nil).Return(nil, fmt.Errorf("no task"))
			con.EXPECT().Spec(gomock.Any()).Return(nil, fmt.Errorf("spec error"))

			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			err := service.Restart(ctx, cid, options)
			Expect(err).Should(Not(BeNil()))
		})
	})
})
