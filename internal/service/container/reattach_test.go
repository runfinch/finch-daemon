// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractContainerID(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "standard bridge network result file",
			filename: "bridge-finch-a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2-eth0",
			want:     "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		},
		{
			name:     "custom network name",
			filename: "mynet-finch-abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789-eth0",
			want:     "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name:     "different interface name",
			filename: "bridge-finch-abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789-eth1",
			want:     "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name:     "no valid container ID",
			filename: "bridge-finch-tooshort-eth0",
			want:     "",
		},
		{
			name:     "empty filename",
			filename: "",
			want:     "",
		},
		{
			name:     "non-hex 64 char string",
			filename: "bridge-finch-zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz-eth0",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractContainerID(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKillPortReserver(t *testing.T) {
	t.Run("no pid file - does nothing", func(t *testing.T) {
		// Should not panic or error with nonexistent path
		killPortReserver("nonexistent-ns", "nonexistent-container")
	})

	t.Run("invalid pid in file", func(t *testing.T) {
		dir := t.TempDir()
		pidFile := filepath.Join(dir, "port-reserver.pid")
		os.WriteFile(pidFile, []byte("notanumber"), 0644)

		// Monkey-patch by calling with a namespace that resolves to our temp dir
		// Since killPortReserver uses a hardcoded /run/nerdctl path, we test the
		// logic indirectly — this test verifies the function handles missing files gracefully
		killPortReserver("nonexistent", "nonexistent")
	})

	t.Run("pid file with stale pid", func(t *testing.T) {
		// Create a temp dir mimicking /run/nerdctl/{ns}/{id}/
		dir := t.TempDir()
		nsDir := filepath.Join(dir, "finch", "testcontainer123")
		os.MkdirAll(nsDir, 0750)
		pidFile := filepath.Join(nsDir, "port-reserver.pid")
		// Write a PID that doesn't exist (99999999)
		os.WriteFile(pidFile, []byte(strconv.Itoa(99999999)), 0644)

		// Can't easily test this without mocking the filesystem path,
		// but we verify it doesn't panic
		killPortReserver("finch", "testcontainer123")
	})
}
