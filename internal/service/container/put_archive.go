// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// ExtractArchiveInContainer extracts the given tar archive to the specified location in the
// filesystem of this container. The given path must be of a directory in the
// container. If it is not, the error will be an errdefs.InvalidFormat. If
// noOverwriteDirNonDir is true then it will be an error if unpacking the
// given content would cause an existing directory to be replaced with a non-
// directory and vice versa.
// setting uidgid has no effect due to https://github.com/docker/docs/issues/17533#issuecomment-1588223056
// Fix once the bug is resolved.

func (s *service) ExtractArchiveInContainer(ctx context.Context, opts *types.PutArchiveOptions, body io.ReadCloser) error {
	con, err := s.getContainer(ctx, opts.ContainerId)
	if err != nil {
		return err
	}
	// First check if the mount is a volume and readonly or in a readonly rootfs
	err = s.isReadOnlyMount(ctx, con, opts.Path)
	if err != nil {
		return err
	}
	var (
		root     string
		cleanup  func()
		filePath string
	)

	task, err := con.Task(ctx, nil)
	if err != nil {
		// if the task is simply not found, we should try to mount the snapshot. any other type of error from Task() is fatal here.
		if !cerrdefs.IsNotFound(err) {
			s.logger.Errorf("Error getting task from container %s: %v", opts.ContainerId, err)
			return err
		}
		root, cleanup, err = s.mountSnapshotForContainer(ctx, con)
		if err != nil {
			s.logger.Errorf("Could not mount snapshot: %v", err)
			return err
		}
	} else {
		var status containerd.Status
		status, err = task.Status(ctx)
		if err != nil {
			s.logger.Errorf("Could not get task status: %v", err)
			return err
		}
		if status.Status == containerd.Running {
			pid := task.Pid()
			// containerPath to the container's root filesystem. reference:
			// https://github.com/containerd/nerdctl/blob/774b6e9ab69fadbcffb60297791db3f036231abf/pkg/containerutil/cp_linux.go#L44
			root = fmt.Sprintf("/proc/%d/root", pid)
		} else {
			root, cleanup, err = s.mountSnapshotForContainer(ctx, con)
			if err != nil {
				s.logger.Errorf("Could not mount snapshot: %v", err)
				return err
			}
		}
	}
	if cleanup != nil {
		defer cleanup()
	}
	// TODO: ideally this should use securejoin.SecureJoin() to sanitize inputs but securejoin can't work with afero.MemMapFs yet
	// so it makes testing difficult
	filePath = path.Join(root, opts.Path)

	stat, err := s.fs.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = errdefs.NewNotFound(err)
			return err
		}
		s.logger.Errorf("Error statting %s: %s", filePath, err)
		return err
	}
	if !stat.IsDir() {
		return errdefs.NewInvalidFormat(fmt.Errorf("extraction point: %s is not a directory", filePath))
	}
	tarOptions := &archive.TarOptions{
		NoOverwriteDirNonDir: opts.Overwrite,
		IDMap:                idtools.IdentityMapping{},
	}
	return s.tarExtractor.ExtractCompressed(body, filePath, tarOptions)
}

func (s *service) isReadOnlyMount(ctx context.Context, con containerd.Container, containerPath string) error {
	filePath := filepath.Clean(containerPath)
	spec, err := con.Spec(ctx)
	if err != nil {
		return err
	}
	if spec.Root.Readonly {
		return errdefs.NewForbidden(fmt.Errorf("container rootfs: %s is marked read-only", spec.Root.Path))
	}
	for _, mount := range spec.Mounts {
		for _, option := range mount.Options {
			// Check if path to copy is marked read-only
			if option == "ro" && isParentDir(filePath, mount.Destination) {
				return errdefs.NewForbidden(fmt.Errorf("mount point %s is marked read-only", filePath))
			}
		}
	}
	return nil
}

func isParentDir(filePath, potentialParent string) bool {
	if filePath == potentialParent {
		return true
	}
	if filePath == "/" {
		return false
	}
	return isParentDir(path.Dir(filePath), potentialParent)
}
