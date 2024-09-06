// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	ncTypes "github.com/containerd/nerdctl/pkg/api/types"
	"github.com/containerd/nerdctl/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

var _ = Describe("Container List API", func() {
	var (
		mockCtrl   *gomock.Controller
		logger     *mocks_logger.Logger
		service    *mocks_container.MockService
		h          *handler
		rr         *httptest.ResponseRecorder
		globalOpts ncTypes.GlobalCommandOptions
		resp       []types.ContainerListItem
		respJSON   []byte
		err        error
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		globalOpts = ncTypes.GlobalCommandOptions(c)
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		resp = []types.ContainerListItem{
			{
				Id:    "id1",
				Names: []string{"/name1"},
			},
			{
				Id:    "id2",
				Names: []string{"/name2"},
			},
		}
		respJSON, err = json.Marshal(resp)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return containers and 200 status code upon success with all query parameters", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/json?all=true&limit=1&size=true&filters={\"status\": [\"paused\"]}", nil)
			Expect(err).Should(BeNil())
			listOpts := ncTypes.ContainerListOptions{
				GOptions: globalOpts,
				All:      true,
				LastN:    1,
				Truncate: true,
				Size:     true,
				Filters:  []string{"status=paused"},
			}
			service.EXPECT().List(gomock.Any(), listOpts).Return(resp, nil)

			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(respJSON))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return containers and 200 status code upon success with no query parameter", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/json", nil)
			Expect(err).Should(BeNil())
			listOpts := ncTypes.ContainerListOptions{
				GOptions: globalOpts,
				All:      false,
				LastN:    0,
				Truncate: true,
				Size:     false,
				Filters:  nil,
			}
			service.EXPECT().List(gomock.Any(), listOpts).Return(resp, nil)

			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(respJSON))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 400 status code when there is error parsing all", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/json?all=invalid", nil)
			Expect(err).Should(BeNil())
			errorMsg := fmt.Sprintf("invalid query parameter \\\"all\\\": %s", fmt.Errorf("strconv.ParseBool: parsing \\\"invalid\\\": invalid syntax"))

			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})
		It("should return 400 status code when there is error parsing limit", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/json?limit=invalid", nil)
			Expect(err).Should(BeNil())
			errorMsg := fmt.Sprintf("invalid query parameter \\\"limit\\\": %s", fmt.Errorf("strconv.ParseInt: parsing \\\"invalid\\\": invalid syntax"))

			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})
		It("should return 400 status code when there is error parsing size", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/json?size=invalid", nil)
			Expect(err).Should(BeNil())
			errorMsg := fmt.Sprintf("invalid query parameter \\\"size\\\": %s", fmt.Errorf("strconv.ParseBool: parsing \\\"invalid\\\": invalid syntax"))

			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})
		It("should return 400 status code when there is error parsing filters", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/json?filters=invalid", nil)
			Expect(err).Should(BeNil())
			errorMsg := fmt.Sprintf("invalid query parameter \\\"filters\\\": %s", fmt.Errorf("invalid character 'i' looking for beginning of value"))

			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})
		It("should return 500 status code when service returns error", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/json", nil)
			Expect(err).Should(BeNil())
			listOpts := ncTypes.ContainerListOptions{
				GOptions: globalOpts,
				All:      false,
				LastN:    0,
				Truncate: true,
				Size:     false,
				Filters:  nil,
			}
			errorMsg := "error from ListContainers"
			service.EXPECT().List(gomock.Any(), listOpts).Return(nil, fmt.Errorf("%s", errorMsg))

			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "` + errorMsg + `"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
