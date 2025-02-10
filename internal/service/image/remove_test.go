// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/api/handlers/image"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to image remove API.
var _ = Describe("Image Remove API", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		ncClient *mocks_backend.MockNerdctlImageSvc
		name     string
		img      images.Image
		service  image.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
		name = "test-image"
		img = images.Image{
			Name: name,
			Target: ocispec.Descriptor{
				Digest: "test-digest",
			},
		}
		service = NewService(cdClient, ncClient, logger)
	})
	Context("service", func() {
		It("should successfully remove the image", func() {
			// search image method returns one image
			ncClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				1, 1, []*images.Image{&img}, nil)

			cdClient.EXPECT().GetUsedImages(gomock.Any()).Return(
				make(map[string]string),
				make(map[string]string),
				nil)
			cdClient.EXPECT().DeleteImage(gomock.Any(), gomock.Any()).Return(nil)
			cdClient.EXPECT().GetImageDigests(gomock.Any(), gomock.Any()).Return([]digest.Digest{"test-digest"}, nil)

			// service should return inspect object
			untagged, deleted, err := service.Remove(ctx, name, false)
			Expect(err).Should(BeNil())
			Expect(untagged).Should(HaveLen(1))
			Expect(untagged).Should(ContainElement("test-image:test-digest"))
			Expect(deleted).Should(HaveLen(1))
			Expect(deleted).Should(ContainElement("test-digest"))
		})
		It("should return NotFound error if image was not found", func() {
			// search image method returns no image
			ncClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				0, 0, []*images.Image{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// service should return a NotFound error
			untagged, deleted, err := service.Remove(ctx, name, false)
			Expect(untagged).Should(HaveLen(0))
			Expect(deleted).Should(HaveLen(0))
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return an error if multiple images were found for the given Id", func() {
			// search image method returns multiple images
			ncClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				2, 1, []*images.Image{&img, &img}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// service should return an error
			untagged, deleted, err := service.Remove(ctx, name, false)
			Expect(untagged).Should(HaveLen(0))
			Expect(deleted).Should(HaveLen(0))
			Expect(err).Should(HaveOccurred())
		})
		It("should return an error image is being used by a running container", func() {
			// search image method returns one image
			ncClient.EXPECT().SearchImage(gomock.Any(), name).Return(1, 1, []*images.Image{&img}, nil)

			cdClient.EXPECT().GetUsedImages(gomock.Any()).Return(
				make(map[string]string),
				map[string]string{"test-image": "test-running-container"},
				nil)

			// service should return inspect object
			untagged, deleted, err := service.Remove(ctx, name, false)
			Expect(err).Should(Not(BeNil()))
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
			Expect(untagged).Should(HaveLen(0))
			Expect(deleted).Should(HaveLen(0))
		})
		It("should return an error image is being used by a stopped container", func() {
			// search image method returns one image
			ncClient.EXPECT().SearchImage(gomock.Any(), name).Return(1, 1, []*images.Image{&img}, nil)

			cdClient.EXPECT().GetUsedImages(gomock.Any()).Return(
				map[string]string{"test-image": "test-stopped-container"},
				make(map[string]string),
				nil)

			// service should return inspect object
			untagged, deleted, err := service.Remove(ctx, name, false)
			Expect(err).Should(HaveOccurred())
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
			Expect(untagged).Should(HaveLen(0))
			Expect(deleted).Should(HaveLen(0))
		})
		It("should successfully remove the image used by stopped container with force flag", func() {
			// search image method returns one image
			ncClient.EXPECT().SearchImage(gomock.Any(), name).Return(1, 1, []*images.Image{&img}, nil)

			cdClient.EXPECT().GetUsedImages(gomock.Any()).Return(
				map[string]string{"test-image": "test-stopped-container"},
				make(map[string]string),
				nil)
			cdClient.EXPECT().DeleteImage(gomock.Any(), gomock.Any()).Return(nil)
			cdClient.EXPECT().GetImageDigests(gomock.Any(), gomock.Any()).Return([]digest.Digest{"test-digest"}, nil)

			// service should return inspect object
			untagged, deleted, err := service.Remove(ctx, name, true)
			Expect(err).Should(BeNil())
			Expect(untagged).Should(HaveLen(1))
			Expect(untagged).Should(ContainElement("test-image:test-digest"))
			Expect(deleted).Should(HaveLen(1))
			Expect(deleted).Should(ContainElement("test-digest"))
		})
	})
})
