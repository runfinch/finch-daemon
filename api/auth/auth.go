// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

// Copied from https://github.com/moby/moby/blob/master/api/types/registry/authconfig.go
// TODO: Revisit this as this to remove things that don't make sense to Finch.
//       Likely we should move this file to some auth package as it should be used by more than /push
//       (e.g., pulling an image from a private registry).

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/pkg/errors"
)

// AuthHeader is the name of the header used to send encoded registry
// authorization credentials for registry operations (push/pull).
const AuthHeader = "X-Registry-Auth"

// DecodeAuthConfig decodes base64url encoded (RFC4648, section 5) JSON
// authentication information as sent through the X-Registry-Auth header.
//
// This function always returns an AuthConfig, even if an error occurs. It is up
// to the caller to decide if authentication is required, and if the error can
// be ignored.
//
// For details on base64url encoding, see:
// - RFC4648, section 5:   https://tools.ietf.org/html/rfc4648#section-5
func DecodeAuthConfig(authEncoded string) (*dockertypes.AuthConfig, error) {
	if authEncoded == "" {
		return &dockertypes.AuthConfig{}, nil
	}

	authJSON := base64.NewDecoder(base64.URLEncoding, strings.NewReader(authEncoded))
	return decodeAuthConfigFromReader(authJSON)
}

// DecodeAuthConfigBody decodes authentication information as sent as JSON in the
// body of a request. This function is to provide backward compatibility with old
// clients and API versions. Current clients and API versions expect authentication
// to be provided through the X-Registry-Auth header.
//
// Like DecodeAuthConfig, this function always returns an AuthConfig, even if an
// error occurs. It is up to the caller to decide if authentication is required,
// and if the error can be ignored.
func DecodeAuthConfigBody(rdr io.ReadCloser) (*dockertypes.AuthConfig, error) {
	return decodeAuthConfigFromReader(rdr)
}

func decodeAuthConfigFromReader(rdr io.Reader) (*dockertypes.AuthConfig, error) {
	authConfig := &dockertypes.AuthConfig{}
	if err := json.NewDecoder(rdr).Decode(authConfig); err != nil {
		// always return an (empty) AuthConfig to increase compatibility with
		// the existing API.
		return &dockertypes.AuthConfig{}, errors.Wrap(err, "invalid X-Registry-Auth header")
	}
	return authConfig, nil
}
