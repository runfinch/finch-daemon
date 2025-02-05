// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"fmt"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/oci"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/spf13/afero"

	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Exec API ", func() {
	var (
		ctx          context.Context
		mockCtrl     *gomock.Controller
		logger       *mocks_logger.Logger
		cdClient     *mocks_backend.MockContainerdClient
		ncClient     *mocks_backend.MockNerdctlContainerSvc
		con          *mocks_container.MockContainer
		task         *mocks_container.MockTask
		proc         *mocks_container.MockProcess
		service      container.Service
		fs           afero.Fs
		tarCreator   *mocks_archive.MockTarCreator
		tarExtractor *mocks_archive.MockTarExtractor
		execConfig   types.ExecConfig
		pspec        *specs.Process
	)
	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		fs = afero.NewMemMapFs()
		tarCreator = mocks_archive.NewMockTarCreator(mockCtrl)
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		proc = mocks_container.NewMockProcess(mockCtrl)
		service = NewService(
			cdClient,
			mockNerdctlService{ncClient, nil},
			logger,
			fs,
			tarCreator,
			tarExtractor,
		)
		execConfig = types.ExecConfig{
			User:         "foo",
			Privileged:   true,
			Tty:          true,
			ConsoleSize:  &[2]uint{123, 321},
			AttachStdin:  true,
			AttachStderr: true,
			AttachStdout: true,
			Detach:       false,
			DetachKeys:   "foo",
			Env:          []string{"foo=bar", "bar=baz"},
			WorkingDir:   "path/to/dir",
			Cmd:          []string{"foo", "bar"},
		}
		pspec = &specs.Process{
			Terminal:    execConfig.Tty,
			ConsoleSize: &specs.Box{Height: 123, Width: 321},
			User: specs.User{
				UID:            123,
				GID:            123,
				AdditionalGids: []uint32{1, 2, 3},
			},
			Args:        execConfig.Cmd,
			CommandLine: "",
			Env:         execConfig.Env,
			Cwd:         execConfig.WorkingDir,
			Capabilities: &specs.LinuxCapabilities{
				Bounding:    []string{"foo", "bar"},
				Permitted:   []string{"foo", "bar"},
				Inheritable: []string{"foo", "bar"},
				Effective:   []string{"foo", "bar"},
			},
		}
	})
	Context("service", func() {
		It("should not return any errors on success", func() {
			var eid string
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: containerd.Running}, nil)
			task.EXPECT().Exec(ctx, gomock.Any(), pspec, gomock.Any()).DoAndReturn(
				func(ctx context.Context, execId string, pspec *specs.Process, ioCreate cio.Creator) (containerd.Process, error) {
					eid = execId
					fifos := &cio.FIFOSet{
						Config: cio.Config{
							Stdin:    "stdin",
							Stdout:   "stdout",
							Stderr:   "stderr",
							Terminal: execConfig.Tty,
						},
					}
					cdClient.EXPECT().NewFIFOSetInDir(defaults.DefaultFIFODir, execId, execConfig.Tty).Return(fifos, nil)
					cdClient.EXPECT().NewDirectCIO(ctx, fifos).Return(&cio.DirectIO{}, nil)

					_, err := ioCreate(execId)
					Expect(err).Should(BeNil())

					return proc, nil
				})

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).Should(BeNil())
			Expect(execId).Should(Equal(fmt.Sprintf("123/%s", eid)))
		})
		It("should not create fifos for stdio when they aren't supposed to be attached", func() {
			execConfig.AttachStdin = false
			execConfig.AttachStderr = false
			execConfig.AttachStdout = false

			var eid string
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: containerd.Running}, nil)
			task.EXPECT().Exec(ctx, gomock.Any(), pspec, gomock.Any()).DoAndReturn(
				func(ctx context.Context, execId string, pspec *specs.Process, ioCreate cio.Creator) (containerd.Process, error) {
					eid = execId
					fifos := &cio.FIFOSet{
						Config: cio.Config{
							Stdin:    "stdin",
							Stdout:   "stdout",
							Stderr:   "stderr",
							Terminal: execConfig.Tty,
						},
					}
					expectFifos := &cio.FIFOSet{
						Config: cio.Config{
							Stdin:    "",
							Stdout:   "",
							Stderr:   "",
							Terminal: execConfig.Tty,
						},
					}
					cdClient.EXPECT().NewFIFOSetInDir(defaults.DefaultFIFODir, execId, execConfig.Tty).Return(fifos, nil)
					cdClient.EXPECT().NewDirectCIO(ctx, expectFifos).Return(&cio.DirectIO{}, nil)

					_, err := ioCreate(execId)
					Expect(err).Should(BeNil())

					return proc, nil
				})

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).Should(BeNil())
			Expect(execId).Should(Equal(fmt.Sprintf("123/%s", eid)))
		})
		It("should return a NotFound error if the container is not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return(nil, cerrdefs.ErrNotFound)
			logger.EXPECT().Errorf("failed to search container: %s. error: %s", "123", gomock.Any())

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through other errors from getContainer", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{}, nil)
			logger.EXPECT().Debugf("no such container: %s", "123")

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(err.Error()).Should(Equal("no such container: 123"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through errors from container.Spec", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(nil, errors.New("spec error"))

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("spec error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through errors from container.Info", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, errors.New("info error"))

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("info error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should throw an error if any oci.SpecOpts throw an error", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					return errors.New("withUser error")
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(1)

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("withUser error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through errors from GetCurrentCapabilities", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return(nil, errors.New("getCaps error"))

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("getCaps error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should return a Conflict error if the task is not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, cerrdefs.ErrNotFound)

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
			Expect(err.Error()).Should(Equal("container 123 is not running"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through any other error from container.Task", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, errors.New("task error"))

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("task error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should return a Conflict error if the task status is not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{}, cerrdefs.ErrNotFound)

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
			Expect(err.Error()).Should(Equal("container 123 is not running"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through any other error from task.Status", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{}, errors.New("status error"))

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("status error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should return a Conflict error if the status is not running", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: containerd.Stopped}, nil)

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
			Expect(err.Error()).Should(Equal("container 123 is not running"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through any errors from task.Exec", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: containerd.Running}, nil)
			task.EXPECT().Exec(ctx, gomock.Any(), pspec, gomock.Any()).Return(nil, errors.New("exec error"))

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("exec error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through errors from NewFIFOSetInDir", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: containerd.Running}, nil)
			task.EXPECT().Exec(ctx, gomock.Any(), pspec, gomock.Any()).DoAndReturn(
				func(ctx context.Context, execId string, pspec *specs.Process, ioCreate cio.Creator) (containerd.Process, error) {
					cdClient.EXPECT().NewFIFOSetInDir(defaults.DefaultFIFODir, execId, execConfig.Tty).Return(nil, errors.New("fifo error"))

					_, err := ioCreate(execId)
					return nil, err
				})

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("fifo error"))
			Expect(execId).Should(BeEmpty())
		})
		It("should pass through errors from NewDirectCIO", func() {
			cdClient.EXPECT().SearchContainer(ctx, "123").Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&oci.Spec{Process: &specs.Process{}}, nil)
			cdClient.EXPECT().OCISpecWithUser(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.UID = 123
					spec.Process.User.GID = 123
					return nil
				})
			cdClient.EXPECT().OCISpecWithAdditionalGIDs(execConfig.User).Return(
				func(_ context.Context, _ oci.Client, _ *containers.Container, spec *oci.Spec) error {
					spec.Process.User.AdditionalGids = []uint32{1, 2, 3}
					return nil
				})
			con.EXPECT().Info(ctx).Return(containers.Container{}, nil)
			cdClient.EXPECT().GetClient().Return(&containerd.Client{}).Times(3)
			cdClient.EXPECT().GetCurrentCapabilities().Return([]string{"foo", "bar"}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: containerd.Running}, nil)
			task.EXPECT().Exec(ctx, gomock.Any(), pspec, gomock.Any()).DoAndReturn(
				func(ctx context.Context, execId string, pspec *specs.Process, ioCreate cio.Creator) (containerd.Process, error) {
					fifos := &cio.FIFOSet{
						Config: cio.Config{
							Stdin:    "stdin",
							Stdout:   "stdout",
							Stderr:   "stderr",
							Terminal: execConfig.Tty,
						},
					}
					cdClient.EXPECT().NewFIFOSetInDir(defaults.DefaultFIFODir, execId, execConfig.Tty).Return(fifos, nil)
					cdClient.EXPECT().NewDirectCIO(ctx, fifos).Return(nil, errors.New("cio error"))

					_, err := ioCreate(execId)
					return nil, err
				})

			execId, err := service.ExecCreate(ctx, "123", execConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("cio error"))
			Expect(execId).Should(BeEmpty())
		})
	})
})
