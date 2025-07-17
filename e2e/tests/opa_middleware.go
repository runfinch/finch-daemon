// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// OpaMiddlewareTest tests the OPA functionality.
func OpaMiddlewareTest(opt *option.Option) {
	Describe("test opa middleware functionality", func() {
		var (
			uClient                     *http.Client
			version                     string
			wantContainerName           string
			containerCreateOptions      types.ContainerCreateRequest
			createUrl                   string
			unimplementedUnspecifiedUrl string
			unimplementedSpecifiedUrl   string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			wantContainerName = fmt.Sprintf("/%s", testContainerName)
			// set default container containerCreateOptions
			containerCreateOptions = types.ContainerCreateRequest{}
			containerCreateOptions.Image = defaultImage
			createUrl = client.ConvertToFinchUrl(version, "/containers/create")
			unimplementedUnspecifiedUrl = client.ConvertToFinchUrl(version, "/secrets")
			unimplementedSpecifiedUrl = client.ConvertToFinchUrl(version, "/swarm")
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})
		It("should allow GET version API request", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl("", "/version"))
			Expect(err).ShouldNot(HaveOccurred())
			jd := json.NewDecoder(res.Body)
			var v types.VersionInfo
			err = jd.Decode(&v)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v.Version).ShouldNot(BeNil())
			Expect(v.ApiVersion).Should(Equal("1.43"))
			fmt.Println(version)
		})

		It("shold allow GET containers API request", func() {
			id := command.StdoutStr(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")
			want := []types.ContainerListItem{
				{
					Id:    id[:12],
					Names: []string{wantContainerName},
				},
			}

			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/json"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			var got []types.ContainerListItem
			err = json.NewDecoder(res.Body).Decode(&got)
			Expect(err).Should(BeNil())
			Expect(len(got)).Should(Equal(2))
			got = filterContainerList(got)
			Expect(got).Should(ContainElements(want))
		})

		It("shold disallow POST containers/create API request", func() {
			containerCreateOptions.Cmd = []string{"echo", "hello world"}

			reqBody, err := json.Marshal(containerCreateOptions)
			Expect(err).Should(BeNil())

			fmt.Println("createUrl = ", createUrl)
			res, _ := uClient.Post(createUrl, "application/json", bytes.NewReader(reqBody))

			Expect(res.StatusCode).Should(Equal(http.StatusForbidden))
		})

		It("should fail unimplemented API calls, fail via daemon", func() {
			fmt.Println("incompatibleUrl = ", unimplementedUnspecifiedUrl)
			res, _ := uClient.Get(unimplementedUnspecifiedUrl)

			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
		})

		It("should fail non implemented API calls,even if specified in the rego file", func() {
			fmt.Println("incompatibleUrl = ", unimplementedSpecifiedUrl)
			res, _ := uClient.Get(unimplementedSpecifiedUrl)

			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
		})

		// Add this test to OpaMiddlewareTest function
		It("should handle rego file permissions correctly", func() {
			// Create a temporary rego file with overly permissive permissions
			tmpDir := GinkgoT().TempDir()

			regoPath := filepath.Join(tmpDir, "test.rego")
			regoContent := []byte(`package finch.authz
			default allow = false`)

			var err error
			err = os.WriteFile(regoPath, regoContent, 0644)
			Expect(err).NotTo(HaveOccurred())

			// Try to start daemon with overly permissive file
			cmd := exec.Command(GetFinchDaemonExe(), //nolint:gosec // G204: This is a test file with controlled inputs
				"--socket-addr", "/run/test.sock",
				"--pidfile", "/run/test.pid",
				"--rego-file", regoPath,
				"--experimental")
			err = cmd.Run()

			// Should fail due to permissions
			Expect(err).To(HaveOccurred())

			// For the second test with skip-check:
			cmd = exec.Command(GetFinchDaemonExe(), //nolint:gosec // G204: This is a test file with controlled inputs
				"--socket-addr", "/run/test.sock",
				"--pidfile", "/run/test.pid",
				"--rego-file", regoPath,
				"--experimental",
				"--skip-rego-perm-check")

			// Start the process in background
			err = cmd.Start()
			Expect(err).NotTo(HaveOccurred())

			// Give it a moment to initialize
			time.Sleep(1 * time.Second)

			// Kill the process
			err = cmd.Process.Kill()
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
