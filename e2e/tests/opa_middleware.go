// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

		It("should not allow updates to the rego file", func() {
			regoFilePath := os.Getenv("REGO_FILE_PATH")
			Expect(regoFilePath).NotTo(BeEmpty(), "REGO_FILE_PATH environment variable should be set")

			fileInfo, err := os.Stat(regoFilePath)
			Expect(err).NotTo(HaveOccurred(), "Failed to get Rego file info")

			// Check file permissions
			mode := fileInfo.Mode()
			Expect(mode.Perm()).To(Equal(os.FileMode(0400)), "Rego file should be read-only (0400)")
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
	})
}
