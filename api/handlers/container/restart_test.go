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

	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Restart API ", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_container.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		req      *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodPost, "/containers/123/restart", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "123"})
	})
	Context("handler", func() {
		It("should return 204 as success response", func() {
			// service mock returns nil to mimic handler started the container successfully.
			service.EXPECT().Restart(gomock.Any(), "123", gomock.Any()).Return(nil)

			//handler should return success message with 204 status code.
			h.restart(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNoContent))
		})

		It("should return 404 not found response", func() {
			// service mock returns not found error to mimic user trying to start container that does not exist
			service.EXPECT().Restart(gomock.Any(), "123", gomock.Any()).Return(
				errdefs.NewNotFound(fmt.Errorf("container not found")))
			logger.EXPECT().Debugf("Restart container API responding with error code. Status code %d, Message: %s", 404, "container not found")

			//handler should return 404 status code with an error msg.
			h.restart(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(rr.Body).Should(MatchJSON(`{"message": "container not found"}`))
		})
		It("should return 500 internal error response", func() {
			// service mock return error to mimic a user trying to start a container with an id that has
			// multiple containers with same prefix.
			service.EXPECT().Restart(gomock.Any(), "123", gomock.Any()).Return(
				fmt.Errorf("multiple IDs found with provided prefix"))
			logger.EXPECT().Debugf("Restart container API responding with error code. Status code %d, Message: %s", 500, "multiple IDs found with provided prefix")

			//handler should return 500 status code with an error msg.
			h.restart(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "multiple IDs found with provided prefix"}`))
		})
	})
})
