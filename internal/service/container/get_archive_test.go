// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"fmt"
	"os"
	pathutil "path"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/mount"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_ecc"
	"github.com/runfinch/finch-daemon/mocks/mocks_http"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Get Archive API", func() {
	var (
		ctx           context.Context
		mockCtrl      *gomock.Controller
		logger        *mocks_logger.Logger
		cdClient      *mocks_backend.MockContainerdClient
		ncClient      *mocks_backend.MockNerdctlContainerSvc
		fs            afero.Fs
		tarCreator    *mocks_archive.MockTarCreator
		con           *mocks_container.MockContainer
		task          *mocks_container.MockTask
		cid           string
		mockPid       uint32
		mockPath      string
		mockWriter    *mocks_http.MockResponseWriter
		mockCmd       *mocks_ecc.MockExecCmd
		s             *service
		containerPath string
	)
	BeforeEach(func() {
		ctx = context.Background()
		cid = "test123"
		mockPid = 1234
		mockPath = "/path/to/files"
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		fs = afero.NewMemMapFs()
		tarCreator = mocks_archive.NewMockTarCreator(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		mockWriter = mocks_http.NewMockResponseWriter(mockCtrl)
		mockCmd = mocks_ecc.NewMockExecCmd(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()
		s = &service{
			client:           cdClient,
			nctlContainerSvc: mockNerdctlService{ncClient, nil},
			logger:           logger,
			fs:               fs,
			tarCreator:       tarCreator,
		}
		containerPath = pathutil.Join(fmt.Sprintf("/proc/%d/root", mockPid), mockPath)
	})
	Context("GetPathToFilesInContainer", func() {
		It("should not return any errors on success", func() {
			err := fs.MkdirAll(pathutil.Dir(containerPath), os.ModeDir)
			Expect(err).Should(BeNil())

			_, err = fs.Create(containerPath)
			Expect(err).Should(BeNil())

			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "running"}, nil)
			task.EXPECT().Pid().Return(mockPid)

			path, cleanup, err := s.GetPathToFilesInContainer(ctx, cid, mockPath)
			Expect(err).Should(BeNil())
			Expect(path).Should(Equal(containerPath))
			Expect(cleanup).Should(BeNil())
		})
		It("should pass through errors from getContainer", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return(nil, fmt.Errorf("getContainer error"))
			logger.EXPECT().Errorf("failed to search container: %s. error: %s", cid, "getContainer error")
			logger.EXPECT().Errorf("Error getting container: %s", gomock.Any())

			path, cleanup, err := s.GetPathToFilesInContainer(ctx, cid, mockPath)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("getContainer error"))
			Expect(path).Should(BeEmpty())
			Expect(cleanup).Should(BeNil())
		})
		It("should return a NotFound error if the file does not exist inside the container", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "running"}, nil)
			task.EXPECT().Pid().Return(mockPid)

			path, cleanup, err := s.GetPathToFilesInContainer(ctx, cid, mockPath)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(path).Should(Equal(containerPath))
			Expect(cleanup).Should(BeNil())
		})
		It("should mount snapshot layers if the container has no task", func() {
			mockSnapKey := "123"
			mockMounts := []mount.Mount{
				{},
			}
			var mountDir string
			var snapPath string
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, cerrdefs.ErrNotFound)
			con.EXPECT().Info(ctx).Return(containers.Container{SnapshotKey: mockSnapKey}, nil)
			cdClient.EXPECT().ListSnapshotMounts(ctx, mockSnapKey).Return(mockMounts, nil)
			cdClient.EXPECT().MountAll(mockMounts, gomock.Any()).DoAndReturn(func(mounts []mount.Mount, tmpDir string) error {
				mountDir = tmpDir
				snapPath = pathutil.Join(mountDir, mockPath)
				err := fs.MkdirAll(snapPath, os.ModeDir)
				Expect(err).Should(BeNil())
				return nil
			})

			path, cleanup, err := s.GetPathToFilesInContainer(ctx, cid, mockPath)

			Expect(err).Should(BeNil())
			Expect(path).Should(HavePrefix(pathutil.Join(os.TempDir(), "mount-snapshot")))
			Expect(path).Should(HaveSuffix(mockPath))
			_, err = fs.Stat(snapPath)
			Expect(err).Should(BeNil())
			Expect(cleanup).ShouldNot(BeNil())

			cdClient.EXPECT().Unmount(gomock.Any(), 0).Return(nil)

			cleanup()
			_, err = fs.Stat(snapPath)
			Expect(err).ShouldNot(BeNil())
			Expect(errors.Is(err, os.ErrNotExist)).Should(BeTrue())
			_, err = fs.Stat(mountDir)
			Expect(err).ShouldNot(BeNil())
			Expect(errors.Is(err, os.ErrNotExist)).Should(BeTrue())
		})
		It("should mount snapshot layers if the container is not running", func() {
			mockSnapKey := "123"
			mockMounts := []mount.Mount{
				{},
			}
			var mountDir string
			var snapPath string
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "stopped"}, nil)
			con.EXPECT().Info(ctx).Return(containers.Container{SnapshotKey: mockSnapKey}, nil)
			cdClient.EXPECT().ListSnapshotMounts(ctx, mockSnapKey).Return(mockMounts, nil)
			cdClient.EXPECT().MountAll(mockMounts, gomock.Any()).DoAndReturn(func(mounts []mount.Mount, tmpDir string) error {
				mountDir = tmpDir
				snapPath = pathutil.Join(mountDir, mockPath)
				err := fs.MkdirAll(snapPath, os.ModeDir)
				Expect(err).Should(BeNil())
				return nil
			})

			path, cleanup, err := s.GetPathToFilesInContainer(ctx, cid, mockPath)

			Expect(err).Should(BeNil())
			Expect(path).Should(HavePrefix(pathutil.Join(os.TempDir(), "mount-snapshot")))
			Expect(path).Should(HaveSuffix(mockPath))
			_, err = fs.Stat(snapPath)
			Expect(err).Should(BeNil())
			Expect(cleanup).ShouldNot(BeNil())

			cdClient.EXPECT().Unmount(gomock.Any(), 0).Return(nil)

			cleanup()
			_, err = fs.Stat(snapPath)
			Expect(err).ShouldNot(BeNil())
			Expect(errors.Is(err, os.ErrNotExist)).Should(BeTrue())
			_, err = fs.Stat(mountDir)
			Expect(err).ShouldNot(BeNil())
			Expect(errors.Is(err, os.ErrNotExist)).Should(BeTrue())
		})
		It("should cleanup the tempdir if it fails to mount the snapshot layers", func() {
			mockSnapKey := "123"
			mockMounts := []mount.Mount{
				{},
			}
			var mountDir string
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, cerrdefs.ErrNotFound)
			con.EXPECT().Info(ctx).Return(containers.Container{SnapshotKey: mockSnapKey}, nil)
			cdClient.EXPECT().ListSnapshotMounts(ctx, mockSnapKey).Return(mockMounts, nil)
			cdClient.EXPECT().MountAll(mockMounts, gomock.Any()).DoAndReturn(func(mounts []mount.Mount, tmpDir string) error {
				mountDir = tmpDir
				return fmt.Errorf("MountAll error")
			})
			logger.EXPECT().Errorf("Could not mount snapshot: %s", gomock.Any())

			path, cleanup, err := s.GetPathToFilesInContainer(ctx, cid, mockPath)

			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("MountAll error"))
			Expect(path).Should(BeEmpty())
			_, err = fs.Stat(mountDir)
			Expect(err).Should(BeNil())
			Expect(cleanup).ShouldNot(BeNil())

			cdClient.EXPECT().Unmount(gomock.Any(), 0).Return(nil)

			cleanup()
			_, err = fs.Stat(mountDir)
			Expect(err).ShouldNot(BeNil())
			Expect(errors.Is(err, os.ErrNotExist)).Should(BeTrue())
		})
	})
	Context("WriteFilesAsTarArchive", func() {
		It("should return no error on success", func() {
			tarCreator.EXPECT().CreateTarCommand(mockPath, false).Return(mockCmd, nil)
			mockCmd.EXPECT().SetStdout(mockWriter)
			mockCmd.EXPECT().Run().Return(nil)

			err := s.WriteFilesAsTarArchive(mockPath, mockWriter, false)
			Expect(err).Should(BeNil())
		})
		It("should pass through errors from CreateTarCommand", func() {
			tarCreator.EXPECT().CreateTarCommand(mockPath, false).Return(nil, fmt.Errorf("CreateTarCommand error"))

			err := s.WriteFilesAsTarArchive(mockPath, mockWriter, false)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("CreateTarCommand error"))
		})
		It("should pass through errors from cmd.Run", func() {
			tarCreator.EXPECT().CreateTarCommand(mockPath, false).Return(mockCmd, nil)
			mockCmd.EXPECT().SetStdout(mockWriter)
			mockCmd.EXPECT().Run().Return(fmt.Errorf("cmd.Run error"))

			err := s.WriteFilesAsTarArchive(mockPath, mockWriter, false)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("cmd.Run error"))
		})
	})
})
