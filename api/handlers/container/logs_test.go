// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/containerd/nerdctl/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"

	hj "github.com/getlantern/httptest"
)

var _ = Describe("Container Logs API", func() {
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
		req, _ = http.NewRequest(http.MethodGet, "/containers/123/logs?stdout=1&stderr=1", nil)
		h = newHandler(service, &c, logger)
		rr = hj.NewRecorder(nil)
	})
	Context("handler", func() {
		It("should return an error if the user does not set stdout or stderr", func() {
			expErrCode := http.StatusBadRequest
			expErrMsg := "you must choose at least one stream"
			req, _ = http.NewRequest(http.MethodGet, "/containers/123/logs", nil)

			h.logs(rr, req)

			Expect(rr.Code()).Should(Equal(expErrCode))
			Expect(rr.Body()).Should(MatchJSON(`{"message": "` + expErrMsg + `"}`))
		})
		It("should return an internal server error when failing to hijack the conn", func() {
			// define expected values, setup mock, create request using ErrorResponseRecorder defined below
			expErrCode := http.StatusInternalServerError
			expErrMsg := errRRErrMsg
			rrErr := newErrorResponseRecorder()

			h.logs(rrErr, req)

			Expect(rrErr.Code()).Should(Equal(expErrCode))
			Expect(rrErr.Body()).Should(MatchJSON(`{"message": "` + expErrMsg + `"}`))
		})
		It("should return an internal server error for default errors", func() {
			// define expected values, setup mock, create request
			expErrCode := http.StatusInternalServerError
			expErrMsg := "error"
			service.EXPECT().Logs(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(fmt.Errorf("%s", expErrMsg))

			h.logs(rr, req)

			rrBody := (*(rr.Body())).String()
			Expect(rrBody).Should(Equal(fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: %s\r\n\r\n%s\r\n", expErrCode,
				http.StatusText(expErrCode), "application/vnd.docker.raw-stream", expErrMsg)))
		})
		It("should return a 404 error for container not found", func() {
			// define expected values, setup mock, create request
			expErrCode := http.StatusNotFound
			expErrMsg := fmt.Sprintf("no container is found given the string: %s", "123")
			service.EXPECT().Logs(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errdefs.NewNotFound(fmt.Errorf("%s", expErrMsg)))

			h.logs(rr, req)

			rrBody := (*(rr.Body())).String()
			Expect(rrBody).Should(Equal(fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: %s\r\n\r\n%s\r\n", expErrCode,
				http.StatusText(expErrCode), "application/vnd.docker.raw-stream", expErrMsg)))
		})
		It("should succeed upon no errors in service.Attach and close the connection", func() {
			service.EXPECT().Logs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			h.logs(rr, req)

			Expect(rr.Closed()).Should(BeTrue())
		})
		It("should handle the url variable parsing correctly", func() {
			cid := "test_con"
			vars := map[string]string{
				"id": cid,
			}
			expectedOpts := &types.LogsOptions{
				GetStreams: func() (io.Writer, io.Writer, chan os.Signal, func(), error) {
					conn, _, err := rr.Hijack()
					Expect(err).Should(BeNil())
					return conn, conn, nil, nil, nil
				},
				Stdout:     true,
				Stderr:     true,
				Follow:     true,
				Since:      10,
				Until:      11,
				Timestamps: true,
				Tail:       "all",
				MuxStreams: true,
			}
			service.EXPECT().Logs(gomock.Any(), cid, logsOptsEqualTo(expectedOpts)).Return(nil)
			req, _ = http.NewRequest(http.MethodGet, "/containers/"+cid+"/logs?"+
				"stdout=1&"+
				"stderr=1&"+
				"follow=1&"+
				"since=10&"+
				"until=11&"+
				"timestamps=1&"+
				"tail=all", nil)
			req = mux.SetURLVars(req, vars)

			h.logs(rr, req)

			Expect(rr.Closed()).Should(BeTrue())
		})
	})
})

// logsOptsMatcher is adapted from container create to be a wrapper type to
// compare logs option structs when we cannot define what GetStreams will be.
type logsOptsMatcher struct {
	obj        *types.LogsOptions
	mismatches []string
}

func logsOptsEqualTo(object *types.LogsOptions) *logsOptsMatcher {
	return &logsOptsMatcher{
		obj:        object,
		mismatches: []string{},
	}
}

func (e *logsOptsMatcher) Matches(x interface{}) bool {
	y := x.(*types.LogsOptions)

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
	if e.obj.Stdout != y.Stdout {
		e.mismatches = append(e.mismatches, "Stdout")
	}
	if e.obj.Stderr != y.Stderr {
		e.mismatches = append(e.mismatches, "Stderr")
	}
	if e.obj.Follow != y.Follow {
		e.mismatches = append(e.mismatches, "Follow")
	}
	if e.obj.Since != y.Since {
		e.mismatches = append(e.mismatches, "Since")
	}
	if e.obj.Until != y.Until {
		e.mismatches = append(e.mismatches, "Until")
	}
	if e.obj.Timestamps != y.Timestamps {
		e.mismatches = append(e.mismatches, "Timestamps")
	}
	if e.obj.Tail != y.Tail {
		e.mismatches = append(e.mismatches, "Tail")
	}
	if e.obj.MuxStreams != y.MuxStreams {
		e.mismatches = append(e.mismatches, "MuxStreams")
	}

	if len(e.mismatches) > 0 {
		return false
	}
	return true
}

func (e *logsOptsMatcher) String() string {
	return strings.Join(e.mismatches, ",")
}
