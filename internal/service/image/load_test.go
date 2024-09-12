// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/image"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

// Unit tests related to image load API.
var _ = Describe("Image Load API", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		ncClient *mocks_backend.MockNerdctlImageSvc
		name     string
		inStream io.Reader
		service  image.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
		name = "/tmp"
		inStream = strings.NewReader("")
		service = NewService(cdClient, ncClient, logger)
	})
	Context("service", func() {
		It("should return no errors upon success", func() {
			ncClient.EXPECT().GetDataStore().
				Return(name, nil)
			ncClient.EXPECT().LoadImage(gomock.Any(), gomock.Any(), nil, gomock.Any()).
				Return(nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

			// service should return no error
			err := service.Load(ctx, inStream, nil, false)
			Expect(err).Should(BeNil())
		})
		It("should return an error if load image method returns an error", func() {
			ncClient.EXPECT().GetDataStore().
				Return(name, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			ncClient.EXPECT().LoadImage(gomock.Any(), gomock.Any(), nil, gomock.Any()).
				Return(errors.New("error message"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			// service should return an error
			err := service.Load(ctx, inStream, nil, false)
			Expect(err).ShouldNot(BeNil())
		})
		It("should return an error if get datastore method returns an error", func() {
			ncClient.EXPECT().GetDataStore().
				Return(name, errors.New("error message"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			// service should return an error
			err := service.Load(ctx, inStream, nil, false)
			Expect(err).ShouldNot(BeNil())
		})
		It("should return an error if import stream is nil", func() {
			// service should return an error
			err := service.Load(ctx, nil, nil, false)
			Expect(err).ShouldNot(BeNil())
		})
	})
})
