// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_builder"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Build API", func() {
	var (
		mockCtrl   *gomock.Controller
		logger     *mocks_logger.Logger
		service    *mocks_builder.MockService
		ncBuildSvc *mocks_backend.MockNerdctlBuilderSvc
		h          *handler
		rr         *httptest.ResponseRecorder
		stream     io.Writer
		req        *http.Request
		result     []types.BuildResult
		auxMsg     []*json.RawMessage
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_builder.NewMockService(mockCtrl)
		ncBuildSvc = mocks_backend.NewMockNerdctlBuilderSvc(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger, ncBuildSvc)
		rr = httptest.NewRecorder()
		stream = response.NewStreamWriter(rr)
		req, _ = http.NewRequest(http.MethodPost, "/build", nil)
		result = []types.BuildResult{
			{ID: "image1"},
			{ID: "image2"},
		}

		auxMsg = []*json.RawMessage{}
		for _, image := range result {
			auxData, err := json.Marshal(image)
			Expect(err).Should(BeNil())
			rawMsg := json.RawMessage(auxData)
			auxMsg = append(auxMsg, &rawMsg)
		}
	})
	Context("handler", func() {
		It("should return 200 as success response", func() {
			// service mock returns nil to mimic service built the image successfully.
			service.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()

			h.build(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))

			// expected stream output
			scanner := bufio.NewScanner(rr.Body)
			outputs := []response.StreamResponse{}
			for scanner.Scan() {
				var streamResp response.StreamResponse
				err := json.Unmarshal(scanner.Bytes(), &streamResp)
				Expect(err).Should(BeNil())
				outputs = append(outputs, streamResp)
			}
			Expect(len(outputs)).Should(Equal(len(result)))
			for i, aux := range auxMsg {
				Expect(outputs[i]).Should(Equal(response.StreamResponse{Aux: aux}))
			}
		})

		It("should return 500 error", func() {
			// service mock returns not found error to mimic image build failed
			service.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				nil, errdefs.NewNotFound(fmt.Errorf("some error")))
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			h.build(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "some error"}`))
		})
		It("should return error due to buildkit failure", func() {
			req = httptest.NewRequest(http.MethodPost, "/build", nil)
			logger.EXPECT().Warnf("Failed to get buildkit host: %v", gomock.Any())
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("", fmt.Errorf("some error")).AnyTimes()
			h.build(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "some error"}`))
		})
		It("should set the buildkit host", func() {
			req = httptest.NewRequest(http.MethodPost, "/build", nil)
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.BuildKitHost).Should(Equal("mocked-value"))
		})
		It("should fail to get build options due to buildkit error", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("", fmt.Errorf("some error"))
			logger.EXPECT().Warnf("Failed to get buildkit host: %v", gomock.Any())
			req = httptest.NewRequest(http.MethodPost, "/build", nil)

			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(Not(BeNil()))
			Expect(buildOption).Should(BeNil())
		})
		It("should set the tag query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?t=tag1&t=tag2", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Tag).Should(ContainElements("tag1", "tag2"))
		})
		It("should set the platform query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?platform=amd64/x86_64", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Platform).Should(ContainElements("amd64/x86_64"))
		})
		It("should set the dockerfile query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?dockerfile=mydockerfile", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.File).Should(Equal("mydockerfile"))
		})
		It("should set the rm query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?rm=false", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Rm).Should(BeFalse())
		})
		It("should set the rm query param to default if invalid value is provided", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?rm=WrongType", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Rm).Should(BeTrue())
		})
		It("should set all the default value for the query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Tag).Should(HaveLen(0))
			Expect(buildOption.Platform).Should(HaveLen(0))
			Expect(buildOption.File).Should(Equal("Dockerfile"))
			Expect(buildOption.Rm).Should(BeTrue())
		})
	})
})
