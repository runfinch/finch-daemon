// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/events"
	"github.com/runfinch/finch-daemon/api/handlers/builder"
	"github.com/runfinch/finch-daemon/mocks/mocks_archive"
	"github.com/runfinch/finch-daemon/mocks/mocks_backend"
	"github.com/runfinch/finch-daemon/mocks/mocks_container"
	"github.com/runfinch/finch-daemon/mocks/mocks_ecc"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
)

// Unit tests related to Build API.
var _ = Describe("Build API ", func() {
	var (
		ctx          context.Context
		mockCtrl     *gomock.Controller
		logger       *mocks_logger.Logger
		tarExtractor *mocks_archive.MockTarExtractor
		mockCmd      *mocks_ecc.MockExecCmd
		cdClient     *mocks_backend.MockContainerdClient
		ncBuilderSvc *mocks_backend.MockNerdctlBuilderSvc
		ncImgSvc     *mocks_backend.MockNerdctlImageSvc
		con          *mocks_container.MockContainer
		cid          string
		service      builder.Service
		buildOption  types.BuilderBuildOptions
		req          *http.Request
		rr           *httptest.ResponseRecorder
	)
	BeforeEach(func() {
		ctx = context.Background()
		// initialize the mocks
		mockCtrl = gomock.NewController(GinkgoT())
		rr = httptest.NewRecorder()
		logger = mocks_logger.NewLogger(mockCtrl)
		tarExtractor = mocks_archive.NewMockTarExtractor(mockCtrl)
		mockCmd = mocks_ecc.NewMockExecCmd(mockCtrl)
		cdClient = mocks_backend.NewMockContainerdClient(mockCtrl)
		ncBuilderSvc = mocks_backend.NewMockNerdctlBuilderSvc(mockCtrl)
		ncImgSvc = mocks_backend.NewMockNerdctlImageSvc(mockCtrl)
		con = mocks_container.NewMockContainer(mockCtrl)
		con.EXPECT().ID().Return(cid).AnyTimes()
		mockCmd.EXPECT().GetDir().Return(fmt.Sprintf("%s/%s", os.TempDir(), cid)).AnyTimes()
		mockCmd.EXPECT().SetStderr(gomock.Any()).AnyTimes()

		service = NewService(cdClient, mockNerdctlService{ncBuilderSvc, ncImgSvc}, logger, tarExtractor)
		buildOption = types.BuilderBuildOptions{}
		req, _ = http.NewRequest(http.MethodPost, "/build", nil)
	})
	Context("service", func() {
		It("should successfully build image", func() {
			// set up the mock
			ncBuilderSvc.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
			tarExtractor.EXPECT().ExtractInTemp(gomock.Any(), gomock.Any()).
				Return(mockCmd, nil)
			tarExtractor.EXPECT().Cleanup(gomock.Any())
			mockCmd.EXPECT().Run().Return(nil)

			// service should not return any error
			_, err := service.Build(ctx, &buildOption, req.Body)
			Expect(err).Should(BeNil())
		})
		It("should fail building image due to temp folder creation failure", func() {
			// set up the mock
			mockErr := fmt.Errorf("failed to create temp folder")
			tarExtractor.EXPECT().ExtractInTemp(gomock.Any(), gomock.Any()).Return(nil, mockErr)
			logger.EXPECT().Warnf("Failed to extract build context. Error: %v", mockErr)
			// service should return error
			_, err := service.Build(ctx, &buildOption, req.Body)
			Expect(err).Should(Equal(mockErr))
		})
		It("should fail building image due to tar extraction failure", func() {
			// set up the mock
			mockCmd.EXPECT().Run().Return(fmt.Errorf("tar error"))
			tarExtractor.EXPECT().ExtractInTemp(gomock.Any(), gomock.Any()).Return(mockCmd, nil)
			tarExtractor.EXPECT().Cleanup(gomock.Any())
			logger.EXPECT().Warnf("Failed to extract build context in temp folder. Dir: %s, Error: %s, Stderr: %s",
				gomock.Any(), gomock.Any(), gomock.Any())
			// service should return error
			_, err := service.Build(ctx, &buildOption, req.Body)
			Expect(err).Should(Not(BeNil()))
			Expect(err.Error()).Should(Equal("failed to extract build context in temp folder"))
		})
		It("should fail building image due create image", func() {
			// set up the mock
			mockCmd.EXPECT().Run().Return(fmt.Errorf("tar error"))
			tarExtractor.EXPECT().ExtractInTemp(gomock.Any(), gomock.Any()).Return(mockCmd, nil)
			tarExtractor.EXPECT().Cleanup(gomock.Any())
			logger.EXPECT().Warnf("Failed to extract build context in temp folder. Dir: %s, Error: %s, Stderr: %s",
				gomock.Any(), gomock.Any(), gomock.Any())
			// service should return error
			_, err := service.Build(ctx, &buildOption, req.Body)
			Expect(err.Error()).Should(Equal("failed to extract build context in temp folder"))
		})
		It("should fail building image due build error from nerdctl", func() {
			// set up the mock
			mockCmd.EXPECT().Run().Return(nil)
			errExpected := fmt.Errorf("nerdctl error")
			ncBuilderSvc.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any()).Return(errExpected)
			tarExtractor.EXPECT().ExtractInTemp(gomock.Any(), gomock.Any()).Return(mockCmd, nil)
			tarExtractor.EXPECT().Cleanup(gomock.Any())
			// service should return err
			_, err := service.Build(ctx, &buildOption, req.Body)
			Expect(err).Should(Equal(errExpected))
		})
		It("should successfully tag image after build", func() {
			tag := "test-tag"
			imageId := "test-image"
			buildOption.Tag = []string{tag}
			buildOption.Stdout = rr

			// set up mocks
			ncBuilderSvc.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
			expectPublishTagEvent(mockCtrl, tag).Return(&events.Event{ID: imageId}, nil)
			tarExtractor.EXPECT().ExtractInTemp(gomock.Any(), gomock.Any()).
				Return(mockCmd, nil)
			tarExtractor.EXPECT().Cleanup(gomock.Any())
			mockCmd.EXPECT().Run().Return(nil)

			// service should not return any error
			result, err := service.Build(ctx, &buildOption, req.Body)
			Expect(err).Should(BeNil())
			Expect(result).Should(HaveLen(1))
			Expect(result[0].ID).Should(Equal(imageId))

			// should stream output response with "Successfully built {id}"
			data, err := io.ReadAll(rr.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(data)).Should(ContainSubstring(fmt.Sprintf("Successfully built %s", imageId)))
		})
	})
})

// expectPublishTagEvent creates a new mocked object for publishTagEvent function
// with expected input parameters.
func expectPublishTagEvent(ctrl *gomock.Controller, tag string) *mockPublishTagEvent {
	m := &mockPublishTagEvent{ctrl: ctrl, expectedTag: tag}
	ctrl.RecordCall(m, "PublishTagEvent", m.expectedTag)
	publishTagEventFunc = func(_ *service, _ context.Context, tag string) (*events.Event, error) {
		m.PublishTagEvent(tag)
		return m.outputEvent, m.outputErr
	}
	return m
}

func (m *mockPublishTagEvent) PublishTagEvent(tag string) {
	m.ctrl.Call(m, "PublishTagEvent", tag)
}

func (m *mockPublishTagEvent) Return(event *events.Event, err error) {
	m.outputEvent = event
	m.outputErr = err
}

type mockPublishTagEvent struct {
	expectedTag string
	outputEvent *events.Event
	outputErr   error
	ctrl        *gomock.Controller
}
