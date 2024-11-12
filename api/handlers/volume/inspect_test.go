// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_volume"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Volume Inspect API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_volume.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		req      *http.Request
		volName  string
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_volume.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		var err error
		volName = "test-volume"
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/volumes/%s", volName), nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"name": volName})
	})
	Context("handler", func() {
		It("should successfully return volume details", func() {
			resp := native.Volume{
				Name:       "test-volume",
				Mountpoint: "/path/to/test-volume",
				Labels:     nil,
				Size:       100,
			}
			service.EXPECT().Inspect(volName).Return(&resp, nil)

			// handler should return response object with 200 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"Name": "test-volume", "Mountpoint": "/path/to/test-volume", "Size": 100}`))
		})
		It("should return 404 status code if service returns not found error", func() {
			service.EXPECT().Inspect(volName).Return(nil, errdefs.NewNotFound(fmt.Errorf("not found")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// handler should return not found error msg
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "not found"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Inspect(volName).Return(nil, fmt.Errorf("some error"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// handler should return error message
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "some error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
