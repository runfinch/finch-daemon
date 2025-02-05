// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
)

var _ = Describe("Network Inspect API ", func() {
	var (
		mockCtrl *gomock.Controller
		service  *mocks_network.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		handler  *handler
		conf     *config.Config
		logger   *mocks_logger.Logger
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		// initialize mocks
		service = mocks_network.NewMockService(mockCtrl)
		conf = &config.Config{}
		logger = mocks_logger.NewLogger(mockCtrl)
		handler = newHandler(service, conf, logger)
		rr = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "/networks", nil)
	})
	Context("handler", func() {
		It("should return a 200 when there is no error", func() {
			service.EXPECT().List(gomock.Any()).Return(nil, nil)

			handler.list(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Body.String()).Should(Equal("null\n"))
		})
		It("should return a 500 when there is any other error", func() {
			service.EXPECT().List(gomock.Any()).Return(nil, nil)

			handler.list(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Body.String()).Should(Equal("null\n"))
		})
	})
})
