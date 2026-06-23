// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
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
