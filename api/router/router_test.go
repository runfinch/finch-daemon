// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_system"
	"github.com/runfinch/finch-daemon/version"
)

// TestRouterFunctions is the entry point for unit tests in the router package.
func TestRouterFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - Router functions")
}

// Unit tests for the version middleware.
var _ = Describe("version middleware test", func() {
	var (
		opts     *Options
		h        http.Handler
		rr       *httptest.ResponseRecorder
		expected types.VersionInfo
		sysSvc   *mocks_system.MockService
	)

	// TODO: rethink the unit test cases for the router.
	BeforeEach(func() {
		mockCtrl := gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()
		c := config.Config{}
		sysSvc = mocks_system.NewMockService(mockCtrl)
		opts = &Options{
			Config:           &c,
			ContainerService: nil,
			ImageService:     nil,
			NetworkService:   nil,
			SystemService:    sysSvc,
			BuilderService:   nil,
			VolumeService:    nil,
			NerdctlWrapper:   nil,
		}
		h = New(opts)
		rr = httptest.NewRecorder()
		expected = types.VersionInfo{
			Platform: struct {
				Name string
			}{},
			Version:       "0.0.1",
			ApiVersion:    "1.43",
			MinAPIVersion: "1.35",
			GitCommit:     "abcd",
			Os:            "linux",
			Arch:          "x86",
			KernelVersion: "kernel-123",
			Experimental:  true,
			Components: []types.ComponentVersion{
				{
					Name:    "containerd",
					Version: "v1.7.1",
					Details: map[string]string{
						"GitCommit": "1677a17964311325ed1c31e2c0a3589ce6d5c30d",
					},
				},
			},
		}
		sysSvc.EXPECT().GetVersion(gomock.Any()).Return(&expected, nil).AnyTimes()
	})
	It("should return a 400 error for versions below the min supported", func() {
		testVer := "1.11"
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost/v%s/version", testVer), nil)

		h.ServeHTTP(rr, req)

		Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		Expect(rr.Body.String()).Should(MatchJSON(fmt.Sprintf(
			`{"message": "your api version, v%s, is below the minimum supported version, v%s"}`, testVer,
			version.MinimumApiVersion)))
	})
	It("should return a 400 error for versions above the default supported", func() {
		testVer := "1.99"
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost/v%s/version", testVer), nil)

		h.ServeHTTP(rr, req)

		Expect(rr).Should(HaveHTTPStatus(http.StatusBadRequest))
		Expect(rr.Body.String()).Should(MatchJSON(fmt.Sprintf(
			`{"message": "your api version, v%s, is newer than the server's version, v%s"}`, testVer,
			version.DefaultApiVersion)))
	})
	It("should parse a versioned route correctly and return 200 success", func() {
		testVer := "1.40"
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost/v%s/version", testVer), nil)

		h.ServeHTTP(rr, req)

		Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		jd := json.NewDecoder(rr.Body)
		var v types.VersionInfo
		err := jd.Decode(&v)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(v).Should(Equal(expected))
	})
	It("should parse a non-versioned route correctly and return 200 success", func() {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost/version", nil)

		h.ServeHTTP(rr, req)

		Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		jd := json.NewDecoder(rr.Body)
		var v types.VersionInfo
		err := jd.Decode(&v)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(v).Should(Equal(expected))
	})
})
