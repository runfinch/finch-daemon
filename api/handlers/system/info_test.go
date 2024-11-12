// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_system"
)

var _ = Describe("System Info API Handler", func() {
	const (
		path = "/info"
	)

	var (
		mockController   *gomock.Controller
		service          *mocks_system.MockService
		handler          *handler
		responseRecorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		service = mocks_system.NewMockService(mockController)
		cfg := config.Config{}
		handler = newHandler(service, &cfg, nil, nil)
		responseRecorder = httptest.NewRecorder()
	})

	When("getting the system info", func() {
		It("should return 200 Ok and the system info", func() {
			expected := &dockercompat.Info{}
			service.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(expected, nil)

			httpRequest, err := http.NewRequest(http.MethodGet, path, nil)
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			handler.info(responseRecorder, httpRequest)
			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusOK))
		})
	})

	When("getting the system info causes an error", func() {
		It("should return 500 Internal Server Error and a message", func() {
			expected := errors.New("get system info error")
			service.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(nil, expected)

			httpRequest, err := http.NewRequest(http.MethodGet, path, nil)
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			handler.info(responseRecorder, httpRequest)
			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(responseRecorder.Body.String()).Should(MatchJSON(fmt.Sprintf(`{"message": "%v"}`, expected)))
		})
	})
})
