// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/nerdctl/v2/pkg/logging/jsonfile"
	"github.com/sirupsen/logrus"
)

// containerLogPath returns the json-file log path for a container.
// Matches nerdctl's convention: {dataStore}/containers/{ns}/{id}/{id}-json.log.
func containerLogPath(dataStore, ns, containerID string) string {
	return filepath.Join(dataStore, "containers", ns, containerID, containerID+"-json.log")
}

// startLogCopiers starts goroutines that read container stdout/stderr from
// FIFO read ends and write json-file formatted log entries. The file handle
// is closed when copiers finish (on container exit); the file persists on
// disk until the container is removed.
func startLogCopiers(logPath string, dio *cio.DirectIO) {
	if dio == nil {
		return
	}

	// Ensure the log directory exists (first start for this container).
	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		logrus.Warnf("failed to create log directory %s: %v", filepath.Dir(logPath), err)
		return
	}

	// O_APPEND so container restarts append to the same file rather than overwrite.
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		logrus.Warnf("failed to open log file %s: %v", logPath, err)
		return
	}

	// Mutex guards concurrent writes from the stdout and stderr goroutines.
	// Each write creates a new json.Encoder to avoid shared encoder state.
	var mu sync.Mutex
	var wg sync.WaitGroup

	// copyStream reads lines from a FIFO read end and writes json-file entries.
	copyStream := func(reader io.ReadCloser, stream string) {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			entry := jsonfile.Entry{
				Log:    scanner.Text() + "\n",
				Stream: stream,
				Time:   time.Now().UTC(),
			}
			mu.Lock()
			json.NewEncoder(logFile).Encode(entry) //nolint:errcheck // best-effort log writing
			mu.Unlock()
		}
	}

	// Spawn a copier goroutine for each available stream.
	if dio.Stdout != nil {
		wg.Add(1)
		go copyStream(dio.Stdout, "stdout")
	}
	if dio.Stderr != nil {
		wg.Add(1)
		go copyStream(dio.Stderr, "stderr")
	}

	// Close the file handle once both copiers have drained their FIFOs.
	go func() {
		wg.Wait()
		logFile.Sync()  //nolint:errcheck // best-effort flush
		logFile.Close() //nolint:errcheck // best-effort close
	}()
}
