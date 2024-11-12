// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

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
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Put Archive API", func() {
	var (
		mockCtrl       *gomock.Controller
		service        *mocks_container.MockService
		conf           *config.Config
		logger         *mocks_logger.Logger
		r              *mux.Router
		rr             *httptest.ResponseRecorder
		req            *http.Request
		putArchiveOpts *types.PutArchiveOptions
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		service = mocks_container.NewMockService(mockCtrl)
		conf = &config.Config{}
		logger = mocks_logger.NewLogger(mockCtrl)
		r = mux.NewRouter()
		RegisterHandlers(types.VersionedRouter{Router: r}, service, conf, logger)
		rr = httptest.NewRecorder()
		putArchiveOpts = &types.PutArchiveOptions{
			ContainerId: "123",
			Path:        "/home",
		}
	})
	Context("handler", func() {
		It("should return 200 as success response", func() {
			req, _ = http.NewRequest(http.MethodPut, "/containers/123/archive?path=%2Fhome", nil)
			putArchiveOpts.ContainerId = "123"
			putArchiveOpts.Path = "/home"
			service.EXPECT().ExtractArchiveInContainer(gomock.Any(), putArchiveOpts, gomock.Any()).Return(nil)
			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusOK))
		})
		It("should return 403 as response on forbidden error from service", func() {
			req, _ = http.NewRequest(http.MethodPut, "/containers/123/archive?path=%2Fhome", nil)
			e := errdefs.NewForbidden(errors.New("forbidden"))
			service.EXPECT().ExtractArchiveInContainer(gomock.Any(), putArchiveOpts, gomock.Any()).Return(e)
			logger.EXPECT().Errorf("error handling request %v", e)
			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusForbidden))
		})
		It("should return 404 as response on not found error from service", func() {
			req, _ = http.NewRequest(http.MethodPut, "/containers/123/archive?path=%2Fhome", nil)
			e := errdefs.NewNotFound(errors.New("not found"))
			service.EXPECT().ExtractArchiveInContainer(gomock.Any(), putArchiveOpts, gomock.Any()).Return(e)
			logger.EXPECT().Errorf("error handling request %v", e)
			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusNotFound))
		})
		It("should return 400 as response on invalid format error from service", func() {
			req, _ = http.NewRequest(http.MethodPut, "/containers/123/archive?path=%2Fhome", nil)
			e := errdefs.NewInvalidFormat(errors.New("invalid format"))
			service.EXPECT().ExtractArchiveInContainer(gomock.Any(), putArchiveOpts, gomock.Any()).Return(e)
			logger.EXPECT().Errorf("error handling request %v", e)
			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusBadRequest))
		})
		It("should return 400 as response on empty path", func() {
			req, _ = http.NewRequest(http.MethodPut, "/containers/123/archive?path=", nil)
			logger.EXPECT().Error("error handling request, bad path")
			r.ServeHTTP(rr, req)
			Expect(rr.Code).Should(Equal(http.StatusBadRequest))
		})
	})
})
