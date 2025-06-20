// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package credentialrouter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	dockertypes "github.com/docker/cli/cli/config/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/runfinch/finch-daemon/pkg/credential"
)

// TestRouterFunctions is the entry point for unit tests in the router package.
func TestRouterFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Router functions")
}

// Unit tests for the CreateCredentialHandler function.
var _ = Describe("CreateCredentialHandler test", func() {
	var (
		mockCtrl          *gomock.Controller
		mockLogger        *mocks_logger.Logger
		rr                *httptest.ResponseRecorder
		credCache         *credential.CredentialCache
		credentialService *credential.CredentialService
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLogger = mocks_logger.NewLogger(mockCtrl)
		rr = httptest.NewRecorder()

		// Create a real credential service and cache for the handler
		credCache = credential.NewCredentialCache()
		credentialService = credential.NewCredentialService(mockLogger, credCache)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should set up credential routes correctly", func() {
		// Store test credentials
		buildID := "test-build-id"
		serverAddr := "registry.example.com"
		authConfig := dockertypes.AuthConfig{
			Username: "testuser",
			Password: "testpass",
		}

		err := credentialService.StoreAuthConfigs(
			context.Background(),
			buildID,
			map[string]dockertypes.AuthConfig{serverAddr: authConfig},
		)
		Expect(err).Should(BeNil())
		mock_auth_middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}
		handler, err := CreateCredentialHandler(credentialService, mockLogger, mock_auth_middleware)
		Expect(err).Should(BeNil())

		requestBody := fmt.Sprintf(`{"buildID": "%s", "serverAddr": "%s"}`, buildID, serverAddr)
		req, _ := http.NewRequest(http.MethodGet, "/finch/credentials", bytes.NewBufferString(requestBody))

		handler.ServeHTTP(rr, req)

		Expect(rr.Code).Should(Equal(http.StatusOK))
	})
})
