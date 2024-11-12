// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_exec"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Exec Resize API", func() {
	var (
		mockCtrl *gomock.Controller
		service  *mocks_exec.MockService
		conf     config.Config
		logger   *mocks_logger.Logger
		h        *handler
		rr       *httptest.ResponseRecorder
		req      *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		service = mocks_exec.NewMockService(mockCtrl)
		logger = mocks_logger.NewLogger(mockCtrl)
		h = newHandler(service, &conf, logger)
		rr = httptest.NewRecorder()
		var err error
		req, err = http.NewRequest(http.MethodPost, "/exec/123/exec-123/resize?h=123&w=123", nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"id": "123/exec-123"})
	})
	Context("handler", func() {
		It("should return 200 on successful resize", func() {
			service.EXPECT().Resize(gomock.Any(), &types.ExecResizeOptions{
				ConID:  "123",
				ExecID: "exec-123",
				Height: 123,
				Width:  123,
			}).Return(nil)

			h.resize(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 404 if the exec instance is not found", func() {
			service.EXPECT().Resize(gomock.Any(), &types.ExecResizeOptions{
				ConID:  "123",
				ExecID: "exec-123",
				Height: 123,
				Width:  123,
			}).Return(errdefs.NewNotFound(errors.New("not found")))

			h.resize(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(rr.Body).Should(MatchJSON(`{"message": "not found"}`))
		})
		It("should return 500 on any other error", func() {
			service.EXPECT().Resize(gomock.Any(), &types.ExecResizeOptions{
				ConID:  "123",
				ExecID: "exec-123",
				Height: 123,
				Width:  123,
			}).Return(errors.New("inspect error"))

			h.resize(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "inspect error"}`))
		})
		It("should return 400 if h is not specified", func() {
			badReq, err := http.NewRequest(http.MethodPost, "/exec/123/exec-123/resize?w=123", nil)
			Expect(err).Should(BeNil())
			badReq = mux.SetURLVars(badReq, map[string]string{"id": "123/exec-123"})

			h.resize(rr, badReq)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body).Should(MatchJSON(`{"message": "query parameter h required"}`))
		})
		It("should return 400 if w is not specified", func() {
			badReq, err := http.NewRequest(http.MethodPost, "/exec/123/exec-123/resize?h=123", nil)
			Expect(err).Should(BeNil())
			badReq = mux.SetURLVars(badReq, map[string]string{"id": "123/exec-123"})

			h.resize(rr, badReq)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body).Should(MatchJSON(`{"message": "query parameter w required"}`))
		})
		It("should return 400 if a query param is not an int", func() {
			badReq, err := http.NewRequest(http.MethodPost, "/exec/123/exec-123/resize?h=foo&w=123", nil)
			Expect(err).Should(BeNil())
			badReq = mux.SetURLVars(badReq, map[string]string{"id": "123/exec-123"})

			h.resize(rr, badReq)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body).Should(MatchJSON(`{"message": "h must be an integer"}`))
		})
	})
	Context("getQueryParamInt", func() {
		It("should correctly get h", func() {
			height, err := getQueryParamInt(req, "h")
			Expect(err).Should(BeNil())
			Expect(height).Should(Equal(123))
		})
		It("should return error if the query param does not exist", func() {
			_, err := getQueryParamInt(req, "none")
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("query parameter none required"))
		})
		It("should return error if the query param is not an integer", func() {
			badReq, err := http.NewRequest(http.MethodPost, "/exec/123/exec-123/resize?foo=bar", nil)
			Expect(err).Should(BeNil())

			_, err = getQueryParamInt(badReq, "foo")
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("foo must be an integer"))
		})
	})
})
