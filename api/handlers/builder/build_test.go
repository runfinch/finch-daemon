// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/containerd/nerdctl/v2/pkg/config"
	dockertypes "github.com/docker/cli/cli/config/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/runfinch/finch-daemon/api/auth"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_builder"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/credential"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// encodeRegistryConfig encodes auth configurations to a base64-encoded JSON string.
func encodeRegistryConfig(authConfigs map[string]dockertypes.AuthConfig) (string, error) {
	configJSON, err := json.Marshal(authConfigs)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(configJSON), nil
}

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
		// No need for credService in this test
		c := config.Config{}
		credCache := credential.NewCredentialCache()
		credService := credential.NewCredentialService(logger, credCache)
		h = newHandler(service, &c, logger, ncBuildSvc, credService)
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
			// service mock returns build results to mimic service built the image successfully.
			service.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx interface{}, options interface{}, reader interface{}, buildID string) ([]types.BuildResult, error) {
					return result, nil
				})
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
			service.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
				func(ctx interface{}, options interface{}, reader interface{}, buildID string) ([]types.BuildResult, error) {
					return nil, errdefs.NewNotFound(fmt.Errorf("some error"))
				})
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			h.build(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(rr.Body).Should(MatchJSON(`{"message": "some error"}`))
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("", fmt.Errorf("some error")).AnyTimes()
			h.build(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
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
		It("should set the CacheFrom query param as map", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?cachefrom={\"image1\":\"tag1\",\"image2\":\"tag2\"}", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.CacheFrom).Should(ContainElements("image1=tag1", "image2=tag2"))
		})

		It("should set the CacheFrom query param as array", func() {
			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()
			req = httptest.NewRequest(http.MethodPost, "/build?cachefrom=[\"image1:tag1\",\"image2:tag2\"]", nil)
			buildOption, err := h.getBuildOptions(rr, req, stream)
			Expect(err).Should(BeNil())
			Expect(buildOption.CacheFrom).Should(ContainElements("image1:tag1", "image2:tag2"))
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

		It("should store credentials uniquely and clean up after build is complete", func() {
			// Create a real credential cache and service
			credCache := credential.NewCredentialCache()
			credService := credential.NewCredentialService(logger, credCache)

			// Create build handler with the real credential service
			h := newHandler(service, &config.Config{}, logger, ncBuildSvc, credService)

			// Setup auth configs for the test
			authConfigs1 := map[string]dockertypes.AuthConfig{
				"registry1.example.com": {
					Username: "user1",
					Password: "pass1",
				},
				"registry2.example.com": {
					Username: "user2",
					Password: "pass2",
				},
			}

			authConfigs2 := map[string]dockertypes.AuthConfig{
				"registry1.example.com": {
					Username: "different-user1",
					Password: "different-pass1",
				},
				"registry2.example.com": {
					Username: "different-user2",
					Password: "different-pass2",
				},
			}

			buildId, err := credService.GenerateBuildID()
			Expect(err).Should(BeNil())

			err = credService.StoreAuthConfigs(context.TODO(), buildId, authConfigs2)
			Expect(err).Should(BeNil())

			registryAuthHeader1, err := encodeRegistryConfig(authConfigs1)
			Expect(err).Should(BeNil())

			registryAuthHeader2, err := encodeRegistryConfig(authConfigs2)
			Expect(err).Should(BeNil())

			Expect(len(credCache.Entries)).Should(Equal(1), "Credential cache should hold the entry made")

			// First build request
			req1 := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader("test-body-1"))
			req1.Header.Set(auth.RegistryConfigHeader, registryAuthHeader1)
			rr1 := httptest.NewRecorder()

			var capturedBuildID1 string
			service.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx interface{}, options interface{}, reader interface{}, buildID string) ([]types.BuildResult, error) {
					capturedBuildID1 = buildID
					Expect(buildID).ShouldNot(BeEmpty())

					auth1, err := credService.GetCredentials(context.Background(), buildID, "registry1.example.com")
					Expect(err).Should(BeNil())
					Expect(auth1.Username).Should(Equal("user1"))
					Expect(auth1.Password).Should(Equal("pass1"))

					auth2, err := credService.GetCredentials(context.Background(), buildID, "registry2.example.com")
					Expect(err).Should(BeNil())
					Expect(auth2.Username).Should(Equal("user2"))
					Expect(auth2.Password).Should(Equal("pass2"))

					Expect(len(credCache.Entries)).Should(Equal(2), "Credential cache should contain both creds")

					return result, nil
				})

			ncBuildSvc.EXPECT().GetBuildkitHost().Return("mocked-value", nil).AnyTimes()

			h.build(rr1, req1)
			Expect(rr1).Should(HaveHTTPStatus(http.StatusOK))
			Expect(len(credCache.Entries)).Should(Equal(1), "Credential cache should clear one cache entry after first build")

			// Second build request
			req2 := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader("test-body-2"))
			req2.Header.Set(auth.RegistryConfigHeader, registryAuthHeader2)
			rr2 := httptest.NewRecorder()

			var capturedBuildID2 string
			service.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx interface{}, options interface{}, reader interface{}, buildID string) ([]types.BuildResult, error) {
					capturedBuildID2 = buildID
					Expect(buildID).ShouldNot(BeEmpty())

					// Verify credentials from second request
					auth1, err := credService.GetCredentials(context.Background(), buildID, "registry1.example.com")
					Expect(err).Should(BeNil())
					Expect(auth1.Username).Should(Equal("different-user1"))
					Expect(auth1.Password).Should(Equal("different-pass1"))

					auth2, err := credService.GetCredentials(context.Background(), buildID, "registry2.example.com")
					Expect(err).Should(BeNil())
					Expect(auth2.Username).Should(Equal("different-user2"))
					Expect(auth2.Password).Should(Equal("different-pass2"))

					Expect(len(credCache.Entries)).Should(Equal(2), "Credential cache should contain both creds")

					return result, nil
				})

			h.build(rr2, req2)
			Expect(rr2).Should(HaveHTTPStatus(http.StatusOK))

			// Verify buildIDs are unique
			Expect(capturedBuildID1).ShouldNot(BeEmpty())
			Expect(capturedBuildID2).ShouldNot(BeEmpty())
			Expect(capturedBuildID1).ShouldNot(Equal(capturedBuildID2), "Build IDs should be unique")

			Expect(len(credCache.Entries)).Should(Equal(1), "Credential cache should have one entry left")

			credService.RemoveCredentials(buildId)
			Expect(len(credCache.Entries)).Should(Equal(0), "Credential cache should be clear now")
		})
	})
})
