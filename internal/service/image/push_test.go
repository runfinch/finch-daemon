// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	"github.com/containerd/platforms"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/api/handlers/image"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to image push API.
var _ = Describe("Image Push API ", func() {
	var (
		ctx       context.Context
		mockCtrl  *gomock.Controller
		logger    *mocks_logger.Logger
		cdClient  *mocks_backend.MockContainerdClient
		ncClient  *mocks_backend.MockNerdctlImageSvc
		name      string
		tag       string
		digest    string
		domain    string
		rawRef    string
		pushRef   string
		authCfg   dockertypes.AuthConfig
		authCreds dockerconfigresolver.AuthCreds
		resolver  remotes.Resolver
		tracker   docker.StatusTracker
		service   image.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
		name = "public.ecr.aws/test-image/test-image"
		digest = "test-digest"
		tag = "test-tag"
		domain = "public.ecr.aws"
		rawRef = fmt.Sprintf("%s:%s", name, tag)
		pushRef = fmt.Sprintf("%s-tmp-reduced-platform", name)
		authCfg = dockertypes.AuthConfig{
			Username: "test-user",
			Password: "test-password",
		}
		authCreds = func(_ string) (string, string, error) {
			return authCfg.Username, authCfg.Password, nil
		}
		resolver = &mockResolver{}
		tracker = docker.NewInMemoryTracker()

		service = NewService(cdClient, ncClient, logger)

		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})
	Context("service", func() {
		It("should return no errors upon success", func() {
			pushImage := &images.Image{Name: pushRef}
			image := &mockImage{
				ImageName:   name,
				ImageDigest: digest,
				ImageTag:    tag,
				ImageSize:   256,
			}
			expected := types.PushResult{
				Tag:    tag,
				Digest: digest,
				Size:   256,
			}

			// expected backend calls
			cdClient.EXPECT().ParseDockerRef(rawRef).
				Return(name, domain, nil)
			cdClient.EXPECT().DefaultPlatformStrict().
				Return(nil)
			cdClient.EXPECT().ConvertImage(gomock.Any(), pushImage.Name, name, gomock.Any()).
				Return(pushImage, nil)
			cdClient.EXPECT().DeleteImage(gomock.Any(), pushImage.Name).
				Return(nil)
			expectGetAuthCreds(mockCtrl, domain, authCfg).
				Return(authCreds, nil)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).
				Return(resolver, tracker, nil)
			ncClient.EXPECT().PushImage(gomock.Any(), resolver, tracker, nil, pushImage.Name, name, nil).
				Return(nil)
			cdClient.EXPECT().GetImage(gomock.Any(), name).
				Return(image, nil)

			// service should return no error
			result, err := service.Push(ctx, name, tag, &authCfg, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(*result).Should(Equal(expected))
		})
		It("should return error due to malformed name", func() {
			// service should return error
			result, err := service.Push(ctx, "malformed:/image:name", "malformed:tag", &authCfg, nil)
			Expect(err).Should(HaveOccurred())
			Expect(result).Should(BeNil())
		})
		It("should return an error if image reference is invalid", func() {
			expectedError := fmt.Errorf("invalid image reference")

			// expected backend calls
			cdClient.EXPECT().ParseDockerRef(rawRef).
				Return("", "", fmt.Errorf("invalid image reference"))

			// service should return invalid reference error
			result, err := service.Push(ctx, name, tag, &authCfg, nil)
			Expect(err.Error()).Should(ContainSubstring(expectedError.Error()))
			Expect(result).Should(BeNil())
		})
		It("should return errors due to image conversion", func() {
			expectedError := fmt.Errorf("convert image failed")

			// expected backend calls
			cdClient.EXPECT().ParseDockerRef(rawRef).
				Return(name, domain, nil)
			cdClient.EXPECT().DefaultPlatformStrict().
				Return(nil)
			cdClient.EXPECT().ConvertImage(gomock.Any(), pushRef, name, gomock.Any()).
				Return(nil, expectedError)

			// service should return error
			result, err := service.Push(ctx, name, tag, &authCfg, nil)
			Expect(err.Error()).Should(ContainSubstring(expectedError.Error()))
			Expect(result).Should(BeNil())
		})
		It("should return a not found errors if image does not exist", func() {
			expectedError := fmt.Errorf("image `%s`: not found", name)

			// expected backend calls
			cdClient.EXPECT().ParseDockerRef(rawRef).
				Return(name, domain, nil)
			cdClient.EXPECT().DefaultPlatformStrict().
				Return(nil)
			cdClient.EXPECT().ConvertImage(gomock.Any(), pushRef, name, gomock.Any()).
				Return(nil, expectedError)

			// service should return error
			result, err := service.Push(ctx, name, tag, &authCfg, nil)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(result).Should(BeNil())
		})
		It("should return an error if credentials are invalid", func() {
			pushImage := &images.Image{Name: pushRef}
			expectedError := fmt.Errorf("invalid credentials")

			// expected backend calls
			cdClient.EXPECT().ParseDockerRef(rawRef).
				Return(name, domain, nil)
			cdClient.EXPECT().DefaultPlatformStrict().
				Return(nil)
			cdClient.EXPECT().ConvertImage(gomock.Any(), pushRef, name, gomock.Any()).
				Return(pushImage, nil)
			cdClient.EXPECT().DeleteImage(gomock.Any(), pushImage.Name).
				Return(nil)
			expectGetAuthCreds(mockCtrl, domain, authCfg).
				Return(nil, expectedError)

			// service should return error
			result, err := service.Push(ctx, name, tag, &authCfg, nil)
			Expect(err.Error()).Should(ContainSubstring(expectedError.Error()))
			Expect(result).Should(BeNil())
		})
		It("should fail due to resolver error", func() {
			pushImage := &images.Image{Name: pushRef}
			expectedError := fmt.Errorf("resolver error")

			// expected backend calls
			cdClient.EXPECT().ParseDockerRef(rawRef).
				Return(name, domain, nil)
			cdClient.EXPECT().DefaultPlatformStrict().
				Return(nil)
			cdClient.EXPECT().ConvertImage(gomock.Any(), pushRef, name, gomock.Any()).
				Return(pushImage, nil)
			cdClient.EXPECT().DeleteImage(gomock.Any(), pushImage.Name).
				Return(nil)
			expectGetAuthCreds(mockCtrl, domain, authCfg).
				Return(authCreds, nil)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).
				Return(nil, nil, expectedError)

			// service should return error
			result, err := service.Push(ctx, name, tag, &authCfg, nil)
			Expect(err.Error()).Should(ContainSubstring(expectedError.Error()))
			Expect(result).Should(BeNil())
		})
		It("should return an error upon service failure", func() {
			pushImage := &images.Image{Name: pushRef}
			expectedError := fmt.Errorf("failed to push image")

			// expected backend calls
			cdClient.EXPECT().ParseDockerRef(rawRef).
				Return(name, domain, nil)
			cdClient.EXPECT().DefaultPlatformStrict().
				Return(nil)
			cdClient.EXPECT().ConvertImage(gomock.Any(), pushRef, name, gomock.Any()).
				Return(pushImage, nil)
			cdClient.EXPECT().DeleteImage(gomock.Any(), pushImage.Name).
				Return(nil)
			expectGetAuthCreds(mockCtrl, domain, authCfg).
				Return(authCreds, nil)
			ncClient.EXPECT().GetDockerResolver(gomock.Any(), domain, gomock.Not(gomock.Nil())).
				Return(resolver, tracker, nil)
			ncClient.EXPECT().PushImage(gomock.Any(), resolver, tracker, nil, pushImage.Name, name, nil).
				Return(expectedError)

			// service should return error
			result, err := service.Push(ctx, name, tag, &authCfg, nil)
			Expect(err.Error()).Should(ContainSubstring(expectedError.Error()))
			Expect(result).Should(BeNil())
		})
	})

	// TODO: need to add an authenticated push unit test.
})

