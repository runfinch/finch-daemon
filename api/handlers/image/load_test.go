// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_image"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

var _ = Describe("Image Load API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_image.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		name     string
		req      *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_image.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		name = "test-image"
		var err error
		req, err = http.NewRequest(http.MethodPost, "/images/load", nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"name": name})

		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})
	Context("handler", func() {
		It("should return 200 status code upon success", func() {
			service.EXPECT().Load(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(nil)

			// handler should return 200 status code
			h.load(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Load(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message
			h.load(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
