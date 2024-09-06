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

	"github.com/containerd/nerdctl/pkg/config"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/auth"
	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_image"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

var _ = Describe("Image Push API", func() {
	var (
		mockCtrl *gomock.Controller
		logger   *mocks_logger.Logger
		service  *mocks_image.MockService
		h        *handler
		req      *http.Request
		rr       *httptest.ResponseRecorder
		name     string
		tag      string
		result   *types.PushResult
		auxMsg   json.RawMessage
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
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		var err error
		req, err = http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/images/%s/push?tag=%s", name, tag),
			nil,
		)
		Expect(err).ShouldNot(HaveOccurred())
		req = mux.SetURLVars(req, map[string]string{"name": name})
		authB64 := base64.StdEncoding.EncodeToString([]byte(`{"username": "test-user", "password": "test-password"}`))
		req.Header.Set(auth.AuthHeader, authB64)
		result = &types.PushResult{
			Tag:    tag,
			Digest: "test-digest",
			Size:   256,
		}
		auxData, err := json.Marshal(result)
		Expect(err).ShouldNot(HaveOccurred())
		auxMsg = json.RawMessage(auxData)
	})
	Context("handler", func() {
		It("should return 200 status code and stream output upon success", func() {
			// expected decoded auth config
			expectedAuthCfg := dockertypes.AuthConfig{
				Username: "test-user",
				Password: "test-password",
			}

			// stream messages
			streamMsg := []string{"Pushing image", "Pushed"}

			service.EXPECT().Push(gomock.Any(), name, tag, gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, name, tag string, authCfg *dockertypes.AuthConfig, outStream io.Writer) (*types.PushResult, error) {
					Expect(authCfg.Username).Should(Equal(expectedAuthCfg.Username))
					Expect(authCfg.Password).Should(Equal(expectedAuthCfg.Password))
					for _, msg := range streamMsg {
						outStream.Write([]byte(msg))
					}
					return result, nil
				})

			// handler should return 200 status code
			h.push(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))

			// expected output stream
			scanner := bufio.NewScanner(rr.Body)
			outputs := []response.StreamResponse{}
			for scanner.Scan() {
				var stream response.StreamResponse
				err := json.Unmarshal(scanner.Bytes(), &stream)
				Expect(err).Should(BeNil())
				outputs = append(outputs, stream)
			}
			Expect(len(outputs)).Should(Equal(len(streamMsg) + 1))
			for i, msg := range streamMsg {
				Expect(outputs[i]).Should(Equal(response.StreamResponse{Stream: msg}))
			}
			Expect(outputs[len(outputs)-1]).Should(Equal(response.StreamResponse{Aux: &auxMsg}))
		})
		It("should return 500 status code due to invalid auth header", func() {
			req.Header.Set(auth.AuthHeader, "Invalid token")

			// handler should return 500 status code
			h.push(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
		It("should return 404 status code if image could not be resolved", func() {
			service.EXPECT().Push(
				gomock.Any(),
				name,
				tag,
				gomock.Any(),
				gomock.Any(),
			).Return(nil, errdefs.NewNotFound(fmt.Errorf("no such image")))

			// handler should return error message with 404 status code
			h.push(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "no such image"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusNotFound))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Push(
				gomock.Any(),
				name,
				tag,
				gomock.Any(),
				gomock.Any(),
			).Return(nil, fmt.Errorf("some error"))

			// handler should return error message with 500 status code
			h.push(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "some error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
		It("should return 200 status code but return auth error message as stream response", func() {
			streamMsg := "Pushing image"
			errMsg := "auth error"

			// pass empty auth header.
			req.Header.Set(auth.AuthHeader, "")
			service.EXPECT().Push(
				gomock.Any(),
				name,
				tag,
				gomock.Any(),
				gomock.Any(),
			).DoAndReturn(func(ctx context.Context, name, tag string, authCfg *dockertypes.AuthConfig, outStream io.Writer) (*types.PushResult, error) {
				// username and password should be empty
				Expect(authCfg.Username).Should(BeEmpty())
				Expect(authCfg.Password).Should(BeEmpty())
				// mimic service is trying to push the image and failed with auth error.
				outStream.Write([]byte(streamMsg))
				return nil, fmt.Errorf("%s", errMsg)
			})

			// handler should return error message with 500 status code
			h.push(rr, req)
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
			data, err := io.ReadAll(rr.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(data)).Should(ContainSubstring(streamMsg))
			Expect(string(data)).Should(And(ContainSubstring(errMsg), ContainSubstring(`"errorDetail"`)))
		})
	})
})
