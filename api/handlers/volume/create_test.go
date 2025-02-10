// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_volume"
)

var _ = Describe("Create Volume API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_volume.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		req      *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_volume.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		var err error
		Expect(err).Should(BeNil())
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return volume list object and 200 status code upon success", func() {
			// setup mocks
			response := &native.Volume{Name: "NewVolume"}
			service.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			reqJson := []byte(`{"Name": "NewVolume"}`)
			req, _ = http.NewRequest(http.MethodPost, "/volumes/create", bytes.NewBuffer(reqJson))
			// call the API to check if it returns the error generated from the list method
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Body).Should(MatchJSON(`{"Name": "NewVolume", "Mountpoint": ""}`))
		})
		It("should return 500 status code if volume name is missing", func() {
			// setup mocks
			service.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error from create api"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			reqJson := []byte(`{}`)
			req, _ = http.NewRequest(http.MethodPost, "/volumes/create", bytes.NewBuffer(reqJson))
			// call the API to check if it returns the error generated from the list method
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from create api"}`))
		})
		It("should return 500 status code if service returns an error message", func() {
			// setup mocks
			service.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error from create api"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())
			reqJson := []byte(`{"Name": "NewVolume"}`)
			req, _ = http.NewRequest(http.MethodPost, "/volumes/create", bytes.NewBuffer(reqJson))
			// call the API to check if it returns the error generated from the list method
			h.create(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "error from create api"}`))
		})
	})
})
