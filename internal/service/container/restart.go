// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"io"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

func (s *service) Restart(ctx context.Context, cid string, options types.ContainerRestartOptions) error {
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}

	stopOptions := types.ContainerStopOptions{
		Stdout:   io.Discard,
		Stderr:   io.Discard,
		Timeout:  options.Timeout,
		Signal:   "SIGTERM",
		GOptions: options.GOption,
	}

	// restart the container and if error occurs then return error otherwise return nil
	// swallow IsNotModified error on StopContainer for already stopped container, simply call StartContainer
	s.logger.Debugf("restarting container: %s", cid)
	if err := s.nctlContainerSvc.StopContainer(ctx, con.ID(), stopOptions); err != nil && !errdefs.IsNotModified(err) {
		s.logger.Errorf("Failed to stop container: %s. Error: %v", cid, err)
		return err
	}

	startContainerOptions := types.ContainerStartOptions{
		Stdout:     options.Stdout,
		GOptions:   options.GOption,
		DetachKeys: "",
		Attach:     false,
	}

	if err = s.nctlContainerSvc.StartContainer(ctx, cid, startContainerOptions); err != nil {
		s.logger.Errorf("Failed to start container: %s. Error: %v", cid, err)
		return err
	}
	s.logger.Debugf("successfully restarted: %s", cid)
	return nil
}
