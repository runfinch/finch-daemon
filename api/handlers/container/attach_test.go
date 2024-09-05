// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/moby/moby/api/server/httputils"

	"github.com/runfinch/finch-daemon/api/types"

	"github.com/containerd/nerdctl/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_http"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"

	hj "github.com/getlantern/httptest"
)

const errRRErrMsg = "error hijacking the connection"

var _ = Describe("Container Attach API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_container.MockService
		h        *handler
		rr       *hj.HijackableResponseRecorder
		req      *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_container.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = hj.NewRecorder(nil)
	})
	Context("handler", func() {
		It("should return an internal server error when failing to hijack the conn", func() {
			// define expected values, setup mock, create request using ErrorResponseRecorder defined below
			expErrCode := http.StatusInternalServerError
			expErrMsg := errRRErrMsg
			rrErr := newErrorResponseRecorder()
			req, _ = http.NewRequest(http.MethodPost, "/containers/123", nil)

			h.attach(rrErr, req)

			Expect(rrErr.Code()).Should(Equal(expErrCode))
			Expect(rrErr.Body()).Should(MatchJSON(`{"message": "` + expErrMsg + `"}`))
		})
		It("should return an internal server error for default errors", func() {
			// define expected values, setup mock, create request
			expErrCode := http.StatusInternalServerError
			expErrMsg := "error"
			service.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(fmt.Errorf("%s", expErrMsg))
			req, _ = http.NewRequest(http.MethodPost, "/containers/123", nil)

			h.attach(rr, req)

			rrBody := (*(rr.Body())).String()
			Expect(rrBody).Should(Equal(fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: %s\r\n\r\n%s\r\n", expErrCode,
				http.StatusText(expErrCode), "application/vnd.docker.raw-stream", expErrMsg)))
		})
		It("should return a 404 error for container not found", func() {
			// define expected values, setup mock, create request
			expErrCode := http.StatusNotFound
			expErrMsg := fmt.Sprintf("no container is found given the string: %s", "123")
			service.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errdefs.NewNotFound(fmt.Errorf("%s", expErrMsg)))
			req, _ = http.NewRequest(http.MethodPost, "/containers/123", nil)

			h.attach(rr, req)

			rrBody := (*(rr.Body())).String()
			Expect(rrBody).Should(Equal(fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: %s\r\n\r\n%s\r\n", expErrCode,
				http.StatusText(expErrCode), "application/vnd.docker.raw-stream", expErrMsg)))
		})
		It("should succeed upon no errors in service.Attach and close the connection", func() {
			service.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			req, _ = http.NewRequest(http.MethodPost, "/containers/123", nil)

			h.attach(rr, req)

			Expect(rr.Closed()).Should(BeTrue())
		})
		It("should handle the url variable parsing correctly", func() {
			cid := "test_con"
			vars := map[string]string{
				"id": cid,
			}
			expectedOpts := &types.AttachOptions{
				GetStreams: func() (io.Writer, io.Writer, chan os.Signal, func(), error) {
					conn, _, err := rr.Hijack()
					Expect(err).Should(BeNil())
					return conn, conn, nil, nil, nil
				},
				UseStdin:   true,
				UseStdout:  true,
				UseStderr:  true,
				Logs:       true,
				Stream:     true,
				MuxStreams: true,
			}
			service.EXPECT().Attach(gomock.Any(), cid, attachOptsEqualTo(expectedOpts)).Return(nil)
			req, _ = http.NewRequest(http.MethodPost, "/containers/"+cid+"/attach?"+
				"stdin=1&"+
				"stdout=1&"+
				"stderr=1&"+
				"logs=1&"+
				"stream=1", nil)
			req = mux.SetURLVars(req, vars)

			h.attach(rr, req)

			Expect(rr.Closed()).Should(BeTrue())
		})
	})
	Context("testing the checkUpgradeStatus helper function", func() {
		var (
			defHeader      func(ct string) string
			upgradedHeader func(ct string) string
			defContentType string
			muxContentType string
		)
		BeforeEach(func() {
			defHeader = func(ct string) string {
				return "HTTP/1.1 200 OK\r\n" +
					"Content-Type: " + ct + "\r\n\r\n"
			}
			upgradedHeader = func(ct string) string {
				return "HTTP/1.1 101 UPGRADED\r\n" +
					"Content-Type: " + ct + "\r\n" +
					"Connection: Upgrade\r\n" +
					"Upgrade: tcp\r\n\r\n"
			}
			defContentType = "application/vnd.docker.raw-stream"
			muxContentType = "application/vnd.docker.multiplexed-stream"
		})
		It("should not return an upgraded header if upgrade is false", func() {
			ct, r := checkUpgradeStatus(context.Background(), false)

			Expect(r).Should(Equal(defHeader(defContentType)))
			Expect(ct).Should(Equal(defContentType))
		})
		It("should return an upgraded header without mux if version < 1.42 & upgrade is true", func() {
			ctx := context.WithValue(context.Background(), httputils.APIVersionKey{}, "1.41")
			ct, r := checkUpgradeStatus(ctx, true)

			Expect(r).Should(Equal(upgradedHeader(defContentType)))
			Expect(ct).Should(Equal(defContentType))
		})
		It("should return an upgraded header with mux if version = 1.42 & upgrade is true", func() {
			ctx := context.WithValue(context.Background(), httputils.APIVersionKey{}, "1.42")
			ct, r := checkUpgradeStatus(ctx, true)

			Expect(r).Should(Equal(upgradedHeader(muxContentType)))
			Expect(ct).Should(Equal(muxContentType))
		})
		It("should return an upgraded header with mux if version > 1.42 & upgrade is true", func() {
			ctx := context.WithValue(context.Background(), httputils.APIVersionKey{}, "1.43")
			ct, r := checkUpgradeStatus(ctx, true)

			Expect(r).Should(Equal(upgradedHeader(muxContentType)))
			Expect(ct).Should(Equal(muxContentType))
		})
	})
	Context("testing the checkConnection helper function", func() {
		var mockConn *mocks_http.MockConn
		BeforeEach(func() {
			mockConn = mocks_http.NewMockConn(mockCtrl)
			mockConn.EXPECT().Close().Do(func() {
				mockConn = nil
			})
		})
		It("should close the connection when there is an io.EOF error", func() {
			mockConn.EXPECT().Read(gomock.Any()).Return(0, io.EOF)
			checkConnection(mockConn, func() {})
			Expect(mockConn).Should(BeNil())
		})
		It("shouldn't close the connection when there are no errors", func() {
			mockConn.EXPECT().Read(gomock.Any()).Return(0, nil)
			checkConnection(mockConn, func() {})
			Expect(mockConn).ShouldNot(BeNil())
		})
		It("shouldn't close the connection when there is a non io.EOF error", func() {
			mockConn.EXPECT().Read(gomock.Any()).Return(0, errors.New("not io.EOF"))
			checkConnection(mockConn, func() {})
			Expect(mockConn).ShouldNot(BeNil())
		})
	})
})

