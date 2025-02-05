// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"encoding/json"
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

var _ = Describe("Container Exec API ", func() {
	var (
		mockCtrl   *gomock.Controller
		logger     *mocks_logger.Logger
		service    *mocks_container.MockService
		h          *handler
		rr         *httptest.ResponseRecorder
		req        *http.Request
		execConfig *types.ExecConfig
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		execConfig = &types.ExecConfig{
			User:         "foo",
			Privileged:   false,
			Tty:          true,
			ConsoleSize:  &[2]uint{123, 123},
			AttachStdin:  false,
			AttachStderr: true,
			AttachStdout: true,
			Detach:       false,
			DetachKeys:   "ctrl-C",
			Env:          []string{"FOO=bar"},
			WorkingDir:   "/foo/bar",
			Cmd:          []string{"foo", "bar"},
		}
		bodyBytes, err := json.Marshal(execConfig)
		Expect(err).Should(BeNil())
		req, err = http.NewRequest(http.MethodPost, "/containers/123/exec", bytes.NewReader(bodyBytes))
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"id": "123"})
	})
	Context("handler", func() {
		It("should return 201 on successful exec", func() {
			service.EXPECT().ExecCreate(gomock.Any(), "123", *execConfig).Return("exec123", nil)

			h.exec(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(rr.Body).Should(MatchJSON(`{"Id": "exec123"}`))
		})
		It("should return 404 if the container is not found", func() {
			service.EXPECT().ExecCreate(gomock.Any(), "123", *execConfig).Return("", errdefs.NewNotFound(fmt.Errorf("not found")))

			h.exec(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(rr.Body).Should(MatchJSON(`{"message": "not found"}`))
		})
		It("should return 409 if the container is not running", func() {
			service.EXPECT().ExecCreate(gomock.Any(), "123", *execConfig).Return("", errdefs.NewConflict(fmt.Errorf("not running")))

			h.exec(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusConflict))
			Expect(rr.Body).Should(MatchJSON(`{"message": "not running"}`))
		})
		It("should return 500 on any other error", func() {
			service.EXPECT().ExecCreate(gomock.Any(), "123", *execConfig).Return("", fmt.Errorf("exec create error"))

			h.exec(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "exec create error"}`))
		})
		It("should return 400 if the request body is not an ExecConfig", func() {
			var err error
			req, err = http.NewRequest(http.MethodPost, "/containers/123/exec", bytes.NewReader([]byte("foo")))
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "123"})

			h.exec(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body).Should(MatchJSON(`{"message": "unable to parse request body: invalid character 'o' in literal false (expecting 'a')"}`))
		})
		It("should return 400 if the request body is empty", func() {
			var err error
			req, err = http.NewRequest(http.MethodPost, "/containers/123/exec", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "123"})

			h.exec(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(rr.Body).Should(MatchJSON(`{"message": "request body should not be empty"}`))
		})
	})
})
