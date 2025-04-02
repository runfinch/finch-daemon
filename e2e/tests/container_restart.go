// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/e2e/client"
)

// ContainerRestart tests the `POST containers/{id}/restart` API.
func ContainerRestart(opt *option.Option) {
	Describe("restart a container", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage,
				"/bin/sh", "-c", `date; sleep infinity`)
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})

		FIt("should start and restart the container", func() {
			containerShouldBeRunning(opt, testContainerName)

			// use location to ensure all times are UTC since
			// the default location is different on different platforms
			lo, _ := time.LoadLocation("UTC")
			before := time.Now().In(lo).Round(0)

			restartRelativeUrl := fmt.Sprintf("/containers/%s/restart", testContainerName)
			res, err := uClient.Post(client.ConvertToFinchUrl(version, restartRelativeUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))

			logsRelativeUrl := fmt.Sprintf("/containers/%s/logs", testContainerName)
			opts := "?stdout=1" +
				"&stderr=0" +
				"&follow=0" +
				"&tail=0" +
				"&since=0" +
				"&tail=all"
			res, err = uClient.Get(client.ConvertToFinchUrl(version, logsRelativeUrl+opts))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())

			// get the second date from the container logs, which are newline delimited
			// and strip newlines, spaces, and non-printable characters
			dateStr := strings.TrimFunc(strings.Split(string(body), "\n")[1], func(r rune) bool {
				return !unicode.IsGraphic(r) || unicode.IsSpace(r)
			})

			fmt.Printf("\nbefore: %s\nbody: %s\ndateStr: %s\n", before, string(body), dateStr)

			date, err := time.ParseInLocation(time.UnixDate, dateStr, lo)
			Expect(err).Should(BeNil())
			Expect(before.Before(date)).Should(BeTrue())
		})
		It("should fail to restart container that does not exist", func() {
			// restart a container that does not exist
			relativeUrl := client.ConvertToFinchUrl(version, "/containers/container-does-not-exist/restart")
			res, err := uClient.Post(relativeUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			var errResponse response.Error
			err = json.NewDecoder(res.Body).Decode(&errResponse)
			Expect(err).Should(BeNil())
			Expect(errResponse.Message).Should(Not(BeEmpty()))
		})
		It("should restart a stopped container", func() {
			containerShouldBeRunning(opt, testContainerName)

			stopRelativeUrl := fmt.Sprintf("/containers/%s/stop", testContainerName)
			res, err := uClient.Post(client.ConvertToFinchUrl(version, stopRelativeUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldNotBeRunning(opt, testContainerName)

			restartRelativeUrl := fmt.Sprintf("/containers/%s/restart", testContainerName)
			res, err = uClient.Post(client.ConvertToFinchUrl(version, restartRelativeUrl), "application/json", nil)
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			containerShouldBeRunning(opt, testContainerName)
		})
		It("should restart the container with timeout", func() {
			containerShouldBeRunning(opt, testContainerName)

			// stop the container with a timeout of 5 seconds
			now := time.Now()
			restartRelativeUrl := fmt.Sprintf("/containers/%s/restart?t=5", testContainerName)
			res, err := uClient.Post(client.ConvertToFinchUrl(version, restartRelativeUrl), "application/json", nil)
			later := time.Now()
			Expect(err).Should(BeNil())
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
			elapsed := later.Sub(now)
			Expect(elapsed.Seconds()).Should(BeNumerically(">", 4.0))
			Expect(elapsed.Seconds()).Should(BeNumerically("<", 10.0))
			containerShouldBeRunning(opt, testContainerName)
		})
	})
}
