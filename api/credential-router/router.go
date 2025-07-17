// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package credentialrouter

import (
	"net/http"
	"os"

	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	credentialhandler "github.com/runfinch/finch-daemon/api/credential"
	"github.com/runfinch/finch-daemon/pkg/credential"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

// CreateCredentialHandler creates a dedicated HTTP handler for the credential socket.
func CreateCredentialHandler(credentialService *credential.CredentialService, logger flog.Logger, authMiddleware func(http.Handler) http.Handler) (http.Handler, error) {
	r := mux.NewRouter()
	r.Use(authMiddleware)

	// Register the credential handler
	credentialhandler.RegisterHandlers(r, credentialService, logger)
	return ghandlers.LoggingHandler(os.Stderr, r), nil
}
