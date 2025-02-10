// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

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
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_volume"
)

// TestVolumesHandler function is the entry point of volumes handler package's unit test using ginkgo.
func TestVolumesHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Volumes APIs Handler")
}

// Unit tests related to check RegisterHandlers() has configured the endpoint properly for volume related APIs.
var _ = Describe("Volumes API ", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_volume.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		conf     config.Config
		router   *mux.Router
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_volume.NewMockService(mockCtrl)
		router = mux.NewRouter()
		RegisterHandlers(types.VersionedRouter{Router: router}, service, &conf, logger)
		rr = httptest.NewRecorder()
	})
	Context("handler", func() {
		It("should call volumes list method", func() {
			// setup mocks
			service.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("error from list api"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			req, _ = http.NewRequest(http.MethodGet, "/volumes", nil)
			// call the API to check if it returns the error generated from the list method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from list api"}`))
		})
		It("should call volumes create method", func() {
			// setup mocks
			service.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error from create api"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			reqJson := []byte(`{"Name": "NewVolume"}`)
			req, _ = http.NewRequest(http.MethodPost, "/volumes/create", bytes.NewBuffer(reqJson))
			// call the API to check if it returns the error
			// generated from the create method
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from create api"}`))
		})
	})
})
