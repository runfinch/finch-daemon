// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"
	"fmt"

	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/volume"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Remove volume API", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		ncClient *mocks_backend.MockNerdctlVolumeSvc
		name     string
		s        volume.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		ncClient = mocks_backend.NewMockNerdctlVolumeSvc(mockCtrl)
		logger := mocks_logger.NewLogger(mockCtrl)
		name = "test-volume"
		s = NewService(ncClient, logger)
	})
	Context("service", func() {
		It("should remove the volume successfully", func() {
			ncClient.EXPECT().RemoveVolume(ctx, name, false /* force */, gomock.Any()).Return(nil)
			err := s.Remove(ctx, name, false)
			Expect(err).Should(BeNil())
		})
		It("should return not found error", func() {
			// mock mimics not found error occurred in the nerdctl client
			ncClient.EXPECT().RemoveVolume(ctx, name, false /* force */, gomock.Any()).
				Return(fmt.Errorf("not found"))
			err := s.Remove(ctx, name, false)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return in conflict error", func() {
			// mock mimics not found error occurred in the nerdctl client
			ncClient.EXPECT().RemoveVolume(ctx, name, false /* force */, gomock.Any()).
				Return(fmt.Errorf("volume %q is in use", name))
			err := s.Remove(ctx, name, false)
			Expect(errdefs.IsConflict(err)).Should(BeTrue())
		})
		It("should return generic error", func() {
			// mock mimics not found error occurred in the nerdctl client
			ncClient.EXPECT().RemoveVolume(ctx, name, false /* force */, gomock.Any()).
				Return(fmt.Errorf("some error"))
			err := s.Remove(ctx, name, false)
			Expect(errdefs.IsConflict(err)).Should(BeFalse())
			Expect(errdefs.IsNotFound(err)).Should(BeFalse())
		})
	})
})
