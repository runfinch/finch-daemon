// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"bytes"
	"context"
	"errors"

	"github.com/containerd/containerd/v2/core/images"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/runfinch/finch-daemon/api/handlers/image"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Image Export API", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		ncClient *mocks_backend.MockNerdctlImageSvc
		service  image.Service
		name     string
	)
	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
		name = "test-image"
		service = NewService(cdClient, ncClient, logger)
	})
	Context("service", func() {
		It("should return no errors upon success", func() {
			img := images.Image{Name: name}
			cdClient.EXPECT().SearchImage(gomock.Any(), name).
				Return([]images.Image{img}, nil)
			ncClient.EXPECT().ExportImage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil)

			var buf bytes.Buffer
			err := service.Export(ctx, name, nil, &buf)
			Expect(err).Should(BeNil())
		})
		It("should return NotFound error if image not found", func() {
			cdClient.EXPECT().SearchImage(gomock.Any(), name).
				Return([]images.Image{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			var buf bytes.Buffer
			err := service.Export(ctx, name, nil, &buf)
			Expect(err).ShouldNot(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return error if ExportImage fails", func() {
			img := images.Image{Name: name}
			cdClient.EXPECT().SearchImage(gomock.Any(), name).
				Return([]images.Image{img}, nil)
			ncClient.EXPECT().ExportImage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("export error"))

			var buf bytes.Buffer
			err := service.Export(ctx, name, nil, &buf)
			Expect(err).ShouldNot(BeNil())
		})
	})
})
