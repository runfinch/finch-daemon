// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"errors"

	"github.com/containerd/containerd/images"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/image"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to image inspect API.
var _ = Describe("Image Inspect API ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		ncClient *mocks_backend.MockNerdctlImageSvc
		name     string
		img      images.Image
		inspect  dockercompat.Image
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
		img = images.Image{Name: name}
		inspect = dockercompat.Image{
			ID:          name,
			RepoTags:    []string{"test-image:latest"},
			RepoDigests: []string{"test-image@test-digest"},
			Size:        100,
		}

		service = NewService(cdClient, ncClient, logger)
	})
	Context("service", func() {
		It("should return the inspect object upon success", func() {
			// search image method returns one image
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{img}, nil)

			ncClient.EXPECT().InspectImage(gomock.Any(), img).Return(
				&inspect, nil)

			// service should return inspect object
			result, err := service.Inspect(ctx, name)
			Expect(*result).Should(Equal(inspect))
			Expect(err).Should(BeNil())
		})
		It("should return NotFound error if image was not found", func() {
			// search image method returns no image
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// service should return a NotFound error
			result, err := service.Inspect(ctx, name)
			Expect(result).Should(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should succeed if multiple images were found for the given Id", func() {
			// search image method returns multiple images
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{img, img}, nil)

			ncClient.EXPECT().InspectImage(gomock.Any(), img).Return(
				&inspect, nil)

			// service should return an error
			result, err := service.Inspect(ctx, name)
			Expect(err).Should(BeNil())
			Expect(*result).Should(Equal(inspect))
		})
		It("should return an error if search image method failed", func() {
			// search image method returns no image
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				nil, errors.New("error message"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			// service should return an error
			result, err := service.Inspect(ctx, name)
			Expect(result).Should(BeNil())
			Expect(err).ShouldNot(BeNil())
		})
		It("should return an error if the backend inspect method failed", func() {
			// search image method returns one image
			cdClient.EXPECT().SearchImage(gomock.Any(), name).Return(
				[]images.Image{img}, nil)

			ncClient.EXPECT().InspectImage(gomock.Any(), img).Return(
				nil, errors.New("error message"))

			// service should return an error
			result, err := service.Inspect(ctx, name)
			Expect(result).Should(BeNil())
			Expect(err).ShouldNot(BeNil())
		})
	})
})
