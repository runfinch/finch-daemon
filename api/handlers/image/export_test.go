// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/runfinch/finch-daemon/mocks/mocks_image"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Image Export API", func() {
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
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/images/%s/get", name), nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"name": name})
	})
	Context("handler", func() {
		It("should return 200 status code upon success", func() {
			service.EXPECT().Export(
				gomock.Any(),
				name,
				gomock.Any(),
				gomock.Any(),
			).Return(nil)

			h.export(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Header().Get("Content-Type")).Should(Equal("application/x-tar"))
		})
		It("should return 404 status code if image not found", func() {
			service.EXPECT().Export(gomock.Any(), name, gomock.Any(), gomock.Any()).Return(errdefs.NewNotFound(fmt.Errorf("no such image")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any(), gomock.Any())

			h.export(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 500 status code if service returns an error", func() {
			service.EXPECT().Export(gomock.Any(), name, gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any(), gomock.Any())

			h.export(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
		It("should return 400 status code for invalid platform JSON", func() {
			req.URL.RawQuery = url.Values{"platform": {"invalid-json"}}.Encode()

			h.export(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})
		It("should parse valid platform JSON and pass to service", func() {
			platformJSON := `{"os":"linux","architecture":"amd64"}`
			req.URL.RawQuery = url.Values{"platform": {platformJSON}}.Encode()
			service.EXPECT().Export(gomock.Any(), name, gomock.Any(), gomock.Any()).Return(nil)

			h.export(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
	})
})
