// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"io"
	"strconv"

	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/moby/moby/pkg/stdcopy"

	"github.com/runfinch/finch-daemon/api/types"
)

// Logs attaches the stdout and stderr to the container using nerdctl logs.
func (s *service) Logs(ctx context.Context, cid string, opts *types.LogsOptions) error {
	// fetch container
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}
	s.logger.Infof("getting logs for container: %s", con.ID())

	// set up io streams
	outStream, errStream, stopChannel, printSuccessResp, err := opts.GetStreams()
	if err != nil {
		return err
	}

	if opts.MuxStreams {
		errStream = stdcopy.NewStdWriter(errStream, stdcopy.Stderr)
		outStream = stdcopy.NewStdWriter(outStream, stdcopy.Stdout)
	}
	var stdout, stderr io.Writer
	if opts.Stdout {
		stdout = outStream
	}
	if opts.Stderr {
		stderr = errStream
	}

	tail := uint64(0)
	if opts.Tail != "all" && len(opts.Tail) != 0 {
		if tail, err = strconv.ParseUint(opts.Tail, 10, 16); err != nil {
			return err
		}
	}

	// assign until "" if zero is returned as until = 0 (is default to docker to show everything)
	// but nerdctl will interpret that as a time of 0
	until := strconv.FormatInt(opts.Until, 10)
	if until == "0" {
		until = ""
	}

	// assemble log options and call attachLogs (based off of nerdctl's container.Logs)
	logOpts := ncTypes.ContainerLogsOptions{
		Stdout:     stdout,
		Stderr:     stderr,
		GOptions:   ncTypes.GlobalCommandOptions{},
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
		Tail:       uint(tail),
		Since:      strconv.FormatInt(opts.Since, 10),
		Until:      until,
	}
	err = s.attachLogs(ctx, con, logOpts, stopChannel, printSuccessResp)
	if err != nil {
		s.logger.Debugf("failed to retrieve logs for the container: %s", cid)
		return err
	}
	return nil
}
