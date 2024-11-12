// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/v2/core/remotes"
	"github.com/containerd/containerd/v2/core/remotes/docker"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to image pull API.
var _ = Describe("Image Pull API ", func() {
	Context("service", func() {
		var (
			ctx         context.Context
			mockCtrl    *gomock.Controller
			logger      *mocks_logger.Logger
			cdClient    *mocks_backend.MockContainerdClient
			ncClient    *mocks_backend.MockNerdctlImageSvc
			name        string
			tag         string
			platform    string
			imageRef    string
			domain      string
			ociPlatform ocispec.Platform
			authCfg     dockertypes.AuthConfig
			authCreds   dockerconfigresolver.AuthCreds
			resolver    remotes.Resolver
			s           service
		)
		BeforeEach(func() {
			ctx = context.Background()
			// initialize mocks
			mockCtrl = gomock.NewController(GinkgoT())
			logger = mocks_logger.NewLogger(mockCtrl)
			cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
			ncClient = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
			name = "public.ecr.aws/test-image/test-image"
			tag = "test-tag"
			imageRef = fmt.Sprintf("%s:%s", name, tag)
			domain = "public.ecr.aws"
			platform = "linux/amd64"
			ociPlatform = ocispec.Platform{
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
			resolver = &mockResolver{}

			s = service{
				client:       cdClient,
				nctlImageSvc: ncClient,
				logger:       logger,
			}
		})

		It("should return no errors upon success", func() {
			// expected backend calls
			cdClient.EXPECT().ParsePlatform(platform).Return(
				ociPlatform, nil,
			)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				resolver, nil, nil,
			)
			ncClient.EXPECT().PullImage(gomock.Any(), nil, nil, resolver, imageRef, []ocispec.Platform{ociPlatform}).Return(
				nil, nil,
			)

			// service should return no error
			err := s.Pull(ctx, name, tag, platform, &authCfg, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return no errors when reference spec includes a digest spec", func() {
			tag := "sha256:7ea94d4e7f346a9328a9ff053ab149e3c99c1737f8d251094e7cc38664c3d4b9"
			imageRef := fmt.Sprintf("%s@%s", name, tag)

			// expected backend calls
			cdClient.EXPECT().ParsePlatform(platform).Return(
				ociPlatform, nil,
			)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				resolver, nil, nil,
			)
			ncClient.EXPECT().PullImage(gomock.Any(), nil, nil, resolver, imageRef, []ocispec.Platform{ociPlatform}).Return(
				nil, nil,
			)

			// service should return no error
			err := s.Pull(ctx, name, tag, platform, &authCfg, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should use default platform if not specified", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				authCreds, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).Return(
				resolver, nil, nil,
			)
			ncClient.EXPECT().PullImage(gomock.Any(), nil, nil, resolver, imageRef, []ocispec.Platform{ociPlatform}).Return(
				nil, nil,
			)

			// service should return no error
			err := s.Pull(ctx, name, tag, "", &authCfg, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should succeed without authentication", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Nil()).Return(
				resolver, nil, nil,
			)
			ncClient.EXPECT().PullImage(gomock.Any(), nil, nil, resolver, imageRef, []ocispec.Platform{ociPlatform}).Return(
				nil, nil,
			)

			// service should return no error
			err := s.Pull(ctx, name, tag, "", nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error if platform is invalid", func() {
			// expected backend calls
			cdClient.EXPECT().ParsePlatform(platform).Return(
				ocispec.Platform{}, fmt.Errorf("invalid platform"),
			)

			// service should return invalid platform error
			err := s.Pull(ctx, name, tag, platform, nil, nil)
			Expect(err).Should(HaveOccurred())
		})
		It("should return an error if image reference is invalid", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				"", "", fmt.Errorf("invalid image reference"),
			)

			// service should return invalid reference error
			err := s.Pull(ctx, name, tag, "", nil, nil)
			Expect(err).Should(HaveOccurred())
		})
		It("should return an error if credentials are invalid", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			expectGetAuthCreds(mockCtrl, domain, authCfg).Return(
				nil, fmt.Errorf("invalid credentials"),
			)

			// service should return invalid credentials error
			err := s.Pull(ctx, name, tag, "", &authCfg, nil)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("invalid credentials"))
		})
		It("should fail due to resolver error", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Nil()).Return(
				nil, nil, fmt.Errorf("resolver error"),
			)

			// service should return resolver error
			err := s.Pull(ctx, name, tag, "", nil, nil)
			Expect(err).Should(HaveOccurred())
		})
		It("should return an error upon service failure", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Nil()).Return(
				resolver, nil, nil,
			)
			ncClient.EXPECT().PullImage(gomock.Any(), nil, nil, resolver, imageRef, []ocispec.Platform{ociPlatform}).Return(
				nil, fmt.Errorf("service error"),
			)

			// service should return service error
			err := s.Pull(ctx, name, tag, "", nil, nil)
			Expect(err).Should(HaveOccurred())
		})
		It("should return a not found error if authorization failed", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Nil()).Return(
				resolver, nil, nil,
			)
			ncClient.EXPECT().PullImage(gomock.Any(), nil, nil, resolver, imageRef, []ocispec.Platform{ociPlatform}).Return(
				nil, docker.ErrInvalidAuthorization,
			)

			// service should return not found error
			err := s.Pull(ctx, name, tag, "", nil, nil)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return a not found error if image cannot be resolved", func() {
			// expected backend calls
			cdClient.EXPECT().DefaultPlatformSpec().Return(ociPlatform)
			cdClient.EXPECT().ParseDockerRef(imageRef).Return(
				imageRef, domain, nil,
			)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Nil()).Return(
				resolver, nil, nil,
			)
			ncClient.EXPECT().PullImage(gomock.Any(), nil, nil, resolver, imageRef, []ocispec.Platform{ociPlatform}).Return(
				nil, cerrdefs.ErrNotFound,
			)

			// service should return not found error
			err := s.Pull(ctx, name, tag, "", nil, nil)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
	})
})
