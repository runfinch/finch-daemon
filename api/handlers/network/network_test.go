// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containerd/nerdctl/pkg/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
)

// TestNetworkHandler function is the entry point of network handler package's unit test using ginkgo.
func TestNetworkHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Network APIs Handler")
}

var _ = Describe("Network API ", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_network.MockService
		rr       *httptest.ResponseRecorder
		req      *http.Request
		conf     *config.Config
		router   *mux.Router
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		// initialize mocks
		service = mocks_network.NewMockService(mockCtrl)
		conf = &config.Config{}
		router = mux.NewRouter()
		logger = mocks_logger.NewLogger(mockCtrl)
		RegisterHandlers(types.VersionedRouter{Router: router}, service, conf, logger)
		rr = httptest.NewRecorder()
	})
	Context("handler", func() {
		When("POST /networks/create", func() {
			It("should call network create handler", func() {
				const (
					networkName = "test-network"
				)

				logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).MinTimes(1)
				service.EXPECT().Create(gomock.Any(), gomock.Any()).MaxTimes(1)

				jsonBytes := []byte(fmt.Sprintf(`{"Name": "%s"}`, networkName))
				req, err := http.NewRequest(http.MethodPost, "/networks/create", bytes.NewReader(jsonBytes))
				Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

				router.ServeHTTP(rr, req)
			})
		})

		When("GET /networks/{id}", func() {
			It("should call network inspect handler", func() {
				// setup mocks
				service.EXPECT().Inspect(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error from Inspect"))
				req, _ = http.NewRequest(http.MethodGet, "/networks/123", nil)
				// call the API to check if it returns the error generated from Inspect method
				router.ServeHTTP(rr, req)
				Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
				Expect(rr.Body.String()).Should(MatchJSON(`{"message": "error from Inspect"}`))
			})
		})
		It("should call the network list handler using /networks", func() {
			// setup mocks
			expErr := "error from List"
			service.EXPECT().List(gomock.Any()).Return(nil, fmt.Errorf("%s", expErr))
			req, _ = http.NewRequest(http.MethodGet, "/networks", nil)
			// call api and check if it returns error
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body.String()).Should(MatchJSON(fmt.Sprintf(`{"message": "%s"}`, expErr)))
		})
		It("should call the network list handler using /networks/", func() {
			// setup mocks
			expErr := "error from List"
			service.EXPECT().List(gomock.Any()).Return(nil, fmt.Errorf("%s", expErr))
			req, _ = http.NewRequest(http.MethodGet, "/networks/", nil)
			// call api and check if it returns error
			router.ServeHTTP(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body.String()).Should(MatchJSON(fmt.Sprintf(`{"message": "%s"}`, expErr)))
		})
	})
})
