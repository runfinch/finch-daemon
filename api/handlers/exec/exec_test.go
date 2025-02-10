// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_exec"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

// TestExecHandler is the entry point of the exec handler package's unit tests.
func TestExecHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Exec APIs Handler")
}

// Unit tests related to checking whether RegisterHandlers() has correctly configured the endpoints.
var _ = Describe("Exec API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_exec.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		conf     config.Config
		router   *mux.Router
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_exec.NewMockService(mockCtrl)
		router = mux.NewRouter()
		RegisterHandlers(types.VersionedRouter{Router: router}, service, &conf, logger)
		rr = httptest.NewRecorder()
	})
	Context("handlers", func() {
		It("should call exec start method", func() {
			req, _ = http.NewRequest(http.MethodPost, "/exec/123/exec-123/start", bytes.NewReader([]byte(`{"detach": true}`)))
			service.EXPECT().Inspect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			service.EXPECT().Start(gomock.Any(), gomock.Any()).Return(errors.New("start error"))

			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "start error"}`))
		})
		It("should call exec resize method", func() {
			req, _ = http.NewRequest(http.MethodPost, "/exec/123/exec-123/resize?h=123&w=123", nil)
			service.EXPECT().Resize(gomock.Any(), gomock.Any()).Return(errors.New("resize error"))

			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "resize error"}`))
		})
		It("should call exec inspect method", func() {
			service.EXPECT().Inspect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("inspect error"))
			req, _ = http.NewRequest(http.MethodGet, "/exec/123/exec-123/json", nil)

			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "inspect error"}`))
		})
	})
})
