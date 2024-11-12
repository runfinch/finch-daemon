// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/v2/pkg/config"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/mocks/mocks_image"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Image Pull API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_image.MockService
		h        *handler
		rr       *httptest.ResponseRecorder
		name     string
		tag      string
		platform string
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_image.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		name = "test-image"
		tag = "test-tag"
		platform = "test-platform"
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})
	Context("handler", func() {
		It("should return 200 status code upon success", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s", name, tag),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				"",
				gomock.Any(),
				gomock.Any(),
			).Return(nil)

			// handler should return 200 status code
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 200 status code upon success with platform specification", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s&platform=%s", name, tag, platform),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				platform,
				gomock.Any(),
				gomock.Any(),
			).Return(nil)

			// handler should return 200 status code
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 200 status code upon success with authentication", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s", name, tag),
				nil,
			)
			Expect(err).Should(BeNil())
			authB64 := base64.StdEncoding.EncodeToString([]byte(`{"username": "test-user", "password": "test-password"}`))
			req.Header.Set("X-Registry-Auth", authB64)

			// expected decoded auth config
			authCfg := dockertypes.AuthConfig{
				Username: "test-user",
				Password: "test-password",
			}

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				"",
				&authCfg,
				gomock.Any(),
			).Return(nil)

			// handler should return 200 status code
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 200 status code upon success when digest is specified", func() {
			tag := "sha256:7ea94d4e7f346a9328a9ff053ab149e3c99c1737f8d251094e7cc38664c3d4b9"
			nameWithDigest := fmt.Sprintf("%s@%s", name, tag)
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s", nameWithDigest, tag),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				"",
				gomock.Any(),
				gomock.Any(),
			).Return(nil)

			// handler should return 200 status code
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 404 status code if image could not be resolved", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s&platform=%s", name, tag, platform),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				platform,
				gomock.Any(),
				gomock.Any(),
			).Return(errdefs.NewNotFound(fmt.Errorf("no such image")))

			// handler should return error message with 404 status code
			h.pull(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "no such image"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 500 status code if service returns an error message", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s&platform=%s", name, tag, platform),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				platform,
				gomock.Any(),
				gomock.Any(),
			).Return(fmt.Errorf("error"))

			// handler should return error message with 500 status code
			h.pull(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
		It("should log warnings if unsupported parameters are specified", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&fromSrc=%s&tag=%s&change=abcd", name, name, tag),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				"",
				gomock.Any(),
				gomock.Any(),
			).Return(nil)

			logger.EXPECT().Warn(gomock.Any()).Times(2)

			// handler should return error message with 400 status code
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 400 status code if image is not specified", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				"/images/create",
				nil,
			)
			Expect(err).Should(BeNil())

			// handler should return error message with 400 status code
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})
		It("should return 400 status code if tag is not specified", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s", name),
				nil,
			)
			Expect(err).Should(BeNil())

			// handler should return error message with 400 status code
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		})
		It("should stream pull updates upon success", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s", name, tag),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				"",
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(func(_ context.Context, _, _, _ string, _ *dockertypes.AuthConfig, sw io.Writer) error {
				sw.Write([]byte("this message should be ignored\n"))
				sw.Write([]byte("resolved image\n")) // the streamwriter will start streaming when a message contains substring "resolved"
				sw.Write([]byte("pulling image\n"))
				sw.Write([]byte("pulling complete\n"))
				return nil
			})

			// handler should return 200 status code with streamed updates
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))

			// expected stream output
			scanner := bufio.NewScanner(rr.Body)
			outputs := []response.StreamResponse{}
			for scanner.Scan() {
				var stream response.StreamResponse
				err = json.Unmarshal(scanner.Bytes(), &stream)
				Expect(err).Should(BeNil())
				outputs = append(outputs, stream)
			}
			Expect(len(outputs)).Should(Equal(4))
			Expect(outputs[0]).Should(Equal(response.StreamResponse{Stream: "resolved image\n"}))
			Expect(outputs[1]).Should(Equal(response.StreamResponse{Stream: "pulling image\n"}))
			Expect(outputs[2]).Should(Equal(response.StreamResponse{Stream: "pulling complete\n"}))
			Expect(outputs[3]).Should(Equal(response.StreamResponse{Stream: fmt.Sprintf("Pulled %s:%s\n", name, tag)}))
		})
		It("should send 200 status code and display an error message after streaming", func() {
			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/images/create?fromImage=%s&tag=%s", name, tag),
				nil,
			)
			Expect(err).Should(BeNil())

			service.EXPECT().Pull(
				gomock.Any(),
				name,
				tag,
				"",
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(func(_ context.Context, _, _, _ string, _ *dockertypes.AuthConfig, sw io.Writer) error {
				sw.Write([]byte("this message should be ignored\n"))
				sw.Write([]byte("resolved image\n")) // the streamwriter will start streaming when a message contains substring "resolved"
				sw.Write([]byte("pulling image\n"))
				return fmt.Errorf("error pulling")
			})

			// handler should return 200 status code with streamed updates
			h.pull(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))

			// expected stream output
			scanner := bufio.NewScanner(rr.Body)
			outputs := []response.StreamResponse{}
			for scanner.Scan() {
				var stream response.StreamResponse
				err = json.Unmarshal(scanner.Bytes(), &stream)
				Expect(err).Should(BeNil())
				outputs = append(outputs, stream)
			}
			Expect(len(outputs)).Should(Equal(3))
			Expect(outputs[0]).Should(Equal(response.StreamResponse{Stream: "resolved image\n"}))
			Expect(outputs[1]).Should(Equal(response.StreamResponse{Stream: "pulling image\n"}))
			Expect(outputs[2]).Should(Equal(response.StreamResponse{
				Error:        &jsonmessage.JSONError{Code: http.StatusInternalServerError, Message: "error pulling"},
				ErrorMessage: "error pulling",
			}))
		})
	})
})
