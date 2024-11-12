// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_builder"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

// TestBuilderHandler function is the entry point of builder handler package's unit test using ginkgo.
func TestBuilderHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Build APIs Handler")
}

// Unit tests related to check RegisterHandlers() has configured the endpoint properly for build related API.
var _ = Describe("Build API ", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_builder.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		conf     config.Config
		router   *mux.Router
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_builder.NewMockService(mockCtrl)
		router = mux.NewRouter()
		ncBuildSvc := mocks_backend.NewMockNerdctlBuilderSvc(mockCtrl)
		RegisterHandlers(types.VersionedRouter{Router: router}, service, &conf, logger, ncBuildSvc)
		rr = httptest.NewRecorder()
		ncBuildSvc.EXPECT().GetBuildkitHost().Return("", nil).AnyTimes()
	})
	Context("handler", func() {
		It("should call build method", func() {
			// setup mocks
			service.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error from build api"))
			req, _ = http.NewRequest(http.MethodPost, "/build", nil)
			// call the API to check if it returns the error generated from the build method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{ "message": "error from build api"}`))
		})
	})
})
