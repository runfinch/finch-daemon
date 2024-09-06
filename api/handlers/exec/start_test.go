// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/pkg/config"
	hj "github.com/getlantern/httptest"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_exec"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Exec Start API", func() {
	var (
		mockCtrl  *gomock.Controller
		service   *mocks_exec.MockService
		conf      config.Config
		logger    *mocks_logger.Logger
		h         *handler
		rr        *httptest.ResponseRecorder
		opts      *types.ExecStartCheck
		req       *http.Request
		startOpts *types.ExecStartOptions
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		service = mocks_exec.NewMockService(mockCtrl)
		logger = mocks_logger.NewLogger(mockCtrl)
		h = newHandler(service, &conf, logger)
		rr = httptest.NewRecorder()
	})
	Context("handler", func() {
		Context("bad request", func() {
			It("should return 400 if the request body is empty", func() {
				var err error
				req, err = http.NewRequest(http.MethodPost, "/exec/123/exec-123/start", nil)
				Expect(err).Should(BeNil())
				req = mux.SetURLVars(req, map[string]string{"id": "123/exec-123"})

				h.start(rr, req)
				Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
				Expect(rr.Body).Should(MatchJSON(`{"message": "body should not be empty"}`))
			})
			It("should return 400 if the body reader returns an error", func() {
				var err error
				req, err = http.NewRequest(http.MethodPost, "/exec/123/exec-123/start", NewErrorReader())
				Expect(err).Should(BeNil())
				req = mux.SetURLVars(req, map[string]string{"id": "123/exec-123"})

				h.start(rr, req)
				Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
				Expect(rr.Body).Should(MatchJSON(`{"message": "unable to parse request body: read error"}`))
			})
			It("should return 400 if the request body is not an ExecStartCheck", func() {
				var err error
				req, err = http.NewRequest(http.MethodPost, "/exec/123/exec-123/start", bytes.NewReader([]byte("foo")))
				Expect(err).Should(BeNil())
				req = mux.SetURLVars(req, map[string]string{"id": "123/exec-123"})

				h.start(rr, req)
				Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
				Expect(rr.Body).Should(MatchJSON(`{"message": "unable to parse request body: invalid character 'o' in literal false (expecting 'a')"}`))
			})
		})
		Context("detach == true", func() {
			BeforeEach(func() {
				opts = &types.ExecStartCheck{
					Detach:      true,
					Tty:         false,
					ConsoleSize: &[2]uint{123, 123},
				}
				reqBody, err := json.Marshal(opts)
				Expect(err).Should(BeNil())
				req, err = http.NewRequest(http.MethodPost, "/exec/123/exec-123/start", bytes.NewReader(reqBody))
				Expect(err).Should(BeNil())
				req = mux.SetURLVars(req, map[string]string{"id": "123/exec-123"})
				startOpts = &types.ExecStartOptions{
					ExecStartCheck: opts,
					ConID:          "123",
					ExecID:         "exec-123",
					Stdin:          nil,
					Stdout:         nil,
					Stderr:         nil,
				}
			})
			It("should return 200 on successful start", func() {
				service.EXPECT().Start(gomock.Any(), startOpts).Return(nil)

				h.start(rr, req)
				Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			})
			It("should return 404 if the exec instance is not found", func() {
				service.EXPECT().Start(gomock.Any(), startOpts).Return(errdefs.NewNotFound(fmt.Errorf("not found")))

				h.start(rr, req)
				Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
				Expect(rr.Body).Should(MatchJSON(`{"message": "not found"}`))
			})
			It("should return 500 if it fails to start", func() {
				service.EXPECT().Start(gomock.Any(), startOpts).Return(fmt.Errorf("failed to start"))

				h.start(rr, req)
				Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
				Expect(rr.Body).Should(MatchJSON(`{"message": "failed to start"}`))
			})
		})
		XContext("detach == false", func() {
			var rr *hj.HijackableResponseRecorder

			BeforeEach(func() {
				opts = &types.ExecStartCheck{
					Detach:      false,
					Tty:         true,
					ConsoleSize: &[2]uint{123, 123},
				}
				reqBody, err := json.Marshal(opts)
				Expect(err).Should(BeNil())
				req, err = http.NewRequest(http.MethodPost, "/exec/123/exec-123/start", bytes.NewReader(reqBody))
				Expect(err).Should(BeNil())
				req = mux.SetURLVars(req, map[string]string{"id": "123/exec-123"})
				rr = hj.NewRecorder(nil)
			})
			It("should return 500 if Hijacking the connection fails", func() {
				hijackErrMsg := "error hijacking the connection"
				errRR := newResponseRecorderWithMockHijack(rr, func() (net.Conn, *bufio.ReadWriter, error) {
					return nil, nil, fmt.Errorf("%s", hijackErrMsg)
				})

				h.start(errRR, req)
				Expect(rr.Code()).Should(Equal(http.StatusInternalServerError))
				Expect(rr.Body().Bytes()).Should(MatchJSON(fmt.Sprintf(`{"message": "%s"}`, hijackErrMsg)))
			})
			It("should return 404 if the exec instance is not found", func() {
				service.EXPECT().Start(gomock.Any(), execStartOptionsWithIdsAndCheck("123", "exec-123", opts)).Return(errdefs.NewNotFound(fmt.Errorf("not found")))

				h.start(rr, req)
				rrBody := (*(rr.Body())).String()
				Expect(rrBody).Should(Equal("HTTP/1.1 404 Not Found\r\nContent-Type: application/json\r\n\r\n{\"message\":\"not found\"}\r\n"))
			})
			It("should return 409 if the container is not running", func() {
				service.EXPECT().Start(gomock.Any(), execStartOptionsWithIdsAndCheck("123", "exec-123", opts)).Return(errdefs.NewConflict(fmt.Errorf("not running")))

				h.start(rr, req)
				rrBody := (*(rr.Body())).String()
				Expect(rrBody).Should(Equal("HTTP/1.1 409 Conflict\r\nContent-Type: application/json\r\n\r\n{\"message\":\"not running\"}\r\n"))
			})
			It("should return 500 on other errors", func() {
				service.EXPECT().Start(gomock.Any(), execStartOptionsWithIdsAndCheck("123", "exec-123", opts)).Return(fmt.Errorf("start error"))

				h.start(rr, req)
				rrBody := (*(rr.Body())).String()
				Expect(rrBody).Should(Equal("HTTP/1.1 500 Internal Server Error\r\nContent-Type: application/json\r\n\r\n{\"message\":\"start error\"}\r\n"))
			})
			It("should correctly upgrade if requested", func() {
				req.Header.Set("Upgrade", "foo")
				service.EXPECT().Start(gomock.Any(), execStartOptionsWithIdsAndCheck("123", "exec-123", opts)).DoAndReturn(
					func(ctx context.Context, execStartOptions *types.ExecStartOptions) error {
						execStartOptions.SuccessResponse()
						return nil
					})

				h.start(rr, req)

				rrBody := (*(rr.Body())).String()
				Expect(rrBody).Should(Equal(fmt.Sprintf("HTTP/1.1 101 UPGRADED\r\nContent-Type: %s\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n",
					"application/vnd.docker.raw-stream")))
			})
			It("should stream output from the started process", func() {
				service.EXPECT().Start(gomock.Any(), execStartOptionsWithIdsAndCheck("123", "exec-123", opts)).DoAndReturn(
					func(ctx context.Context, execStartOptions *types.ExecStartOptions) error {
						execStartOptions.SuccessResponse()
						rrBody := (*(rr.Body())).String()
						Expect(rrBody).Should(Equal("HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n"))
						rr.Body().Reset()

						responses := [3]string{"foo", "bar", "baz"}
						for i, response := range responses {
							fmt.Fprint(execStartOptions.Stdout, response)
							rrBody = (*(rr.Body())).String()
							Expect(rrBody).Should(Equal(responses[i]))
							rr.Body().Reset()
						}
						return nil
					})

				h.start(rr, req)
			})
			It("should print any errors from Start to the connection if the success response has been returned", func() {
				service.EXPECT().Start(gomock.Any(), execStartOptionsWithIdsAndCheck("123", "exec-123", opts)).DoAndReturn(
					func(ctx context.Context, execStartOptions *types.ExecStartOptions) error {
						execStartOptions.SuccessResponse()
						return fmt.Errorf("start error")
					})

				h.start(rr, req)
				rrBody := (*(rr.Body())).String()
				Expect(rrBody).Should(Equal("HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\nstart error\r\n"))
			})
		})
	})
})

