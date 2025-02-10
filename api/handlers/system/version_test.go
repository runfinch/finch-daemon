// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_system"
)

var _ = Describe("Version API ", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_system.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		req      *http.Request
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_system.NewMockService(mockCtrl)
		ncClient := mocks_backend.NewMockNerdctlSystemSvc(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, ncClient, logger)
		rr = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "/version", nil)
	})
	Context("handler", func() {
		It("should return 200 as success response", func() {
			// service mock returns nil to mimic handler generated the version info successfully.
			expectedVersion := types.VersionInfo{
				Platform: struct {
					Name string
				}{},
				Version:       "0.0.1",
				ApiVersion:    "1.43",
				MinAPIVersion: "1.35",
				GitCommit:     "abcd",
				Os:            "linux",
				Arch:          "x86",
				KernelVersion: "kernel-123",
				Experimental:  true,
				Components: []types.ComponentVersion{
					{
						Name:    "containerd",
						Version: "v1.7.1",
						Details: map[string]string{
							"GitCommit": "1677a17964311325ed1c31e2c0a3589ce6d5c30d",
						},
					},
				},
			}
			service.EXPECT().GetVersion(gomock.Any()).Return(&expectedVersion, nil)

			h.version(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			jd := json.NewDecoder(rr.Body)
			var v types.VersionInfo
			err := jd.Decode(&v)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v).Should(Equal(expectedVersion))
		})

		It("should return 500 internal error response", func() {
			// service mock returns not found error to mimic version info could not generate due internal error.
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Return().AnyTimes()
			service.EXPECT().GetVersion(gomock.Any()).Return(nil, fmt.Errorf("some error"))

			h.version(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "some error"}`))
		})
	})
})
