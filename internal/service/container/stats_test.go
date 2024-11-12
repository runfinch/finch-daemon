// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	"github.com/containerd/containerd"
	cTypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/events"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/typeurl/v2"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	eventtype "github.com/runfinch/finch-daemon/api/events"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/mocks/mocks_statsutil"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// Unit tests related to container stats API.
var _ = Describe("Container Stats API ", func() {
	var (
		ctx         context.Context
		mockCtrl    *gomock.Controller
		logger      *mocks_logger.Logger
		cdClient    *mocks_backend.MockContainerdClient
		ncClient    *mocks_backend.MockNerdctlContainerSvc
		stats       *mocks_statsutil.MockStatsUtil
		con         *mocks_container.MockContainer
		task        *mocks_container.MockTask
		cid         string
		cname       string
		removeCh    chan *events.Envelope
		removeErrCh chan error
		s           service
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize mocks
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		stats = mocks_statsutil.NewMockStatsUtil(mockCtrl)
		cid = "123"
		cname = "/test"
		con = mocks_container.NewMockContainer(mockCtrl)
		task = mocks_container.NewMockTask(mockCtrl)
		removeCh = make(chan *events.Envelope)
		removeErrCh = make(chan error)
		con.EXPECT().ID().Return(cid).AnyTimes()
		con.EXPECT().Labels(gomock.Any()).Return(map[string]string{labels.Name: "test"}, nil).AnyTimes()
		cdClient.EXPECT().GetContainerRemoveEvent(gomock.Any(), con).Return(removeCh, removeErrCh).AnyTimes()
		s = service{
			client:           cdClient,
			nctlContainerSvc: mockNerdctlService{ncClient, nil},
			logger:           logger,
			stats:            stats,
		}
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})
	Context("service", func() {
		It("should return NotFound error if container was not found", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{}, nil)

			// service should return NotFound error
			statsCh, err := s.Stats(ctx, cid)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
			Expect(statsCh).Should(BeNil())
		})
		It("should return empty stats objects if task was not found", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(nil, cerrdefs.ErrNotFound).MinTimes(1)

			// service should return the stats channel
			ctx, cancel := context.WithCancel(ctx)
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 2 ticks for stats channel to be populated
			time.Sleep(time.Second * 2)
			cancel()

			// check returnted stats objects
			expected := types.StatsJSON{ID: cid, Name: cname}
			num := 0
			for st := range statsCh {
				Expect(*st).Should(Equal(expected))
				num += 1
			}
			// should tick 1 or 2 times in 2 seconds
			Expect(num).Should(Or(Equal(1), Equal(2)))
		})
		It("should return empty stats objects if there was an error in getting container status", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(task, nil).MinTimes(1)
			task.EXPECT().Status(gomock.Any()).Return(
				containerd.Status{}, fmt.Errorf("error getting status")).MinTimes(1)
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).MinTimes(1)

			// service should return the stats channel
			ctx, cancel := context.WithCancel(ctx)
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 2 ticks for stats channel to be populated
			time.Sleep(time.Second * 2)
			cancel()

			// check returned stats objects
			expected := types.StatsJSON{ID: cid, Name: cname}
			num := 0
			for st := range statsCh {
				Expect(*st).Should(Equal(expected))
				num += 1
			}
			// should tick 1 or 2 times in 2 seconds
			Expect(num).Should(Or(Equal(1), Equal(2)))
		})
		It("should return empty stats objects if task metrics are not in the correct format", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(task, nil).MinTimes(1)
			task.EXPECT().Status(gomock.Any()).Return(
				containerd.Status{Status: containerd.Running}, nil).MinTimes(1)

			// define an invalid metrics type
			data := eventtype.Event{}
			anydata, err := typeurl.MarshalAny(&data)
			Expect(err).Should(BeNil())
			metrics := &cTypes.Metric{Data: &anypb.Any{TypeUrl: anydata.GetTypeUrl(), Value: anydata.GetValue()}}
			task.EXPECT().Metrics(gomock.Any()).Return(metrics, nil).MinTimes(1)
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).MinTimes(1)

			// service should return the stats channel
			ctx, cancel := context.WithCancel(ctx)
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 2 ticks for stats channel to be populated
			time.Sleep(time.Second * 2)
			cancel()

			// check returned stats objects
			expected := types.StatsJSON{ID: cid, Name: cname}
			num := 0
			for st := range statsCh {
				Expect(*st).Should(Equal(expected))
				num += 1
			}
			// should tick 1 or 2 times in 2 seconds
			Expect(num).Should(Or(Equal(1), Equal(2)))
		})
		It("should return empty stats objects if there was an error collecting network stats", func() {
			pid := 457
			netNS := native.NetNS{
				Interfaces: []native.NetInterface{},
			}
			metrics, _ := getDummyMetricsV1()

			// setup mocks
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(task, nil).MinTimes(1)
			task.EXPECT().Status(gomock.Any()).Return(
				containerd.Status{Status: containerd.Running}, nil).MinTimes(1)
			task.EXPECT().Pid().Return(uint32(pid)).MinTimes(1)
			ncClient.EXPECT().InspectNetNS(gomock.Any(), pid).Return(&netNS, nil).MinTimes(1)
			task.EXPECT().Metrics(gomock.Any()).Return(metrics, nil).MinTimes(1)
			stats.EXPECT().GetSystemCPUUsage().Return(uint64(1000), nil).MinTimes(1)
			stats.EXPECT().GetNumberOnlineCPUs().Return(uint32(3), nil).MinTimes(1)

			// setup error mock
			stats.EXPECT().CollectNetworkStats(pid, netNS.Interfaces).Return(
				nil, fmt.Errorf("error collecting network stats")).MinTimes(1)
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).MinTimes(1)

			// service should return the stats channel
			ctx, cancel := context.WithCancel(ctx)
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 2 ticks for stats channel to be populated
			time.Sleep(time.Second * 2)
			cancel()

			// check returned stats objects
			expected := types.StatsJSON{ID: cid, Name: cname}
			num := 0
			for st := range statsCh {
				Expect(*st).Should(Equal(expected))
				num += 1
			}
			// should tick 1 or 2 times in 2 seconds
			Expect(num).Should(Or(Equal(1), Equal(2)))
		})
		It("should return empty stats objects for a container that is not running", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(task, nil).MinTimes(1)
			task.EXPECT().Status(gomock.Any()).Return(
				containerd.Status{Status: containerd.Stopped}, nil).MinTimes(1)

			// service should return the stats channel
			ctx, cancel := context.WithCancel(ctx)
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 2 ticks for stats channel to be populated
			time.Sleep(time.Second * 2)
			cancel()

			// check returned stats objects
			expected := types.StatsJSON{ID: cid, Name: cname}
			num := 0
			for st := range statsCh {
				Expect(*st).Should(Equal(expected))
				num += 1
			}
			// should tick 1 or 2 times in 2 seconds
			Expect(num).Should(Or(Equal(1), Equal(2)))
		})
		It("should stop sending updates after container is removed", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(nil, cerrdefs.ErrNotFound).MinTimes(1)

			// service should return the stats channel
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 2 ticks for stats channel to be populated
			// and then send container remove event
			time.Sleep(time.Second * 2)
			removeCh <- &events.Envelope{}

			// check returned stats objects
			expected := types.StatsJSON{ID: cid, Name: cname}
			num := 0
			for st := range statsCh {
				Expect(*st).Should(Equal(expected))
				num += 1
			}
			// should tick 1 or 2 times in 2 seconds
			Expect(num).Should(Or(Equal(1), Equal(2)))
		})
		It("should stop sending updates and log error after container is removed with error", func() {
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(nil, cerrdefs.ErrNotFound).MinTimes(1)

			// service should return the stats channel
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 2 ticks for stats channel to be populated
			// and then send container remove event
			time.Sleep(time.Second * 2)
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			removeErrCh <- fmt.Errorf("error removing container")

			// check returned stats objects
			expected := types.StatsJSON{ID: cid, Name: cname}
			num := 0
			for st := range statsCh {
				Expect(*st).Should(Equal(expected))
				num += 1
			}
			// should tick 1 or 2 times in 2 seconds
			Expect(num).Should(Or(Equal(1), Equal(2)))
		})
		It("should return expected stats objects from given metrics", func() {
			pid := 456
			netNS := native.NetNS{
				Interfaces: []native.NetInterface{},
			}
			metrics1, expected1 := getDummyMetricsV1()
			metrics2, expected2 := getDummyMetricsV2()
			expected := []*types.StatsJSON{expected1, expected2}

			// setup mocks
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return(
				[]containerd.Container{con}, nil)
			con.EXPECT().Task(gomock.Any(), nil).Return(task, nil).MinTimes(2)
			task.EXPECT().Status(gomock.Any()).Return(
				containerd.Status{Status: containerd.Running}, nil).MinTimes(2)
			task.EXPECT().Pid().Return(uint32(pid)).MinTimes(2)
			ncClient.EXPECT().InspectNetNS(gomock.Any(), pid).Return(&netNS, nil).MinTimes(2)

			// expect calls to Metrics() to return dummy cgroups data
			task.EXPECT().Metrics(gomock.Any()).Return(metrics1, nil)
			task.EXPECT().Metrics(gomock.Any()).Return(metrics2, nil)
			// subsequent calls will return the first metric
			task.EXPECT().Metrics(gomock.Any()).Return(metrics1, nil).AnyTimes()

			// mock statsutil
			netStats := dockertypes.NetworkStats{
				RxBytes:   20,
				TxBytes:   30,
				RxPackets: 10,
				TxPackets: 5,
				RxErrors:  1,
				TxErrors:  2,
				RxDropped: 5,
				TxDropped: 10,
			}
			stats.EXPECT().GetSystemCPUUsage().Return(uint64(2500), nil).MinTimes(2)
			stats.EXPECT().GetNumberOnlineCPUs().Return(uint32(3), nil).MinTimes(2)
			stats.EXPECT().CollectNetworkStats(pid, netNS.Interfaces).Return(
				map[string]dockertypes.NetworkStats{"eth0": netStats}, nil).MinTimes(2)

			// service should return the stats channel
			ctx, cancel := context.WithCancel(ctx)
			statsCh, err := s.Stats(ctx, cid)
			Expect(err).Should(BeNil())

			// wait 3 ticks for stats channel to be populated
			time.Sleep(time.Second * 3)
			cancel()

			// expected stats
			for _, exp := range expected {
				exp.ID = cid
				exp.Name = cname
				exp.Networks = map[string]dockertypes.NetworkStats{"eth0": netStats}
			}

			// check returned stats objects
			num := 0
			prev := &types.StatsJSON{}
			for st := range statsCh {
				var exp *types.StatsJSON
				if num < 2 {
					exp = expected[num]
				} else {
					exp = expected[0]
				}
				exp.Read = st.Read
				exp.PreRead = prev.Read
				exp.PreCPUStats = prev.CPUStats
				Expect(*st).Should(Equal(*exp))
				prev = exp
				num += 1
			}
			// should tick 2 or 3 times in 3 seconds
			Expect(num).Should(Or(Equal(2), Equal(3)))
		})
	})
})

