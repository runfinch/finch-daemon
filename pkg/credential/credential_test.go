// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package credential

import (
	"context"
	"fmt"
	"sync"
	"testing"

	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/runfinch/finch-daemon/mocks/mocks_logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCredentialService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks_logger.NewLogger(ctrl)
	cache := NewCredentialCache()
	service := NewCredentialService(logger, cache)

	assert.NotNil(t, service, "CredentialService should not be nil")
	assert.NotNil(t, cache, "CredentialCache should not be nil")
	assert.NotNil(t, cache.Entries, "Entries map should be initialized")
	assert.Empty(t, cache.Entries, "Entries map should be empty")
	assert.Equal(t, cache, service.cache, "Service cache should match the provided cache")
}

func TestGenerateBuildID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks_logger.NewLogger(ctrl)
	cache := NewCredentialCache()
	service := NewCredentialService(logger, cache)

	id1, err1 := service.GenerateBuildID()
	assert.NoError(t, err1, "GenerateBuildID should not return an error")
	assert.NotEmpty(t, id1, "Generated ID should not be empty")

	id2, err2 := service.GenerateBuildID()
	assert.NoError(t, err2, "Second GenerateBuildID should not return an error")
	assert.NotEmpty(t, id2, "Second generated ID should not be empty")
	assert.NotEqual(t, id1, id2, "Generated IDs should be unique")
}

func TestStoreAuthConfigs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks_logger.NewLogger(ctrl)
	cache := NewCredentialCache()
	service := NewCredentialService(logger, cache)
	ctx := context.Background()
	buildID := "test-build-id"

	authConfigs := map[string]dockertypes.AuthConfig{
		"registry1.example.com": {
			Username: "user1",
			Password: "pass1",
		},
		"registry2.example.com": {
			Username: "user2",
			Password: "pass2",
		},
	}

	err := service.StoreAuthConfigs(ctx, buildID, authConfigs)
	assert.NoError(t, err, "StoreAuthConfigs should not return an error")

	// Verify the configs were stored correctly
	entry, exists := cache.Entries[buildID]
	assert.True(t, exists, "Entry should exist for the build ID")
	assert.Equal(t, authConfigs, entry.credentials, "Stored credentials should match the provided configs")

	// Test overwriting existing entry
	newAuthConfigs := map[string]dockertypes.AuthConfig{
		"registry3.example.com": {
			Username: "user3",
			Password: "pass3",
		},
	}

	err = service.StoreAuthConfigs(ctx, buildID, newAuthConfigs)
	assert.NoError(t, err, "StoreAuthConfigs should not return an error when overwriting")

	// Verify the configs were updated
	entry, exists = cache.Entries[buildID]
	assert.True(t, exists, "Entry should still exist for the build ID")
	assert.Equal(t, newAuthConfigs, entry.credentials, "Stored credentials should be updated")
}

func TestGetCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks_logger.NewLogger(ctrl)
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	cache := NewCredentialCache()
	service := NewCredentialService(logger, cache)
	ctx := context.Background()
	buildID := "test-build-id"

	authConfigs := map[string]dockertypes.AuthConfig{
		"https://registry.example.com/v2/": {
			Username:      "user1",
			Password:      "pass1",
			ServerAddress: "https://registry.example.com/v2/",
		},
		"https://index.docker.io:443/v1/": {
			Username:      "user2",
			Password:      "pass2",
			ServerAddress: "https://index.docker.io:443/v1/",
		},
		"https://gcr.io/v1/": {
			Username:      "user3",
			Password:      "pass3",
			ServerAddress: "https://gcr.io/v1/",
		},
	}
	err := service.StoreAuthConfigs(ctx, buildID, authConfigs)
	assert.NoError(t, err, "StoreAuthConfigs should succeed")

	testCases := []struct {
		name         string
		requestURL   string
		expectUser   string
		expectPass   string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "exact match",
			requestURL:  "https://registry.example.com/v2/",
			expectUser:  "user1",
			expectPass:  "pass1",
			expectError: false,
		},
		{
			name:        "different scheme",
			requestURL:  "http://registry.example.com/v2/",
			expectUser:  "user1",
			expectPass:  "pass1",
			expectError: false,
		},
		{
			name:        "missing scheme",
			requestURL:  "registry.example.com/v2/",
			expectUser:  "user1",
			expectPass:  "pass1",
			expectError: false,
		},
		{
			name:        "missing port",
			requestURL:  "https://index.docker.io/v1/",
			expectUser:  "user2",
			expectPass:  "pass2",
			expectError: false,
		},
		{
			name:        "missing path",
			requestURL:  "https://gcr.io",
			expectUser:  "user3",
			expectPass:  "pass3",
			expectError: false,
		},
		{
			name:         "no match",
			requestURL:   "https://quay.io",
			expectUser:   "",
			expectPass:   "",
			expectError:  true,
			errorMessage: "no credentials found for server",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// For the invalid build ID test case, use a different build ID
			// Store the credentials
			testBuildID := buildID

			auth, err := service.GetCredentials(ctx, testBuildID, tc.requestURL)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectUser, auth.Username)
				assert.Equal(t, tc.expectPass, auth.Password)
			}
		})
	}
}

func TestRemoveCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks_logger.NewLogger(ctrl)
	cache := NewCredentialCache()
	service := NewCredentialService(logger, cache)
	ctx := context.Background()
	buildID := "test-build-id"

	// Store some test credentials
	authConfigs := map[string]dockertypes.AuthConfig{
		"registry.example.com": {
			Username: "user1",
			Password: "pass1",
		},
	}

	err := service.StoreAuthConfigs(ctx, buildID, authConfigs)
	assert.NoError(t, err, "StoreAuthConfigs should succeed")

	_, exists := cache.Entries[buildID]
	assert.True(t, exists, "Entry should exist before removal")

	err = service.RemoveCredentials(buildID)
	assert.NoError(t, err, "RemoveCredentials should succeed")

	_, exists = cache.Entries[buildID]
	assert.False(t, exists, "Entry should not exist after removal")

	err = service.RemoveCredentials("non-existent-id")
	assert.Error(t, err, "RemoveCredentials should fail for non-existent ID")
	assert.Contains(t, err.Error(), "no credentials found")
}

func TestParallelCredentialOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks_logger.NewLogger(ctrl)
	cache := NewCredentialCache()
	service := NewCredentialService(logger, cache)
	ctx := context.Background()

	const numBuildIDs = 1000
	const numRegistriesPerBuild = 5
	const numConcurrentOperations = 100

	buildIDs := make([]string, numBuildIDs)
	for i := 0; i < numBuildIDs; i++ {
		id, err := service.GenerateBuildID()
		assert.NoError(t, err)
		buildIDs[i] = id
	}

	var wg sync.WaitGroup

	t.Run("Parallel Store", func(t *testing.T) {
		wg.Add(numBuildIDs)

		for i := 0; i < numBuildIDs; i++ {
			go func(buildIdx int) {
				defer wg.Done()
				buildID := buildIDs[buildIdx]

				authConfigs := make(map[string]dockertypes.AuthConfig)
				for j := 0; j < numRegistriesPerBuild; j++ {
					registry := fmt.Sprintf("registry%d-%d.example.com", buildIdx, j)
					authConfigs[registry] = dockertypes.AuthConfig{
						Username:      fmt.Sprintf("user%d-%d", buildIdx, j),
						Password:      fmt.Sprintf("pass%d-%d", buildIdx, j),
						ServerAddress: registry,
					}
				}

				err := service.StoreAuthConfigs(ctx, buildID, authConfigs)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// Verify all entries were created
		assert.Equal(t, numBuildIDs, len(cache.Entries), "All entries should be stored correctly")
	})

	// Add stress test for concurrent GetCredentials
	t.Run("Parallel Get", func(t *testing.T) {
		wg = sync.WaitGroup{}
		wg.Add(numConcurrentOperations)

		for i := 0; i < numConcurrentOperations; i++ {
			go func(operationIdx int) {
				defer wg.Done()

				for j := 0; j < numBuildIDs/numConcurrentOperations; j++ {
					buildIdx := (operationIdx*numBuildIDs/numConcurrentOperations + j) % numBuildIDs
					buildID := buildIDs[buildIdx]

					for k := 0; k < numRegistriesPerBuild; k++ {
						registry := fmt.Sprintf("registry%d-%d.example.com", buildIdx, k)
						auth, err := service.GetCredentials(ctx, buildID, registry)
						assert.NoError(t, err)
						assert.Equal(t, fmt.Sprintf("user%d-%d", buildIdx, k), auth.Username)
						assert.Equal(t, fmt.Sprintf("pass%d-%d", buildIdx, k), auth.Password)
					}
				}
			}(i)
		}

		wg.Wait()
	})

	// Add stress test for concurrent Remove operations
	t.Run("Parallel Remove", func(t *testing.T) {
		wg = sync.WaitGroup{}
		wg.Add(numConcurrentOperations)

		for i := 0; i < numConcurrentOperations; i++ {
			go func(operationIdx int) {
				defer wg.Done()

				startIdx := operationIdx * numBuildIDs / numConcurrentOperations
				endIdx := (operationIdx + 1) * numBuildIDs / numConcurrentOperations

				for buildIdx := startIdx; buildIdx < endIdx; buildIdx++ {
					buildID := buildIDs[buildIdx]
					err := service.RemoveCredentials(buildID)
					assert.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify all entries were removed
		assert.Equal(t, 0, len(cache.Entries), "All entries should be removed correctly")
	})
}
