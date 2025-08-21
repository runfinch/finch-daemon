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
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"go.uber.org/mock/gomock"

	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

var _ = Describe("Container Top API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_container.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
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

	Context("top handler", func() {
		It("should return 200 OK with process list using default ps args", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/id1/top", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			// Mock logger call
			logger.EXPECT().Debugf("calling nerdctl top with the following option : %s", "")

			// Mock successful response with default ps output
			defaultPsOutput := "UID PID PPID C STIME TTY TIME CMD\nroot 1 0 0 10:00 ? 00:00:00 sleep Infinity\n"
			service.EXPECT().Top(gomock.Any(), "id1", gomock.Any()).DoAndReturn(
				func(ctx context.Context, cid string, opts ncTypes.ContainerTopOptions) error {
					Expect(opts.PsArgs).Should(Equal(""))                // Empty string passed to service
					_, err := opts.Stdout.Write([]byte(defaultPsOutput)) // Service returns output as if "-ef" was used
					return err
				})

			h.top(rr, req)
			expectedResponse := `{
				"Titles": ["UID", "PID", "PPID", "C", "STIME", "TTY", "TIME", "CMD"],
				"Processes": [["root", "1", "0", "0", "10:00", "?", "00:00:00", "sleep Infinity"]]
			}`
			Expect(rr.Body).Should(MatchJSON(expectedResponse))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})

		It("should return 200 OK with custom ps args", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/id1/top?ps_args=-o pid,ppid,cmd", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			// Mock logger call
			logger.EXPECT().Debugf("calling nerdctl top with the following option : %s", "-o pid,ppid,cmd")

			customPsOutput := "PID PPID CMD\n1 0 sleep Infinity\n"
			service.EXPECT().Top(gomock.Any(), "id1", gomock.Any()).DoAndReturn(
				func(ctx context.Context, cid string, opts ncTypes.ContainerTopOptions) error {
					Expect(opts.PsArgs).Should(Equal("-o pid,ppid,cmd"))
					_, err := opts.Stdout.Write([]byte(customPsOutput))
					return err
				})

			h.top(rr, req)
			expectedResponse := `{
				"Titles": ["PID", "PPID", "CMD"],
				"Processes": [["1", "0", "sleep Infinity"]]
			}`
			Expect(rr.Body).Should(MatchJSON(expectedResponse))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})

		It("should return 400 when container ID is missing", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers//top", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": ""})

			h.top(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "must specify a container ID"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should return 404 when service returns a not found error", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/id1/top", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			// Mock logger call
			logger.EXPECT().Debugf("calling nerdctl top with the following option : %s", "")

			service.EXPECT().Top(gomock.Any(), "id1", gomock.Any()).Return(errdefs.NewNotFound(fmt.Errorf("not found")))

			h.top(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "not found"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})

		It("should return 409 when service returns a conflict error", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/id1/top", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			// Mock logger call
			logger.EXPECT().Debugf("calling nerdctl top with the following option : %s", "")

			service.EXPECT().Top(gomock.Any(), "id1", gomock.Any()).Return(errdefs.NewConflict(fmt.Errorf("conflict")))

			h.top(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "conflict"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusConflict))
		})

		It("should return 400 when service returns an invalid argument error", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/id1/top?ps_args=--invalid", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			// Mock logger call
			logger.EXPECT().Debugf("calling nerdctl top with the following option : %s", "--invalid")

			service.EXPECT().Top(gomock.Any(), "id1", gomock.Any()).Return(fmt.Errorf("unknown argument --invalid"))

			h.top(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "unknown argument --invalid"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should return 500 when service returns an internal error", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/id1/top", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			// Mock logger call
			logger.EXPECT().Debugf("calling nerdctl top with the following option : %s", "")

			service.EXPECT().Top(gomock.Any(), "id1", gomock.Any()).Return(fmt.Errorf("unexpected error"))

			h.top(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "unexpected error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})

		It("should return 500 when output format is invalid", func() {
			req, err := http.NewRequest(http.MethodGet, "/containers/id1/top", nil)
			Expect(err).Should(BeNil())
			req = mux.SetURLVars(req, map[string]string{"id": "id1"})

			// Mock logger call
			logger.EXPECT().Debugf("calling nerdctl top with the following option : %s", "")

			service.EXPECT().Top(gomock.Any(), "id1", gomock.Any()).DoAndReturn(
				func(ctx context.Context, cid string, opts ncTypes.ContainerTopOptions) error {
					// Write invalid output format
					_, err := opts.Stdout.Write([]byte("invalid output"))
					return err
				})

			h.top(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "invalid top output format"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})

var _ = Describe("parseTopOutput", func() {
	It("should parse standard process output correctly", func() {
		input := []string{
			"UID PID PPID C STIME TTY TIME CMD",
			"root 1 0 0 10:00 ? 00:00:00 sleep infinity",
			"root 7 1 0 10:01 ? 00:00:00 nginx -g daemon off",
		}

		titles, processes, err := parseTopOutput(input)
		Expect(err).Should(BeNil())
		Expect(titles).Should(Equal([]string{"UID", "PID", "PPID", "C", "STIME", "TTY", "TIME", "CMD"}))
		Expect(processes).Should(Equal([][]string{
			{"root", "1", "0", "0", "10:00", "?", "00:00:00", "sleep infinity"},
			{"root", "7", "1", "0", "10:01", "?", "00:00:00", "nginx -g daemon off"},
		}))
	})

	It("should handle command field with spaces correctly", func() {
		input := []string{
			"PID COMMAND",
			"1 /usr/local/bin/python3 -m http.server 8080",
		}

		titles, processes, err := parseTopOutput(input)
		Expect(err).Should(BeNil())
		Expect(titles).Should(Equal([]string{"PID", "COMMAND"}))
		Expect(processes).Should(Equal([][]string{
			{"1", "/usr/local/bin/python3 -m http.server 8080"},
		}))
	})

	It("should return error for empty titles", func() {
		input := []string{
			"",
			"1 sleep infinity",
		}
		titles, processes, err := parseTopOutput(input)
		Expect(err).Should(MatchError("no titles found"))
		Expect(titles).Should(BeNil())
		Expect(processes).Should(BeNil())
	})

	It("should handle empty process lines correctly", func() {
		input := []string{
			"PID CMD",
			"",
			"1 sleep infinity",
			"",
			"2 nginx",
		}
		titles, processes, err := parseTopOutput(input)
		Expect(err).Should(BeNil())
		Expect(titles).Should(Equal([]string{"PID", "CMD"}))
		Expect(processes).Should(Equal([][]string{
			{"1", "sleep infinity"},
			{"2", "nginx"},
		}))
	})
})
