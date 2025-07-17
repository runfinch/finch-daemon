// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config provides shared configuration used throughout the application
package config

import "sync"

const (
	DefaultFinchAddr = "/run/finch.sock"
	DefaultCredentialAddr = "/run/finch-credential.sock"
	DefaultNamespace = "finch"
	DefaultConfigPath = "/etc/finch/finch.toml"
	DefaultPidFile = "/run/finch.pid"
)

var (
	credentialAddr string = DefaultCredentialAddr
	mu            sync.RWMutex
)

// SetCredentialAddr sets the credential address to be used at runtime.
func SetCredentialAddr(addr string) {
	mu.Lock()
	defer mu.Unlock()
	credentialAddr = addr
}

// GetCredentialAddr returns the current credential address.
func GetCredentialAddr() string {
	mu.RLock()
	defer mu.RUnlock()
	return credentialAddr
}
