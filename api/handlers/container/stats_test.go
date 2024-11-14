// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/containerd/nerdctl/v2/pkg/config"
	dockertypes "github.com/docker/docker/api/types/container"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Stats API ", func() {
	var (
		mockCtrl  *gomock.Controller
		logger    *mocks_logger.Logger
		service   *mocks_container.MockService
		cid       string
		statsData types.StatsJSON
		h         *handler
		rr        *httptest.ResponseRecorder
		req       *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		cid = "123"
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/containers/%s/stats", cid), nil)
		req = mux.SetURLVars(req, map[string]string{"id": cid})
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		// create a dummy stats object
		statsData = types.StatsJSON{
			ID:   cid,
			Name: "test-container",
		}
		statsData.PidsStats = dockertypes.PidsStats{Current: 10, Limit: 20}
		statsData.CPUStats = types.CPUStats{
			CPUUsage: dockertypes.CPUUsage{
				TotalUsage:        1000,
				UsageInKernelmode: 500,
				UsageInUsermode:   250,
				PercpuUsage:       []uint64{1, 2, 3, 4},
			},
			SystemUsage: 2500,
			OnlineCPUs:  3,
		}
		statsData.MemoryStats = dockertypes.MemoryStats{
			Usage:    250,
			Limit:    1000,
			MaxUsage: 500,
			Failcnt:  50,
		}
	})
	Context("handler", func() {
		It("should return 404 if container was not found", func() {
			service.EXPECT().Stats(gomock.Any(), cid).Return(
				nil, errdefs.NewNotFound(fmt.Errorf("no such container")))

			// handler should return 404 status code with an error msg.
			h.stats(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(rr.Body).Should(MatchJSON(`{"message": "no such container"}`))
		})
		It("should fail with 500 status code for service error messages", func() {
			service.EXPECT().Stats(gomock.Any(), cid).Return(
				nil, fmt.Errorf("internal error"))

			// handler should return 500 status code with an error msg.
			h.stats(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "internal error"}`))
		})
		It("should show stats once and exit when streaming is disabled", func() {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/containers/%s/stats?stream=false", cid), nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": cid})
			statsCh := make(chan *types.StatsJSON, 20)
			service.EXPECT().Stats(gomock.Any(), cid).Return(
				statsCh, nil)

			// populate stats channel with 10 stats objects
			statsCh <- &statsData
			for i := 2; i <= 10; i++ {
				st := types.StatsJSON{}
				st.Read = time.Now()
				statsCh <- &st
			}

			// handler should return 200 status code with a single JSON stats object.
			h.stats(rr, req)
			expectedJSON, err := json.Marshal(statsData)
			Expect(err).Should(BeNil())
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Body).Should(MatchJSON(expectedJSON))
		})
		It("should log an error and exit gracefully when stats cannot be received from the channel", func() {
			statsCh := make(chan *types.StatsJSON, 10)
			service.EXPECT().Stats(gomock.Any(), cid).Return(
				statsCh, nil)
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			// close the channel so there is nothing to read
			close(statsCh)

			// handler should return 200 status code with an empty body.
			h.stats(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Body.String()).Should(BeEmpty())
		})
		It("should stream stats", func() {
			statsCh := make(chan *types.StatsJSON, 20)
			service.EXPECT().Stats(gomock.Any(), cid).Return(
				statsCh, nil)

			// setup a goroutine to populate stats channel with 10 objects
			allStats := []*types.StatsJSON{}
			go func() {
				defer close(statsCh)
				for i := 1; i <= 10; i++ {
					st := types.StatsJSON{
						ID:   cid,
						Name: statsData.Name,
					}
					st.PidsStats = statsData.PidsStats
					st.CPUStats = statsData.CPUStats
					st.Read = time.Now()
					statsCh <- &st
					allStats = append(allStats, &st)
					time.Sleep(time.Millisecond * 100)
				}
			}()

			// handler should return 200 status code with 10 JSON stats objects.
			h.stats(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			for _, st := range allStats {
				stJSON, err := json.Marshal(*st)
				Expect(err).Should(BeNil())
				body, err := rr.Body.ReadBytes('\n')
				Expect(err).Should(BeNil())
				Expect(body).Should(MatchJSON(stJSON))
			}
		})
	})
})
