// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package distribution

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/config"
	registrytypes "github.com/docker/docker/api/types/registry"
	"go.uber.org/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/runfinch/finch-daemon/mocks/mocks_distribution"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// TestDistributionHandler function is the entry point of distribution handler package's unit test using ginkgo.
func TestDistributionHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Distribution APIs Handler")
}

var _ = Describe("Distribution Inspect API", func() {
	var (
		mockCtrl       *gomock.Controller
		logger         *mocks_logger.Logger
		service        *mocks_distribution.MockService
		h              *handler
		rr             *httptest.ResponseRecorder
		name           string
		req            *http.Request
		ociPlatformAmd ocispec.Platform
		ociPlatformArm ocispec.Platform
		resp           registrytypes.DistributionInspect
		respJSON       []byte
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		logger = mocks_logger.NewLogger(mockCtrl)
		service = mocks_distribution.NewMockService(mockCtrl)
		c := config.Config{}
		h = newHandler(service, &c, logger)
		rr = httptest.NewRecorder()
		name = "test-image"
		var err error
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/distribution/%s/json", name), nil)
		Expect(err).Should(BeNil())
		req = mux.SetURLVars(req, map[string]string{"name": name})
		ociPlatformAmd = ocispec.Platform{
			Architecture: "amd64",
			OS:           "linux",
		}
		ociPlatformArm = ocispec.Platform{
			Architecture: "amd64",
			OS:           "linux",
		}
		resp = registrytypes.DistributionInspect{
			Descriptor: ocispec.Descriptor{
				MediaType:   ocispec.MediaTypeImageManifest,
				Digest:      "sha256:9bae60c369e612488c2a089c38737277a4823a3af97ec6866c3b4ad05251bfa5",
				Size:        2,
				URLs:        []string{},
				Annotations: map[string]string{},
				Data:        []byte{},
				Platform:    &ociPlatformAmd,
			},
			Platforms: []ocispec.Platform{
				ociPlatformAmd,
				ociPlatformArm,
			},
		}
		respJSON, err = json.Marshal(resp)
		Expect(err).Should(BeNil())
	})
	Context("handler", func() {
		It("should return inspect object and 200 status code upon success", func() {
			service.EXPECT().Inspect(gomock.Any(), name, gomock.Any()).Return(&resp, nil)

			// handler should return response object with 200 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(respJSON))
			Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		})
		It("should return 403 status code if image resolution fails due to lack of credentials", func() {
			service.EXPECT().Inspect(gomock.Any(), name, gomock.Any()).Return(nil, errdefs.NewUnauthenticated(fmt.Errorf("access denied")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message with 404 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "access denied"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusForbidden))
		})
		It("should return 403 status code if image was not found", func() {
			service.EXPECT().Inspect(gomock.Any(), name, gomock.Any()).Return(nil, errdefs.NewNotFound(fmt.Errorf("no such image")))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message with 404 status code
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "no such image"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusForbidden))
		})
		It("should return 500 status code if service returns an error message", func() {
			service.EXPECT().Inspect(gomock.Any(), name, gomock.Any()).Return(nil, fmt.Errorf("error"))
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

			// handler should return error message
			h.inspect(rr, req)
			Expect(rr.Body).Should(MatchJSON(`{"message": "error"}`))
			Expect(rr).Should(HaveHTTPStatus(http.StatusInternalServerError))
		})
	})
})
