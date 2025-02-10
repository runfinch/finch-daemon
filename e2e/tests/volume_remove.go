// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
)

// VolumeRemove tests the volume remove API.
func VolumeRemove(opt *option.Option) {
	Describe("Remove volume API", func() {
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
		It("should remove a volume", func() {
			command.Run(opt, "volume", "create", testVolumeName)
			volumeShouldExist(opt, testVolumeName)
			apiUrl := client.ConvertToFinchUrl(version, "/volumes/"+testVolumeName)
			req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
			Expect(err).Should(BeNil())
			res, err := uClient.Do(req)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			volumeShouldNotExist(opt, testVolumeName)
		})
		It("should remove a volume with force=true", func() {
			command.Run(opt, "volume", "create", testVolumeName)
			volumeShouldExist(opt, testVolumeName)
			apiUrl := client.ConvertToFinchUrl(version, "/volumes/"+testVolumeName+"?force=true")
			req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
			Expect(err).Should(BeNil())
			res, err := uClient.Do(req)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			volumeShouldNotExist(opt, testVolumeName)
		})
		It("should fail to remove a volume that is in use", func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, "-v", testVolumeName+":/data",
				defaultImage, "sleep", "infinity")
			apiUrl := client.ConvertToFinchUrl(version, "/volumes/"+testVolumeName)
			req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
			Expect(err).Should(BeNil())
			res, err := uClient.Do(req)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
			volumeShouldExist(opt, testVolumeName)
		})
	})
}
