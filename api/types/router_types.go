// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"net/http"

	"github.com/gorilla/mux"
)

// VersionedRouter wraps the router to provide a router that can redirect routes to either a versioned
// path or a non-versioned path.
type VersionedRouter struct {
	Router *mux.Router
	prefix string
}

// HandleFunc replaces the router.HandleFunc function to create a new handler and direct certain paths
// to certain functions.
func (vr *VersionedRouter) HandleFunc(path string, f func(http.ResponseWriter, *http.Request), methods ...string) {
	vr.Router.HandleFunc(vr.prefix+path, f).Methods(methods...)
	vr.Router.HandleFunc("/v{version}"+vr.prefix+path, f).Methods(methods...)
}

// SetPrefix sets the prefix of the route, so that any specified routes can just specify the endpoint only.
func (vr *VersionedRouter) SetPrefix(prefix string) {
	vr.prefix = prefix
}
