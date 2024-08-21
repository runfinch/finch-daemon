// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// Status writes the supplied HTTP status code to response header.
func Status(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

// JSON writes data as JSON object(s) to response with the supplied HTTP status code.
func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to JSON-encode response: %v", err)
	}
}
