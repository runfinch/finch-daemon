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
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Image Remove API", func() {
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
		req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("/images/%s", name), nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"name": name})
	})
	Context("handler", func() {
		It("should return  200 status code upon success", func() {
			service.EXPECT().Remove(gomock.Any(), name, false).Return([]string{"12345"}, []string{"12345"}, nil)

			// handler should return response object with 200 status code
			h.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 404 status code if image was not found", func() {
			service.EXPECT().Remove(gomock.Any(), name, false).Return(nil, nil, errdefs.NewNotFound(fmt.Errorf("no such image")))

			// handler should return error message with 404 status code
			h.remove(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "no such image"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 409 status code if image is being used", func() {
			service.EXPECT().Remove(gomock.Any(), name, false).Return(nil, nil,
				errdefs.NewConflict(fmt.Errorf("in use")))

			// handler should return error message with 409 status code
			h.remove(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "in use"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusConflict))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Remove(gomock.Any(), name, false).Return(nil, nil, fmt.Errorf("error"))

			// handler should return error message
			h.remove(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
		It("should pass force flag as true to service", func() {
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/images/%s?force=true", name), nil)
			req = mux.SetURLVars(req, map[string]string{"name": name})

			Expect(err).Should(BeNil())
			service.EXPECT().Remove(gomock.Any(), name, true).Return([]string{"12345"}, []string{"12345"}, nil)

			// handler should return response object with 200 status code
			h.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should pass force flag as false to service for invalid value", func() {
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/images/%s?force=asdf", name), nil)
			req = mux.SetURLVars(req, map[string]string{"name": name})

			Expect(err).Should(BeNil())
			service.EXPECT().Remove(gomock.Any(), name, false).Return([]string{"12345"}, []string{"12345"}, nil)

			// handler should return response object with 200 status code
			h.remove(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
	})
})
