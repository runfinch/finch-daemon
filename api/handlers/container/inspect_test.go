// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Inspect API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_container.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		cid      string
		req      *http.Request
		resp     types.Container
		respJSON []byte
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		cid = "123"
		var err error
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/containers/%s/json", cid), nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"id": "123"})
		resp = types.Container{
			ID:    cid,
			Image: "test-image",
			Name:  "/test-container",
		}
		respJSON, err = json.Marshal(resp)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return inspect object and 200 status code upon success", func() {
			service.EXPECT().Inspect(gomock.Any(), cid).Return(&resp, nil)

			// handler should return response object with 200 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(respJSON))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 404 status code if container was not found", func() {
			service.EXPECT().Inspect(gomock.Any(), cid).Return(nil, errdefs.NewNotFound(fmt.Errorf("no such container")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message with 404 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "no such container"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Inspect(gomock.Any(), cid).Return(nil, fmt.Errorf("error"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
