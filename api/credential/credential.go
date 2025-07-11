// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package credential contains functions and structures related to credential management APIs
package credential

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/runfinch/finch-daemon/api/response"
	credService "github.com/runfinch/finch-daemon/pkg/credential"
	"github.com/runfinch/finch-daemon/pkg/flog"
	"github.com/shirou/gopsutil/v3/process"
)

type handler struct {
	service *credService.CredentialService
	logger  flog.Logger
}

// BuildRequestAuthMiddleware checks peercreds for incoming request for build registry credentials.
func BuildRequestAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		netConn := GetConn(r)
		logger := flog.NewLogrus()

		// Validate the connection credentials
		if err := validatePeerCreds(netConn); err != nil {
			logger.Errorf("failed to get connection %v", err)
			response.SendErrorResponse(w, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RegisterHandlers sets up the credential handlers.
func RegisterHandlers(r *mux.Router, service *credService.CredentialService, logger flog.Logger) {
	h := newHandler(service, logger)
	r.HandleFunc("/finch/credentials", h.getCredentials).Methods(http.MethodGet)
}

func newHandler(service *credService.CredentialService, logger flog.Logger) *handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

// getCredentials handles credential requests from the credential helper.
func (h *handler) getCredentials(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var req struct {
		BuildID    string `json:"buildID"`
		ServerAddr string `json:"serverAddr"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request")
		response.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to decode request"))
		return
	}

	// Validate the request
	if req.BuildID == "" {
		h.logger.Errorf("Request rejected: missing build ID")
		response.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("request rejected: missing build ID"))
		return
	}

	if req.ServerAddr == "" {
		h.logger.Errorf("Request rejected: missing server address")
		response.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("request rejected: missing server address"))
		return
	}

	// Get the credentials
	authConfig, err := h.service.GetCredentials(r.Context(), req.BuildID, req.ServerAddr)
	if err != nil {
		h.logger.Errorf("Failed to get credentials")
		response.SendErrorResponse(w, http.StatusNotFound, fmt.Errorf("failed to get credentials"))
		return
	}

	// Return the full AuthConfig object (which is already dockertypes.AuthConfig)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authConfig); err != nil {
		h.logger.Errorf("Failed to encode response")
		response.SendErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to encode response"))
		return
	}
}

// validatePeerCreds validates peer credentials of a connection.
// This function implements security check to ensure that credential requests
// are only accepted from authorized processes within the finch ecosystem:
//
//  1. UID/GID matching: Verifies the connecting process runs under the same user/group
//     as the daemon, preventing cross-user access attempts.
//
//  2. Parent-parent PID matching: Implements process ancestry validation by checking that
//     the parent parent's of the connecting process is the current daemon process.
//     This works because the typical process hierarchy during a build is:
//     finch-daemon (this process) -> buildkit -> credential-helper -> connection.
//
//     By matching the parent's parent PID against our own PID,
//     we ensure the request originates from a legitimate build process spawned by
//     this daemon instance, preventing:
//     - Processes from other UID/GID impersonating credential helpers.
//     - Unauthorized access from unrelated processes with matching UID/GID.
func validatePeerCreds(conn net.Conn) error {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return errors.New("connection is not a Unix socket connection")
	}

	rawConn, err := unixConn.SyscallConn()
	if err != nil {
		return err
	}

	var (
		pid     int
		uid     int
		gid     int
		connErr error
	)

	err = rawConn.Control(func(fd uintptr) {
		uid, gid, pid, connErr = getPeerCredentials(int(fd))
	})

	if err != nil || connErr != nil {
		return fmt.Errorf("failed to get peer credentials")
	}

	currentUID := os.Getuid()
	currentGID := os.Getgid()

	if uid != currentUID {
		return errors.New("unauthorized access")
	}

	if gid != currentGID {
		return errors.New("unauthorized access")
	}

	// Validate process ancestry: connecting process should be a descendant of this daemon
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return errors.New("internal error")
	}

	// Get parent process (typically buildkit)
	pp, err := p.Parent()
	if err != nil {
		return fmt.Errorf("internal error")
	}

	// Get parent's parent process (should match current finch-daemon pid)
	ppp, err := pp.Parent()
	if err != nil {
		return fmt.Errorf("internal error")
	}

	// Verify the great-grandparent is this daemon process
	if int(ppp.Pid) != os.Getpid() {
		return errors.New("unauthorized access")
	}

	return nil
}

// getPeerCredentials gets the credentials from a socket connection.
func getPeerCredentials(fd int) (int, int, int, error) {
	// GetsockoptUcred is used to get the uid, gid and pid of the established connection.
	cred, err := syscall.GetsockoptUcred(fd, syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return 0, 0, 0, err
	}

	return int(cred.Uid), int(cred.Gid), int(cred.Pid), nil
}

type contextKey struct {
	key string
}

var ConnContextKey = &contextKey{"cred-conn-ctx"}

func SetConn(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, ConnContextKey, c)
}
func GetConn(r *http.Request) net.Conn {
	return r.Context().Value(ConnContextKey).(net.Conn)
}
