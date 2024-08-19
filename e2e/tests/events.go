// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/common-tests/option"
	"github.com/runfinch/finch-daemon/e2e/client"
	eventtype "github.com/runfinch/finch-daemon/pkg/api/events"
)

// SystemEvents tests streaming container events
func SystemEvents(opt *option.Option) {
	Describe("stream container events", func() {
		var (
			uClient *http.Client
			version string
			tagUrl  string
			imageId string
		)
		BeforeEach(func() {
			imageId = pullImage(opt, defaultImage)
			uClient = client.NewClient(GetDockerHostUrl())
			version = GetDockerApiVersion()
			tagUrl = client.ConvertToFinchUrl(version, fmt.Sprintf("/images/%s/tag?repo=test&tag=test", defaultImage))
		})
		AfterEach(func() {
			removeImage(opt, defaultImage)
		})
		It("should successfully stream image tag events", func() {
			relativeUrl := "/events"
			finchUrl := client.ConvertToFinchUrl(version, relativeUrl)

			res, err := uClient.Get(finchUrl)
			Expect(err).Should(BeNil())

			_, err = uClient.Post(tagUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			defer removeImage(opt, "test:test")

			scanner := bufio.NewScanner(res.Body)
			scanner.Scan()
			read := scanner.Text()
			Expect(read).ShouldNot(BeEmpty())

			event := &eventtype.Event{}
			err = json.Unmarshal([]byte(read), event)
			Expect(err).Should(BeNil())

			Expect(event.Type).Should(Equal("image"))
			Expect(event.Action).Should(Equal("tag"))
			Expect(event.ID).Should(ContainSubstring(imageId))
			Expect(event.Actor.Attributes["name"]).Should(Equal("test:test"))
		})
		It("should receive image events when filtering for image events", func() {
			// TODO: once we've added more event types, ensure that different event types are not received
			relativeUrl := `/events?filters={"type":["image"]}`
			finchUrl := client.ConvertToFinchUrl(version, relativeUrl)

			res, err := uClient.Get(finchUrl)
			Expect(err).Should(BeNil())

			_, err = uClient.Post(tagUrl, "application/json", nil)
			Expect(err).Should(BeNil())
			defer removeImage(opt, "test:test")

			scanner := bufio.NewScanner(res.Body)
			scanner.Scan()
			read := scanner.Text()
			Expect(read).ShouldNot(BeEmpty())

			event := &eventtype.Event{}
			err = json.Unmarshal([]byte(read), event)
			Expect(err).Should(BeNil())

			Expect(event.Type).Should(Equal("image"))
			Expect(event.Action).Should(Equal("tag"))
			Expect(event.ID).Should(ContainSubstring(imageId))
			Expect(event.Actor.Attributes["name"]).Should(Equal("test:test"))
		})
	})
}
