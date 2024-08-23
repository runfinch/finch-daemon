// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/docker/docker/pkg/jsonmessage"
)

// StreamResponse holds the stream and error messages, if any, for events and status updates.
//
// From https://github.com/moby/moby/blob/v24.0.4/pkg/jsonmessage/jsonmessage.go#L145-L158
type StreamResponse struct {
	Stream string `json:"stream,omitempty"`
	// TODO: Status          string        `json:"status,omitempty"`
	// TODO: Progress        *JSONProgress `json:"progressDetail,omitempty"`
	// TODO: ProgressMessage string        `json:"progress,omitempty"` // deprecated
	// TODO: ID              string        `json:"id,omitempty"`
	// TODO: From            string        `json:"from,omitempty"`
	// TODO: Time            int64         `json:"time,omitempty"`
	// TODO: TimeNano        int64         `json:"timeNano,omitempty"`
	Error        *jsonmessage.JSONError `json:"errorDetail,omitempty"`
	ErrorMessage string                 `json:"error,omitempty"` // deprecated
	// Aux contains out-of-band data, such as digests for push signing and image id after building.
	Aux *json.RawMessage `json:"aux,omitempty"`
}

// StreamWriter allows to write stdout, stderr output as json steaming data in http response.
// This struct implements the Write() function defined in Writer interface which allows StreamWriter
// to capture the output of stdout and stderr and send it to json stream as per docker api spec.
type StreamWriter struct {
	responseWriter http.ResponseWriter
	jsonEncoder    *json.Encoder
	flusher        http.Flusher
	initializer    sync.Once
}

// NewStreamWriter creates a new StreamWriter.
func NewStreamWriter(w http.ResponseWriter) *StreamWriter {
	sw := &StreamWriter{
		responseWriter: w,
		jsonEncoder:    json.NewEncoder(w),
	}
	// check if the writer implements http.Flusher interface and assign it to flusher
	if flusher, ok := w.(http.Flusher); ok && flusher != nil {
		sw.flusher = flusher
	}
	return sw
}

// Write function is implementation of Writer interface. This function converts the stdout and stderr output into
// json stream and send it though http response.
func (sw *StreamWriter) Write(b []byte) (n int, err error) {
	// set response header and status code only once
	sw.initializer.Do(func() {
		sw.responseWriter.Header().Set("Content-Type", "application/json")
		sw.responseWriter.WriteHeader(http.StatusOK)
	})

	err = sw.jsonEncoder.Encode(StreamResponse{Stream: string(b)})
	if err != nil {
		return 0, err
	}
	if sw.flusher != nil {
		// flush after each write so the client can receive status
		// updates as they happen
		sw.flusher.Flush()
	}
	return len(b), nil
}

// WriteError sends a Docker-compatible error message and status code as a
// JSON stream to the http responseWriter.
// If header is not already set, it will set the header and send error message as a response.Error.
func (sw *StreamWriter) WriteError(code int, err error) error {
	// set response header and status code only once
	firstMessage := false
	sw.initializer.Do(func() {
		SendErrorResponse(sw.responseWriter, code, err)
		firstMessage = true
	})

	if firstMessage {
		return nil
	}

	jsonErr := jsonmessage.JSONError{
		Code:    code,
		Message: err.Error(),
	}
	return sw.jsonEncoder.Encode(StreamResponse{
		Error:        &jsonErr,
		ErrorMessage: err.Error(),
	})
}

// WriteAux sends raw data as a Docker-compatible auxiliary response,
// such as digests for pushed image or image id after building.
func (sw *StreamWriter) WriteAux(data []byte) error {
	// set response header and status code only once
	sw.initializer.Do(func() {
		sw.responseWriter.Header().Set("Content-Type", "application/json")
		sw.responseWriter.WriteHeader(http.StatusOK)
	})

	aux := json.RawMessage(data)
	err := sw.jsonEncoder.Encode(StreamResponse{Aux: &aux})
	if err != nil {
		return err
	}
	if sw.flusher != nil {
		sw.flusher.Flush()
	}

	return nil
}

type pullJobWriter struct {
	StreamWriter
	resolved bool
	mx       sync.Mutex
}

// NewPullJobWriter is an extension of StreamWriter that sends image pull status updates from
// nerdctl to http response as a JSON stream. It ensures that the image is resolved before
// writing anything to the response body, so that any resolver errors can be sent to the client
// directly with an appropriate status code.
func NewPullJobWriter(w http.ResponseWriter) *pullJobWriter {
	pw := &pullJobWriter{
		StreamWriter: StreamWriter{
			responseWriter: w,
			jsonEncoder:    json.NewEncoder(w),
		},
		resolved: false,
	}
	// check if the writer implements http.Flusher interface and assign it to flusher
	if flusher, ok := w.(http.Flusher); ok && flusher != nil {
		pw.flusher = flusher
	}
	return pw
}

// Write function is implementation of Writer interface, similar to that of StreamWriter.
// However, until the image is resolved, nothing is written to the response body and header remains unset.
func (pw *pullJobWriter) Write(b []byte) (n int, err error) {
	if !pw.IsResolved() {
		str := string(b)
		if strings.Contains(str, "resolved") {
			pw.Resolve()
		} else {
			return 0, nil
		}
	}

	return pw.StreamWriter.Write(b)
}

func (pw *pullJobWriter) IsResolved() bool {
	pw.mx.Lock()
	defer pw.mx.Unlock()
	return pw.resolved
}

func (pw *pullJobWriter) Resolve() {
	pw.mx.Lock()
	defer pw.mx.Unlock()
	pw.resolved = true
}
