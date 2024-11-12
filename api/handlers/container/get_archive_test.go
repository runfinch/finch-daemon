// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
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

var _ = Describe("Container Get Archive API", func() {
	var (
		mockCtrl *gomock.Controller
		service  *mocks_container.MockService
		conf     *config.Config
		logger   *mocks_logger.Logger
		mockPath string
		r        *mux.Router
		rr       *httptest.ResponseRecorder
		req      *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		service = mocks_container.NewMockService(mockCtrl)
		conf = &config.Config{}
		logger = mocks_logger.NewLogger(mockCtrl)
		mockPath = "./mockPath"
		r = mux.NewRouter()
		RegisterHandlers(types.VersionedRouter{Router: r}, service, conf, logger)
		rr = httptest.NewRecorder()
	})
	Context("handler", func() {
		It("should return 200 as success response", func() {
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/archive?path=%2Fhome", nil)
			service.EXPECT().GetPathToFilesInContainer(gomock.Any(), "123", "/home").Return(mockPath, nil, nil)
			service.EXPECT().WriteFilesAsTarArchive(mockPath, gomock.Any(), false).Return(nil)

			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusOK))
		})
		It("should return 400 if the path is not specified", func() {
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/archive", nil)

			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusBadRequest))
			Expect(rr.Body).Should(MatchJSON(`{"message": "must specify a file or directory path"}`))
		})
		It("should return 404 if CopyFilesFromContainer returns a NotFound error", func() {
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/archive?path=%2Fhome", nil)
			service.EXPECT().GetPathToFilesInContainer(gomock.Any(), "123", "/home").Return("", nil, errdefs.NewNotFound(fmt.Errorf("not found")))
			logger.EXPECT().Debugf("Responding with error. Error code: %d, Message: %s", http.StatusNotFound, "not found")

			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusNotFound))
			Expect(rr.Body).Should(MatchJSON(`{"message": "not found"}`))
		})
		It("should return 500 if CopyFilesFromContainer returns any other error", func() {
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/archive?path=%2Fhome", nil)
			service.EXPECT().GetPathToFilesInContainer(gomock.Any(), "123", "/home").Return("", nil, fmt.Errorf("internal error"))

			logger.EXPECT().Debugf("Responding with error. Error code: %d, Message: %s", http.StatusInternalServerError, "internal error")

			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "internal error"}`))
		})
		It("should run a cleanup function returned from GetPathToFilesInContainer", func() {
			cleanupHasRun := false
			cleanup := func() {
				cleanupHasRun = true
			}
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/archive?path=%2Fhome", nil)
			service.EXPECT().GetPathToFilesInContainer(gomock.Any(), "123", "/home").Return(mockPath, cleanup, nil)
			service.EXPECT().WriteFilesAsTarArchive(mockPath, gomock.Any(), false).Return(nil)

			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			Expect(cleanupHasRun).Should(BeTrue())
		})
	})
})