func getDummyMetricsV1() (*cTypes.Metric, *types.StatsJSON) {
	// containerd task metrics
	data := v1.Metrics{
		Pids: &v1.PidsStat{Current: 10, Limit: 20},
		CPU: &v1.CPUStat{
			Usage: &v1.CPUUsage{
				Total:  1000,
				Kernel: 500,
				User:   250,
				PerCPU: []uint64{1, 2, 3, 4},
			},
		},
		Memory: &v1.MemoryStat{
			Usage: &v1.MemoryEntry{
				Limit:   1000,
				Usage:   250,
				Max:     500,
				Failcnt: 50,
			},
		},
	}
	anydata, err := typeurl.MarshalAny(&data)
	Expect(err).Should(BeNil())
	m := cTypes.Metric{Data: &anypb.Any{TypeUrl: anydata.GetTypeUrl(), Value: anydata.GetValue()}}

	// expected stats object for dummy metrics
	expected := types.StatsJSON{}
	expected.PidsStats = dockertypes.PidsStats{Current: 10, Limit: 20}
	expected.CPUStats = types.CPUStats{
		CPUUsage: dockertypes.CPUUsage{
			TotalUsage:        1000,
			UsageInKernelmode: 500,
			UsageInUsermode:   250,
			PercpuUsage:       []uint64{1, 2, 3, 4},
		},
		SystemUsage: 2500,
		OnlineCPUs:  3,
	}
	expected.MemoryStats = dockertypes.MemoryStats{
		Usage:    250,
		Limit:    1000,
		MaxUsage: 500,
		Failcnt:  50,
	}

	return &m, &expected
}

