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
		ctx       context.Context
		mockCtrl  *gomock.Controller
		ncClient  *mocks_backend.MockNerdctlVolumeSvc
		logger    *mocks_logger.Logger
		name      string
		volume    native.Volume
		volumeMap map[string]native.Volume
		s         service
	)
	BeforeEach(func() {
		// initialize mocks
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		ncClient = mocks_backend.NewMockNerdctlVolumeSvc(mockCtrl)
		logger = mocks_logger.NewLogger(mockCtrl)
		name = "test-volume"
		volume = native.Volume{Name: name}
		s = service{
			nctlVolumeSvc: ncClient,
			logger:        logger,
		}
		volumeMap = make(map[string]native.Volume)
		volumeMap[name] = volume
	})
	Context("ListVolumes", func() {
		It("should return the volume(s) if it was found", func() {
			ncClient.EXPECT().ListVolumes(gomock.Any(), gomock.Any()).Return(
				volumeMap, nil)

			result, err := s.List(ctx, nil)
			Expect(err).Should(BeNil())
			Expect(result.Volumes).ShouldNot(BeEmpty())
			Expect(result.Volumes[0]).Should(Equal(volume))
		})
		It("should return an error if ListVolumes errors", func() {
			ncClient.EXPECT().ListVolumes(gomock.Any(), gomock.Any()).Return(
				nil, errors.New("fake"))
			logger.EXPECT().Errorf("failed to list volumes: %v", errors.New("fake"))

			result, err := s.List(ctx, nil)
			Expect(err).Should(Not(BeNil()))
			Expect(result).Should(BeNil())
		})
	})
})
