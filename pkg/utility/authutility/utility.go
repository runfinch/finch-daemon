// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package authutility

import (
	"fmt"
	"strings"

	dockertypes "github.com/docker/cli/cli/config/types"

	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	"github.com/runfinch/finch-daemon/internal/backend"
)

const IndexServer = "https://index.docker.io/v1/"

// GetAuthCreds returns authentication credentials resolver function from image reference domain and auth config.
func GetAuthCreds(refDomain string, containerdClient backend.ContainerdClient, ac dockertypes.AuthConfig) (dockerconfigresolver.AuthCreds, error) {
	// return nil if no credentials specified
	if ac.Username == "" && ac.Password == "" && ac.IdentityToken == "" && ac.RegistryToken == "" {
		return nil, nil
	}

	// domain expected by the authcreds function
	// DefaultHost converts "docker.io" to "registry-1.docker.io"
	expectedDomain, err := containerdClient.DefaultDockerHost(refDomain)
	if err != nil {
		return nil, err
	}

	// ensure that server address matches the image reference domain
	sa := ac.ServerAddress
	if sa != "" {
		saHostname := convertToHostname(sa)
		// "registry-1.docker.io" can show up as "https://index.docker.io/v1/" in ServerAddress
		if expectedDomain == "registry-1.docker.io" {
			if saHostname != refDomain && sa != IndexServer {
				return nil, fmt.Errorf("specified server address %s does not match the image reference domain %s", sa, refDomain)
			}
		} else if saHostname != refDomain {
			return nil, fmt.Errorf("specified server address %s does not match the image reference domain %s", sa, refDomain)
		}
	}

	// return auth creds function
	return func(domain string) (string, string, error) {
		if domain != expectedDomain {
			return "", "", fmt.Errorf("expected domain %s, but got %s", expectedDomain, domain)
		}
		if ac.IdentityToken != "" {
			return "", ac.IdentityToken, nil
		} else {
			return ac.Username, ac.Password, nil
		}
	}, nil
}

// convertToHostname converts a registry url which has http|https prepended
// to just an hostname.
// Copied from github.com/docker/docker/registry.ConvertToHostname to reduce dependencies.
func convertToHostname(url string) string {
	stripped := url
	if strings.HasPrefix(url, "http://") {
		stripped = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		stripped = strings.TrimPrefix(url, "https://")
	}

	hostName, _, _ := strings.Cut(stripped, "/")
	return hostName
}
