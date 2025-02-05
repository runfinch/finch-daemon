// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_volume"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Volume Remove API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_volume.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		req      *http.Request
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
		req, err = http.NewRequest(http.MethodDelete, "/volumes/test-volume", nil)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return 204 status code upon success", func() {
			service.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			// handler should return response object with 200 status code
			h.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNoContent))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// handler should return error message
			h.remove(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
		It("should return 404 status code if service returns an not found error", func() {
			service.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errdefs.NewNotFound(fmt.Errorf("not found")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// handler should return error message
			h.remove(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "not found"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 409 status code if service returns volume is in use error", func() {
			service.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errdefs.NewConflict(fmt.Errorf("in use")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// handler should return error message
			h.remove(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "in use"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusConflict))
		})
	})
})
