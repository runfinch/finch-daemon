// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
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
	"github.com/runfinch/finch-daemon/mocks/mocks_image"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

// TestImageHandler function is the entry point of image handler package's unit test using ginkgo.
func TestImageHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Image APIs Handler")
}

// Unit tests related to check RegisterHandlers() has configured the endpoint properly for image related APIs.
var _ = Describe("Image API ", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_image.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		conf     config.Config
		router   *mux.Router
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_image.NewMockService(mockCtrl)
		router = mux.NewRouter()
		RegisterHandlers(types.VersionedRouter{Router: router}, service, &conf, logger)
		rr = httptest.NewRecorder()
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})
	Context("handler", func() {
		It("should call image inspect method", func() {
			// setup mocks
			service.EXPECT().Inspect(gomock.Any(), "test-image").Return(nil, errors.New("error from inspect api"))
			req, _ = http.NewRequest(http.MethodGet, "/images/test-image/json", nil)
			// call the API to check if it returns the error generated from the inspect method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from inspect api"}`))
		})
		It("should call image pull method", func() {
			// setup mocks
			service.EXPECT().Pull(
				gomock.Any(),
				"test-image",
				"test-tag",
				"test-platform",
				gomock.Any(),
				gomock.Any(),
			).Return(errors.New("error from pull api"))
			req, _ = http.NewRequest(http.MethodPost, "/images/create?fromImage=test-image&tag=test-tag&platform=test-platform", nil)
			// call the API to check if it returns the error generated from the pull method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from pull api"}`))
		})
	})
})
