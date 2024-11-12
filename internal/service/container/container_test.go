// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"testing"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

type mockNerdctlService struct {
	*mocks_backend.MockNerdctlContainerSvc
	*mocks_backend.MockNerdctlNetworkSvc
}

// TestContainerService is the entry point of container service package's unit tests using ginkgo.
func TestContainerService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Container APIs Service")
}

var _ = Describe("Container API service common ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		ncClient *mocks_backend.MockNerdctlContainerSvc
		con      *mocks_container.MockContainer
		cid      string
		s        service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		cid = "123"
		con = mocks_container.NewMockContainer(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()
		s = service{
			client:           cdClient,
			nctlContainerSvc: mockNerdctlService{ncClient, nil},
			logger:           logger,
		}
	})
	Context("getContainer", func() {
		It("should return the container object if it was found", func() {
			// search method returns one container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(Equal(con))
			Expect(err).Should(BeNil())
		})
		It("should return an error if search container method fails", func() {
			// search method returns no container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				nil, errors.New("search container error"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
		It("should return NotFound error if no container was found", func() {
			// search method returns no container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return an error if multiple containers were found", func() {
			// search method returns two containers
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			result, err := s.getContainer(ctx, cid)
			Expect(result).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
	})
})
