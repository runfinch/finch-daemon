// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_archive"

	"github.com/containerd/containerd"
	"github.com/containerd/nerdctl/pkg/labels"
	"github.com/containerd/nerdctl/pkg/labels/k8slabels"
	"github.com/containerd/typeurl/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/handlers/container"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Container Logs API ", func() {
	var (
		ctx          context.Context
		mockCtrl     *gomock.Controller
		logger       *mocks_logger.Logger
		cdClient     *mocks_backend.MockContainerdClient
		ncClient     *mocks_backend.MockNerdctlContainerSvc
		tarExtractor *mocks_archive.MockTarExtractor
		service      container.Service

		mockWriter   *bytes.Buffer
		stopChannel  chan os.Signal
		setupStreams func() (io.Writer, io.Writer, chan os.Signal, func(), error)
		cid          string
	)
	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		logger = mocks_logger.NewLogger(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncClient = mocks_backend.NewMockNerdctlContainerSvc(mockCtrl)
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)
		service = NewService(cdClient, mockNerdctlService{ncClient, nil}, logger, nil, nil, tarExtractor)

		mockWriter = new(bytes.Buffer)
		stopChannel = make(chan os.Signal, 1)
		signal.Notify(stopChannel, syscall.SIGTERM, syscall.SIGINT)
		setupStreams = func() (io.Writer, io.Writer, chan os.Signal, func(), error) {
			return mockWriter, mockWriter, stopChannel, func() {}, nil
		}
		cid = "test-container"
	})
	Context("service", func() {
		It("should return an error if opts.GetStreams returns an error", func() {
			// set up expected mocks, errors and the setupstreams to return an error
			con := mocks_container.NewMockContainer(mockCtrl)
			logger.EXPECT().Infof("getting logs for container: %s", cid).Return()
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return([]containerd.Container{con}, nil)
			con.EXPECT().ID().Return(cid)
			expErr := fmt.Errorf("error")
			setupStreams = func() (io.Writer, io.Writer, chan os.Signal, func(), error) {
				return nil, nil, nil, nil, expErr
			}
			// set up options
			opts := types.LogsOptions{
				GetStreams: setupStreams,
				Stdout:     true,
				Stderr:     true,
				Follow:     true,
				Since:      0,
				Until:      0,
				Timestamps: false,
				Tail:       "",
				MuxStreams: false,
			}

			// run function and assertions
			err := service.Logs(ctx, cid, &opts)
			Expect(err).Should(Equal(expErr))
		})
		It("should return an error if the datastore cannot be found", func() {
			// set up mocks and expected errors
			expErr := "error data store not found"
			con := mocks_container.NewMockContainer(mockCtrl)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return([]containerd.Container{con}, nil)
			logger.EXPECT().Infof("getting logs for container: %s", cid).Return()
			con.EXPECT().ID().Return(cid)
			ncClient.EXPECT().GetDataStore().Return("", fmt.Errorf("%s", expErr))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Return()
			// set up options
			opts := types.LogsOptions{
				GetStreams: setupStreams,
				Stdout:     true,
				Stderr:     true,
				Follow:     true,
				Since:      0,
				Until:      0,
				Timestamps: false,
				Tail:       "",
				MuxStreams: true,
			}

			// run function and assertions
			err := service.Logs(ctx, cid, &opts)
			Expect(err.Error()).Should(ContainSubstring(expErr))
		})
		It("should return a not found error if a container can't be found", func() {
			// set up mocks
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return([]containerd.Container{}, nil)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Return()

			// set up options
			opts := types.LogsOptions{
				GetStreams: setupStreams,
				Stdout:     true,
				Stderr:     true,
				Follow:     true,
				Since:      0,
				Until:      0,
				Timestamps: false,
				Tail:       "",
				MuxStreams: true,
			}

			// run function and assertions
			err := service.Logs(ctx, cid, &opts)
			Expect(errdefs.IsNotFound(err)).Should(BeTrue())
		})
		It("should successfully return logs without follow", func() {
			// set up mocks
			con := mocks_container.NewMockContainer(mockCtrl)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return([]containerd.Container{con}, nil)
			logger.EXPECT().Infof("getting logs for container: %s", cid).Return()
			con.EXPECT().ID().Return(cid)
			ncClient.EXPECT().GetDataStore().Return("", nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{labels.Namespace: "test"}, nil)
			// construct typeURL.Any object as getLogPath calls it to unmarshall and get a value
			type testJSONObj struct{ LogPath string }
			testJSON := &testJSONObj{LogPath: ""}
			testAny, _ := typeurl.MarshalAny(testJSON)
			// continue setting up mocks
			con.EXPECT().Extensions(gomock.Any()).Return(map[string]typeurl.Any{
				k8slabels.ContainerMetadataExtension: testAny,
			}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)
			con.EXPECT().ID().Return(cid)
			ncClient.EXPECT().LoggingInitContainerLogViewer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			ncClient.EXPECT().LoggingPrintLogsTo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			// set up options
			opts := types.LogsOptions{
				GetStreams: setupStreams,
				Stdout:     true,
				Stderr:     true,
				Follow:     false,
				Since:      0,
				Until:      0,
				Timestamps: false,
				Tail:       "",
				MuxStreams: true,
			}

			// run function and assertions
			err := service.Logs(ctx, cid, &opts)
			Expect(err).Should(BeNil())
		})
		It("should successfully log a container with follow=1, but not follow when stopped", func() {
			// set up mocks
			con := mocks_container.NewMockContainer(mockCtrl)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return([]containerd.Container{con}, nil)
			logger.EXPECT().Infof("getting logs for container: %s", cid).Return()
			con.EXPECT().ID().Return(cid)
			ncClient.EXPECT().GetDataStore().Return("", nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{labels.Namespace: "test"}, nil)
			// construct typeURL.Any object as getLogPath calls it to unmarshall and get a value
			type testJSONObj struct{ LogPath string }
			testJSON := &testJSONObj{LogPath: ""}
			testAny, _ := typeurl.MarshalAny(testJSON)
			// continue setting up mocks
			con.EXPECT().Extensions(gomock.Any()).Return(map[string]typeurl.Any{
				k8slabels.ContainerMetadataExtension: testAny,
			}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Stopped)
			con.EXPECT().Task(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			con.EXPECT().ID().Return(cid)
			ncClient.EXPECT().LoggingInitContainerLogViewer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			ncClient.EXPECT().LoggingPrintLogsTo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			// set up options
			opts := types.LogsOptions{
				GetStreams: setupStreams,
				Stdout:     true,
				Stderr:     true,
				Follow:     true,
				Since:      0,
				Until:      0,
				Timestamps: false,
				Tail:       "",
				MuxStreams: true,
			}

			// run function and assertions
			err := service.Logs(ctx, cid, &opts)
			Expect(err).Should(BeNil())
		})
		It("should return an error with follow and a running container when failed to get wait channel", func() {
			// set up expected error and mocks
			expErr := "error task wait channel"
			con := mocks_container.NewMockContainer(mockCtrl)
			cdClient.EXPECT().SearchContainer(gomock.Any(), gomock.Any()).Return([]containerd.Container{con}, nil)
			logger.EXPECT().Infof("getting logs for container: %s", cid).Return()
			con.EXPECT().ID().Return(cid)
			ncClient.EXPECT().GetDataStore().Return("", nil)
			con.EXPECT().Labels(gomock.Any()).Return(map[string]string{labels.Namespace: "test"}, nil)
			// construct typeURL.Any object as getLogPath calls it to unmarshall and get a value
			type testJSONObj struct{ LogPath string }
			testJSON := &testJSONObj{LogPath: ""}
			testAny, _ := typeurl.MarshalAny(testJSON)
			// continue setting up mocks
			con.EXPECT().Extensions(gomock.Any()).Return(map[string]typeurl.Any{
				k8slabels.ContainerMetadataExtension: testAny,
			}, nil)
			cdClient.EXPECT().GetContainerStatus(gomock.Any(), gomock.Any()).Return(containerd.Running)
			cdClient.EXPECT().GetContainerTaskWait(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, fmt.Errorf("%s", expErr))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Return()

			// set up options
			opts := types.LogsOptions{
				GetStreams: setupStreams,
				Stdout:     true,
				Stderr:     true,
				Follow:     true,
				Since:      0,
				Until:      0,
				Timestamps: false,
				Tail:       "",
				MuxStreams: true,
			}

			// run function and assertions
			err := service.Logs(ctx, cid, &opts)
			Expect(err.Error()).Should(ContainSubstring(expErr))
		})
	})
})
