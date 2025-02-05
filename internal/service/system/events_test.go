// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/v2/core/events"
	"github.com/containerd/typeurl/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	eventtype "github.com/runfinch/finch-daemon/api/events"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
)

var _ = Describe("Events API ", func() {
	var (
		mockCtrl    *gomock.Controller
		ctx         context.Context
		client      *mocks_backend.MockContainerdClient
		s           *service
		mockEventCh chan *events.Envelope
		mockErrCh   chan error
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		client = mocks_backend.NewMockContainerdClient(mockCtrl)
		s = &service{
			client: client,
		}
	})
	Context("service", func() {
		It("should stream events to a channel", func() {
			mockEventCh = make(chan *events.Envelope)
			mockErrCh = make(chan error)

			client.EXPECT().SubscribeToEvents(ctx, containerdFiltersFromAPIFilters(map[string][]string{})).Return(mockEventCh, mockErrCh)

			ch, _ := s.SubscribeEvents(ctx, map[string][]string{})

			event := &eventtype.Event{
				Type:   "test",
				Action: "test",
				Actor: eventtype.EventActor{
					Id:         "123",
					Attributes: map[string]string{"test": "test"},
				},
			}
			eventAny, err := typeurl.MarshalAny(event)
			Expect(err).Should(BeNil())

			now := time.Now()

			mockEventCh <- &events.Envelope{
				Timestamp: now,
				Topic:     "/dockercompat/image/tag",
				Namespace: "finch",
				Event:     eventAny,
			}

			gotEvent := <-ch
			Expect(gotEvent).ShouldNot(BeNil())
			Expect(gotEvent.Type).Should(Equal(event.Type))
			Expect(gotEvent.Action).Should(Equal(event.Action))
			Expect(gotEvent.Actor).Should(Equal(event.Actor))
			Expect(gotEvent.Time).Should(Equal(now.Unix()))
			Expect(gotEvent.TimeNano).Should(Equal(now.UnixNano()))
			Expect(gotEvent.Scope).Should(Equal("local"))

			close(mockEventCh)
			close(mockErrCh)
		})
		It("should forward errors to a channel", func() {
			mockEventCh = make(chan *events.Envelope)
			mockErrCh = make(chan error)

			client.EXPECT().SubscribeToEvents(ctx, containerdFiltersFromAPIFilters(map[string][]string{})).Return(mockEventCh, mockErrCh)

			_, ch := s.SubscribeEvents(ctx, map[string][]string{})

			err := fmt.Errorf("mock error")
			mockErrCh <- err

			gotErr := <-ch
			Expect(gotErr).Should(Equal(err))

			close(mockEventCh)
			close(mockErrCh)
		})
	})
	Context("containerdFiltersFromAPIFilters", func() {
		It("should return the docker compatible filter if no other filters are provided", func() {
			filters := containerdFiltersFromAPIFilters(map[string][]string{})
			Expect(len(filters)).Should(Equal(1))
			Expect(filters[0]).Should(Equal(fmt.Sprintf(`topic~="/%s/*"`, eventtype.CompatibleTopicPrefix)))
		})
		It("should include a more specific filter if a type filter is provided", func() {
			filters := containerdFiltersFromAPIFilters(map[string][]string{"type": {"test"}})
			Expect(len(filters)).Should(Equal(1))
			Expect(filters[0]).Should(Equal(fmt.Sprintf(`topic~="/%s/%s/*"`, eventtype.CompatibleTopicPrefix, "test")))
		})
		It("should be able to support multiple type filters", func() {
			filters := containerdFiltersFromAPIFilters(map[string][]string{"type": {"test1", "test2"}})
			Expect(len(filters)).Should(Equal(2))
			Expect(filters[0]).Should(Equal(fmt.Sprintf(`topic~="/%s/%s/*"`, eventtype.CompatibleTopicPrefix, "test1")))
			Expect(filters[1]).Should(Equal(fmt.Sprintf(`topic~="/%s/%s/*"`, eventtype.CompatibleTopicPrefix, "test2")))
		})
	})
})
