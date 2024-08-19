// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/option"

	"github.com/runfinch/finch-daemon/e2e/client"
)

// ContainerLogs tests the `POST containers/logs` API.
func ContainerLogs(opt *option.Option) {
	Describe("get logs from a container", func() {
		var (
			uClient *http.Client
			version string
		)
		BeforeEach(func() {
			// create a custom client to use http over unix sockets
			uClient = client.NewClient(GetDockerHostUrl())
			// get the docker api version that will be tested
			version = GetDockerApiVersion()
			// run container in detached mode, outputting 1, 2, 3 in different lines
			command.Run(opt, "run", "-d", "--name", testContainerName, defaultImage,
				"/bin/sh", "-c", `for VAR in 1 2 3; do echo $VAR; done; sleep infinity`)
		})
		AfterEach(func() {
			command.RemoveAll(opt)
		})
		It("should return a 404 status if the container is not found", func() {
			// create url and options
			notFoundName := "doesnt-exist"
			relativeUrl := fmt.Sprintf("/containers/%s/logs", notFoundName)
			opts := "?stdout=1" +
				"&stderr=1"

			// call the endpoint
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl+opts))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())

			// make assertions
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
			Expect(string(body)).Should(ContainSubstring(`no such container: ` + notFoundName))
		})
		It("should return a bad request and do nothing when stdout and stderr are false", func() {
			// create url and options
			relativeUrl := fmt.Sprintf("/containers/%s/logs", testContainerName)
			opts := "?stdout=0" +
				"&stderr=0"

			// call the endpoint
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl+opts))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())

			// make assertions
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
			Expect(string(body)).Should(MatchJSON(`{"message":"you must choose at least one stream"}`))
		})
		It("should succeed logging a running container without following", func() {
			// create url and options
			relativeUrl := fmt.Sprintf("/containers/%s/logs", testContainerName)
			opts := "?stdout=1" +
				"&stderr=1" +
				"&follow=0" +
				"&tail=0"

			// wait for container to run & echo the output, then call endpoint
			time.Sleep(1 * time.Second)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl+opts))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())

			// make assertions
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			// response body is made up of the stream format explained here:
			// https://docs.docker.com/engine/api/v1.43/#tag/Container/operation/ContainerAttach
			// basically, header := [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
			Expect(body[8]).Should(Equal(byte('1')))
			Expect(body[18]).Should(Equal(byte('2')))
			Expect(body[28]).Should(Equal(byte('3')))
		})
		It("should succeed in logging the last 1 line given a tail", func() {
			// create url and options
			relativeUrl := fmt.Sprintf("/containers/%s/logs", testContainerName)
			opts := "?stdout=1" +
				"&stderr=1" +
				"&follow=0" +
				"&tail=1"

			// wait for container to run & echo the output, then call endpoint
			time.Sleep(1 * time.Second)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl+opts))
			Expect(err).Should(BeNil())
			body, err := io.ReadAll(res.Body)
			Expect(err).Should(BeNil())

			// make assertions
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			// response body is made up of the stream format explained here:
			// https://docs.docker.com/engine/api/v1.43/#tag/Container/operation/ContainerAttach
			// basically, header := [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
			Expect(body[8]).Should(Equal(byte('3')))
		})
		It("should succeed in logging with follow enabled", func() {
			altCtrName := "ctr-test2"
			command.Run(opt, "run", "-d", "--name", altCtrName, defaultImage,
				"/bin/sh", "-c", `for VAR in 1 2 3; do echo $VAR; done; sleep 2; for VAR in a b c; do echo $VAR; done`)
			// create url and options
			relativeUrl := fmt.Sprintf("/containers/%s/logs", altCtrName)
			opts := "?stdout=1" +
				"&stderr=1" +
				"&tail=1" +
				"&follow=1"

			// wait for container to reach steady state, then call endpoint
			time.Sleep(1 * time.Second)
			res, err := uClient.Get(client.ConvertToFinchUrl(version, relativeUrl+opts))
			Expect(err).Should(BeNil())

			time.Sleep(4 * time.Second)
			body, err := io.ReadAll(res.Body)
			_ = res.Body.Close()

			// make assertions
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			Expect(body[8]).Should(Equal(byte('3')))
			Expect(body[18]).Should(Equal(byte('a')))
			Expect(body[28]).Should(Equal(byte('b')))
			Expect(body[38]).Should(Equal(byte('c')))
		})
	})
}
