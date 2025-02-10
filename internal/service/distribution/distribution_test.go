// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package distribution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_remotes"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// TestImageHandler function is the entry point of image service package's unit test using ginkgo.
func TestDistributionService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Distribution APIs Service")
}

// Unit tests related to distribution inspect API.
var _ = Describe("Distribution Inspect API ", func() {
	Context("service", func() {
		var (
			ctx                     context.Context
			mockCtrl                *gomock.Controller
			logger                  *mocks_logger.Logger
			cdClient                *mocks_backend.MockContainerdClient
			ncClient                *mocks_backend.MockNerdctlImageSvc
			mockResolver            *mocks_remotes.MockResolver
			mockFetcher             *mocks_remotes.MockFetcher
			name                    string
			tag                     string
			imageRef                string
			domain                  string
			ociPlatformAmd          ocispec.Platform
			ociPlatformArm          ocispec.Platform
			authCfg                 dockertypes.AuthConfig
			authCreds               dockerconfigresolver.AuthCreds
			imageIndexDescriptor    ocispec.Descriptor
			imageDescriptor1        ocispec.Descriptor
			imageDescriptor2        ocispec.Descriptor
			imageIndex              ocispec.Index
			image                   ocispec.Image
			imageManifestDescriptor ocispec.Descriptor
			imageManifest           ocispec.Manifest
			imageManifestBytes      []byte
			imageBytes              []byte
			imageIndexBytes         []byte
			s                       service
		)
		BeforeEach(func() {
			ctx = context.Background()
			// initialize mocks
			mockCtrl = gomock.NewController(GinkgoT())
			logger = mocks_logger.NewLogger(mockCtrl)
			cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
			ncClient = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
			mockResolver = mocks_remotes.NewMockResolver(mockCtrl)
			mockFetcher = mocks_remotes.NewMockFetcher(mockCtrl)
			name = "public.ecr.aws/test-image/test-image"
			tag = "test-tag"
			imageRef = fmt.Sprintf("%s:%s", name, tag)
			domain = "public.ecr.aws"
			ociPlatformAmd = ocispec.Platform{
				Architecture: "amd64",
				OS:           "linux",
			}
			ociPlatformArm = ocispec.Platform{
				Architecture: "amd64",
				OS:           "linux",
			}
			authCfg = dockertypes.AuthConfig{
				Username: "test-user",
				Password: "test-password",
			}
			authCreds = func(_ string) (string, string, error) {
				return authCfg.Username, authCfg.Password, nil
			}

			imageIndexDescriptor = ocispec.Descriptor{
				MediaType:   ocispec.MediaTypeImageIndex,
				Digest:      "sha256:9bae60c369e612488c2a089c38737277a4823a3af97ec6866c3b4ad05251bfa5",
				Size:        2,
				URLs:        nil,
				Annotations: nil,
				Data:        nil,
				Platform:    nil,
			}

			imageDescriptor1 = ocispec.Descriptor{
				MediaType:   ocispec.MediaTypeImageManifest,
				Digest:      "sha256:deadbeef",
				Size:        2,
				URLs:        []string{},
				Annotations: map[string]string{},
				Data:        []byte{},
				Platform:    &ociPlatformAmd,
			}

			imageDescriptor2 = ocispec.Descriptor{
				MediaType:   ocispec.MediaTypeImageManifest,
				Digest:      "sha256:decafbad",
				Size:        2,
				URLs:        []string{},
				Annotations: map[string]string{},
				Data:        []byte{},
				Platform:    &ociPlatformArm,
			}

			imageIndex = ocispec.Index{
				Versioned: specs.Versioned{
					SchemaVersion: 1,
				},
				MediaType: ocispec.MediaTypeImageIndex,
				Manifests: []ocispec.Descriptor{
					imageDescriptor1,
					imageDescriptor2,
				},
			}
			b, err := json.Marshal(imageIndex)
			Expect(err).ShouldNot(HaveOccurred())
			imageIndexBytes = b

			imageManifestDescriptor = ocispec.Descriptor{
				MediaType:   ocispec.MediaTypeImageManifest,
				Digest:      "sha256:9b13590c9a50929020dc76a30ad813e42514a4e34de2f04f5a088f5a1320367c",
				Size:        2,
				URLs:        nil,
				Annotations: nil,
				Data:        nil,
			}

			imageManifest = ocispec.Manifest{
				MediaType: ocispec.MediaTypeImageManifest,
				Config: ocispec.Descriptor{
					MediaType:    ocispec.MediaTypeImageManifest,
					Digest:       "sha256:58cc9abebfec4b5ee95157d060207f7bc302516e6d84a0d83a560a1f7ed00e6e",
					Size:         2,
					URLs:         []string{},
					Annotations:  map[string]string{},
					Data:         []byte{},
					Platform:     &ocispec.Platform{},
					ArtifactType: "",
				},
			}
			b, err = json.Marshal(imageManifest)
			Expect(err).ShouldNot(HaveOccurred())
			imageManifestBytes = b

			image = ocispec.Image{
				Platform: ociPlatformAmd,
			}
			b, err = json.Marshal(image)
			Expect(err).ShouldNot(HaveOccurred())
			imageBytes = b

			s = service{
				client:       cdClient,
				nctlImageSvc: ncClient,
				logger:       logger,
			}
		})

		It("should return an error when canonicalization fails due to invalid input", func() {
			inspect, err := s.Inspect(ctx, "sdfsdfsdfsdf:invalid@digest", &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(errdefs.IsInvalidFormat(err)).Should(BeTrue())
			Expect(inspect).Should(BeNil())
		})

		It("should return an error when ParseDockerRef fails due to invalid input", func() {
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				"", "", fmt.Errorf("parsing failed"),
			)

			inspect, err := s.Inspect(ctx, imageRef, &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(errdefs.IsInvalidFormat(err)).Should(BeTrue())
			Expect(inspect).Should(BeNil())
		})

		It("should return an error when getAuthCredsFunc fails", func() {
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				nil, fmt.Errorf("invalid credentials"),
			)

			inspect, err := s.Inspect(ctx, imageRef, &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("invalid credentials"))
			Expect(inspect).Should(BeNil())
		})

		It("should return an error when getting the docker resolver fails", func() {
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				mockResolver, nil, fmt.Errorf("resolver error"),
			)

			inspect, err := s.Inspect(ctx, imageRef, &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("resolver error"))
			Expect(inspect).Should(BeNil())
		})

		It("should return an error when Resolving fails", func() {
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)

			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				mockResolver, nil, nil,
			)

			mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", ocispec.Descriptor{}, fmt.Errorf("failed to resolve"))

			inspect, err := s.Inspect(ctx, imageRef, &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to resolve"))
			Expect(inspect).Should(BeNil())
		})

		It("should return an error when creating Fetcher fails", func() {
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)

			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				mockResolver, nil, nil,
			)

			mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", imageIndexDescriptor, nil)
			mockResolver.EXPECT().Fetcher(gomock.Any(), imageRef).Return(nil, fmt.Errorf("failed to create fetcher"))

			inspect, err := s.Inspect(ctx, imageRef, &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to create fetcher"))
			Expect(inspect).Should(BeNil())
		})

		It("should return an error when Fetcher fails to Fetch", func() {
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)

			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				mockResolver, nil, nil,
			)

			mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", imageIndexDescriptor, nil)
			mockResolver.EXPECT().Fetcher(gomock.Any(), imageRef).Return(mockFetcher, nil)
			mockFetcher.EXPECT().Fetch(gomock.Any(), imageIndexDescriptor).Return(nil, fmt.Errorf("fetcher failed to fetch"))

			inspect, err := s.Inspect(ctx, imageRef, &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("fetcher failed to fetch"))
			Expect(inspect).Should(BeNil())
		})

		It("should return an error when reading manifest fails", func() {
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)

			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				mockResolver, nil, nil,
			)

			mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", imageIndexDescriptor, nil)
			mockResolver.EXPECT().Fetcher(gomock.Any(), imageRef).Return(mockFetcher, nil)
			imageIndexRc := io.NopCloser(&mockReader{
				err: fmt.Errorf("failed to read"),
			})
			mockFetcher.EXPECT().Fetch(gomock.Any(), imageIndexDescriptor).Return(imageIndexRc, nil)

			inspect, err := s.Inspect(ctx, imageRef, &authCfg)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to read"))
			Expect(inspect).Should(BeNil())
		})

		When("Image index", func() {
			It("should return expected response upon success", func() {
				cdClient.EXPECT().ParseDockerRef(imageRef).Return(
					imageRef, domain, nil,
				)
				expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
					authCreds, nil,
				)

				ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
					mockResolver, nil, nil,
				)

				mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", imageIndexDescriptor, nil)
				mockResolver.EXPECT().Fetcher(gomock.Any(), imageRef).Return(mockFetcher, nil)
				imageIndexRc := io.NopCloser(strings.NewReader(string(imageIndexBytes)))
				mockFetcher.EXPECT().Fetch(gomock.Any(), imageIndexDescriptor).Return(imageIndexRc, nil)

				inspectRes, err := s.Inspect(ctx, imageRef, &authCfg)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(inspectRes).ShouldNot(BeNil())
				Expect(inspectRes.Descriptor).Should(Equal(imageIndexDescriptor))
				Expect(inspectRes.Platforms).Should(HaveLen(2))
				Expect(inspectRes.Platforms).Should(ContainElements(ociPlatformAmd, ociPlatformArm))
			})
		})

		When("Image", func() {
			It("should return expected response upon success", func() {
				cdClient.EXPECT().ParseDockerRef(imageRef).Return(
					imageRef, domain, nil,
				)
				expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
					authCreds, nil,
				)

				ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
					mockResolver, nil, nil,
				)

				mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", imageManifestDescriptor, nil)
				mockResolver.EXPECT().Fetcher(gomock.Any(), imageRef).Return(mockFetcher, nil)
				imageManifestRc := io.NopCloser(strings.NewReader(string(imageManifestBytes)))
				mockFetcher.EXPECT().Fetch(gomock.Any(), imageManifestDescriptor).Return(imageManifestRc, nil)

				imageRc := io.NopCloser(strings.NewReader(string(imageBytes)))
				// gomock.Any() used for second argument because comparing maps compares addresses, which
				// will never be equal due to (un)marshalling
				mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(imageRc, nil)

				inspectRes, err := s.Inspect(ctx, imageRef, &authCfg)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(inspectRes).ShouldNot(BeNil())
				Expect(inspectRes.Descriptor).Should(Equal(imageManifestDescriptor))
				Expect(inspectRes.Platforms).Should(HaveLen(1))
				Expect(inspectRes.Platforms).Should(ContainElement(ociPlatformAmd))
			})

			It("should return an error when image Fetcher fails to Fetch", func() {
				cdClient.EXPECT().ParseDockerRef(imageRef).Return(
					imageRef, domain, nil,
				)
				expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
					authCreds, nil,
				)

				ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
					mockResolver, nil, nil,
				)

				mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", imageManifestDescriptor, nil)
				mockResolver.EXPECT().Fetcher(gomock.Any(), imageRef).Return(mockFetcher, nil)
				imageManifestRc := io.NopCloser(strings.NewReader(string(imageManifestBytes)))
				mockFetcher.EXPECT().Fetch(gomock.Any(), imageManifestDescriptor).Return(imageManifestRc, nil)
				mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("image fetcher failed to fetch"))

				inspect, err := s.Inspect(ctx, imageRef, &authCfg)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("image fetcher failed to fetch"))
				Expect(inspect).Should(BeNil())
			})

			It("should return an error when reading image fails", func() {
				cdClient.EXPECT().ParseDockerRef(imageRef).Return(
					imageRef, domain, nil,
				)
				expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
					authCreds, nil,
				)

				ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
					mockResolver, nil, nil,
				)

				mockResolver.EXPECT().Resolve(gomock.Any(), imageRef).Return("", imageManifestDescriptor, nil)
				mockResolver.EXPECT().Fetcher(gomock.Any(), imageRef).Return(mockFetcher, nil)
				imageManifestRc := io.NopCloser(strings.NewReader(string(imageManifestBytes)))
				mockFetcher.EXPECT().Fetch(gomock.Any(), imageManifestDescriptor).Return(imageManifestRc, nil)

				imageRc := io.NopCloser(&mockReader{
					err: fmt.Errorf("failed to read image"),
				})
				mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(imageRc, nil)

				inspect, err := s.Inspect(ctx, imageRef, &authCfg)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to read image"))
				Expect(inspect).Should(BeNil())
			})
		})
	})
})

