// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
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
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

// TestContainerHandler function is the entry point of container handler package's unit test using ginkgo.
func TestContainerHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Container APIs Handler")
}

// Unit tests related to check RegisterHandlers() has configured the endpoint properly for containers related API.
var _ = Describe("Container API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_container.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		conf     config.Config
		router   *mux.Router
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		router = mux.NewRouter()
		RegisterHandlers(types.VersionedRouter{Router: router}, service, &conf, logger)
		rr = httptest.NewRecorder()
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})
	Context("handlers", func() {
		It("should call container delete method", func() {
			// setup mocks
			service.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error from delete api"))
			req, _ = http.NewRequest(http.MethodDelete, "/containers/123", nil)
			// call the API to check if it returns the error generated from the remove method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from delete api"}`))
		})
		It("should call container start method", func() {
			// setup mocks
			service.EXPECT().Start(gomock.Any(), gomock.Any()).Return(fmt.Errorf("error from start api"))
			req, _ = http.NewRequest(http.MethodPost, "/containers/123/start", nil)
			// call the API to check if it returns the error generated from start method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from start api"}`))
		})
		It("should call container stop method", func() {
			// setup mocks
			service.EXPECT().Stop(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error from stop api"))
			req, _ = http.NewRequest(http.MethodPost, "/containers/123/stop", nil)
			// call the API to check if it returns the error generated from stop method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from stop api"}`))
		})
		It("should call container restart method", func() {
			// setup mocks
			service.EXPECT().Restart(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error from restart api"))
			req, _ = http.NewRequest(http.MethodPost, "/containers/123/restart", nil)
			// call the API to check if it returns the error generated from restart method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from restart api"}`))
		})
		It("should call container create method", func() {
			// setup mocks
			body := []byte(`{"Image": "test-image"}`)
			service.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", fmt.Errorf("error from create api"))
			req, _ = http.NewRequest(http.MethodPost, "/containers/create", bytes.NewReader(body))
			// call the API to check if it returns the error generated from create method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from create api"}`))
		})
		It("should call container attach method", func() {
			req, _ = http.NewRequest(http.MethodPost, "/containers/123/attach", nil)

			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr).Should(HaveHTTPBody(`{"message":"the response writer is not a http.Hijacker"}` + "\n"))
		})
		It("should call container inspect method", func() {
			// setup mocks
			service.EXPECT().Inspect(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error from inspect api"))
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/json", nil)
			// call the API to check if it returns the error generated from inspect method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from inspect api"}`))
		})
		It("should call container list method", func() {
			// setup mocks
			service.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error from list api"))
			req, _ = http.NewRequest(http.MethodGet, "/containers/json", nil)
			// call the API to check if it returns the error generated from list method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from list api"}`))
		})
		It("should call container rename method", func() {
			// setup mocks
			service.EXPECT().Rename(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error from rename api"))
			req, _ = http.NewRequest(http.MethodPost, "/containers/123/rename", nil)
			// call the API to check if it returns the error generated from list method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from rename api"}`))
		})
		It("should call container logs method", func() {
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/logs", nil)

			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body).Should(MatchJSON(`{"message":"you must choose at least one stream"}`))
		})
		It("should call container stats method", func() {
			// setup mocks
			service.EXPECT().Stats(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error from stats api"))
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/stats", nil)
			// call the API to check if it returns the error generated from stats method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from stats api"}`))
		})
	})
})
