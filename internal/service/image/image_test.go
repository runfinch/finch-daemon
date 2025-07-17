// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"errors"
	"testing"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/remotes"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	dockertypes "github.com/docker/cli/cli/config/types"
	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/internal/backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// TestImageHandler function is the entry point of image service package's unit test using ginkgo.
func TestImageService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Image APIs Service")
}

var _ = Describe("Image API service common ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		ncClient *mocks_backend.MockNerdctlImageSvc
		name     string
		name2    string
		digest1  digest.Digest
		digest2  digest.Digest
		img      images.Image
		img2     images.Image
		img3     images.Image
		s        service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
		name = "test-image"
		name2 = "test-image-2"
		digest1 = digest.NewDigestFromBytes(digest.SHA256, []byte("123abc"))
		digest2 = digest.NewDigestFromBytes(digest.SHA256, []byte("abc123"))
		img = images.Image{
			Name: name,
			Target: ocispec.Descriptor{
				Digest: digest1,
			},
		}
		img2 = images.Image{
			Name: name2,
			Target: ocispec.Descriptor{
				Digest: digest1,
			},
		}
		img3 = images.Image{
			Name: name2,
			Target: ocispec.Descriptor{
				Digest: digest2,
			},
		}
		s = service{
			client:       cdClient,
			nctlImageSvc: ncClient,
			logger:       logger,
		}
	})
	Context("getImage", func() {
		It("should return the containerd image if it was found", func() {
			// search method returns one image
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{img}, nil)

			result, err := s.getImage(ctx, name)
			Expect(*result).Should(Equal(img))
			Expect(err).Should(BeNil())
		})
		It("should return an error if search image method fails", func() {
			// search method returns an error
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				nil, errors.New("search image error"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			result, err := s.getImage(ctx, name)
			Expect(result).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
		It("should return NotFound error if no image was found", func() {
			// search method returns no image
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			result, err := s.getImage(ctx, name)
			Expect(result).Should(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return the first image if multiple images with the same digest were found", func() {
			// search method returns two images
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{img, img2}, nil)

			result, err := s.getImage(ctx, name)
			Expect(err).Should(BeNil())
			Expect(*result).Should(Equal(img))
		})
		It("should return an error if multiple images with different digests were found", func() {
			// search method returns two images with different digests
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{img, img3}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			result, err := s.getImage(ctx, name)
			Expect(err).ShouldNot(BeNil())
			Expect(result).Should(BeNil())
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

// dummy remotes resolver.
type mockResolver struct{}

func (m *mockResolver) Resolve(context.Context, string) (string, ocispec.Descriptor, error) {
	return "", ocispec.Descriptor{}, nil
}

func (m *mockResolver) Fetcher(context.Context, string) (remotes.Fetcher, error) {
	return nil, nil
}

func (m *mockResolver) Pusher(context.Context, string) (remotes.Pusher, error) {
	return nil, nil
}
