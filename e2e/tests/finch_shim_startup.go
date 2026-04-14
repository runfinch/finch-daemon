// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// FinchShimStartup tests that finch-daemon fails fast at startup when finch-shim is absent
// from PATH, rather than silently failing at container create time.
// This test is self-contained: it does not use the shared daemon started by the test suite.
// It starts its own short-lived daemon process with a sanitised PATH.
func FinchShimStartup() {
	Describe("finch-daemon startup without finch-shim", func() {
		It("should exit with a non-zero status and reference finch-shim in the error output", func() {
			daemonExe := GetFinchDaemonExe()
			if _, err := os.Stat(daemonExe); os.IsNotExist(err) {
				Skip("finch-daemon binary not found at " + daemonExe + "; skipping startup test")
			}

			// Build a PATH that contains everything except finch-shim.
			// We do this by constructing a temp dir with symlinks to every binary in PATH
			// except finch-shim, then setting that as the only PATH entry.
			// Simpler approach: just strip any directory that contains finch-shim.
			cleanPath := pathWithoutBinary("finch-shim")

			cmd := exec.Command(daemonExe) //nolint:gosec // path is resolved from a known env var or hardcoded fallback, not user input
			cmd.Env = append(os.Environ(), "PATH="+cleanPath)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			// The daemon should exit quickly once it fails the startup check.
			// We give it 5 seconds; if it hasn't exited by then something is wrong.
			err := runWithTimeout(cmd, 5*time.Second)

			// We expect a non-zero exit — either the process exited with an error,
			// or it timed out (which means it didn't fail fast, also a test failure).
			Expect(err).ShouldNot(BeNil(), "expected daemon to exit with error when finch-shim is absent")

			errOutput := stderr.String()
			Expect(errOutput).Should(ContainSubstring("finch-shim"),
				"expected error output to reference finch-shim, got: %s", errOutput)
		})
	})
}

// pathWithoutBinary returns a PATH string that excludes any directory containing
// the named binary. This lets us simulate finch-shim being absent without modifying
// the filesystem.
func pathWithoutBinary(binary string) string {
	original := os.Getenv("PATH")
	dirs := strings.Split(original, ":")
	var filtered []string
	for _, dir := range dirs {
		candidate := dir + "/" + binary
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			filtered = append(filtered, dir)
		}
	}
	return strings.Join(filtered, ":")
}

// runWithTimeout starts cmd and waits up to timeout for it to exit.
// Returns the exit error if the process exited with non-zero status,
// or a timeout error if it did not exit in time.
func runWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		return &timeoutError{timeout: timeout}
	}
}

type timeoutError struct{ timeout time.Duration }

func (e *timeoutError) Error() string {
	return "process did not exit within " + e.timeout.String()
}
