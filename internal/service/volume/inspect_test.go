// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"fmt"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/volume"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Inspect volume API", func() {
	var (
		mockCtrl *gomock.Controller
		ncClient *mocks_backend.MockNerdctlVolumeSvc
		name     string
		s        volume.Service
	)
	BeforeEach(func() {
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		ncClient = mocks_backend.NewMockNerdctlVolumeSvc(mockCtrl)
		logger := mocks_logger.NewLogger(mockCtrl)
		name = "test-volume"
		s = NewService(ncClient, logger)
	})
	Context("service", func() {
		It("should return the volume details", func() {
			expectedVol := native.Volume{
				Name:       name,
				Mountpoint: "/path/to/test-volume",
				Labels:     nil,
				Size:       100,
			}
			ncClient.EXPECT().GetVolume(name).Return(&expectedVol, nil)
			vol, err := s.Inspect(name)
			Expect(err).Should(BeNil())
			Expect(*vol).Should(Equal(expectedVol))
		})
		It("should return not found error", func() {
			// mock mimics not found error occurred in the nerdctl client
			ncClient.EXPECT().GetVolume(name).Return(nil, fmt.Errorf("not found"))
			vol, err := s.Inspect(name)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(vol).Should(BeNil())
		})
		It("should return generic error", func() {
			// mock mimics error while retrieving volume details
			ncClient.EXPECT().GetVolume(name).Return(nil, fmt.Errorf("some error"))
			vol, err := s.Inspect(name)
			Expect(errdefs.IsNotFound(err)).Should(BeFalse())
			Expect(vol).Should(BeNil())
		})
	})
})
