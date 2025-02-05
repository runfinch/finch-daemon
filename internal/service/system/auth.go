// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/containerd/containerd/v2/core/remotes/docker"
	dockerconfig "github.com/containerd/containerd/v2/core/remotes/docker/config"
	remoteerrs "github.com/containerd/containerd/v2/core/remotes/errors"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/dockerconfigresolver"
	"golang.org/x/net/context/ctxhttp"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// To be consistent with nerdctl: https://github.com/containerd/nerdctl/blob/2b06050d782c27571c98947ac9fa790d5f2d0bde/cmd/nerdctl/login.go#L90
const defaultRegHost = "https://index.docker.io/v1/"

func (s *service) Auth(ctx context.Context, username, password, serverAddr string) (string, error) {
	if serverAddr == "" {
		serverAddr = defaultRegHost
	}

	host, err := dockerconfigresolver.Parse(serverAddr)
	if err != nil {
		return "", fmt.Errorf("failed to parse server address: %v", err)
	}
	// TODO: Support server addr that starts with "http://" (probably useful when testing)
	// Currently TLS is enforced.
	// Check dockerconfigresolver.WithSkipVerifyCerts and dockerconfigresolver.WithPlainHTTP.
	ho, err := dockerconfigresolver.NewHostOptions(ctx, host.CanonicalIdentifier(), dockerconfigresolver.WithAuthCreds(
		func(acArg string) (string, string, error) {
			if acArg == host.CanonicalIdentifier() {
				return username, password, nil
			}
			return "", "", fmt.Errorf("expected acArg to be %q, got %q", host, acArg)
		},
	))
	if err != nil {
		return "", fmt.Errorf("failed to initialize host options: %v", err)
	}

	fetchedRefreshTokens := make(map[string]string)
	// TODO: Support ad-hoc host certs:
	// https://github.com/containerd/nerdctl/blob/1c4029a79bdcb4f728b5bdc534aec36e13a3d2ac/cmd/nerdctl/login.go#L223
	// By default, the host's root CA set is used, so usually this is not needed because
	// established registries's certificates (e.g., ECR, Docker Hub, etc.) should work just fine.
	// This probably comes in handy when testing (e.g., spin up a registry locally with a self-signed certificate).
	ho.AuthorizerOpts = append(ho.AuthorizerOpts, docker.WithFetchRefreshToken(
		func(ctx context.Context, token string, req *http.Request) {
			fetchedRefreshTokens[req.URL.Host] = token
		},
	))
	regHosts, err := dockerconfig.ConfigureHosts(ctx, *ho)(host.CanonicalIdentifier())
	if err != nil {
		return "", fmt.Errorf("failed to configure registry host: %w", err)
	}

	for _, rh := range regHosts {
		if err = loginRegHost(ctx, rh); err != nil {
			log.Printf("failed to log in registry host %s: %v", rh.Host, err)
			continue
		}
		// It's possible that the token is empty:
		// https://github.com/containerd/containerd/blob/5d4276cc34ddd20454caaae23824b73b6c6907c1/remotes/docker/authorizer.go#L126
		return fetchedRefreshTokens[rh.Host], nil
	}
	return "", fmt.Errorf("failed to log in to all the registry hosts, last err: %w", err)
}

func loginRegHost(ctx context.Context, rh docker.RegistryHost) error {
	if rh.Authorizer == nil {
		return errors.New("got nil Authorizer")
	}
	// Why we need to manually add the slash here: https://github.com/containerd/containerd/pull/6470#issuecomment-1020664375
	if rh.Path == "/v2" {
		// What this endpoint is about: https://github.com/opencontainers/distribution-spec/blob/main/spec.md#determining-support
		rh.Path = "/v2/"
	}
	u := url.URL{
		Scheme: rh.Scheme,
		Host:   rh.Host,
		Path:   rh.Path,
	}
	var resps []*http.Response
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		for k, v := range rh.Header.Clone() {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}
		if err := rh.Authorizer.Authorize(ctx, req); err != nil {
			var usErr remoteerrs.ErrUnexpectedStatus
			if errors.As(err, &usErr) {
				if usErr.StatusCode == http.StatusUnauthorized {
					err = errdefs.NewUnauthenticated(err)
				}
			}
			return fmt.Errorf("failed to call rh.Authorizer.Authorize: %w", err)
		}
		resp, err := ctxhttp.Do(ctx, rh.Client, req)
		if err != nil {
			return fmt.Errorf("failed to call rh.Client.Do: %w", err)
		}
		log.Printf("trial %d, status code: %d", i, resp.StatusCode)
		resps = append(resps, resp)
		if resp.StatusCode == http.StatusUnauthorized {
			// TODO: figure out why the first request is always 401, and suddenly the second request will be 200.
			// Maybe AddResponses does some magic.
			if err := rh.Authorizer.AddResponses(ctx, resps); err != nil {
				return fmt.Errorf("failed to call rh.Authorizer.AddResponses: %w", err)
			}
			continue
		}
		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
		return nil
	}
	return errdefs.NewUnauthenticated(errors.New("too many 401 (probably)"))
}
