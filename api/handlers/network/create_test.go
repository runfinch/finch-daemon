// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_network"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

type errorOnRead struct{}

func (eor *errorOnRead) Read(p []byte) (int, error) {
	return 0, errors.New("error on read")
}

var _ = Describe("Network Create API Handler", func() {
	const (
		path          = "/networks/create"
		networkName   = "test-network"
		networkID     = "f2ce5cdfcb34238294c247a218b764347f78e55b0f61d00c6364df0ffe3a1de9"
		networkDriver = "baby"

		anErrorMessageWasReturned = `{"message":\s*".*"}`
	)

	var (
		mockController   *gomock.Controller
		service          *mocks_network.MockService
		nerdctlConfig    *config.Config
		logger           *mocks_logger.Logger
		handler          *handler
		responseRecorder *httptest.ResponseRecorder
	)

	parseableRequestBody := func(request types.NetworkCreateRequest) io.Reader {
		json, err := json.Marshal(request)
		Expect(err).ShouldNot(HaveOccurred(), "crafting request JSON")
		return bytes.NewReader(json)
	}

	simpleRequest := func(opts ...types.NetworkCreateOption) (io.Reader, types.NetworkCreateRequest) {
		request := *types.NewCreateNetworkRequest(networkName, opts...)
		return parseableRequestBody(request), request
	}

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())

		service = mocks_network.NewMockService(mockController)
		nerdctlConfig = &config.Config{}
		logger = mocks_logger.NewLogger(mockController)
		handler = newHandler(service, nerdctlConfig, logger)

		responseRecorder = httptest.NewRecorder()
	})

	When("a network request occurs for a new network", func() {
		It("should return a 201 Created and the network ID", func() {
			reader, expected := simpleRequest()
			request, err := http.NewRequest(http.MethodPost, path, reader)
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			serviceResponse := types.NetworkCreateResponse{ID: networkID}
			service.EXPECT().Create(gomock.Any(), expected).Return(serviceResponse, nil)
			logger.EXPECT().Debugf("Create network '%s'.", networkName)
			logger.EXPECT().Debugf("Network '%s' created.", networkName)
			handler.create(responseRecorder, request)

			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(responseRecorder.Body.String()).Should(MatchJSON(fmt.Sprintf(`{"Id": "%s"}`, networkID)))
		})
	})

	When("a create network request occurs for an already existing network", func() {
		It("should return a 201 Created, the network ID, and a warning that the network already exists", func() {
			reader, expected := simpleRequest()
			request, err := http.NewRequest(http.MethodPost, path, reader)
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			serviceResponse := types.NetworkCreateResponse{ID: networkID, Warning: "network already exists"}
			service.EXPECT().Create(gomock.Any(), expected).Return(serviceResponse, nil)
			logger.EXPECT().Debugf("Create network '%s'.", networkName)
			logger.EXPECT().Debugf("Network '%s' created.", networkName)
			handler.create(responseRecorder, request)

			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusCreated))
			Expect(responseRecorder.Body.String()).Should(MatchJSON(fmt.Sprintf(`{"Id": "%s", "Warning": "network already exists"}`, networkID)))
		})
	})

	When("a create network request occurs with a CNI plugin that is not supported", func() {
		It("should return a 404 Not Found and a message that the plugin was not found", func() {
			reader, expected := simpleRequest(types.WithDriver(networkDriver))
			request, err := http.NewRequest(http.MethodPost, path, reader)
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			pluginNotFoundWrapper := errdefs.NewNotFound(errors.New("unsupported cni plugin"))
			service.EXPECT().Create(gomock.Any(), expected).Return(types.NetworkCreateResponse{}, pluginNotFoundWrapper)
			logger.EXPECT().Debugf("Create network '%s'.", networkName)
			logger.EXPECT().Errorf("Create network '%s' failed for CNI plugin '%s' not supported.", networkName, networkDriver)
			handler.create(responseRecorder, request)

			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusNotFound))
			Expect(responseRecorder.Body.String()).Should(MatchRegexp(anErrorMessageWasReturned))
		})
	})

	When("an error occurs on network create", func() {
		It("should return a 500 Internal Server Error and a message that an error occurred", func() {
			reader, expected := simpleRequest()
			request, err := http.NewRequest(http.MethodPost, path, reader)
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			serviceErr := errors.New("internal server error")
			service.EXPECT().Create(gomock.Any(), expected).Return(types.NetworkCreateResponse{}, serviceErr)
			logger.EXPECT().Debugf("Create network '%s'.", networkName)
			logger.EXPECT().Errorf("Create network '%s' failed: %v.", networkName, serviceErr)
			handler.create(responseRecorder, request)

			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(responseRecorder.Body.String()).Should(MatchRegexp(anErrorMessageWasReturned))
		})
	})

	When("an error occurs on request read", func() {
		It("should return a 500 Internal Server Error and a message that an error occurred", func() {
			request, err := http.NewRequest(http.MethodPost, path, &errorOnRead{})
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).MinTimes(1)
			handler.create(responseRecorder, request)

			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(responseRecorder.Body.String()).Should(MatchRegexp(anErrorMessageWasReturned))
		})
	})

	When("a JSON parsing error occurs", func() {
		It("should return a 400 Bad Request and a message that an error occurred", func() {
			request, err := http.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(`{"Na}`)))
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).MinTimes(1)
			handler.create(responseRecorder, request)

			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusBadRequest))
			Expect(responseRecorder.Body.String()).Should(MatchRegexp(anErrorMessageWasReturned))
		})
	})

	When("an request occurs missing the required network name", func() {
		It("should return a 500 Internal Server Error and a message that an error occurred", func() {
			request, err := http.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(`{}`)))
			Expect(err).ShouldNot(HaveOccurred(), "crafting HTTP request")

			logger.EXPECT().Warn(gomock.Any())
			handler.create(responseRecorder, request)

			Expect(responseRecorder).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(responseRecorder.Body.String()).Should(MatchRegexp(anErrorMessageWasReturned))
		})
	})
})
