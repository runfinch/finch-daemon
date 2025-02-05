// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Network Inspect API ", func() {
	var (
		mockCtrl    *gomock.Controller
		service     *mocks_network.MockService
		rr          *httptest.ResponseRecorder
		req         *http.Request
		handler     *handler
		conf        *config.Config
		logger      *mocks_logger.Logger
		mockNet     *types.NetworkInspectResponse
		mockNetJSON []byte
		nid         string
	)
	BeforeEach(func() {
		nid = "123"
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		// initialize mocks
		service = mocks_network.NewMockService(mockCtrl)
		conf = &config.Config{}
		logger = mocks_logger.NewLogger(mockCtrl)
		handler = newHandler(service, conf, logger)
		rr = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/networks/%s", nid), nil)
		req = mux.SetURLVars(req, map[string]string{"id": nid})
		mockNet = &types.NetworkInspectResponse{
			Name: "name",
			ID:   nid,
			IPAM: dockercompat.IPAM{
				Config: []dockercompat.IPAMConfig{
					{Subnet: "10.5.2.0/24", Gateway: "10.5.2.1"},
				},
			},
			Labels: map[string]string{"label": "value"},
		}
		var err error
		mockNetJSON, err = json.Marshal(mockNet)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return a 200 when there is no error", func() {
			service.EXPECT().Inspect(gomock.Any(), nid).Return(mockNet, nil)

			handler.inspect(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			Expect(rr.Body.String()).Should(MatchJSON(mockNetJSON))
		})
		It("should return a 404 when Inspect returns notFound", func() {
			service.EXPECT().Inspect(gomock.Any(), nid).Return(nil, errdefs.NewNotFound(fmt.Errorf("not found")))

			handler.inspect(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(rr.Body.String()).Should(MatchJSON(`{"message": "not found"}`))
		})
		It("should return a 500 when Inspect returns any other error", func() {
			service.EXPECT().Inspect(gomock.Any(), nid).Return(nil, fmt.Errorf("internal error"))

			handler.inspect(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body.String()).Should(MatchJSON(`{"message": "internal error"}`))
		})
		It("should return a 500 if the network ID is empty", func() {
			req = mux.SetURLVars(req, map[string]string{"id": ""})

			handler.inspect(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body.String()).Should(MatchJSON(`{"message": "id cannot be empty"}`))
		})
	})
})
