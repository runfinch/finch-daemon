// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to container inspect API.
var _ = Describe("Container Inspect API ", func() {
	var (
		ctx      context.Context
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		cdClient *mocks_backend.MockContainerdClient
		ncClient *mocks_backend.MockNerdctlContainerSvc
		con      *mocks_container.MockContainer
		cid      string
		img      string
		inspect  dockercompat.Container
		ret      types.Container
		service  container.Service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		cid = "123"
		img = "test-image"
		inspect = dockercompat.Container{
			ID:      cid,
			Created: "2023-06-01",
			Path:    "/bin/sh",
			Args:    []string{"echo", "hello"},
			Image:   img,
			Name:    "test-cont",
			Config: &dockercompat.Config{
				Hostname:    "test-hostname",
				User:        "test-user",
				AttachStdin: false,
			},
		}
		ret = types.Container{
			ID:      cid,
			Created: "2023-06-01",
			Path:    "/bin/sh",
			Args:    []string{"echo", "hello"},
			Image:   img,
			Name:    "/test-cont",
			Config: &types.ContainerConfig{
				Hostname:    "test-hostname",
				User:        "test-user",
				AttachStdin: false,
				Tty:         false,
				Image:       img,
			},
		}

		service = NewService(cdClient, mockNerdctlService{ncClient, nil}, logger, nil, nil, nil)
	})
	Context("service", func() {
		It("should return the inspect object upon success", func() {
			sizeFlag := false
			// search container method returns one container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			ncClient.EXPECT().InspectContainer(gomock.Any(), con, sizeFlag).Return(
				&inspect, nil)

			con.EXPECT().Labels(gomock.Any()).Return(nil, nil)
			result, err := service.Inspect(ctx, cid, sizeFlag)

			Expect(*result).Should(Equal(ret))
			Expect(err).Should(BeNil())
		})
		It("should return inspect object with HostConfig if it is present in nerdctl's response", func() {
			inspectWithHostConfig := inspect
			inspectWithHostConfig.HostConfig = &dockercompat.HostConfig{
				ContainerIDFile: "test-container-id-file",
				// dockercompat.loggerLogConfig is not exported
				PortBindings: nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "localhost",
							HostPort: "8080",
						},
					},
				},
				ShmSize:        0,
				IpcMode:        "testIpcMode",
				PidMode:        "testPidMode",
				ReadonlyRootfs: false,
				Sysctls: map[string]string{
					"test": "test",
				},
				CPUSetMems:     "testCPUSetMems",
				CPUSetCPUs:     "testCPUSetCPUs",
				CPUShares:      0,
				CPUPeriod:      0,
				Memory:         0,
				MemorySwap:     0,
				OomKillDisable: false,
			}

			retWithHostConfig := ret
			retWithHostConfig.HostConfig = &types.ContainerHostConfig{
				ContainerIDFile: "test-container-id-file",
				PortBindings: nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "localhost",
							HostPort: "8080",
						},
					},
				},
				ShmSize:        0,
				IpcMode:        "testIpcMode",
				PidMode:        "testPidMode",
				ReadonlyRootfs: false,
				Sysctls: map[string]string{
					"test": "test",
				},
				CPUSetMems:     "testCPUSetMems",
				CPUSetCPUs:     "testCPUSetCPUs",
				CPUShares:      0,
				CPUPeriod:      0,
				Memory:         0,
				MemorySwap:     0,
				OomKillDisable: false,
			}

			// search container method returns one container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			ncClient.EXPECT().InspectContainer(gomock.Any(), con, false).Return(
				&inspectWithHostConfig, nil)

			con.EXPECT().Labels(gomock.Any()).Return(nil, nil)
			result, err := service.Inspect(ctx, cid, false)

			Expect(*result).Should(Equal(retWithHostConfig))
			Expect(err).Should(BeNil())
		})
		It("should return NotFound error if container was not found", func() {
			// search container method returns no container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// service should return a NotFound error
			result, err := service.Inspect(ctx, cid, false)
			Expect(result).Should(BeNil())
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should return an error if multiple containers were found for the given Id", func() {
			// search container method returns multiple containers
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con, con}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// service should return an error
			result, err := service.Inspect(ctx, cid, false)
			Expect(result).Should(BeNil())
			Expect(err).ShouldNot(BeNil())
		})
		It("should return an error if search container method failed", func() {
			// search container method returns no container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				nil, errors.New("error message"))
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())

			// service should return an error
			result, err := service.Inspect(ctx, cid, false)
			Expect(result).Should(BeNil())
			Expect(err).ShouldNot(BeNil())
		})
		It("should return an error if the backend inspect method failed", func() {
			// search container method returns no container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			ncClient.EXPECT().InspectContainer(gomock.Any(), con, false).Return(
				nil, errors.New("error message"))

			// service should return an error
			result, err := service.Inspect(ctx, cid, false)
			Expect(result).Should(BeNil())
			Expect(err).ShouldNot(BeNil())
		})
	})
	Context("service with size flag", func() {
		It("should return SizeRw and SizeRootFs when size flag is true", func() {
			sizeFlag := true
			expectedSizeRw := int64(1000)
			expectedSizeRootFs := int64(5000)

			inspectWithSize := inspect
			inspectWithSize.SizeRw = &expectedSizeRw
			inspectWithSize.SizeRootFs = &expectedSizeRootFs

			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			ncClient.EXPECT().InspectContainer(gomock.Any(), con, sizeFlag).Return(
				&inspectWithSize, nil)

			con.EXPECT().Labels(gomock.Any()).Return(nil, nil)
			result, err := service.Inspect(ctx, cid, sizeFlag)
			Expect(err).Should(BeNil())
			Expect(result.SizeRw).ShouldNot(BeNil())
			Expect(*result.SizeRw).Should(Equal(expectedSizeRw))
			Expect(result.SizeRootFs).ShouldNot(BeNil())
			Expect(*result.SizeRootFs).Should(Equal(expectedSizeRootFs))
		})

		It("should not return SizeRw and SizeRootFs when size flag is false", func() {
			sizeFlag := false

			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			ncClient.EXPECT().InspectContainer(gomock.Any(), con, sizeFlag).Return(
				&inspect, nil)

			con.EXPECT().Labels(gomock.Any()).Return(nil, nil)
			result, err := service.Inspect(ctx, cid, sizeFlag)
			Expect(err).Should(BeNil())
			Expect(result.SizeRw).Should(BeZero())
			Expect(result.SizeRootFs).Should(BeZero())
		})
	})

})