func NewErrorReader() io.Reader {
	return &errorReader{}
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

func newResponseRecorderWithMockHijack(wrapped http.ResponseWriter, mockHijack func() (net.Conn, *bufio.ReadWriter, error)) *responseRecorderWithMockHijack {
	h := &responseRecorderWithMockHijack{
		wrapped:    wrapped,
		mockHijack: mockHijack,
	}
	return h
}

type responseRecorderWithMockHijack struct {
	wrapped    http.ResponseWriter
	mockHijack func() (net.Conn, *bufio.ReadWriter, error)
}

func (h *responseRecorderWithMockHijack) Header() http.Header {
	return h.wrapped.Header()
}

func (h *responseRecorderWithMockHijack) Write(bytes []byte) (int, error) {
	return h.wrapped.Write(bytes)
}

func (h *responseRecorderWithMockHijack) WriteHeader(statusCode int) {
	h.wrapped.WriteHeader(statusCode)
}

func (h *responseRecorderWithMockHijack) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h.mockHijack()
}

func execStartOptionsWithIdsAndCheck(conId string, execId string, check *types.ExecStartCheck) gomock.Matcher {
	return &execStartOptionsMatcher{
		conId:  conId,
		execId: execId,
		check:  check,
	}
}

type execStartOptionsMatcher struct {
	conId  string
	execId string
	check  *types.ExecStartCheck
}

func (m *execStartOptionsMatcher) Matches(x interface{}) bool {
	matchOpts, ok := x.(*types.ExecStartOptions)
	if !ok {
		return false
	}
	if m.conId != matchOpts.ConID {
		return false
	}
	if m.execId != matchOpts.ExecID {
		return false
	}
	if m.check.Detach != matchOpts.Detach {
		return false
	}
	if m.check.Tty != matchOpts.Tty {
		return false
	}
	if m.check.ConsoleSize[0] != matchOpts.ConsoleSize[0] {
		return false
	}
	if m.check.ConsoleSize[1] != matchOpts.ConsoleSize[1] {
		return false
	}
	return true
}

func (m *execStartOptionsMatcher) String() string {
	return fmt.Sprintf("*types.ExecStartOptions with ConId: %s, ExecId: %s, ExecStartCheck: %v", m.conId, m.execId, m.check)
}
