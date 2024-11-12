// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/v2/core/events"
	"github.com/containerd/typeurl/v2"
	"github.com/sirupsen/logrus"

	eventtype "github.com/runfinch/finch-daemon/api/events"
)

func (s *service) SubscribeEvents(ctx context.Context, filters map[string][]string) (<-chan *eventtype.Event, <-chan error) {
	sendCh := make(chan *eventtype.Event)
	sendErrCh := make(chan error)

	go func() {
		defer close(sendCh)
		defer close(sendErrCh)

		eventCh, errCh := s.client.SubscribeToEvents(ctx, containerdFiltersFromAPIFilters(filters)...)

		for {
			var e *events.Envelope
			select {
			case e = <-eventCh:
			case err := <-errCh:
				sendErrCh <- err
				return
			case <-ctx.Done():
				if cerr := ctx.Err(); cerr != nil {
					sendErrCh <- cerr
				}
				return
			}
			if e != nil {
				var event *eventtype.Event

				if e.Event != nil {
					v, err := typeurl.UnmarshalAny(e.Event)
					if err != nil {
						logrus.Errorf("error unmarshaling event: %q\n", err)
						continue
					}
					event = v.(*eventtype.Event)
				} else {
					continue
				}

				event.Scope = "local"
				event.Time = e.Timestamp.Unix()
				event.TimeNano = e.Timestamp.UnixNano()

				sendCh <- event
			}
		}
	}()

	return sendCh, sendErrCh
}

func containerdFiltersFromAPIFilters(filters map[string][]string) []string {
	containerdFilters := []string{
		fmt.Sprintf(`topic~="/%s/*"`, eventtype.CompatibleTopicPrefix),
	}

	for filterType, filterList := range filters {
		switch filterType {
		case "type":
			// pop off general topic filter
			containerdFilters = containerdFilters[1:]
			for _, eventType := range filterList {
				containerdFilters = append(containerdFilters, fmt.Sprintf(`topic~="/%s/%s/*"`,
					eventtype.CompatibleTopicPrefix, eventType))
			}
		default:
			// NOP
		}
	}

	return containerdFilters
}
