// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/defaults"
	"github.com/containerd/containerd/oci"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/flagutil"
	"github.com/containerd/nerdctl/v2/pkg/idgen"
	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) ExecCreate(ctx context.Context, cid string, config types.ExecConfig) (string, error) {
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return "", errdefs.NewNotFound(err)
		}
		return "", err
	}

	pspec, err := s.generateExecProcessSpec(ctx, con, config)
	if err != nil {
		return "", err
	}

	task, err := con.Task(ctx, nil)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return "", errdefs.NewConflict(fmt.Errorf("container %s is not running", cid))
		}
		return "", err
	}

	status, err := task.Status(ctx)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return "", errdefs.NewConflict(fmt.Errorf("container %s is not running", cid))
		}
		return "", err
	}
	if status.Status != containerd.Running {
		return "", errdefs.NewConflict(fmt.Errorf("container %s is not running", cid))
	}

	execID := "exec-" + idgen.GenerateID()

	ioCreator := func(id string) (cio.IO, error) {
		fifos, err := s.client.NewFIFOSetInDir(defaults.DefaultFIFODir, id, config.Tty)
		if err != nil {
			return nil, err
		}

		if !config.AttachStdin {
			fifos.Stdin = ""
		}
		if !config.AttachStdout {
			fifos.Stdout = ""
		}
		if !config.AttachStderr {
			fifos.Stderr = ""
		}

		directIO, err := s.client.NewDirectCIO(ctx, fifos)
		if err != nil {
			return nil, err
		}

		return directIO, nil
	}

	// ignore the returned process because we will load it later in exec_start
	_, err = task.Exec(ctx, execID, pspec, ioCreator)
	if err != nil {
		return "", err
	}

	// the task and process keep track of most of the state we care about, but we first need the container to access its task & process by execID.
	return fmt.Sprintf("%s/%s", cid, execID), nil
}

func (s *service) generateExecProcessSpec(ctx context.Context, container containerd.Container, config types.ExecConfig) (*specs.Process, error) {
	spec, err := container.Spec(ctx)
	if err != nil {
		return nil, err
	}

	userOpts, err := s.generateUserOpts(config.User)
	if err != nil {
		return nil, err
	}
	if userOpts != nil {
		c, err := container.Info(ctx)
		if err != nil {
			return nil, err
		}
		for _, opt := range userOpts {
			if err := opt(ctx, s.client.GetClient(), &c, spec); err != nil {
				return nil, err
			}
		}
	}

	pspec := spec.Process
	pspec.Terminal = config.Tty
	if pspec.Terminal && config.ConsoleSize != nil {
		pspec.ConsoleSize = &specs.Box{Height: config.ConsoleSize[0], Width: config.ConsoleSize[1]}
	}
	pspec.Args = config.Cmd

	if config.WorkingDir != "" {
		pspec.Cwd = config.WorkingDir
	}
	envs := config.Env
	pspec.Env = flagutil.ReplaceOrAppendEnvValues(pspec.Env, envs)

	if config.Privileged {
		err = s.setExecCapabilities(pspec)
		if err != nil {
			return nil, err
		}
	}

	return pspec, nil
}

func (s *service) generateUserOpts(user string) ([]oci.SpecOpts, error) {
	var opts []oci.SpecOpts
	if user != "" {
		opts = append(opts, s.client.OCISpecWithUser(user), withResetAdditionalGIDs(), s.client.OCISpecWithAdditionalGIDs(user))
	}
	return opts, nil
}

func (s *service) setExecCapabilities(pspec *specs.Process) error {
	if pspec.Capabilities == nil {
		pspec.Capabilities = &specs.LinuxCapabilities{}
	}
	allCaps, err := s.client.GetCurrentCapabilities()
	if err != nil {
		return err
	}
	pspec.Capabilities.Bounding = allCaps
	pspec.Capabilities.Permitted = pspec.Capabilities.Bounding
	pspec.Capabilities.Inheritable = pspec.Capabilities.Bounding
	pspec.Capabilities.Effective = pspec.Capabilities.Bounding

	// https://github.com/moby/moby/pull/36466/files
	// > `docker exec --privileged` does not currently disable AppArmor
	// > profiles. Privileged configuration of the container is inherited
	return nil
}

func withResetAdditionalGIDs() oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {
		s.Process.User.AdditionalGids = nil
		return nil
	}
}
