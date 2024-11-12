// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/containerd/containerd"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/api/types/cri"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/labels/k8slabels"
	"github.com/containerd/nerdctl/v2/pkg/logging"
	"github.com/moby/moby/pkg/stdcopy"

	"github.com/runfinch/finch-daemon/api/types"
)

// Attach attaches the stdout and stderr to the container using nerdctl logs
//
// TODO: Investigate fully implementing attach. See David's previous MR:
// https://github.com/containerd/nerdctl/pull/2108
func (s *service) Attach(ctx context.Context, cid string, opts *types.AttachOptions) error {
	// fetch container
	con, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}
	s.logger.Debugf("attaching container: %s", con.ID())

	// set up io streams
	outStream, errStream, stopChannel, printSuccessResp, err := opts.GetStreams()
	if err != nil {
		return err
	}

	// if the caller wants neither to stream nor to view logs, return nothing
	if !opts.Stream && !opts.Logs {
		printSuccessResp()
		return nil
	}

	if opts.MuxStreams {
		errStream = stdcopy.NewStdWriter(errStream, stdcopy.Stderr)
		outStream = stdcopy.NewStdWriter(outStream, stdcopy.Stdout)
	}
	// TODO: implement stdin for a full attach implementation
	var (
		// stdin io.Reader
		stdout, stderr io.Writer
	)
	// if opts.UseStdin {
	// stdin = inStream
	// }
	if opts.UseStdout {
		stdout = outStream
	}
	if opts.UseStderr {
		stderr = errStream
	}

	// Logs option determine if we are viewing the full logs (tail = 0) or just the current output
	// with since = 0s. There is no way to use tail = 0 on nerdctl to return no logs
	since := "0s"
	if opts.Logs {
		since = ""
	}

	// assemble log options and call attachLogs (based off of nerdctl's container.Logs)
	logOpts := ncTypes.ContainerLogsOptions{
		Stdout:     stdout,
		Stderr:     stderr,
		GOptions:   ncTypes.GlobalCommandOptions{},
		Follow:     opts.Stream,
		Timestamps: false,
		Tail:       0,
		Since:      since,
		Until:      "",
	}
	err = s.attachLogs(ctx, con, logOpts, stopChannel, printSuccessResp)
	if err != nil {
		s.logger.Debugf("failed to attach to the container: %s", cid)
		return err
	}
	return nil
}

// attachLogs sets up the logs and channels to be attached. Adapted from
// github.com/containerd/nerdctl/pkg/cmd/container.Logs to pass a stop channel
// and a success response message.
func (s *service) attachLogs(
	ctx context.Context,
	con containerd.Container,
	options ncTypes.ContainerLogsOptions,
	stopChannel chan os.Signal,
	printSuccessResp func(),
) error {
	dataStore, err := s.nctlContainerSvc.GetDataStore()
	if err != nil {
		return err
	}

	l, err := con.Labels(ctx)
	if err != nil {
		return err
	}

	logPath, err := getLogPath(ctx, con)
	if err != nil {
		return err
	}

	status := s.client.GetContainerStatus(ctx, con)
	if status != containerd.Running {
		options.Follow = false
		if status == containerd.Stopped {
			// NOTE: This is a temporary workaround to fix the logger issue where strings without newline are not logged:
			// https://github.com/containerd/nerdctl/issues/2313
			// TODO: Remove this logic when the issue is fixed in nerdctl
			// delete old task to shutdown the logger and print buffered data
			task, err := con.Task(ctx, nil)
			if err == nil {
				task.Delete(ctx)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	if options.Follow {
		task, waitCh, err := s.client.GetContainerTaskWait(ctx, nil, con)
		if err != nil {
			return fmt.Errorf("failed to get wait channel for task %#v: %s", task, err)
		}

		// setup goroutine to send stop event if container task finishes:
		go func() {
			<-waitCh
			s.logger.Debugf("container task has finished, sending kill signal to log viewer")

			// NOTE: This is a temporary workaround to fix the logger issue where strings without newline are not logged:
			// https://github.com/containerd/nerdctl/issues/2313
			// TODO: Remove this logic when the issue is fixed in nerdctl
			// delete finished task to shutdown the logger and print buffered data
			task.Delete(ctx)
			time.Sleep(100 * time.Millisecond)

			stopChannel <- os.Interrupt
		}()
	}

	logViewOpts := logging.LogViewOptions{
		ContainerID:       con.ID(),
		Namespace:         l[labels.Namespace],
		DatastoreRootPath: dataStore,
		LogPath:           logPath,
		Follow:            options.Follow,
		Timestamps:        options.Timestamps,
		Tail:              options.Tail,
		Since:             options.Since,
		Until:             options.Until,
	}
	logViewer, err := s.nctlContainerSvc.LoggingInitContainerLogViewer(l, logViewOpts, stopChannel, options.GOptions.Experimental)
	if err != nil {
		return err
	}

	// Print success response to the connection, then return logs
	printSuccessResp()
	return s.nctlContainerSvc.LoggingPrintLogsTo(options.Stdout, options.Stderr, logViewer)
}

// getLogPath gets the log path for the container to be attached. Original from
// github.com/containerd/nerdctl/pkg/cmd/container.getLogPath.
func getLogPath(ctx context.Context, container containerd.Container) (string, error) {
	extensions, err := container.Extensions(ctx)
	if err != nil {
		return "", fmt.Errorf("get extensions for container %s,failed: %#v", container.ID(), err)
	}
	metaData := extensions[k8slabels.ContainerMetadataExtension]
	var meta cri.ContainerMetadata
	if metaData != nil {
		err = meta.UnmarshalJSON(metaData.GetValue())
		if err != nil {
			return "", fmt.Errorf("unmarshal extensions for container %s,failed: %#v", container.ID(), err)
		}
	}

	return meta.LogPath, nil
}
