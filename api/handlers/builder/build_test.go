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
	"go.uber.org/mock/gomock"
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
		It("should set the q query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?q=false", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Quiet).Should(BeFalse())
		})
		It("should set the nocache query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?nocache=true", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.NoCache).Should(BeTrue())
		})
		It("should set the CacheFrom query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?cachefrom={\"image1\":\"tag1\",\"image2\":\"tag2\"}", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.CacheFrom).Should(ContainElements("image1=tag1", "image2=tag2"))
		})

		It("should set the BuildArgs query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?buildargs={\"ARG1\":\"value1\",\"ARG2\":\"value2\"}", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.BuildArgs).Should(ContainElements("ARG1=value1", "ARG2=value2"))
		})

		It("should set the Label query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?labels={\"LABEL1\":\"value1\",\"LABEL2\":\"value2\"}", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Label).Should(ContainElements("LABEL1=value1", "LABEL2=value2"))
		})

		It("should set the NetworkMode query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?networkmode=host", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.NetworkMode).Should(Equal("host"))
		})

		It("should set the Output query param", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?output=type=docker", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.Output).Should(Equal("type=docker"))
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
			Expect(buildOption.Quiet).Should(BeTrue())
			Expect(buildOption.NoCache).Should(BeFalse())
			Expect(buildOption.CacheFrom).Should(BeEmpty())
			Expect(buildOption.BuildArgs).Should(BeEmpty())
			Expect(buildOption.Label).Should(BeEmpty())
			Expect(buildOption.NetworkMode).Should(BeEmpty())
			Expect(buildOption.Output).Should(BeEmpty())
		})
	})
})
