// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	pathutil "path"
	"strings"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/mount"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/spf13/afero"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

type MockReadCloser struct {
	io.Reader
}

func (m *MockReadCloser) Close() error {
	return nil
}

var _ = Describe("Extract in container API", func() {
	var (
		ctx            context.Context
		mockCtrl       *gomock.Controller
		logger         *mocks_logger.Logger
		cdClient       *mocks_backend.MockContainerdClient
		ncClient       *mocks_backend.MockNerdctlContainerSvc
		fs             afero.Fs
		tarExtractor   *mocks_archive.MockTarExtractor
		con            *mocks_container.MockContainer
		task           *mocks_container.MockTask
		cid            string
		mockPid        uint32
		mockPath       string
		putArchiveOpts *types.PutArchiveOptions
		s              *service
		containerPath  string
		mockReader     io.ReadCloser
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
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()
		s = &service{
			client:           cdClient,
			nctlContainerSvc: mockNerdctlService{ncClient, nil},
			logger:           logger,
			fs:               fs,
			tarExtractor:     tarExtractor,
		}
		putArchiveOpts = &types.PutArchiveOptions{
			ContainerId: "test123",
			Path:        "/path/to/files",
			Overwrite:   false,
			CopyUIDGID:  false,
		}
		containerPath = pathutil.Join(fmt.Sprintf("/proc/%d/root", mockPid), mockPath)
		mockReader = &MockReadCloser{strings.NewReader("Test tar archive")}
	})
	Context("ExtractArchiveInContainer", func() {
		It("should not return an error when container is running and path is writeable", func() {
			err := fs.MkdirAll(containerPath, 0o755)
			Expect(err).Should(BeNil())
			_, err = fs.Stat(containerPath)
			Expect(err).Should(BeNil())
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "running"}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			task.EXPECT().Pid().Return(mockPid)
			tarExtractor.EXPECT().ExtractCompressed(mockReader, containerPath, &archive.TarOptions{
				NoOverwriteDirNonDir: false,
				IDMap:                idtools.IdentityMapping{},
			}).Return(nil)
			err = s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(err).Should(BeNil())
		})
		It("should return an error when container is running and volume is read-only", func() {
			err := fs.MkdirAll(containerPath, 0o755)
			Expect(err).Should(BeNil())
			_, err = fs.Stat(containerPath)
			Expect(err).Should(BeNil())
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "running"}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "ro"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			err = s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(errdefs.IsForbiddenError(err)).Should(Equal(true))
		})
		It("should return an error when container is running and rootfs is read-only", func() {
			err := fs.MkdirAll(containerPath, 0o755)
			Expect(err).Should(BeNil())
			_, err = fs.Stat(containerPath)
			Expect(err).Should(BeNil())
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "running"}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: true,
				},
			}, nil)
			err = s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(errdefs.IsForbiddenError(err)).Should(Equal(true))
		})
		It("should return an error when path is not a directory", func() {
			err := fs.MkdirAll(containerPath, 0o755)
			Expect(err).Should(BeNil())

			_, err = fs.Create(containerPath)
			Expect(err).Should(BeNil())

			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "running"}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			task.EXPECT().Pid().Return(mockPid)
			err = s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(errdefs.IsInvalidFormat(err)).Should(Equal(true))
		})
		It("should not return an error when container is stopped and success", func() {
			mockSnapKey := "123"
			mockMounts := []mount.Mount{
				{},
			}
			var mountDir string
			var snapPath string
			err := fs.MkdirAll(containerPath, 0o755)
			Expect(err).Should(BeNil())
			_, err = fs.Stat(containerPath)
			Expect(err).Should(BeNil())
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "stopped"}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			con.EXPECT().Info(ctx).Return(containers.Container{SnapshotKey: mockSnapKey}, nil)
			cdClient.EXPECT().ListSnapshotMounts(ctx, mockSnapKey).Return(mockMounts, nil)
			cdClient.EXPECT().MountAll(mockMounts, gomock.Any()).DoAndReturn(func(mounts []mount.Mount, tmpDir string) error {
				mountDir = tmpDir
				snapPath = pathutil.Join(mountDir, mockPath)
				err := fs.MkdirAll(snapPath, os.ModeDir)
				Expect(err).Should(BeNil())
				return nil
			})
			tarExtractor.EXPECT().ExtractCompressed(mockReader, gomock.Any(), &archive.TarOptions{
				NoOverwriteDirNonDir: false,
				IDMap:                idtools.IdentityMapping{},
			}).Return(nil)
			cdClient.EXPECT().Unmount(gomock.Any(), 0)
			err = s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(err).Should(BeNil())
		})
		It("should not return an error when task is not found", func() {
			mockSnapKey := "123"
			mockMounts := []mount.Mount{
				{},
			}
			var mountDir string
			var snapPath string
			err := fs.MkdirAll(containerPath, 0o755)
			Expect(err).Should(BeNil())
			_, err = fs.Stat(containerPath)
			Expect(err).Should(BeNil())
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, cerrdefs.ErrNotFound)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "stopped"}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			con.EXPECT().Info(ctx).Return(containers.Container{SnapshotKey: mockSnapKey}, nil)
			cdClient.EXPECT().ListSnapshotMounts(ctx, mockSnapKey).Return(mockMounts, nil)
			cdClient.EXPECT().MountAll(mockMounts, gomock.Any()).DoAndReturn(func(mounts []mount.Mount, tmpDir string) error {
				mountDir = tmpDir
				snapPath = pathutil.Join(mountDir, mockPath)
				err := fs.MkdirAll(snapPath, os.ModeDir)
				Expect(err).Should(BeNil())
				return nil
			})
			tarExtractor.EXPECT().ExtractCompressed(mockReader, gomock.Any(), &archive.TarOptions{
				NoOverwriteDirNonDir: false,
				IDMap:                idtools.IdentityMapping{},
			}).Return(nil)
			cdClient.EXPECT().Unmount(gomock.Any(), 0)
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())
			err = s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(err).Should(BeNil())
		})
		It("should return an error when task not found and error not not found", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			con.EXPECT().Task(ctx, nil).Return(nil, errors.New("task finding error"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())
			err := s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(err.Error()).Should(Equal("task finding error"))
		})
		It("should return an error when task status returns an error ", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "unknown"}, errors.New("status error"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			err := s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(err.Error()).Should(Equal("status error"))
		})
		It("should return an error when searching a container returns an error ", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{}, errors.New("search error"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())
			err := s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(err.Error()).Should(Equal("search error"))
		})
		It("should return an error when error mounting snapshot", func() {
			cdClient.EXPECT().SearchContainer(ctx, cid).Return([]containerd.Container{con}, nil)
			con.EXPECT().Task(ctx, nil).Return(task, nil)
			task.EXPECT().Status(ctx).Return(containerd.Status{Status: "stopped"}, nil)
			con.EXPECT().Spec(ctx).Return(&specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: mockPath,
						Type:        "bind",
						Source:      "/path/on/host",
						Options:     []string{"rbind", "rw"},
					},
				},
				Root: &specs.Root{
					Path:     "rootfs",
					Readonly: false,
				},
			}, nil)
			con.EXPECT().Info(ctx).Return(containers.Container{}, errors.New("error finding info"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			err := s.ExtractArchiveInContainer(ctx, putArchiveOpts, mockReader)
			Expect(err.Error()).Should(Equal("error finding info"))
		})
	})
})
