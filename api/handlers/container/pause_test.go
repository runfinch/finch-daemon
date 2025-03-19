// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/finch-daemon/pkg/errdefs"

	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

var _ = Describe("Container Pause API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_container.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		_        ncTypes.GlobalCommandOptions
		_        error
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
	})

	Context("pause handler", func() {
		It("should return 204 No Content on successful pause", func() {
			req, err := http.NewRequest(http.MethodPost, "/containers/id1/pause", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			service.EXPECT().Pause(gomock.Any(), "id1", gomock.Any()).DoAndReturn(
				func(ctx context.Context, cid string, opts ncTypes.ContainerPauseOptions) error {
					return nil
				})

			h.pause(rr, req)
			Expect(rr.Body.String()).Should(BeEmpty())
			Expect(rr).Should(HaveHTTPStatus(http.StatusNoContent))
		})

		It("should return 400 when container ID is missing", func() {
			req, err := http.NewRequest(http.MethodPost, "/containers//pause", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": ""})

			h.pause(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "must specify a container ID"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should return 404 when service returns a not found error", func() {
			req, err := http.NewRequest(http.MethodPost, "/containers/id1/pause", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			service.EXPECT().Pause(gomock.Any(), "id1", gomock.Any()).Return(
				errdefs.NewNotFound(fmt.Errorf("container not found")))

			h.pause(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "container not found"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})

		It("should return 409 when service returns a conflict error", func() {
			req, err := http.NewRequest(http.MethodPost, "/containers/id1/pause", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			service.EXPECT().Pause(gomock.Any(), "id1", gomock.Any()).Return(
				errdefs.NewConflict(fmt.Errorf("container already paused")))

			h.pause(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "container already paused"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusConflict))
		})

		It("should return 500 when service returns an internal error", func() {
			req, err := http.NewRequest(http.MethodPost, "/containers/id1/pause", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			service.EXPECT().Pause(gomock.Any(), "id1", gomock.Any()).Return(
				fmt.Errorf("unexpected internal error"))

			h.pause(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "unexpected internal error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
