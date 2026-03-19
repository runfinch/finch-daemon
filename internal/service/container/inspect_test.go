// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"errors"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	specs "github.com/opencontainers/runtime-spec/specs-go"
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
		It("should return inspect object with HostConfig", func() {
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
				Devices: []dockercompat.DeviceMapping{
					{
						PathOnHost:        "",
						PathInContainer:   "",
						CgroupPermissions: "",
					},
				},
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
				Devices: []types.DeviceMapping{
					{
						PathOnHost:        "",
						PathInContainer:   "",
						CgroupPermissions: "",
					},
				},
			}

			// search container method returns one container
			cdClient.EXPECT().SearchContainer(gomock.Any(), cid).Return(
				[]containerd.Container{con}, nil)

			ncClient.EXPECT().InspectContainer(gomock.Any(), con, false).Return(
				&inspectWithHostConfig, nil)

			con.EXPECT().Labels(gomock.Any()).Return(nil, nil)
			con.EXPECT().Spec(gomock.Any()).Return(&specs.Spec{}, nil)
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

	Context("getHostConfigFromDockerCompat - extended fields", func() {
		It("should return nil for nil input", func() {
			result := getHostConfigFromDockerCompat(nil)
			Expect(result).Should(BeNil())
		})
		It("should map CgroupnsMode, DNS, ExtraHosts, GroupAdd, Tmpfs, UTSMode, Runtime", func() {
			input := &dockercompat.HostConfig{
				CgroupnsMode: "private",
				DNS:          []string{"8.8.8.8", "8.8.4.4"},
				DNSOptions:   []string{"ndots:5"},
				DNSSearch:    []string{"example.com"},
				ExtraHosts:   []string{"host1:10.0.0.1"},
				GroupAdd:     []string{"audio", "video"},
				Tmpfs:        map[string]string{"/tmp": "rw,noexec"},
				UTSMode:      "host",
				Runtime:      "runc",
			}
			result := getHostConfigFromDockerCompat(input)
			Expect(result).ShouldNot(BeNil())
			Expect(result.CgroupnsMode).Should(Equal(types.CgroupnsMode("private")))
			Expect(result.DNS).Should(Equal([]string{"8.8.8.8", "8.8.4.4"}))
			Expect(result.DNSOptions).Should(Equal([]string{"ndots:5"}))
			Expect(result.DNSSearch).Should(Equal([]string{"example.com"}))
			Expect(result.ExtraHosts).Should(Equal([]string{"host1:10.0.0.1"}))
			Expect(result.GroupAdd).Should(Equal([]string{"audio", "video"}))
			Expect(result.Tmpfs).Should(HaveKeyWithValue("/tmp", "rw,noexec"))
			Expect(result.UTSMode).Should(Equal("host"))
			Expect(result.Runtime).Should(Equal("runc"))
		})
		It("should map CPUQuota", func() {
			input := &dockercompat.HostConfig{
				CPUQuota: 50000,
			}
			result := getHostConfigFromDockerCompat(input)
			Expect(result).ShouldNot(BeNil())
			Expect(result.CPUQuota).Should(Equal(int64(50000)))
		})
		It("should map BlkioWeight and blkio device settings", func() {
			input := &dockercompat.HostConfig{
				BlkioSettings: dockercompat.BlkioSettings{
					BlkioWeight: 500,
					BlkioWeightDevice: []*dockercompat.WeightDevice{
						{Path: "/dev/sda", Weight: 200},
					},
					BlkioDeviceReadBps: []*dockercompat.ThrottleDevice{
						{Path: "/dev/sda", Rate: 1048576},
					},
					BlkioDeviceWriteBps: []*dockercompat.ThrottleDevice{
						{Path: "/dev/sda", Rate: 524288},
					},
					BlkioDeviceReadIOps: []*dockercompat.ThrottleDevice{
						{Path: "/dev/sda", Rate: 1000},
					},
					BlkioDeviceWriteIOps: []*dockercompat.ThrottleDevice{
						{Path: "/dev/sda", Rate: 500},
					},
				},
			}
			result := getHostConfigFromDockerCompat(input)
			Expect(result).ShouldNot(BeNil())
			Expect(result.BlkioWeight).Should(Equal(uint16(500)))
			Expect(result.BlkioWeightDevice).Should(HaveLen(1))
			Expect(result.BlkioWeightDevice[0].Path).Should(Equal("/dev/sda"))
			Expect(result.BlkioWeightDevice[0].Weight).Should(Equal(uint16(200)))
			Expect(result.BlkioDeviceReadBps).Should(HaveLen(1))
			Expect(result.BlkioDeviceReadBps[0].Rate).Should(Equal(uint64(1048576)))
			Expect(result.BlkioDeviceWriteBps).Should(HaveLen(1))
			Expect(result.BlkioDeviceWriteBps[0].Rate).Should(Equal(uint64(524288)))
			Expect(result.BlkioDeviceReadIOps).Should(HaveLen(1))
			Expect(result.BlkioDeviceReadIOps[0].Rate).Should(Equal(uint64(1000)))
			Expect(result.BlkioDeviceWriteIOps).Should(HaveLen(1))
			Expect(result.BlkioDeviceWriteIOps[0].Rate).Should(Equal(uint64(500)))
		})
		It("should skip nil blkio weight devices", func() {
			input := &dockercompat.HostConfig{
				BlkioSettings: dockercompat.BlkioSettings{
					BlkioWeightDevice:    []*dockercompat.WeightDevice{nil, {Path: "/dev/sdb", Weight: 100}},
					BlkioDeviceReadBps:   []*dockercompat.ThrottleDevice{nil, {Path: "/dev/sdb", Rate: 2048}},
					BlkioDeviceWriteBps:  []*dockercompat.ThrottleDevice{nil},
					BlkioDeviceReadIOps:  []*dockercompat.ThrottleDevice{nil},
					BlkioDeviceWriteIOps: []*dockercompat.ThrottleDevice{nil},
				},
			}
			result := getHostConfigFromDockerCompat(input)
			Expect(result).ShouldNot(BeNil())
			Expect(result.BlkioWeightDevice).Should(HaveLen(1))
			Expect(result.BlkioWeightDevice[0].Path).Should(Equal("/dev/sdb"))
			Expect(result.BlkioDeviceReadBps).Should(HaveLen(1))
			Expect(result.BlkioDeviceWriteBps).Should(BeNil())
			Expect(result.BlkioDeviceReadIOps).Should(BeNil())
			Expect(result.BlkioDeviceWriteIOps).Should(BeNil())
		})
		It("should filter all-nil blkio device slices to nil output", func() {
			input := &dockercompat.HostConfig{
				BlkioSettings: dockercompat.BlkioSettings{
					BlkioWeightDevice:    []*dockercompat.WeightDevice{nil, nil, nil},
					BlkioDeviceReadBps:   []*dockercompat.ThrottleDevice{nil, nil},
					BlkioDeviceWriteBps:  []*dockercompat.ThrottleDevice{nil},
					BlkioDeviceReadIOps:  []*dockercompat.ThrottleDevice{nil},
					BlkioDeviceWriteIOps: []*dockercompat.ThrottleDevice{nil},
				},
			}
			result := getHostConfigFromDockerCompat(input)
			Expect(result).ShouldNot(BeNil())
			Expect(result.BlkioWeightDevice).Should(BeNil())
			Expect(result.BlkioDeviceReadBps).Should(BeNil())
			Expect(result.BlkioDeviceWriteBps).Should(BeNil())
			Expect(result.BlkioDeviceReadIOps).Should(BeNil())
			Expect(result.BlkioDeviceWriteIOps).Should(BeNil())
		})
		It("should handle empty/zero fields gracefully", func() {
			input := &dockercompat.HostConfig{}
			result := getHostConfigFromDockerCompat(input)
			Expect(result).ShouldNot(BeNil())
			Expect(result.CgroupnsMode).Should(Equal(types.CgroupnsMode("")))
			Expect(result.DNS).Should(BeNil())
			Expect(result.ExtraHosts).Should(BeNil())
			Expect(result.GroupAdd).Should(BeNil())
			Expect(result.Tmpfs).Should(BeNil())
			Expect(result.UTSMode).Should(Equal(""))
			Expect(result.Runtime).Should(Equal(""))
			Expect(result.CPUQuota).Should(Equal(int64(0)))
			Expect(result.BlkioWeight).Should(Equal(uint16(0)))
			Expect(result.BlkioWeightDevice).Should(BeNil())
			Expect(result.BlkioDeviceReadBps).Should(BeNil())
		})
	})

	Context("enrichHostConfigFromSpec", func() {
		It("should be a no-op for nil spec", func() {
			hc := &types.ContainerHostConfig{}
			enrichHostConfigFromSpec(hc, nil, nil)
			Expect(hc.CapAdd).Should(BeNil())
			Expect(hc.Privileged).Should(BeFalse())
		})
		It("should extract capabilities from OCI spec", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					Capabilities: &specs.LinuxCapabilities{
						// Only non-default caps — CapAdd should contain these, CapDrop should contain all defaults
						Bounding: []string{"CAP_NET_ADMIN", "CAP_SYS_TIME"},
					},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.CapAdd).Should(ConsistOf("CAP_NET_ADMIN", "CAP_SYS_TIME"))
			Expect(hc.CapDrop).Should(HaveLen(14)) // all 14 defaults are dropped
			Expect(hc.Privileged).Should(BeFalse())
		})
		It("should compute CapAdd and CapDrop relative to default cap set", func() {
			hc := &types.ContainerHostConfig{}
			// Simulate: --cap-add NET_ADMIN SYS_TIME --cap-drop CHOWN NET_RAW
			// Bounding = defaults - CHOWN - NET_RAW + NET_ADMIN + SYS_TIME
			bounding := []string{
				"CAP_DAC_OVERRIDE", "CAP_FSETID", "CAP_FOWNER", "CAP_MKNOD",
				"CAP_SETGID", "CAP_SETUID", "CAP_SETFCAP", "CAP_SETPCAP",
				"CAP_NET_BIND_SERVICE", "CAP_SYS_CHROOT", "CAP_KILL", "CAP_AUDIT_WRITE",
				"CAP_NET_ADMIN", "CAP_SYS_TIME",
			}
			spec := &specs.Spec{
				Process: &specs.Process{
					Capabilities: &specs.LinuxCapabilities{
						Bounding: bounding,
					},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.CapAdd).Should(ConsistOf("CAP_NET_ADMIN", "CAP_SYS_TIME"))
			Expect(hc.CapDrop).Should(ConsistOf("CAP_CHOWN", "CAP_NET_RAW"))
			Expect(hc.Privileged).Should(BeFalse())
		})
		It("should detect privileged mode with large capability set", func() {
			hc := &types.ContainerHostConfig{}
			// Build a bounding set with 40 capabilities to trigger privileged detection
			caps := make([]string, 40)
			for i := range caps {
				caps[i] = "CAP_FAKE_" + string(rune('A'+i))
			}
			spec := &specs.Spec{
				Process: &specs.Process{
					Capabilities: &specs.LinuxCapabilities{
						Bounding: caps,
					},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Privileged).Should(BeTrue())
		})
		It("should extract PidsLimit from OCI spec", func() {
			hc := &types.ContainerHostConfig{}
			limit := int64(100)
			spec := &specs.Spec{
				Linux: &specs.Linux{
					Resources: &specs.LinuxResources{
						Pids: &specs.LinuxPids{
							Limit: &limit,
						},
					},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.PidsLimit).Should(Equal(int64(100)))
		})
		It("should extract Ulimits from OCI spec rlimits", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					Rlimits: []specs.POSIXRlimit{
						{Type: "RLIMIT_NOFILE", Hard: 65536, Soft: 1024},
					},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Ulimits).Should(HaveLen(1))
			Expect(hc.Ulimits[0].Name).Should(Equal("nofile"))
			Expect(hc.Ulimits[0].Hard).Should(Equal(int64(65536)))
			Expect(hc.Ulimits[0].Soft).Should(Equal(int64(1024)))
		})
		It("should extract annotations filtering out nerdctl/ prefixed ones", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Annotations: map[string]string{
					"custom-annotation":  "value1",
					"nerdctl/internal":   "should-be-filtered",
					"another-annotation": "value2",
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Annotations).Should(HaveLen(2))
			Expect(hc.Annotations).Should(HaveKeyWithValue("custom-annotation", "value1"))
			Expect(hc.Annotations).Should(HaveKeyWithValue("another-annotation", "value2"))
			Expect(hc.Annotations).ShouldNot(HaveKey("nerdctl/internal"))
		})
		It("should extract NetworkMode from container labels", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.Networks: `["my-network"]`,
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.NetworkMode).Should(Equal("my-network"))
		})
		It("should extract AutoRemove from container labels", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.ContainerAutoRemove: "true",
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.AutoRemove).Should(BeTrue())
		})
		It("should extract Binds from container labels", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.Mounts: `[{"Type":"bind","Source":"/host/path","Destination":"/container/path","Mode":"ro"}]`,
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.Binds).Should(HaveLen(1))
			Expect(hc.Binds[0]).Should(Equal("/host/path:/container/path:ro"))
		})
		It("should extract SecurityOpt from OCI spec", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					ApparmorProfile: "docker-default",
					SelinuxLabel:    "system_u:system_r:container_t:s0",
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.SecurityOpt).Should(ContainElement("apparmor=docker-default"))
			Expect(hc.SecurityOpt).Should(ContainElement("label=system_u:system_r:container_t:s0"))
		})
		It("should detect Init when entrypoint is tini", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					Args: []string{"/sbin/tini", "--", "/app"},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Init).ShouldNot(BeNil())
			Expect(*hc.Init).Should(BeTrue())
		})
		It("should not set Init for non-init entrypoints", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					Args: []string{"/bin/sh", "-c", "echo hello"},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Init).Should(BeNil())
		})
		It("should detect Init when entrypoint is docker-init", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					Args: []string{"/usr/bin/docker-init", "--", "/app"},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Init).ShouldNot(BeNil())
			Expect(*hc.Init).Should(BeTrue())
		})
		It("should not set CapAdd when bounding caps are empty", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					Capabilities: &specs.LinuxCapabilities{
						Bounding: []string{},
					},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.CapAdd).Should(BeNil())
			Expect(hc.CapDrop).Should(HaveLen(14)) // all defaults dropped
			Expect(hc.Privileged).Should(BeFalse())
		})
		It("should handle nil Process in spec without panic", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: nil,
				Linux: &specs.Linux{
					Resources: &specs.LinuxResources{},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.CapAdd).Should(BeNil())
			Expect(hc.Ulimits).Should(BeNil())
			Expect(hc.SecurityOpt).Should(BeNil())
			Expect(hc.Init).Should(BeNil())
		})
		It("should handle multiple ulimits", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Process: &specs.Process{
					Rlimits: []specs.POSIXRlimit{
						{Type: "RLIMIT_NOFILE", Hard: 65536, Soft: 1024},
						{Type: "RLIMIT_NPROC", Hard: 4096, Soft: 2048},
						{Type: "RLIMIT_CORE", Hard: 0, Soft: 0},
					},
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Ulimits).Should(HaveLen(3))
			Expect(hc.Ulimits[0].Name).Should(Equal("nofile"))
			Expect(hc.Ulimits[1].Name).Should(Equal("nproc"))
			Expect(hc.Ulimits[1].Hard).Should(Equal(int64(4096)))
			Expect(hc.Ulimits[1].Soft).Should(Equal(int64(2048)))
			Expect(hc.Ulimits[2].Name).Should(Equal("core"))
		})
		It("should set nil Annotations when all are nerdctl-prefixed", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{
				Annotations: map[string]string{
					"nerdctl/one": "a",
					"nerdctl/two": "b",
				},
			}
			enrichHostConfigFromSpec(hc, spec, nil)
			Expect(hc.Annotations).Should(BeNil())
		})
		It("should extract Binds without Mode suffix", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.Mounts: `[{"Type":"bind","Source":"/src","Destination":"/dst","Mode":""}]`,
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.Binds).Should(HaveLen(1))
			Expect(hc.Binds[0]).Should(Equal("/src:/dst"))
		})
		It("should filter out non-bind mount types from Binds", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.Mounts: `[{"Type":"volume","Source":"vol1","Destination":"/data","Mode":""},{"Type":"bind","Source":"/host","Destination":"/cont","Mode":"rw"}]`,
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.Binds).Should(HaveLen(1))
			Expect(hc.Binds[0]).Should(Equal("/host:/cont:rw"))
		})
		It("should handle invalid JSON in Networks label gracefully", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.Networks: `not-valid-json`,
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.NetworkMode).Should(Equal(""))
		})
		It("should handle invalid JSON in Mounts label gracefully", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.Mounts: `{broken`,
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.Binds).Should(BeNil())
		})
		It("should set AutoRemove false for 'false' label value", func() {
			hc := &types.ContainerHostConfig{}
			spec := &specs.Spec{}
			containerLabels := map[string]string{
				labels.ContainerAutoRemove: "false",
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.AutoRemove).Should(BeFalse())
		})
		It("should populate all fields from a fully enriched spec and labels", func() {
			hc := &types.ContainerHostConfig{}
			limit := int64(512)
			spec := &specs.Spec{
				Process: &specs.Process{
					Capabilities: &specs.LinuxCapabilities{
						Bounding: []string{"CAP_NET_ADMIN"},
					},
					Rlimits: []specs.POSIXRlimit{
						{Type: "RLIMIT_NOFILE", Hard: 1024, Soft: 512},
					},
					ApparmorProfile: "my-profile",
					Args:            []string{"/sbin/tini", "--", "/app"},
				},
				Linux: &specs.Linux{
					Resources: &specs.LinuxResources{
						Pids: &specs.LinuxPids{Limit: &limit},
					},
				},
				Annotations: map[string]string{
					"custom": "val",
				},
			}
			containerLabels := map[string]string{
				labels.Networks:            `["bridge"]`,
				labels.ContainerAutoRemove: "true",
				labels.Mounts:              `[{"Type":"bind","Source":"/a","Destination":"/b","Mode":"ro"}]`,
			}
			enrichHostConfigFromSpec(hc, spec, containerLabels)
			Expect(hc.CapAdd).Should(ConsistOf("CAP_NET_ADMIN"))
			Expect(hc.CapDrop).Should(HaveLen(14)) // all 14 defaults dropped since only CAP_NET_ADMIN in bounding
			Expect(hc.Privileged).Should(BeFalse())
			Expect(hc.PidsLimit).Should(Equal(int64(512)))
			Expect(hc.Ulimits).Should(HaveLen(1))
			Expect(hc.Annotations).Should(HaveKeyWithValue("custom", "val"))
			Expect(hc.NetworkMode).Should(Equal("bridge"))
			Expect(hc.AutoRemove).Should(BeTrue())
			Expect(hc.Binds).Should(Equal([]string{"/a:/b:ro"}))
			Expect(hc.SecurityOpt).Should(ContainElement("apparmor=my-profile"))
			Expect(hc.Init).ShouldNot(BeNil())
			Expect(*hc.Init).Should(BeTrue())
		})
	})

	Context("getNetworkName", func() {
		It("should return original name when no labels exist", func() {
			result := getNetworkName(map[string]string{}, "unknown-eth0")
			Expect(result).Should(Equal("unknown-eth0"))
		})
		It("should return original name when labels key is missing", func() {
			result := getNetworkName(nil, "unknown-eth0")
			Expect(result).Should(Equal("unknown-eth0"))
		})
		It("should resolve network name from label by index", func() {
			lab := map[string]string{
				labels.Networks: `["bridge","custom-net"]`,
			}
			Expect(getNetworkName(lab, "unknown-eth0")).Should(Equal("bridge"))
			Expect(getNetworkName(lab, "unknown-eth1")).Should(Equal("custom-net"))
		})
		It("should return original name when index is out of range", func() {
			lab := map[string]string{
				labels.Networks: `["bridge"]`,
			}
			result := getNetworkName(lab, "unknown-eth5")
			Expect(result).Should(Equal("unknown-eth5"))
		})
		It("should return original name for non-prefixed network names", func() {
			lab := map[string]string{
				labels.Networks: `["bridge"]`,
			}
			result := getNetworkName(lab, "some-other-network")
			Expect(result).Should(Equal("some-other-network"))
		})
		It("should return original name when label JSON is invalid", func() {
			lab := map[string]string{
				labels.Networks: `not-json`,
			}
			result := getNetworkName(lab, "unknown-eth0")
			Expect(result).Should(Equal("unknown-eth0"))
		})
		It("should return original name when index is not a number", func() {
			lab := map[string]string{
				labels.Networks: `["bridge"]`,
			}
			result := getNetworkName(lab, "unknown-ethabc")
			Expect(result).Should(Equal("unknown-ethabc"))
		})
	})

	Context("isPrivileged", func() {
		It("should return false for fewer than 38 capabilities", func() {
			caps := make([]string, 37)
			for i := range caps {
				caps[i] = "CAP_FAKE"
			}
			Expect(isPrivileged(caps)).Should(BeFalse())
		})
		It("should return true for exactly 38 capabilities", func() {
			caps := make([]string, 38)
			for i := range caps {
				caps[i] = "CAP_FAKE"
			}
			Expect(isPrivileged(caps)).Should(BeTrue())
		})
		It("should return false for empty capabilities", func() {
			Expect(isPrivileged([]string{})).Should(BeFalse())
		})
		It("should return false for nil capabilities", func() {
			Expect(isPrivileged(nil)).Should(BeFalse())
		})
	})

})
