// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_image"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Image Inspect API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_image.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		name     string
		req      *http.Request
		resp     dockercompat.Image
		respJSON []byte
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
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/images/%s/json", name), nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"name": name})
		resp = dockercompat.Image{
			ID:          name,
			RepoTags:    []string{"test-image:latest"},
			RepoDigests: []string{"test-image@test-digest"},
			Size:        100,
		}
		respJSON, err = json.Marshal(resp)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return inspect object and 200 status code upon success", func() {
			service.EXPECT().Inspect(gomock.Any(), name).Return(&resp, nil)

			// handler should return response object with 200 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(respJSON))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 404 status code if image was not found", func() {
			service.EXPECT().Inspect(gomock.Any(), name).Return(nil, errdefs.NewNotFound(fmt.Errorf("no such image")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message with 404 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "no such image"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Inspect(gomock.Any(), name).Return(nil, fmt.Errorf("error"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
