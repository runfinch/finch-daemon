// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"encoding/json"
	"net/http"

	"github.com/docker/docker/pkg/jsonmessage"
)

// Error implements the error structure used in Docker Engine API.
type Error struct {
	Message string `json:"message"`
}

func NewError(err error) *Error {
	return &Error{
		Message: err.Error(),
	}
}

func NewErrorFromMsg(msg string) *Error {
	return &Error{
		Message: msg,
	}
}

// SendErrorResponse sends a status code and a json-formatted error message (if any).
func SendErrorResponse(w http.ResponseWriter, code int, err error) {
	if code == http.StatusNotModified {
		// only set the status code as no content and not modified does not have response body
		// for more details see https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/304
		w.WriteHeader(code)
	} else {
		JSON(w, code, NewError(err))
	}
}

// SendErrorAsInStreamResponse sends a status code and a json-formatted error message in stream response.
func SendErrorAsInStreamResponse(w http.ResponseWriter, code int, err error) error {
	jsonErr := jsonmessage.JSONError{
		Code:    code,
		Message: err.Error(),
	}
	jw := json.NewEncoder(w)
	return jw.Encode(StreamResponse{
		Error:        &jsonErr,
		ErrorMessage: err.Error(),
	})
}
