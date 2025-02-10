// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"

	"github.com/runfinch/finch-daemon/api/types"
)

func (s *service) Inspect(ctx context.Context, conID string, execId string) (*types.ExecInspect, error) {
	exec, err := s.loadExecInstance(ctx, conID, execId, nil)
	if err != nil {
		return nil, err
	}

	var running bool
	var exitCode int
	status, err := exec.Process.Status(ctx)
	if err != nil {
		s.logger.Warnf("error getting process status for proc %s: %v", exec.Process.ID(), err)
		running = false
		exitCode = 0
	} else {
		running = status.Status == containerd.Running
		exitCode = int(status.ExitStatus)
	}

	inspectResult := &types.ExecInspect{
		ID:            exec.Process.ID(),
		Running:       running,
		ExitCode:      &exitCode,
		ProcessConfig: &types.ExecProcessConfig{},
		CanRemove:     running,
		ContainerID:   exec.Container.ID(),
		DetachKeys:    []byte(""),
		Pid:           int(exec.Process.Pid()),
	}

	if exec.Process.IO() != nil {
		inspectResult.ProcessConfig.Tty = exec.Process.IO().Config().Terminal
		inspectResult.OpenStdin = exec.Process.IO().Config().Stdin != ""
		inspectResult.OpenStdout = exec.Process.IO().Config().Stdout != ""
		inspectResult.OpenStderr = exec.Process.IO().Config().Stderr != ""
	}

	return inspectResult, nil
}
