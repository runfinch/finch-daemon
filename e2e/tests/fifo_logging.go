// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// FifoLogging tests FIFO-based in-process logging behavior.
func FifoLogging(opt *option.Option) {
	Describe("FIFO-based logging", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should capture stdout via FIFO logging and return it via GET /logs", func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage,
				"sh", "-c", "echo fifo-test-output; sleep infinity")
			time.Sleep(1 * time.Second)

			relativeUrl := fmt.Sprintf("/containers/%s/logs?stdout=1&stderr=1&follow=0&tail=0", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			res.Body.Close()

			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			Expect(string(body)).Should(ContainSubstring("fifo-test-output"))
		})

		It("should tag stderr output separately from stdout", func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage,
				"sh", "-c", "echo stdout-line; echo stderr-line >&2; sleep infinity")
			time.Sleep(1 * time.Second)

			// Get only stdout
			stdoutUrl := fmt.Sprintf("/containers/%s/logs?stdout=1&stderr=0&follow=0&tail=0", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, stdoutUrl))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			res.Body.Close()
			Expect(string(body)).Should(ContainSubstring("stdout-line"))
			Expect(string(body)).ShouldNot(ContainSubstring("stderr-line"))

			// Get only stderr
			stderrUrl := fmt.Sprintf("/containers/%s/logs?stdout=0&stderr=1&follow=0&tail=0", testContainerName)
			res, err = uClient.Get(client.ConvertToFinchUrl(version, stderrUrl))
			Expect(err).Should(BeNil())
			body, err = io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			res.Body.Close()
			Expect(string(body)).Should(ContainSubstring("stderr-line"))
			Expect(string(body)).ShouldNot(ContainSubstring("stdout-line"))
		})

		It("should append logs on container restart rather than overwrite", func() {
			// Create and start via HTTP API so both runs go through customStart.
			createUrl := client.ConvertToFinchUrl(version, "/containers/create")
			options := types.ContainerCreateRequest{}
			options.Image = defaultImage
			options.Cmd = []string{"sh", "-c", "echo run-1; sleep infinity"}

			statusCode, ctr := createContainer(uClient, createUrl, testContainerName, options)
			Expect(statusCode).Should(Equal(http.StatusCreated))
			Expect(ctr.ID).ShouldNot(BeEmpty())

			startUrl := fmt.Sprintf("/containers/%s/start", testContainerName)
			res, err := uClient.Post(client.ConvertToFinchUrl(version, startUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			res.Body.Close()

			time.Sleep(1 * time.Second)

			// Restart via HTTP API — this goes through our customStart path.
			restartUrl := fmt.Sprintf("/containers/%s/restart", testContainerName)
			res, err = uClient.Post(client.ConvertToFinchUrl(version, restartUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			res.Body.Close()
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))

			time.Sleep(2 * time.Second)

			logsUrl := fmt.Sprintf("/containers/%s/logs?stdout=1&stderr=1&follow=0&tail=0", testContainerName)
			res, err = uClient.Get(client.ConvertToFinchUrl(version, logsUrl))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			res.Body.Close()

			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			// Both runs should be present (append, not overwrite)
			Expect(strings.Count(string(body), "run-1")).Should(BeNumerically(">=", 2))
		})

		It("should resume logging after daemon restart", Serial, func() {
			// This test restarts the daemon process and verifies log reattach works.
			// It requires the daemon binary to be installed and the test to have
			// permission to kill/start the daemon.

			daemonExe := getFinchDaemonExe()
			socketPath := getSocketPath()

			// Start a long-running container that produces periodic output
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage,
				"sh", "-c", "i=0; while true; do i=$((i+1)); echo tick-$i; sleep 1; done")
			time.Sleep(3 * time.Second)

			// Verify initial logs are captured
			relativeUrl := fmt.Sprintf("/containers/%s/logs?stdout=1&stderr=1&follow=0&tail=0", testContainerName)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			res.Body.Close()
			Expect(string(body)).Should(ContainSubstring("tick-1"))

			// Kill the daemon
			killDaemon()

			// Wait a moment for the container to produce more output while daemon is down
			time.Sleep(2 * time.Second)

			// Restart the daemon
			startDaemon(daemonExe, socketPath)

			// Wait for daemon to be ready
			waitForDaemon(uClient, version)

			// Wait for container to produce more output after reattach
			time.Sleep(3 * time.Second)

			// Verify logs include output from after the restart
			res, err = uClient.Get(client.ConvertToFinchUrl(version, relativeUrl))
			Expect(err).Should(BeNil())
			body, err = io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			res.Body.Close()

			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			// Should have ticks from after the restart (tick numbers > 5)
			Expect(string(body)).Should(ContainSubstring("tick-6"))
		})
	})
}

func getFinchDaemonExe() string {
	exe := os.Getenv("FINCH_DAEMON_EXE")
	if exe != "" {
		return exe
	}
	path, err := exec.LookPath("finch-daemon")
	if err != nil {
		Skip("finch-daemon binary not found in PATH; skipping daemon restart test")
	}
	return path
}

func getSocketPath() string {
	host := os.Getenv("DOCKER_HOST")
	// DOCKER_HOST is like "unix:///run/finch.sock"
	return strings.TrimPrefix(host, "unix://")
}

func killDaemon() {
	// Find and kill the finch-daemon process.
	cmd := exec.Command("pkill", "-TERM", "finch-daemon")
	_ = cmd.Run() // best-effort kill; process may not exist
	time.Sleep(1 * time.Second)
}

func startDaemon(exe, socketPath string) {
	uid := os.Getuid()
	cmd := exec.Command("sudo", exe, "--debug", "--socket-owner", fmt.Sprintf("%d", uid)) //nolint:gosec // args are constructed from known env vars
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	Expect(err).Should(BeNil())
}

func waitForDaemon(uClient *http.Client, version string) {
	pingUrl := client.ConvertToFinchUrl(version, "/_ping")
	for i := 0; i < 30; i++ {
		resp, err := uClient.Get(pingUrl)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	Fail("daemon did not become ready within 15 seconds after restart")
}
