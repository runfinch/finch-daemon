// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
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
	"github.com/runfinch/finch-daemon/mocks/mocks_exec"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Exec Inspect API", func() {
	var (
		mockCtrl    *gomock.Controller
		service     *mocks_exec.MockService
		conf        config.Config
		logger      *mocks_logger.Logger
		h           *handler
		rr          *httptest.ResponseRecorder
		req         *http.Request
		execInspect *types.ExecInspect
		inspectStr  []byte
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		service = mocks_exec.NewMockService(mockCtrl)
		logger = mocks_logger.NewLogger(mockCtrl)
		h = newHandler(service, &conf, logger)
		rr = httptest.NewRecorder()
		var err error
		req, err = http.NewRequest(http.MethodGet, "/exec/123/exec-123/inspect", nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"id": "123/exec-123"})
		execInspect = &types.ExecInspect{
			ID:       "exec-123",
			Running:  true,
			ExitCode: nil,
			ProcessConfig: &types.ExecProcessConfig{
				Tty:        true,
				Entrypoint: "foo",
				Arguments:  []string{"bar", "baz"},
				Privileged: nil,
				User:       "test",
			},
			OpenStdin:   false,
			OpenStdout:  true,
			OpenStderr:  true,
			CanRemove:   true,
			ContainerID: "123",
			DetachKeys:  nil,
			Pid:         123,
		}
		inspectStr, err = json.Marshal(execInspect)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return 200 on successful inspect", func() {
			service.EXPECT().Inspect(gomock.Any(), "123", "exec-123").Return(execInspect, nil)

			h.inspect(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Body).Should(MatchJSON(inspectStr))
		})
		It("should return 404 if the exec instance is not found", func() {
			service.EXPECT().Inspect(gomock.Any(), "123", "exec-123").Return(nil, errdefs.NewNotFound(fmt.Errorf("inspect error")))

			h.inspect(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(rr.Body).Should(MatchJSON(`{"message": "inspect error"}`))
		})
		It("should return 500 on any other error", func() {
			service.EXPECT().Inspect(gomock.Any(), "123", "exec-123").Return(nil, fmt.Errorf("inspect error"))

			h.inspect(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "inspect error"}`))
		})
	})
})
