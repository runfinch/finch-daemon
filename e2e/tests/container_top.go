// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
)

type ContainerTopResponse struct {
	Titles    []string   `json:"Titles"`
	Processes [][]string `json:"Processes"`
}

func ContainerTop(opt *option.Option) {
	Describe("Get list of processes running inside a Container", func() {
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

		It("should return process list with default ps args", func() {
			command.StdoutStr(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")

			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/top", testContainerName)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusOK))

			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			var topResponse ContainerTopResponse
			err = json.Unmarshal(body, &topResponse)
			Expect(err).Should(BeNil())

			// Validate the response structure
			Expect(topResponse.Titles).ShouldNot(BeEmpty())
			Expect(topResponse.Processes).ShouldNot(BeEmpty())

			// Validate that the sleep process is present
			foundSleepProcess := false
			for _, process := range topResponse.Processes {
				processCmd := strings.Join(process, " ")
				if strings.Contains(processCmd, "sleep infinity") {
					foundSleepProcess = true
					break
				}
			}
			Expect(foundSleepProcess).Should(BeTrue())
		})

		It("should return process list with custom ps args", func() {
			command.StdoutStr(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")

			// Call the top API with custom ps args that produce a specific format
			// Using "-o pid,comm" which will return just PID and command name columns
			customArgs := "-o%20pid,comm"
			url := client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/top?ps_args=%s", testContainerName, customArgs))

			res, err := uClient.Get(url)
			Expect(err).Should(BeNil())

			// Parse the response
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			var topResponse ContainerTopResponse
			err = json.Unmarshal(body, &topResponse)
			Expect(err).Should(BeNil())

			// Validate the response structure has exactly the columns we requested
			Expect(topResponse.Titles).Should(HaveLen(2))
			Expect(topResponse.Titles).Should(ConsistOf("PID", "COMMAND"))
			Expect(topResponse.Processes).ShouldNot(BeEmpty())

			// Validate that each process entry has exactly 2 fields
			for _, process := range topResponse.Processes {
				Expect(process).Should(HaveLen(2))
			}
		})

		It("should return 404 for non-existent container", func() {
			// Call the top API with a non-existent container ID
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers/non-existent-container/top"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))

			// Parse the error response
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())
			defer res.Body.Close()

			// Verify error message contains "no such container"
			Expect(string(body)).Should(ContainSubstring("no such container"))
		})

		It("should return 400 for invalid ps args", func() {
			command.StdoutStr(opt, "run", "-d", "--name", testContainerName, defaultImage, "sleep", "infinity")

			// Call the top API with invalid ps args
			invalidArgs := "--invalid-arg"
			res, err := uClient.Get(client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/top?ps_args=%s", testContainerName, invalidArgs)))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})

		It("should return 405 for empty container ID", func() {
			// Call the top API with an empty container ID
			res, err := uClient.Get(client.ConvertToFinchUrl(version, "/containers//top"))
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusMethodNotAllowed))
		})
	})
}
