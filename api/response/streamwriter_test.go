// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/runfinch/finch-daemon/mocks/mocks_http"
)

// TestAPIResponse function is the entry point of api response package's unit test using ginkgo.
func TestAPIResponse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - API Response Utils")
}

var _ = Describe("API Response ", func() {
	var (
		mockCtrl                     *gomock.Controller
		respWriter                   *mocks_http.MockResponseWriter
		respHeader                   http.Header
		msg1, msg2, msg3, auxMsg     []byte
		resp1, resp2, resp3, auxResp []byte
		errMsg                       error
		errResp                      []byte
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		respWriter = mocks_http.NewMockResponseWriter(mockCtrl)
		respHeader = http.Header{}

		// message strings
		msg1 = []byte("stream message 1")
		msg2 = []byte("hello world")
		msg3 = []byte("last message")
		errMsg = fmt.Errorf("got error")
		auxMsg = []byte(`{"message":"aux data"}`)

		// stream responses
		resp1 = []byte(fmt.Sprintf(`{"stream":"%s"}`+"\n", msg1))
		resp2 = []byte(fmt.Sprintf(`{"stream":"%s"}`+"\n", msg2))
		resp3 = []byte(fmt.Sprintf(`{"stream":"%s"}`+"\n", msg3))
		auxResp = []byte(fmt.Sprintf(`{"aux":%s}`+"\n", auxMsg))

		// error response
		var err error
		errResp, err = json.Marshal(StreamResponse{
			Error: &jsonmessage.JSONError{
				Code:    http.StatusInternalServerError,
				Message: errMsg.Error(),
			},
			ErrorMessage: errMsg.Error(),
		})
		Expect(err).Should(BeNil())
		errResp = append(errResp, byte('\n'))
	})

	Context("StreamWriter", func() {
		It("should write a message as a JSON stream object", func() {
			// expected calls to response writer
			respWriter.EXPECT().Header().Return(respHeader)
			respWriter.EXPECT().WriteHeader(http.StatusOK)
			respWriter.EXPECT().Write(resp1).Return(len(msg1), nil)

			// streamwriter should successfully write the message and set header
			sw := NewStreamWriter(respWriter)
			n, err := sw.Write(msg1)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg1)))
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))
		})
		It("should write multiple messages as a JSON stream and set header once", func() {
			// expected calls to response writer
			respWriter.EXPECT().Header().Return(respHeader)
			respWriter.EXPECT().WriteHeader(http.StatusOK)
			respWriter.EXPECT().Write(resp1).Return(len(msg1), nil)
			respWriter.EXPECT().Write(resp2).Return(len(msg2), nil)
			respWriter.EXPECT().Write(resp3).Return(len(msg3), nil)

			// streamwriter should successfully write the messages and set header
			sw := NewStreamWriter(respWriter)
			n, err := sw.Write(msg1)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg1)))
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))

			n, err = sw.Write(msg2)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg2)))

			n, err = sw.Write(msg3)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg3)))
		})
		It("should write and flush messages to a JSON stream and set header once", func() {
			respFlusher := newMockResponseFlusher(mockCtrl)
			// expected calls to response writer
			respFlusher.EXPECT().Header().Return(respHeader)
			respFlusher.EXPECT().WriteHeader(http.StatusOK)
			respFlusher.EXPECT().Write(resp1).Return(len(msg1), nil)
			respFlusher.EXPECT().Write(resp2).Return(len(msg2), nil)
			respFlusher.EXPECT().Write(resp3).Return(len(msg3), nil)

			// streamwriter should successfully write the messages, set header, and flush 3 times
			sw := NewStreamWriter(respFlusher)
			Expect(respFlusher.flushed).Should(Equal(0))
			n, err := sw.Write(msg1)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg1)))
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))
			Expect(respFlusher.flushed).Should(Equal(1))

			n, err = sw.Write(msg2)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg2)))
			Expect(respFlusher.flushed).Should(Equal(2))

			n, err = sw.Write(msg3)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg3)))
			Expect(respFlusher.flushed).Should(Equal(3))
		})
		It("should stream messages and an error, but set header only once", func() {
			// expected calls to response writer
			respWriter.EXPECT().Header().Return(respHeader)
			respWriter.EXPECT().WriteHeader(http.StatusOK)
			respWriter.EXPECT().Write(resp1).Return(len(msg1), nil)
			respWriter.EXPECT().Write(resp2).Return(len(msg2), nil)
			respWriter.EXPECT().Write(errResp).Return(len(errResp), nil)

			// streamwriter should successfully write messages, set header, and write an error response
			sw := NewStreamWriter(respWriter)
			n, err := sw.Write(msg1)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg1)))
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))

			n, err = sw.Write(msg2)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg2)))

			err = sw.WriteError(http.StatusInternalServerError, errMsg)
			Expect(err).Should(BeNil())
		})
		It("should return an error message with status code without streaming", func() {
			var err error
			errResp, err = json.Marshal(NewError(errMsg))
			Expect(err).Should(BeNil())
			errResp = append(errResp, byte('\n'))

			// expected calls to response writer
			respWriter.EXPECT().Header().Return(respHeader)
			respWriter.EXPECT().WriteHeader(http.StatusInternalServerError)
			respWriter.EXPECT().Write(errResp).Return(len(errResp), nil)

			// streamwriter should successfully write an error response with a status code
			sw := NewStreamWriter(respWriter)
			err = sw.WriteError(http.StatusInternalServerError, errMsg)
			Expect(err).Should(BeNil())
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))
		})
		It("should stream messages and aux data, but set header only once", func() {
			respFlusher := newMockResponseFlusher(mockCtrl)
			// expected calls to response writer
			respFlusher.EXPECT().Header().Return(respHeader)
			respFlusher.EXPECT().WriteHeader(http.StatusOK)
			respFlusher.EXPECT().Write(resp1).Return(len(msg1), nil)
			respFlusher.EXPECT().Write(resp2).Return(len(msg2), nil)
			respFlusher.EXPECT().Write(auxResp).Return(len(auxMsg), nil)

			// streamwriter should successfully write messages, set header, and write an aux response
			sw := NewStreamWriter(respFlusher)
			n, err := sw.Write(msg1)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg1)))
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))
			Expect(respFlusher.flushed).Should(Equal(1))

			n, err = sw.Write(msg2)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg2)))
			Expect(respFlusher.flushed).Should(Equal(2))

			err = sw.WriteAux(auxMsg)
			Expect(err).Should(BeNil())
			Expect(respFlusher.flushed).Should(Equal(3))
		})
		It("should return an error when writing aux data failed", func() {
			auxErr := fmt.Errorf("failed to write aux data")
			// expected calls to response writer
			respWriter.EXPECT().Header().Return(respHeader)
			respWriter.EXPECT().WriteHeader(http.StatusOK)
			respWriter.EXPECT().Write(auxResp).Return(0, auxErr)

			// streamwriter should successfully write messages, set header, and write an aux response
			sw := NewStreamWriter(respWriter)
			err := sw.WriteAux(auxMsg)
			Expect(err).Should(Equal(auxErr))
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))
		})
	})

	Context("pullJobWriter", func() {
		It("should write messages as JSON stream objects after being resolved", func() {
			resolvedMsg := []byte("resolved image")
			resolvedResp := []byte(fmt.Sprintf(`{"stream":"%s"}`+"\n", resolvedMsg))

			// expected calls to response writer
			respWriter.EXPECT().Header().Return(respHeader)
			respWriter.EXPECT().WriteHeader(http.StatusOK)
			respWriter.EXPECT().Write(resolvedResp).Return(len(resolvedMsg), nil)
			respWriter.EXPECT().Write(resp2).Return(len(msg2), nil)

			// pulljobwriter should only write messages after being resolved and set header once
			sw := NewPullJobWriter(respWriter)
			n, err := sw.Write(msg1)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(0))
			Expect(respHeader).Should(BeEmpty())
			Expect(sw.IsResolved()).Should(BeFalse())

			n, err = sw.Write(resolvedMsg)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(resolvedMsg)))
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))
			Expect(sw.IsResolved()).Should(BeTrue())

			n, err = sw.Write(msg2)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(len(msg2)))
		})
		It("should ignore all messages if not resolved and send an error message with status code", func() {
			var err error
			errResp, err = json.Marshal(NewError(errMsg))
			Expect(err).Should(BeNil())
			errResp = append(errResp, byte('\n'))

			// pulljobwriter will ignore all the messages
			sw := NewPullJobWriter(respWriter)
			n, err := sw.Write(msg1)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(0))
			Expect(respHeader).Should(BeEmpty())
			Expect(sw.IsResolved()).Should(BeFalse())

			n, err = sw.Write(msg2)
			Expect(err).Should(BeNil())
			Expect(n).Should(Equal(0))
			Expect(respHeader).Should(BeEmpty())
			Expect(sw.IsResolved()).Should(BeFalse())

			// pulljobwriter will send the error message with status code
			respWriter.EXPECT().Header().Return(respHeader)
			respWriter.EXPECT().WriteHeader(http.StatusInternalServerError)
			respWriter.EXPECT().Write(errResp).Return(len(errResp), nil)
			err = sw.WriteError(http.StatusInternalServerError, errMsg)
			Expect(err).Should(BeNil())
			Expect(respHeader).Should(HaveKeyWithValue("Content-Type", []string{"application/json"}))
			Expect(sw.IsResolved()).Should(BeFalse())
		})
	})
})

type mockResponseFlusher struct {
	*mocks_http.MockResponseWriter
	flushed int
}

func newMockResponseFlusher(ctrl *gomock.Controller) *mockResponseFlusher {
	return &mockResponseFlusher{
		MockResponseWriter: mocks_http.NewMockResponseWriter(ctrl),
		flushed:            0,
	}
}

func (m *mockResponseFlusher) Flush() {
	m.flushed += 1
}
