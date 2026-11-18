// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/nerdctl/v2/pkg/logging/jsonfile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// newTestDirectIO creates a DirectIO with the given stdout and stderr readers.
// This works around the fact that DirectIO embeds pipes via an unexported struct,
// so we can't use struct literal syntax directly in tests.
func newTestDirectIO(stdout, stderr io.ReadCloser) *cio.DirectIO {
	dio := &cio.DirectIO{}
	dio.Stdout = stdout
	dio.Stderr = stderr
	return dio
}

var _ = Describe("copier", func() {
	Context("containerLogPath", func() {
		It("should return correct path format", func() {
			result := containerLogPath("/data/store", "finch", "abc123")
			Expect(result).Should(Equal("/data/store/containers/finch/abc123/abc123-json.log"))
		})
	})

	Context("startLogCopiers", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "copier-test-*")
			Expect(err).Should(BeNil())
		})
		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		It("should write stdout in json-file format", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stdoutR, stdoutW := io.Pipe()

			dio := newTestDirectIO(stdoutR, nil)
			startLogCopiers(logPath, dio)

			stdoutW.Write([]byte("hello\n"))
			stdoutW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())

			var entry jsonfile.Entry
			err = json.Unmarshal(data[:len(strings.TrimSpace(string(data)))], &entry)
			Expect(err).Should(BeNil())
			Expect(entry.Log).Should(Equal("hello\n"))
			Expect(entry.Stream).Should(Equal("stdout"))
			Expect(entry.Time).ShouldNot(BeZero())
		})

		It("should tag stderr lines with stderr stream", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stderrR, stderrW := io.Pipe()

			dio := newTestDirectIO(nil, stderrR)
			startLogCopiers(logPath, dio)

			stderrW.Write([]byte("error msg\n"))
			stderrW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())

			var entry jsonfile.Entry
			err = json.Unmarshal([]byte(strings.TrimSpace(string(data))), &entry)
			Expect(err).Should(BeNil())
			Expect(entry.Log).Should(Equal("error msg\n"))
			Expect(entry.Stream).Should(Equal("stderr"))
		})

		It("should handle both stdout and stderr simultaneously without corruption", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stdoutR, stdoutW := io.Pipe()
			stderrR, stderrW := io.Pipe()

			dio := newTestDirectIO(stdoutR, stderrR)
			startLogCopiers(logPath, dio)

			// Write interleaved lines
			for i := 0; i < 10; i++ {
				stdoutW.Write([]byte("out\n"))
				stderrW.Write([]byte("err\n"))
			}
			stdoutW.Close()
			stderrW.Close()
			time.Sleep(200 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())

			// Each line should be valid JSON
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			Expect(len(lines)).Should(Equal(20))
			for _, line := range lines {
				var entry jsonfile.Entry
				err := json.Unmarshal([]byte(line), &entry)
				Expect(err).Should(BeNil(), "corrupted JSON line: %s", line)
				Expect(entry.Stream).Should(SatisfyAny(Equal("stdout"), Equal("stderr")))
			}
		})

		It("should append to existing log file (container restart)", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			// Pre-populate with existing content
			existing := `{"log":"old line\n","stream":"stdout","time":"2026-01-01T00:00:00Z"}` + "\n"
			err := os.WriteFile(logPath, []byte(existing), 0600)
			Expect(err).Should(BeNil())

			stdoutR, stdoutW := io.Pipe()
			dio := newTestDirectIO(stdoutR, nil)
			startLogCopiers(logPath, dio)

			stdoutW.Write([]byte("new line\n"))
			stdoutW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			Expect(len(lines)).Should(Equal(2))
			Expect(lines[0]).Should(ContainSubstring("old line"))
			Expect(lines[1]).Should(ContainSubstring("new line"))
		})

		It("should close log file when all copiers finish", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stdoutR, stdoutW := io.Pipe()
			stderrR, stderrW := io.Pipe()

			dio := newTestDirectIO(stdoutR, stderrR)
			startLogCopiers(logPath, dio)

			stdoutW.Write([]byte("line\n"))
			stdoutW.Close()
			stderrW.Close()
			// Give the closer goroutine time to run
			time.Sleep(200 * time.Millisecond)

			// Verify file was written
			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())
			Expect(string(data)).Should(ContainSubstring("line"))
		})

		It("should handle nil Stdout gracefully (TTY mode with only stderr)", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stderrR, stderrW := io.Pipe()

			dio := newTestDirectIO(nil, stderrR)
			startLogCopiers(logPath, dio)

			stderrW.Write([]byte("tty err\n"))
			stderrW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())
			Expect(string(data)).Should(ContainSubstring("tty err"))
		})

		It("should handle nil Stderr gracefully", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stdoutR, stdoutW := io.Pipe()

			dio := newTestDirectIO(stdoutR, nil)
			startLogCopiers(logPath, dio)

			stdoutW.Write([]byte("only stdout\n"))
			stdoutW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())
			Expect(string(data)).Should(ContainSubstring("only stdout"))
		})

		It("should preserve empty lines", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stdoutR, stdoutW := io.Pipe()

			dio := newTestDirectIO(stdoutR, nil)
			startLogCopiers(logPath, dio)

			// bufio.Scanner splits on \n; an "empty line" is a line with just ""
			stdoutW.Write([]byte("before\n\nafter\n"))
			stdoutW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			// "before", "", "after" → 3 entries
			Expect(len(lines)).Should(Equal(3))

			var entry jsonfile.Entry
			json.Unmarshal([]byte(lines[1]), &entry)
			Expect(entry.Log).Should(Equal("\n"))
		})

		It("should write each line as a separate JSON entry", func() {
			logPath := filepath.Join(tmpDir, "test-json.log")
			stdoutR, stdoutW := io.Pipe()

			dio := newTestDirectIO(stdoutR, nil)
			startLogCopiers(logPath, dio)

			stdoutW.Write([]byte("line1\nline2\nline3\n"))
			stdoutW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			Expect(len(lines)).Should(Equal(3))

			for i, line := range lines {
				var entry jsonfile.Entry
				err := json.Unmarshal([]byte(line), &entry)
				Expect(err).Should(BeNil())
				expected := "line" + string(rune('1'+i)) + "\n"
				Expect(entry.Log).Should(Equal(expected))
			}
		})

		It("should create log directory if it does not exist", func() {
			logPath := filepath.Join(tmpDir, "nested", "dir", "test-json.log")
			stdoutR, stdoutW := io.Pipe()

			dio := newTestDirectIO(stdoutR, nil)
			startLogCopiers(logPath, dio)

			stdoutW.Write([]byte("created\n"))
			stdoutW.Close()
			time.Sleep(100 * time.Millisecond)

			data, err := os.ReadFile(logPath)
			Expect(err).Should(BeNil())
			Expect(string(data)).Should(ContainSubstring("created"))
		})

		It("should not panic if directory creation fails", func() {
			// Use a path under a file (not a directory) to force MkdirAll failure.
			filePath := filepath.Join(tmpDir, "afile")
			os.WriteFile(filePath, []byte("x"), 0600) //nolint:errcheck // test setup
			logPath := filepath.Join(filePath, "subdir", "test-json.log")

			stdoutR, stdoutW := io.Pipe()
			dio := newTestDirectIO(stdoutR, nil)

			// Should not panic — MkdirAll will fail because "afile" is a file, not a dir.
			startLogCopiers(logPath, dio)
			stdoutW.Close()

			// Log file should not have been created.
			_, err := os.Stat(logPath)
			Expect(err).ShouldNot(BeNil())
		})
	})
})
