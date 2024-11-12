// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	containerd "github.com/containerd/containerd/v2/client"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/spf13/afero"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// GetPathToFilesInContainer locates files in a container. If the container is running, it will use the running container's
// /proc filesystem. If the container is not running, it will use the snapshotter to mount its filesystem to a tempdir.
// In the latter case, some cleanup is required, in which case it will return a func() that will handle the cleanup.
func (s *service) GetPathToFilesInContainer(ctx context.Context, cid string, srcPath string) (filePath string, cleanup func(), err error) {
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		s.logger.Errorf("Error getting container: %s", err)
		return
	}

	var root string

	task, err := con.Task(ctx, nil)
	if err != nil {
		// if the task is simply not found, we should try to mount the snapshot. any other type of error from Task() is fatal here.
		if !cerrdefs.IsNotFound(err) {
			s.logger.Errorf("Error getting task from container: %s", err)
			return
		}

		root, cleanup, err = s.mountSnapshotForContainer(ctx, con)
		if err != nil {
			s.logger.Errorf("Could not mount snapshot: %s", err)
			return
		}
	} else {
		var status containerd.Status
		status, err = task.Status(ctx)
		if err != nil {
			s.logger.Errorf("Could not get task status: %s", err)
			return
		}
		if status.Status == containerd.Running {
			pid := task.Pid()
			// path to the container's root filesystem. reference:
			// https://github.com/containerd/nerdctl/blob/774b6e9ab69fadbcffb60297791db3f036231abf/pkg/containerutil/cp_linux.go#L44
			root = fmt.Sprintf("/proc/%d/root", pid)
		} else {
			root, cleanup, err = s.mountSnapshotForContainer(ctx, con)
			if err != nil {
				s.logger.Errorf("Could not mount snapshot: %s", err)
				return
			}
		}
	}

	// TODO: ideally this should use securejoin.SecureJoin() to sanitize inputs but securejoin can't work with afero.MemMapFs yet
	// so it makes testing difficult
	filePath = path.Join(root, srcPath)

	_, err = s.fs.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = errdefs.NewNotFound(err)
			return
		}
		s.logger.Errorf("Error statting %s: %s", filePath, err)
		return
	}

	return
}

func (s *service) WriteFilesAsTarArchive(filePath string, writer io.Writer, slashDot bool) error {
	cmd, err := s.tarCreator.CreateTarCommand(filePath, slashDot)
	if err != nil {
		return err
	}
	cmd.SetStdout(writer)
	return cmd.Run()
}

func (s *service) mountSnapshotForContainer(ctx context.Context, con containerd.Container) (string, func(), error) {
	cinfo, err := con.Info(ctx)
	if err != nil {
		return "", nil, err
	}
	snapKey := cinfo.SnapshotKey

	mounts, err := s.client.ListSnapshotMounts(ctx, snapKey)
	if err != nil {
		return "", nil, err
	}

	tempDir, err := afero.TempDir(s.fs, "", "mount-snapshot")
	if err != nil {
		return "", nil, err
	}

	// Mount the snapshot
	if err := s.client.MountAll(mounts, tempDir); err != nil {
		cleanup := func() {
			s.fs.RemoveAll(tempDir)
		}
		return "", cleanup, err
	}

	cleanup := func() {
		s.client.Unmount(tempDir, 0)
		s.fs.RemoveAll(tempDir)
	}

	return tempDir, cleanup, nil
}
