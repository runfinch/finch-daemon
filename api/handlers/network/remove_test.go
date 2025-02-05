// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Network Remove API ", func() {
	var (
		mockCtrl *gomock.Controller
		service  *mocks_network.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		handler  *handler
		conf     *config.Config
		logger   *mocks_logger.Logger
		nid      string
	)
	BeforeEach(func() {
		nid = "123"
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		// initialize mocks
		service = mocks_network.NewMockService(mockCtrl)
		conf = &config.Config{}
		logger = mocks_logger.NewLogger(mockCtrl)
		handler = newHandler(service, conf, logger)
		rr = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/networks/%s", nid), nil)
		req = mux.SetURLVars(req, map[string]string{"id": nid})
	})
	Context("handler", func() {
		It("should return a 204 when there is no error", func() {
			service.EXPECT().Remove(gomock.Any(), nid).Return(nil)
			handler.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNoContent))
		})
		It("should return a 404 when network is not found", func() {
			service.EXPECT().Remove(gomock.Any(), nid).Return(errdefs.NewNotFound(fmt.Errorf("not found")))
			handler.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(rr.Body.String()).Should(MatchJSON(`{"message": "not found"}`))
		})
		It("should return a 403 when remove returns forbidden error", func() {
			service.EXPECT().Remove(gomock.Any(), nid).Return(errdefs.NewForbidden(fmt.Errorf("forbidden error")))
			handler.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusForbidden))
			Expect(rr.Body.String()).Should(MatchJSON(`{"message": "forbidden error"}`))
		})
		It("should return a 500 for server errors", func() {
			service.EXPECT().Remove(gomock.Any(), nid).Return(fmt.Errorf("server error"))
			handler.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body.String()).Should(MatchJSON(`{"message": "server error"}`))
		})
	})
})
