// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"fmt"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/system"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/version"
)

// Unit tests related to version API.
var _ = Describe("Version API ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		ncClient *mocks_backend.MockNerdctlSystemSvc
		service  system.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlSystemSvc(mockCtrl)
		cdClient := mocks_backend.NewMockContainerdClient(mockCtrl)
		service = NewService(cdClient, ncClient, logger)
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})
	Context("service", func() {
		It("should return version info", func() {
			// set up the mock to mimic return containerd component version from nerdctl function
			cntdComponent := dockercompat.ComponentVersion{
				Name:    "containerd",
				Version: "v1.7.1",
				Details: map[string]string{
					"GitCommit": "1677a17964311325ed1c31e2c0a3589ce6d5c30d",
				},
			}
			serverVersion := dockercompat.ServerVersion{
				Components: []dockercompat.ComponentVersion{
					cntdComponent,
				},
			}
			ncClient.EXPECT().GetServerVersion(ctx).Return(&serverVersion, nil)
			// service should not return any error
			vInfo, err := service.GetVersion(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vInfo).ShouldNot(BeNil())
			Expect(vInfo.Version).Should(Equal(version.Version))
			Expect(vInfo.GitCommit).Should(Equal(version.GitCommit))
			Expect(vInfo.ApiVersion).Should(Equal(version.DefaultApiVersion))
			Expect(vInfo.Platform.Name).ShouldNot(BeEmpty())
			Expect(vInfo.Os).ShouldNot(BeEmpty())
			Expect(vInfo.Arch).ShouldNot(BeEmpty())
			Expect(vInfo.KernelVersion).ShouldNot(BeEmpty())
			Expect(vInfo.Components).ShouldNot(BeEmpty())
			Expect(vInfo.Components[0]).Should(Equal(types.ComponentVersion{
				Name:    "containerd",
				Version: "v1.7.1",
				Details: map[string]string{
					"GitCommit": "1677a17964311325ed1c31e2c0a3589ce6d5c30d",
				},
			}))
		})
		It("should return error", func() {
			// set up the mock to mimic return error from nerdctl function
			expectedErr := fmt.Errorf("some error")
			ncClient.EXPECT().GetServerVersion(gomock.Any()).Return(nil, expectedErr)
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

			vInfo, err := service.GetVersion(ctx)
			Expect(vInfo).Should(BeNil())
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(MatchError(expectedErr))
		})
	})
})