// expectGetAuthCreds creates a new mocked object for getAuthCreds function
// with expected input parameters.
func expectGetAuthCreds(ctrl *gomock.Controller, refDomain string, ac dockertypes.AuthConfig) *mockGetAuthCreds {
	return &mockGetAuthCreds{
		expectedDomain: refDomain,
		expectedAuth:   ac,
		ctrl:           ctrl,
	}
}

type mockGetAuthCreds struct {
	expectedDomain string
	expectedAuth   dockertypes.AuthConfig
	ctrl           *gomock.Controller
}

// Return mocks getAuthCreds function with expected input parameters and returns the passed output values.
func (m *mockGetAuthCreds) Return(creds dockerconfigresolver.AuthCreds, err error) {
	m.ctrl.RecordCall(m, "GetAuthCreds", m.expectedDomain, m.expectedAuth)
	getAuthCredsFunc = func(domain string, _ backend.ContainerdClient, ac dockertypes.AuthConfig) (dockerconfigresolver.AuthCreds, error) {
		m.GetAuthCreds(domain, ac)
		return creds, err
	}
}

func (m *mockGetAuthCreds) GetAuthCreds(domain string, ac dockertypes.AuthConfig) {
	m.ctrl.Call(m, "GetAuthCreds", domain, ac)
}

type mockReader struct {
	err error
}

func (m mockReader) Read([]byte) (int, error) {
	return 0, m.err
}