// dummy containerd image.
type mockImage struct {
	ImageName   string
	ImageDigest string
	ImageTag    string
	ImageSize   int64
}

func (m *mockImage) Name() string {
	return m.ImageName
}

func (m *mockImage) Target() ocispec.Descriptor {
	return ocispec.Descriptor{Digest: digest.Digest(m.ImageDigest), Size: m.ImageSize}
}

func (m *mockImage) Labels() map[string]string {
	return nil
}

func (m *mockImage) Unpack(context.Context, string, ...containerd.UnpackOpt) error {
	return nil
}

func (m *mockImage) RootFS(context.Context) ([]digest.Digest, error) {
	return nil, nil
}

func (m *mockImage) Size(context.Context) (int64, error) {
	return m.ImageSize, nil
}

func (m *mockImage) Usage(context.Context, ...containerd.UsageOpt) (int64, error) {
	return 0, nil
}

func (m *mockImage) Config(context.Context) (ocispec.Descriptor, error) {
	return ocispec.Descriptor{Digest: digest.Digest(m.ImageDigest), Size: m.ImageSize}, nil
}

func (m *mockImage) IsUnpacked(context.Context, string) (bool, error) {
	return false, nil
}

func (m *mockImage) ContentStore() content.Store {
	return nil
}

func (m *mockImage) Metadata() images.Image {
	return images.Image{
		Name:   m.ImageName,
		Target: ocispec.Descriptor{Digest: digest.Digest(m.ImageDigest), Size: m.ImageSize},
	}
}

func (m *mockImage) Platform() platforms.MatchComparer {
	return nil
}

func (m *mockImage) Spec(context.Context) (ocispec.Image, error) {
	return ocispec.Image{}, nil
}
