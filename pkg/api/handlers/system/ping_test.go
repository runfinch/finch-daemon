package system

import (
	"net/http"
	"net/http/httptest"

	"github.com/containerd/nerdctl/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/runfinch/finch-daemon/pkg/version"
)

// Unit tests for the ping api
var _ = Describe("Ping", func() {
	var (
		h  *handler
		rr *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		c := config.Config{}
		h = newHandler(nil, &c, nil, nil)
		rr = httptest.NewRecorder()
	})

	It("should return with an OK status with the API-Version set to the current version", func() {
		h.ping(rr, nil)
		Expect(rr).Should(HaveHTTPStatus(http.StatusOK))
		Expect(rr.Header().Values("API-Version")[0]).Should(Equal(version.DefaultApiVersion))
	})
})
