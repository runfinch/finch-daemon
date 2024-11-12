// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_volume"
)

var _ = Describe("Volume List API", func() {
	var (
		mockCtrl     *gomock.Controller
		logger       *mocks_logger.Logger
		service      *mocks_volume.MockService
		h            *handler
		rr           *httptest.ResponseRecorder
		name         string
		req          *http.Request
		filterString string
		filterReq    *http.Request
		resp         types.VolumesListResponse
		respJSON     []byte
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_volume.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		name = "test-volume"
		var err error
		req, err = http.NewRequest(http.MethodGet, "/volumes", nil)
		Expect(err).Should(BeNil())
		filterString = url.QueryEscape(fmt.Sprintf(`{"name":["%s"]}`, name))
		filterReq, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/volumes?filters=%s", filterString), nil)
		Expect(err).Should(BeNil())
		resp = types.VolumesListResponse{
			Volumes: []native.Volume{
				{
					Name:       name,
					Mountpoint: "/path/to/test-volume",
					Labels:     nil,
					Size:       100,
				},
			},
		}
		respJSON, err = json.Marshal(resp)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return volume list object and 200 status code upon success", func() {
			service.EXPECT().List(gomock.Any(), gomock.Any()).Return(&resp, nil)

			// handler should return response object with 200 status code
			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(respJSON))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return volume list object and 200 status code upon success with filters", func() {
			service.EXPECT().List(gomock.Any(), gomock.Any()).Return(&resp, nil)

			// handler should return response object with 200 status code
			h.list(rr, filterReq)
			Expect(rr.Body).Should(MatchJSON(respJSON))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))

			// handler should return error message
			h.list(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