func newErrorResponseRecorder() *errorResponseRecorder {
	wrapped := httptest.NewRecorder()
	h := &errorResponseRecorder{
		wrapped: wrapped,
	}
	return h
}

type errorResponseRecorder struct {
	wrapped *httptest.ResponseRecorder
}

func (h *errorResponseRecorder) Header() http.Header {
	return h.wrapped.Header()
}

func (h *errorResponseRecorder) Write(bytes []byte) (int, error) {
	return h.wrapped.Write(bytes)
}

func (h *errorResponseRecorder) WriteHeader(statusCode int) {
	h.wrapped.WriteHeader(statusCode)
}

func (h *errorResponseRecorder) Body() *bytes.Buffer {
	return h.wrapped.Body
}

func (h *errorResponseRecorder) Code() int {
	return h.wrapped.Code
}

func (h *errorResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("%s", errRRErrMsg)
}

// attachOptsMatcher is adapted from container create to be a wrapper type to
// compare attach option structs when we cannot define what GetStreams will be.
type attachOptsMatcher struct {
	obj        *types.AttachOptions
	mismatches []string
}

func attachOptsEqualTo(object *types.AttachOptions) *attachOptsMatcher {
	return &attachOptsMatcher{
		obj:        object,
		mismatches: []string{},
	}
}

func (e *attachOptsMatcher) Matches(x interface{}) bool {
	y := x.(*types.AttachOptions)

	gotStdout, gotStderr, _, _, gotErr := y.GetStreams()
	wantStdout, wantStderr, _, _, wantErr := e.obj.GetStreams()
	if wantStdout != gotStdout {
		e.mismatches = append(e.mismatches, "GetStreams() - stdout")
	}
	if wantStderr != gotStderr {
		e.mismatches = append(e.mismatches, "GetStreams() - stderr")
	}
	if gotErr != wantErr {
		e.mismatches = append(e.mismatches, "GetStreams() - error")
	}
	if e.obj.UseStdin != y.UseStdin {
		e.mismatches = append(e.mismatches, "UseStdin")
	}
	if e.obj.UseStdout != y.UseStdout {
		e.mismatches = append(e.mismatches, "UseStdout")
	}
	if e.obj.UseStderr != y.UseStderr {
		e.mismatches = append(e.mismatches, "UseStderr")
	}
	if e.obj.Logs != y.Logs {
		e.mismatches = append(e.mismatches, "Logs")
	}
	if e.obj.Stream != y.Stream {
		e.mismatches = append(e.mismatches, "Stream")
	}

	if len(e.mismatches) > 0 {
		return false
	}
	return true
}

func (e *attachOptsMatcher) String() string {
	return strings.Join(e.mismatches, ",")
}
