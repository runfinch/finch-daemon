// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/signalutil"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// StdinCloser is from https://github.com/containerd/containerd/blob/v1.4.3/cmd/ctr/commands/tasks/exec.go#L181-L194
type StdinCloser struct {
	mu     sync.Mutex
	Stdin  io.ReadCloser
	Closer func()
	closed bool
}

func (s *StdinCloser) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, syscall.EBADF
	}
	n, err := s.Stdin.Read(p)
	if err != nil {
		if s.Closer != nil {
			s.Closer()
			s.closed = true
		}
	}
	return n, err
}

func (s *service) Start(ctx context.Context, options *types.ExecStartOptions) error {
	var attach cio.Attach
	var in io.Reader
	stdinC := &StdinCloser{
		Stdin: options.Stdin,
	}
	if !options.Detach {
		in = stdinC
		attach = cio.NewAttach(cio.WithStreams(in, options.Stdout, options.Stderr))
	}
	exec, err := s.loadExecInstance(ctx, options.ConID, options.ExecID, attach)
	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "task not found"):
			return errdefs.NewConflict(fmt.Errorf("container %s is not running", options.ConID))
		default:
			return err
		}
	}

	taskStatus, err := exec.Task.Status(ctx)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return errdefs.NewConflict(fmt.Errorf("container %s is not running", options.ConID))
		}
		return err
	}
	if taskStatus.Status != containerd.Running {
		return errdefs.NewConflict(fmt.Errorf("container %s is not running", options.ConID))
	}

	stdinC.Closer = func() {
		exec.Process.CloseIO(ctx, containerd.WithStdinCloser)
	}

	statusC, err := exec.Process.Wait(ctx)
	if err != nil {
		return err
	}

	if !options.Detach {
		if options.Tty && options.ConsoleSize != nil {
			if err = exec.Process.Resize(ctx, uint32(options.ConsoleSize[1]), uint32(options.ConsoleSize[0])); err != nil {
				s.logger.Errorf("could not resize console: %v", err)
			}
		}
		sigc := signalutil.ForwardAllSignals(ctx, exec.Process)
		defer signalutil.StopCatch(sigc)
	}

	if options.SuccessResponse != nil {
		options.SuccessResponse()
	}

	if err = exec.Process.Start(ctx); err != nil {
		return err
	}
	if options.Detach {
		return nil
	}

	status := <-statusC
	code, _, err := status.Result()
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("exec failed with exit code %d", code)
	}

	return nil
}