func getDummyMetricsV2() (*cTypes.Metric, *types.StatsJSON) {
	// containerd task metrics
	data := v2.Metrics{
		Pids: &v2.PidsStat{Current: 20, Limit: 40},
		CPU: &v2.CPUStat{
			UsageUsec:  10,
			UserUsec:   7,
			SystemUsec: 20,
		},
		Memory: &v2.MemoryStat{
			Usage:      100,
			UsageLimit: 500,
		},
		MemoryEvents: &v2.MemoryEvents{Oom: 30},
	}
	anydata, err := typeurl.MarshalAny(&data)
	Expect(err).Should(BeNil())
	m := cTypes.Metric{Data: &anypb.Any{TypeUrl: anydata.GetTypeUrl(), Value: anydata.GetValue()}}

	// expected stats object for dummy metrics
	expected := types.StatsJSON{}
	expected.PidsStats = dockertypes.PidsStats{Current: 20, Limit: 40}
	expected.CPUStats = types.CPUStats{
		CPUUsage: dockertypes.CPUUsage{
			TotalUsage:        10000,
			UsageInKernelmode: 20000,
			UsageInUsermode:   7000,
		},
		SystemUsage: 2500,
		OnlineCPUs:  3,
	}
	expected.MemoryStats = dockertypes.MemoryStats{
		Usage:   100,
		Limit:   500,
		Failcnt: 30,
	}

	return &m, &expected
}
