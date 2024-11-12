// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"
	"time"

	"github.com/containerd/containerd"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	ncContainer "github.com/containerd/nerdctl/v2/pkg/cmd/container"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

// Unit tests related to container list API.
var _ = Describe("Container List API ", func() {
	var (
		ctx          context.Context
		mockCtrl     *gomock.Controller
		logger       *mocks_logger.Logger
		cdClient     *mocks_backend.MockContainerdClient
		ncClient     *mocks_backend.MockNerdctlContainerSvc
		listOpts     ncTypes.ContainerListOptions
		created      time.Time
		containers   []ncContainer.ListItem
		tarExtractor *mocks_archive.MockTarExtractor
		service      container.Service
		con          *mocks_container.MockContainer
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		listOpts = ncTypes.ContainerListOptions{}
		created = time.Now()
		containers = []ncContainer.ListItem{
			{ID: "id1", Names: "name1", Image: "img1", CreatedAt: created, Labels: nil},
			{ID: "id2", Names: "name2", Image: "img2", CreatedAt: created, Labels: nil},
		}
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)

		service = NewService(cdClient, mockNerdctlService{ncClient, nil}, logger, nil, nil, tarExtractor)
	})
	Context("service", func() {
		It("should successfully list containers", func() {
			expectedNS := &dockercompat.NetworkSettings{
				DefaultNetworkSettings: dockercompat.DefaultNetworkSettings{
					IPAddress: "ip-test",
				},
			}

			ncClient.EXPECT().ListContainers(ctx, listOpts).Return(
				containers, nil)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]containerd.Container{con}, nil)
			ncClient.EXPECT().InspectContainer(gomock.Any(), gomock.Any()).AnyTimes().Return(
				&dockercompat.Container{
					NetworkSettings: expectedNS,
					Mounts:          nil,
					State: &dockercompat.ContainerState{
						Status: "running",
					},
				}, nil)
			con.EXPECT().Labels(gomock.Any()).AnyTimes().Return(nil, nil)

			want := []types.ContainerListItem{
				{Id: "id1", Names: []string{"/name1"}, Image: "img1", CreatedAt: created.Unix(), State: "running", NetworkSettings: expectedNS, Mounts: nil},
				{Id: "id2", Names: []string{"/name2"}, Image: "img2", CreatedAt: created.Unix(), State: "running", NetworkSettings: expectedNS, Mounts: nil},
			}
			got, err := service.List(ctx, listOpts)
			Expect(err).Should(BeNil())
			Expect(got).Should(Equal(want))
		})
		It("should successfully list zero container", func() {
			ncClient.EXPECT().ListContainers(ctx, listOpts).Return(
				[]ncContainer.ListItem{}, nil)
			want := []types.ContainerListItem{}
			got, err := service.List(ctx, listOpts)
			Expect(err).Should(BeNil())
			Expect(got).Should(Equal(want))
		})
		It("should return error when nerdctl returns error", func() {
			mockErr := errors.New("error while listing containers")
			ncClient.EXPECT().ListContainers(ctx, listOpts).Return(
				[]ncContainer.ListItem{}, mockErr)
			got, err := service.List(ctx, listOpts)
			Expect(err).Should(Equal(mockErr))
			Expect(got).Should(BeEmpty())
		})
	})
})
