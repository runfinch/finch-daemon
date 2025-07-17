// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package credential consists of definition of service structures and methods related to credential management
package credential

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"sync"

	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker-credential-helpers/registryurl"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

// No constants needed as we're removing TTL-based expiration.
type CredentialCache struct {
	Entries map[string]credentialEntry
	Mutex   sync.RWMutex
}

// NewCredentialCache creates a new shared credential cache.
func NewCredentialCache() *CredentialCache {
	return &CredentialCache{
		Entries: make(map[string]credentialEntry),
	}
}

// credentialEntry represents a set of credentials for a build.
type credentialEntry struct {
	credentials map[string]dockertypes.AuthConfig
}

// service implements the credential.Service interface.
// The service uses a shared cache that is passed in from main.
type CredentialService struct {
	cache  *CredentialCache
	logger flog.Logger
}

// NewCredentialService creates a new credential service with a shared cache.
func NewCredentialService(logger flog.Logger, cache *CredentialCache) *CredentialService {
	return &CredentialService{
		cache:  cache,
		logger: logger,
	}
}

// GenerateBuildID creates a cryptographically secure random build ID using crypto/rand.
func (s *CredentialService) GenerateBuildID() (string, error) {
	id := make([]byte, 32)
	_, err := rand.Read(id)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}

// StoreAuthConfigs stores AuthConfig objects for a build ID.
func (s *CredentialService) StoreAuthConfigs(ctx context.Context, buildID string, authConfigs map[string]dockertypes.AuthConfig) error {
	s.cache.Mutex.Lock()
	defer s.cache.Mutex.Unlock()

	s.cache.Entries[buildID] = credentialEntry{
		credentials: authConfigs,
	}
	return nil
}

// GetCredentials retrieves credentials for a build ID and server address.
func (s *CredentialService) GetCredentials(ctx context.Context, buildID string, serverAddr string) (dockertypes.AuthConfig, error) {
	s.cache.Mutex.Lock()
	defer s.cache.Mutex.Unlock()

	entry, exists := s.cache.Entries[buildID]
	if !exists {
		return dockertypes.AuthConfig{}, fmt.Errorf("no credentials found")
	}

	target, err := s.getTarget(serverAddr, entry.credentials)
	if err != nil {
		s.logger.Errorf("Error finding target for server: %v", err)
		return dockertypes.AuthConfig{}, err
	}

	if target == "" {
		if _, fallbackExists := entry.credentials[serverAddr]; !fallbackExists {
			s.logger.Errorf("No credentials found for server")
			return dockertypes.AuthConfig{}, fmt.Errorf("no credentials found for server")
		}
		target = serverAddr
	}

	authConfig, exists := entry.credentials[target]
	if !exists {
		s.logger.Errorf("No credentials found for matched target %s", target)
		return dockertypes.AuthConfig{}, fmt.Errorf("no credentials found for matched target %s", target)
	}

	return authConfig, nil
}

// RemoveCredentials removes credentials for a build ID.
func (s *CredentialService) RemoveCredentials(buildID string) error {
	s.cache.Mutex.Lock()
	defer s.cache.Mutex.Unlock()

	if _, exists := s.cache.Entries[buildID]; !exists {
		return fmt.Errorf("no credentials found")
	}

	delete(s.cache.Entries, buildID)

	return nil
}

func (s *CredentialService) getTarget(serverURL string, creds map[string]dockertypes.AuthConfig) (string, error) {
	server, err := registryurl.Parse(serverURL)
	if err != nil {
		return serverURL, nil
	}

	var targets []string
	for cred := range creds {
		targets = append(targets, cred)
	}

	if target, found := s.findMatch(server, targets, s.exactMatch); found {
		return target, nil
	}

	if target, found := s.findMatch(server, targets, s.approximateMatch); found {
		return target, nil
	}

	return "", nil
}

func (s *CredentialService) exactMatch(serverURL, target url.URL) bool {
	return serverURL.String() == target.String()
}

func (s *CredentialService) approximateMatch(serverURL, target url.URL) bool {
	// Ignore scheme differences by using target's scheme.
	serverURL.Scheme = target.Scheme

	if serverURL.Port() == "" && target.Port() != "" {
		serverURL.Host = serverURL.Host + ":" + target.Port()
	}

	if serverURL.Path == "" {
		serverURL.Path = target.Path
	}
	return s.exactMatch(serverURL, target)
}

// findMatch is a helper function that tries to match a serverURL against a list of targets using the provided matching function.
func (s *CredentialService) findMatch(serverUrl *url.URL, targets []string, matches func(url.URL, url.URL) bool) (string, bool) {
	for _, target := range targets {
		tURL, err := registryurl.Parse(target)
		if err != nil {
			continue
		}
		if matches(*serverUrl, *tURL) {
			return target, true
		}
	}
	return "", false
}
