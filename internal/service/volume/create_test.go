// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"
	"errors"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

var _ = Describe("Volumes API service common ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		ncClient *mocks_backend.MockNerdctlVolumeSvc
		name     string
		volume   native.Volume
		s        service
	)
	BeforeEach(func() {
		// initialize mocks
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlVolumeSvc(mockCtrl)
		name = "test-volume"
		volume = native.Volume{Name: name}
		s = service{
			nctlVolumeSvc: ncClient,
			logger:        logger,
		}
	})
	Context("Create Volume", func() {
		It("case where volume fails on creation", func() {
			ncClient.EXPECT().CreateVolume(gomock.Any(), gomock.Any()).Return(nil, errors.New("fail"))
			logger.EXPECT().Errorf("failed to create volume: %v", gomock.Any())
			result, err := s.Create(ctx, "test", []string{})
			Expect(result).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
		It("success case where volume was created", func() {
			ncClient.EXPECT().CreateVolume(gomock.Any(), gomock.Any()).Return(&volume, nil)
			result, err := s.Create(ctx, name, []string{})
			Expect(result).Should(Equal(&volume))
			Expect(err).Should(BeNil())
		})
	})
})
