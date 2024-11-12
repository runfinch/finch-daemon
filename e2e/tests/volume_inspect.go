// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"net/http"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// VolumeInspect tests volume inspect API - GET /volumes/{volume_name}.
func VolumeInspect(opt *option.Option) {
	Describe("Inspect volume API", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})
		It("should return volume details", func() {
			command.Run(opt, "volume", "create", testVolumeName, "--label", "foo=bar")
			volumeShouldExist(opt, testVolumeName)

			apiUrl := client.ConvertToFinchUrl(version, "/volumes/"+testVolumeName)
			res, err := uClient.Get(apiUrl)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			// Read response and ensure correct volume inspected.
			var volumesResp native.Volume
			err = json.NewDecoder(res.Body).Decode(&volumesResp)
			Expect(err).Should(BeNil())
			Expect(volumesResp.Name).Should(Equal(testVolumeName))
			Expect(*volumesResp.Labels).Should(HaveKeyWithValue("foo", "bar"))
		})
		It("should return not found error", func() {
			// dont create the volume and try to get the details.
			apiUrl := client.ConvertToFinchUrl(version, "/volumes/"+testVolumeName)
			res, err := uClient.Get(apiUrl)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))

			// Read response and ensure not found error is returned.
			var errRes response.Error
			err = json.NewDecoder(res.Body).Decode(&errRes)
			Expect(err).Should(BeNil())
			Expect(errRes.Message).Should(Not(BeEmpty()))
		})
	})
}
