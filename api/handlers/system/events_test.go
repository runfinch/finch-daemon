// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/events"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_system"
)

var _ = Describe("Events API", func() {
	var (
		mockCtrl      *gomock.Controller
		s             *mocks_system.MockService
		logger        *mocks_logger.Logger
		h             *handler
		rr            *httptest.ResponseRecorder
		mockEvent     *events.Event
		mockEventJson []byte
		mockEventCh   chan *events.Event
		mockErrCh     chan error
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		s = mocks_system.NewMockService(mockCtrl)
		logger = mocks_logger.NewLogger(mockCtrl)
		c := config.Config{}
		h = newHandler(s, &c, nil, logger)
		rr = httptest.NewRecorder()
		mockEvent = &events.Event{
			Type:   "test",
			Action: "test",
			Actor: events.EventActor{
				Id: "123",
				Attributes: map[string]string{
					"test": "test",
				},
			},
			Scope:    "test",
			Time:     0,
			TimeNano: 0,
		}
		mockEventJson, _ = json.Marshal(mockEvent)
	})
	It("should return 200 and stream events on success", func() {
		// in order to add events to the channels while the events handler is running, we need to either run the handler
		// or the channel publisher in a separate goroutine as the events handler will block until it returns. because all
		// of the assertions are in the same thread as the channel publisher, it becomes easier to run the handler in a goroutine.
		// however, this can cause a problem where the handler doesn't finish executing before the test function does. this
		// WaitGroup allows us to signal the test function that the handler has finished executing, and we can block in the
		// main thread until the handler returns.
		var waitGroup sync.WaitGroup

		mockEventCh = make(chan *events.Event)
		mockErrCh = make(chan error)

		req, _ := http.NewRequest(http.MethodGet, "/events", nil)

		s.EXPECT().SubscribeEvents(req.Context(), map[string][]string{}).Return(mockEventCh, mockErrCh)
		logger.EXPECT().Debugf("received error, exiting: %s", gomock.Any())

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			h.events(rr, req)
		}()

		mockEventCh <- mockEvent
		time.Sleep(250 * time.Millisecond)

		// repeat to test that streaming is working
		mockEventCh <- mockEvent
		time.Sleep(250 * time.Millisecond)

		// I didn't put these in AfterEach() because they will cause a logger.Debugf in some cases but not all. this means
		// that if we EXPECT() the Debugf call in the spec itself, it will fail, and we can't EXPECT() it in AfterEach().
		close(mockEventCh)
		close(mockErrCh)

		waitGroup.Wait()

		Expect(rr).Should(HaveHTTPStatus(http.StatusOK))

		line, err := rr.Body.ReadBytes('\n')
		Expect(err).Should(BeNil())
		Expect(line).Should(MatchJSON(mockEventJson))

		line, err = rr.Body.ReadBytes('\n')
		Expect(err).Should(BeNil())
		Expect(line).Should(MatchJSON(mockEventJson))
	})
	It("should return 400 if filters are not in the correct format", func() {
		req, _ := http.NewRequest(http.MethodGet, "/events?filters=bad", nil)
		req = mux.SetURLVars(req, map[string]string{"filters": "bad"})

		h.events(rr, req)
		Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		Expect(rr.Body).Should(MatchJSON(`{"message": "invalid filter: invalid character 'b' looking for beginning of value"}`))
	})
	It("should forward filters to the service", func() {
		var waitGroup sync.WaitGroup

		mockEventCh = make(chan *events.Event)
		mockErrCh = make(chan error)

		req, _ := http.NewRequest(http.MethodGet, `/events?filters={"test": ["test"]}`, nil)
		req = mux.SetURLVars(req, map[string]string{"filters": `{"test": ["test"]}`})

		s.EXPECT().SubscribeEvents(req.Context(), map[string][]string{"test": {"test"}}).Return(mockEventCh, mockErrCh)
		logger.EXPECT().Debugf("received error, exiting: %s", gomock.Any())

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			h.events(rr, req)
		}()

		close(mockEventCh)
		close(mockErrCh)

		waitGroup.Wait()

		Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
	})
	It("should stop streaming if an error is received", func() {
		var waitGroup sync.WaitGroup

		mockEventCh = make(chan *events.Event)
		mockErrCh = make(chan error)

		req, _ := http.NewRequest(http.MethodGet, "/events", nil)

		s.EXPECT().SubscribeEvents(req.Context(), map[string][]string{}).Return(mockEventCh, mockErrCh)
		logger.EXPECT().Debugf("received error, exiting: %s", gomock.Any())

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			h.events(rr, req)
		}()

		mockErrCh <- fmt.Errorf("mock error")

		waitGroup.Wait()

		Expect(rr).Should(HaveHTTPStatus(http.StatusOK))

		close(mockEventCh)
		close(mockErrCh)
	})
})
