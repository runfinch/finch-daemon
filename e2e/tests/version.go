// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
	"github.com/runfinch/finch-daemon/pkg/api/types"
)

// SystemVersion tests the `Get /version` API.
func SystemVersion(opt *option.Option) {
	Describe("version API", func() {
		var uClient *http.Client
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			command.RemoveAll(opt)
		})
		It("should successfully get the version info", func() {
			res, err := uClient.Get(client.ConvertToFinchUrl("", "/version"))
			Expect(err).ShouldNot(HaveOccurred())
			jd := json.NewDecoder(res.Body)
			var v types.VersionInfo
			err = jd.Decode(&v)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v.Version).ShouldNot(BeNil())
			Expect(v.Platform.Name).ShouldNot(BeEmpty())
			Expect(v.GitCommit).ShouldNot(BeEmpty())
			Expect(v.ApiVersion).Should(Equal("1.43"))
			Expect(v.MinAPIVersion).Should(Equal("1.35"))
			Expect(v.Components).ShouldNot(BeEmpty())
			Expect(v.Experimental).Should(BeFalse())
			Expect(strings.ToLower(v.Os)).Should(Equal("linux"))
			Expect(strings.ToLower(v.Arch)).Should(Or(Equal("x86_64"), Equal(("aarch64"))))
			Expect(v.KernelVersion).ShouldNot(BeEmpty())
		})
	})
}
