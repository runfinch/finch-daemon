// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// VolumeList tests listing volumes.
func VolumeList(opt *option.Option) {
	Describe("list volumes", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			httpCreateVolume(uClient, version, testVolumeName, map[string]string{"foo": "bar"})
			httpCreateVolume(uClient, version, testVolumeName2, map[string]string{"baz": "biz"})
			volumeShouldExist(opt, testVolumeName)
			volumeShouldExist(opt, testVolumeName2)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})
		It("should list volumes", func() {
			url := client.ConvertToFinchUrl(version, "/volumes")
			res, err := uClient.Get(url)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			// Read response and ensure more than one volume listed.
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()
			var volumesResp types.VolumesListResponse
			err = json.Unmarshal(body, &volumesResp)
			Expect(err).Should(BeNil())
			Expect(len(volumesResp.Volumes)).Should(BeNumerically(">", 0))
		})
		It("should list volumes and filter them", func() {
			urlEncodedJsonFilter := url.QueryEscape(`{"labels": ["foo"]}`)
			relativeUrl := fmt.Sprintf("/volumes?filters=%s", urlEncodedJsonFilter)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			res, err := uClient.Get(url)

			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			// Read response and expect len to be 1, volume name to be
			// the testVolumeName we filtered for.
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			var volumesResp types.VolumesListResponse
			err = json.Unmarshal(body, &volumesResp)
			Expect(err).Should(BeNil())
			Expect(len(volumesResp.Volumes)).Should(Equal(1))
			Expect(volumesResp.Volumes[0].Name).Should(Equal(testVolumeName))
		})
	})
}
