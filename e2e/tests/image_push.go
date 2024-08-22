// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/ffs"
	"github.com/runfinch/common-tests/fnet"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
)

func ImagePush(opt *option.Option) {
	Describe("push an image", func() {
		var (
			buildContext string
			port         int
			uClient      *http.Client
			version      string
		)

		BeforeEach(func() {
			command.RemoveAll(opt)
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()

			buildContext = ffs.CreateBuildContext(fmt.Sprintf(`FROM %s
		CMD ["echo", "bar"]
			`, defaultImage))
			DeferCleanup(os.RemoveAll, buildContext)
			port = fnet.GetFreePort()
			command.Run(opt, "run", "-dp", fmt.Sprintf("%d:5000", port), "--name", "registry", registryImage)
		})

		AfterEach(func() {
			command.RemoveAll(opt)
		})

		It("should push an image successfully", func() {
			name := fmt.Sprintf(`localhost:%d/test-push:tag`, port)
			command.Run(opt, "build", "-t", name, buildContext)
			relativeUrl := fmt.Sprintf("/images/%s/push", name)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			command.Run(opt, "rmi", name)
			command.Run(opt, "pull", name)
			imageShouldExist(opt, name)
		})
		It("should successfully push an image into two different registry", func() {
			name := fmt.Sprintf(`localhost:%d/test-push:tag`, port)
			command.Run(opt, "build", "-t", name, buildContext)
			relativeUrl := fmt.Sprintf("/images/%s/push", name)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			command.Run(opt, "rmi", name)
			command.Run(opt, "pull", name)
			imageShouldExist(opt, name)

			// spin off a second registry and tag the previously built image and push it to the second registry.
			secondRegistryPort := fnet.GetFreePort()
			command.Run(opt, "run", "-dp", fmt.Sprintf("%d:5000", secondRegistryPort), "--name", "second-registry", registryImage)

			name2 := fmt.Sprintf(`localhost:%d/test-push:tag`, secondRegistryPort)
			command.Run(opt, "tag", name, name2)
			relativeUrl = fmt.Sprintf("/images/%s/push", name2)
			url = client.ConvertToFinchUrl(version, relativeUrl)
			resp, err = uClient.Post(url, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			command.Run(opt, "rmi", name2)
			command.Run(opt, "pull", name2)
			imageShouldExist(opt, name2)
		})

		It("should return an error when pushing a nonexistent tag", func() {
			nonexistentTag := fmt.Sprintf(`localhost:%d/nonexistent:tag`, port)
			relativeUrl := fmt.Sprintf("/images/%s/push", nonexistentTag)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)

			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
		})
		It("should fail due to network error", func() {
			// pass the wrong port to mimic network failure
			freePort := fnet.GetFreePort()
			name := fmt.Sprintf(`localhost:%d/test-push:tag`, freePort)
			command.Run(opt, "build", "-t", name, buildContext)
			relativeUrl := fmt.Sprintf("/images/%s/push", name)
			url := client.ConvertToFinchUrl(version, relativeUrl)
			resp, err := uClient.Post(url, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			data, _ := io.ReadAll(resp.Body)
			Expect(string(data)).Should(ContainSubstring(`"errorDetail"`))
		})

		// TODO: add a e2e test that push an image to ecr public which requires authentication.
	})
}
