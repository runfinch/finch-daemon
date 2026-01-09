// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tests contains the exported functions that are meant to be imported as test cases.
//
// It should not export any other thing except for a SubcommandOption struct (e.g., LoginOption) that may be added in the future.
//
// Each file contains one subcommand to test and is named after that subcommand.
// Note that the file names are not suffixed with _test so that they can appear in Go Doc.
package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/runfinch/common-tests/ffs"
	"github.com/runfinch/common-tests/fnet"

	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/e2e/client"
)

const (
	alpineImage              = "public.ecr.aws/docker/library/alpine:latest"
	newerAlpineImage         = "public.ecr.aws/docker/library/alpine:3.21"
	olderAlpineImage         = "public.ecr.aws/docker/library/alpine:3.13"
	amazonLinux2Image        = "public.ecr.aws/amazonlinux/amazonlinux:2"
	nginxImage               = "public.ecr.aws/docker/library/nginx:latest"
	testImageName            = "test:tag"
	nonexistentImageName     = "ne-repo:ne-tag"
	nonexistentContainerName = "ne-ctr"
	testContainerName        = "ctr-test"
	testContainerName2       = "ctr-test-2"
	testVolumeName           = "testVol"
	testVolumeName2          = "anotherTestVol"
	registryImage            = "public.ecr.aws/docker/library/registry:latest"
	localRegistryName        = "local-registry"
	testUser                 = "testUser"
	testPassword             = "testPassword"
	sha256RegexFull          = "^sha256:[a-f0-9]{64}$"
	bridgeNetwork            = "bridge"
	testNetwork              = "test-network"
)

var defaultImage = alpineImage

// CGMode is the cgroups mode of the host system.
// We copy the struct from containerd/cgroups [1] instead of using it as a library
// because it only builds on linux,
// while we don't really need the functions that make it only build on linux
// (e.g., determine the cgroup version of the current host).
//
// [1] https://github.com/containerd/cgroups/blob/cc78c6c1e32dc5bde018d92999910fdace3cfa27/utils.go#L38-L50
type CGMode int

const (
	// Unavailable cgroup mountpoint.
	Unavailable CGMode = iota
	// Legacy cgroups v1.
	Legacy
	// Hybrid with cgroups v1 and v2 controllers mounted.
	Hybrid
	// Unified with only cgroups v2 mounted.
	Unified
)

// SetupLocalRegistry can be invoked before running the tests to save time when pulling defaultImage.
//
// It spins up a local registry, tags the alpine image, pushes the tagged image to local registry,
// and changes defaultImage to be the one pushed to local registry.
//
// After all the tests are done, invoke CleanupLocalRegistry to clean up the local registry.
func SetupLocalRegistry() {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	httpRemoveAll(uClient, version)

	hostPort := fnet.GetFreePort()
	options := types.ContainerCreateRequest{
		ContainerConfig: types.ContainerConfig{
			Image: registryImage,
		},
		HostConfig: types.ContainerHostConfig{
			PortBindings: nat.PortMap{
				"5000/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", hostPort)}},
			},
		},
	}

	// Try to create the registry container — skip registry setup if OPA blocks container create.
	containerID, ok := httpTryCreateContainer(uClient, version, localRegistryName, options)
	if !ok {
		return
	}
	httpStartContainer(uClient, version, containerID)

	// Wait for the registry to be ready before pushing.
	waitForRegistry(hostPort)

	httpPullImage(uClient, version, alpineImage)
	defaultImage = fmt.Sprintf("localhost:%d/alpine:latest", hostPort)
	httpTagImage(uClient, version, alpineImage, defaultImage)
	httpPushImage(uClient, version, defaultImage)
	httpRemoveImage(uClient, version, alpineImage)
}

// waitForRegistry polls the registry's /v2/ endpoint until it returns 200 or times out.
// This is a direct TCP connection to the registry container's exposed port, not through the daemon.
func waitForRegistry(hostPort int) {
	url := fmt.Sprintf("http://localhost:%d/v2/", hostPort)
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(url) //nolint:noctx // registry readiness poll does not need a context
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	gomega.Expect(false).To(gomega.BeTrue(), fmt.Sprintf("registry at localhost:%d never became ready", hostPort))
}

// httpTryCreateContainer attempts to create a container and returns (id, true) on success,
// or ("", false) if the request is forbidden (e.g. OPA middleware blocks container create).
func httpTryCreateContainer(uClient *http.Client, version, containerName string, options types.ContainerCreateRequest) (string, bool) {
	url := client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/create?name=%s", containerName))
	reqBody, err := json.Marshal(options)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(url, "application/json", bytes.NewReader(reqBody))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return "", false
	}
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))
	var ctr struct {
		ID string `json:"Id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&ctr)
	gomega.Expect(err).Should(gomega.BeNil())
	return ctr.ID, true
}

// CleanupLocalRegistry removes the local registry container and image. It's used together with SetupLocalRegistry,
// and should be invoked after running all the tests.
func CleanupLocalRegistry() {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()

	// Remove container by name - httpRemoveContainerForce handles the case where container doesn't exist
	httpRemoveContainerForce(uClient, version, localRegistryName)

	// Remove all images via HTTP API
	images := httpListImages(uClient, version)
	for _, img := range images {
		if img.ID != "" {
			httpRemoveImageForce(uClient, version, img.ID)
		}
	}
}

func pullImage(imageName string) string {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	httpPullImage(uClient, version, imageName)
	images := httpListImages(uClient, version)
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if imageTagMatches(tag, imageName) {
				return img.ID
			}
		}
	}
	gomega.Expect(false).To(gomega.BeTrue(), fmt.Sprintf("pulled image %s not found", imageName))
	return ""
}

func removeImage(imageName string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	httpRemoveImageForce(uClient, version, imageName)
	images := httpListImages(uClient, version)
	for _, img := range images {
		for _, tag := range img.RepoTags {
			gomega.Expect(tag).NotTo(gomega.Equal(imageName))
		}
	}
}

// Helper functions for HTTP API calls to replace command.Run usage

// httpPullImage pulls an image using the HTTP API.
// Passes fromImage and tag as separate query params so the pull handler
// can correctly handle images like "localhost:PORT/alpine:latest" where
// the colon in host:port must not be treated as a name/tag separator.
func httpPullImage(uClient *http.Client, version, imageName string) {
	// Split on the last ':' to separate repo from tag, but only if the part
	// after the last ':' looks like a tag (no '/' in it).
	repo := imageName
	tag := ""
	if lastColon := strings.LastIndex(imageName, ":"); lastColon > 0 {
		candidate := imageName[lastColon+1:]
		if !strings.Contains(candidate, "/") {
			repo = imageName[:lastColon]
			tag = candidate
		}
	}
	var relativeUrl string
	if tag != "" {
		relativeUrl = fmt.Sprintf("/images/create?fromImage=%s&tag=%s", repo, tag)
	} else {
		relativeUrl = fmt.Sprintf("/images/create?fromImage=%s", imageName)
	}
	u := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(u, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Read body to completion to ensure image is fully pulled
	_, _ = io.Copy(io.Discard, resp.Body)
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
}

// httpTagImage tags an image using the HTTP API.
func httpTagImage(uClient *http.Client, version, sourceImage, targetImage string) {
	// Parse the target image to get repo and tag
	// targetImage is e.g. "localhost:12345/alpine:latest"
	// Split on the LAST colon to separate tag from repo
	lastColon := strings.LastIndex(targetImage, ":")
	repo := targetImage
	tag := "latest"
	if lastColon > 0 {
		repo = targetImage[:lastColon]
		tag = targetImage[lastColon+1:]
	}
	relativeUrl := fmt.Sprintf("/images/%s/tag?repo=%s&tag=%s",
		sourceImage, url.QueryEscape(repo), url.QueryEscape(tag))
	u := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(u, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))
}

// httpPushImage pushes an image using the HTTP API.
func httpPushImage(uClient *http.Client, version, imageName string) {
	httpPushImageWithAuth(uClient, version, imageName, "")
}

// httpPushImageWithAuth pushes an image using the HTTP API with an optional base64-encoded X-Registry-Auth header.
func httpPushImageWithAuth(uClient *http.Client, version, imageName, registryAuth string) {
	relativeUrl := fmt.Sprintf("/images/%s/push", imageName)
	u := client.ConvertToFinchUrl(version, relativeUrl)
	req, err := http.NewRequest(http.MethodPost, u, nil)
	gomega.Expect(err).Should(gomega.BeNil())
	req.Header.Set("Content-Type", "application/json")
	if registryAuth != "" {
		req.Header.Set("X-Registry-Auth", registryAuth)
	}
	resp, err := uClient.Do(req)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Read body to completion
	_, _ = io.Copy(io.Discard, resp.Body)
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
}

// httpRemoveImage removes an image using the HTTP API.
func httpRemoveImage(uClient *http.Client, version, imageName string) {
	relativeUrl := fmt.Sprintf("/images/%s", imageName)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Do(req)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.SatisfyAny(
		gomega.Equal(http.StatusOK),
		gomega.Equal(http.StatusNoContent),
		gomega.Equal(http.StatusNotFound)))
}

// httpRemoveImageForce removes an image with force using the HTTP API.
func httpRemoveImageForce(uClient *http.Client, version, imageName string) {
	relativeUrl := fmt.Sprintf("/images/%s?force=true", imageName)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Do(req)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Ignore status code since we're cleaning up
}

// httpRunContainer creates and starts a container using the HTTP API (replaces command.Run "run" -d).
func httpRunContainer(uClient *http.Client, version, containerName, image string, cmd []string) string {
	options := types.ContainerCreateRequest{
		ContainerConfig: types.ContainerConfig{
			Image: image,
			Cmd:   cmd,
		},
	}
	containerID := httpCreateContainer(uClient, version, containerName, options)
	httpStartContainer(uClient, version, containerID)
	return containerID
}

// httpRunContainerWithOptions creates and starts a container with custom options using the HTTP API.
func httpRunContainerWithOptions(uClient *http.Client, version, containerName string, options types.ContainerCreateRequest) string {
	containerID := httpCreateContainer(uClient, version, containerName, options)
	httpStartContainer(uClient, version, containerID)
	return containerID
}

// httpCreateContainer creates a container using the HTTP API.
func httpCreateContainer(uClient *http.Client, version, containerName string, options types.ContainerCreateRequest) string {
	url := client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/create?name=%s", containerName))
	reqBody, err := json.Marshal(options)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(url, "application/json", bytes.NewReader(reqBody))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))
	var ctr struct {
		ID string `json:"Id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&ctr)
	gomega.Expect(err).Should(gomega.BeNil())
	return ctr.ID
}

// httpStartContainer starts a container using the HTTP API.
func httpStartContainer(uClient *http.Client, version, containerID string) {
	relativeUrl := fmt.Sprintf("/containers/%s/start", containerID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.SatisfyAny(
		gomega.Equal(http.StatusNoContent),
		gomega.Equal(http.StatusNotModified)))
}

// httpStopContainerWithTimeout stops a container with a timeout using the HTTP API.
func httpStopContainerWithTimeout(uClient *http.Client, version, containerID string, timeout int) {
	relativeUrl := fmt.Sprintf("/containers/%s/stop?t=%d", containerID, timeout)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.SatisfyAny(
		gomega.Equal(http.StatusNoContent),
		gomega.Equal(http.StatusNotModified)))
}

// httpKillContainerWithSignal kills a container with a specific signal using the HTTP API.
func httpKillContainerWithSignal(uClient *http.Client, version, containerID, signal string) {
	relativeUrl := fmt.Sprintf("/containers/%s/kill?signal=%s", containerID, signal)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.SatisfyAny(
		gomega.Equal(http.StatusNoContent),
		gomega.Equal(http.StatusNotFound),
		gomega.Equal(http.StatusConflict)))
}

// httpPauseContainer pauses a container using the HTTP API.
func httpPauseContainer(uClient *http.Client, version, containerID string) {
	relativeUrl := fmt.Sprintf("/containers/%s/pause", containerID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusNoContent))
}

// httpRemoveContainer removes a container using the HTTP API.
func httpRemoveContainer(uClient *http.Client, version, containerID string) {
	relativeUrl := fmt.Sprintf("/containers/%s", containerID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Do(req)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.SatisfyAny(
		gomega.Equal(http.StatusOK),
		gomega.Equal(http.StatusNoContent),
		gomega.Equal(http.StatusNotFound)))
}

// httpRemoveContainerForce removes a container with force using the HTTP API.
func httpRemoveContainerForce(uClient *http.Client, version, containerID string) {
	relativeUrl := fmt.Sprintf("/containers/%s?force=true", containerID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Do(req)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Ignore status code since we're cleaning up
}

// httpWaitContainer waits for a container to stop using the HTTP API.
func httpWaitContainer(uClient *http.Client, version, containerID string) {
	relativeUrl := fmt.Sprintf("/containers/%s/wait", containerID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
}

// httpCreateVolume creates a volume using the HTTP API.
func httpCreateVolume(uClient *http.Client, version, volumeName string, labels map[string]string) {
	url := client.ConvertToFinchUrl(version, "/volumes/create")
	reqBody := map[string]interface{}{
		"Name":   volumeName,
		"Labels": labels,
	}
	body, err := json.Marshal(reqBody)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(url, "application/json", bytes.NewReader(body))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Accept both 200 (OK) and 201 (Created) as success
	gomega.Expect(resp.StatusCode).Should(gomega.SatisfyAny(
		gomega.Equal(http.StatusOK),
		gomega.Equal(http.StatusCreated)))
}

// httpRemoveVolume removes a volume using the HTTP API.
func httpRemoveVolume(uClient *http.Client, version, volumeName string) {
	relativeUrl := fmt.Sprintf("/volumes/%s", volumeName)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Do(req)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Ignore status code since we're cleaning up
}

// httpCreateNetwork creates a network using the HTTP API.
func httpCreateNetwork(uClient *http.Client, version, networkName string) string {
	url := client.ConvertToFinchUrl(version, "/networks/create")
	reqBody := map[string]interface{}{
		"Name": networkName,
	}
	body, err := json.Marshal(reqBody)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(url, "application/json", bytes.NewReader(body))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))
	var result struct {
		ID string `json:"Id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	gomega.Expect(err).Should(gomega.BeNil())
	return result.ID
}

// httpRemoveNetwork removes a network using the HTTP API.
func httpRemoveNetwork(uClient *http.Client, version, networkID string) {
	relativeUrl := fmt.Sprintf("/networks/%s", networkID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Do(req)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Ignore status code since we're cleaning up
}

// httpContainerLogs gets the logs of a container using the HTTP API.
// The Docker logs API returns a multiplexed stream with 8-byte frame headers
// ([stream_type(1), 0, 0, 0, size(4 big-endian)]), so we demultiplex it here.
func httpContainerLogs(uClient *http.Client, version, containerID string) string {
	relativeUrl := fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true", containerID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	var buf strings.Builder
	hdr := make([]byte, 8)
	for {
		_, err := io.ReadFull(resp.Body, hdr)
		if err != nil {
			break
		}
		size := uint32(hdr[4])<<24 | uint32(hdr[5])<<16 | uint32(hdr[6])<<8 | uint32(hdr[7])
		frame := make([]byte, size)
		_, err = io.ReadFull(resp.Body, frame)
		if err != nil {
			break
		}
		buf.Write(frame)
	}
	return buf.String()
}

// httpExecContainer creates and starts an exec instance in a container using the HTTP API.
func httpExecContainer(uClient *http.Client, version, containerID string, cmd []string) string {
	// Create exec instance
	createUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/exec", containerID))
	createReq := map[string]interface{}{
		"Cmd":          cmd,
		"AttachStdout": true,
		"AttachStderr": true,
	}
	body, err := json.Marshal(createReq)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(createUrl, "application/json", bytes.NewReader(body))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))

	var execResp struct {
		ID string `json:"Id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&execResp)
	gomega.Expect(err).Should(gomega.BeNil())

	// Start exec instance
	startUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/exec/%s/start", execResp.ID))
	startReq := map[string]interface{}{
		"Detach": false,
	}
	startBody, err := json.Marshal(startReq)
	gomega.Expect(err).Should(gomega.BeNil())
	startResp, err := uClient.Post(startUrl, "application/json", bytes.NewReader(startBody))
	gomega.Expect(err).Should(gomega.BeNil())
	defer startResp.Body.Close()

	// Demultiplex the Docker multiplexed stream (8-byte frame headers).
	var buf strings.Builder
	hdr := make([]byte, 8)
	for {
		_, err := io.ReadFull(startResp.Body, hdr)
		if err != nil {
			break
		}
		size := uint32(hdr[4])<<24 | uint32(hdr[5])<<16 | uint32(hdr[6])<<8 | uint32(hdr[7])
		frame := make([]byte, size)
		_, err = io.ReadFull(startResp.Body, frame)
		if err != nil {
			break
		}
		buf.Write(frame)
	}
	return buf.String()
}

// httpListContainers lists containers using the HTTP API.
// When all is true, includes stopped containers.
func httpListContainers(uClient *http.Client, version string, all bool) []types.ContainerListItem {
	relativeUrl := fmt.Sprintf("/containers/json?all=%t", all)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	var containers []types.ContainerListItem
	err = json.NewDecoder(resp.Body).Decode(&containers)
	gomega.Expect(err).Should(gomega.BeNil())
	return containers
}

// httpListImages lists images using the HTTP API.
func httpListImages(uClient *http.Client, version string) []types.ImageSummary {
	url := client.ConvertToFinchUrl(version, "/images/json")
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	var images []types.ImageSummary
	err = json.NewDecoder(resp.Body).Decode(&images)
	gomega.Expect(err).Should(gomega.BeNil())
	return images
}

// httpListVolumes lists volumes using the HTTP API.
func httpListVolumes(uClient *http.Client, version string) types.VolumesListResponse {
	url := client.ConvertToFinchUrl(version, "/volumes")
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	var volumes types.VolumesListResponse
	err = json.NewDecoder(resp.Body).Decode(&volumes)
	gomega.Expect(err).Should(gomega.BeNil())
	return volumes
}

// httpListNetworks lists networks using the HTTP API.
// Returns an empty slice if the request is forbidden (e.g. OPA middleware blocks GET /networks).
func httpListNetworks(uClient *http.Client, version string) []*types.NetworkInspectResponse {
	url := client.ConvertToFinchUrl(version, "/networks")
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// OPA policy may block GET /networks with 403 — treat as empty list for cleanup purposes
	if resp.StatusCode == http.StatusForbidden {
		return nil
	}
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	var networks []*types.NetworkInspectResponse
	err = json.NewDecoder(resp.Body).Decode(&networks)
	gomega.Expect(err).Should(gomega.BeNil())
	return networks
}

// httpRemoveAll removes all containers, non-default images, volumes, and non-default networks.
// Ignores individual deletion failures to ensure best-effort cleanup.
// Skips the local-registry container so it persists across test AfterEach cleanups.
func httpRemoveAll(uClient *http.Client, version string) {
	// Remove all containers (including stopped), but preserve the local registry.
	containers := httpListContainers(uClient, version, true)
	for _, c := range containers {
		isRegistry := false
		for _, name := range c.Names {
			if name == localRegistryName || name == "/"+localRegistryName {
				isRegistry = true
				break
			}
		}
		if isRegistry {
			continue
		}
		httpRemoveContainerForce(uClient, version, c.Id)
	}

	// Remove all non-default images, but preserve the local registry image and defaultImage.
	images := httpListImages(uClient, version)
	for _, img := range images {
		preserve := false
		for _, tag := range img.RepoTags {
			if tag == registryImage || tag == defaultImage {
				preserve = true
				break
			}
		}
		if preserve {
			continue
		}
		for _, tag := range img.RepoTags {
			httpRemoveImageForce(uClient, version, tag)
		}
	}

	// Remove all volumes
	vols := httpListVolumes(uClient, version)
	for _, v := range vols.Volumes {
		httpRemoveVolume(uClient, version, v.Name)
	}

	// Remove non-default networks
	networks := httpListNetworks(uClient, version)
	for _, n := range networks {
		if n.Name != "bridge" && n.Name != "host" && n.Name != "none" {
			httpRemoveNetwork(uClient, version, n.ID)
		}
	}
}

// httpRemoveAllImages removes all non-default images using the HTTP API.
func httpRemoveAllImages(uClient *http.Client, version string) {
	images := httpListImages(uClient, version)
	for _, img := range images {
		for _, tag := range img.RepoTags {
			httpRemoveImageForce(uClient, version, tag)
		}
	}
}

// httpInspectContainer inspects a container using the HTTP API.
func httpInspectContainer(uClient *http.Client, version, id string) types.Container {
	relativeUrl := fmt.Sprintf("/containers/%s/json", id)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	var container types.Container
	err = json.NewDecoder(resp.Body).Decode(&container)
	gomega.Expect(err).Should(gomega.BeNil())
	return container
}

// httpInspectNetwork inspects a network using the HTTP API.
func httpInspectNetwork(uClient *http.Client, version, id string) types.NetworkInspectResponse {
	relativeUrl := fmt.Sprintf("/networks/%s", id)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	var network types.NetworkInspectResponse
	err = json.NewDecoder(resp.Body).Decode(&network)
	gomega.Expect(err).Should(gomega.BeNil())
	return network
}

// httpCreateNetworkWithLabels creates a network with labels using the HTTP API.
func httpCreateNetworkWithLabels(uClient *http.Client, version, networkName string, labels map[string]string) string {
	url := client.ConvertToFinchUrl(version, "/networks/create")
	reqBody := map[string]interface{}{
		"Name":   networkName,
		"Labels": labels,
	}
	body, err := json.Marshal(reqBody)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(url, "application/json", bytes.NewReader(body))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))
	var result struct {
		ID string `json:"Id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	gomega.Expect(err).Should(gomega.BeNil())
	return result.ID
}

// httpStartContainerAttach starts a container with attach, captures stdout, and waits for completion.
// The attach stream uses Docker's multiplexed format (8-byte frame headers) when TTY=false,
// so we demultiplex it the same way as httpContainerLogs and httpExecContainer.
func httpStartContainerAttach(uClient *http.Client, version, id string) string {
	// Attach to container stdout (logs=1 ensures buffered output is included for fast-exiting containers)
	attachUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/attach?stdout=1&stream=1&logs=1", id))
	attachResp, err := uClient.Post(attachUrl, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())

	// Read the attach stream in a goroutine so we don't miss output from fast-exiting containers.
	type result struct{ s string }
	ch := make(chan result, 1)
	go func() {
		defer attachResp.Body.Close()
		var buf strings.Builder
		hdr := make([]byte, 8)
		for {
			_, err := io.ReadFull(attachResp.Body, hdr)
			if err != nil {
				break
			}
			size := uint32(hdr[4])<<24 | uint32(hdr[5])<<16 | uint32(hdr[6])<<8 | uint32(hdr[7])
			frame := make([]byte, size)
			_, err = io.ReadFull(attachResp.Body, frame)
			if err != nil {
				break
			}
			buf.Write(frame)
		}
		ch <- result{buf.String()}
	}()

	// Start the container
	httpStartContainer(uClient, version, id)

	return (<-ch).s
}

// httpExecContainerWithExitCode creates and starts an exec instance, returning output and exit code.
func httpExecContainerWithExitCode(uClient *http.Client, version, containerID string, cmd []string) (string, int) {
	// Create exec instance
	createUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/containers/%s/exec", containerID))
	createReq := map[string]interface{}{
		"Cmd":          cmd,
		"AttachStdout": true,
		"AttachStderr": true,
	}
	body, err := json.Marshal(createReq)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(createUrl, "application/json", bytes.NewReader(body))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))

	var execResp struct {
		ID string `json:"Id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&execResp)
	gomega.Expect(err).Should(gomega.BeNil())

	// Start exec instance
	startUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/exec/%s/start", execResp.ID))
	startReq := map[string]interface{}{
		"Detach": false,
	}
	startBody, err := json.Marshal(startReq)
	gomega.Expect(err).Should(gomega.BeNil())
	startResp, err := uClient.Post(startUrl, "application/json", bytes.NewReader(startBody))
	gomega.Expect(err).Should(gomega.BeNil())
	defer startResp.Body.Close()

	// Demultiplex the Docker multiplexed stream (8-byte frame headers).
	var buf strings.Builder
	hdr := make([]byte, 8)
	for {
		_, err := io.ReadFull(startResp.Body, hdr)
		if err != nil {
			break
		}
		size := uint32(hdr[4])<<24 | uint32(hdr[5])<<16 | uint32(hdr[6])<<8 | uint32(hdr[7])
		frame := make([]byte, size)
		_, err = io.ReadFull(startResp.Body, frame)
		if err != nil {
			break
		}
		buf.Write(frame)
	}

	// Inspect exec for exit code
	inspectUrl := client.ConvertToFinchUrl(version, fmt.Sprintf("/exec/%s/json", execResp.ID))
	inspectResp, err := uClient.Get(inspectUrl)
	gomega.Expect(err).Should(gomega.BeNil())
	defer inspectResp.Body.Close()
	var execInspect struct {
		ExitCode int `json:"ExitCode"`
	}
	err = json.NewDecoder(inspectResp.Body).Decode(&execInspect)
	gomega.Expect(err).Should(gomega.BeNil())

	return buf.String(), execInspect.ExitCode
}

// httpBuildImage builds an image from a build context directory using the HTTP API.
func httpBuildImage(uClient *http.Client, version, tag, buildContextDir string) {
	// Create tar archive of the build context
	tarReader, err := createTarFromBuildContext(buildContextDir)
	gomega.Expect(err).Should(gomega.BeNil())

	relativeUrl := fmt.Sprintf("/build?t=%s", tag)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/x-tar", tarReader)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Read body to completion to ensure build finishes
	_, _ = io.Copy(io.Discard, resp.Body)
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
}

// httpRegistryLogin authenticates with a registry using the HTTP API and writes credentials to Docker config.
func httpRegistryLogin(uClient *http.Client, version, registry, user, password string) {
	url := client.ConvertToFinchUrl(version, "/auth")
	reqBody := map[string]string{
		"username":      user,
		"password":      password,
		"serveraddress": registry,
	}
	body, err := json.Marshal(reqBody)
	gomega.Expect(err).Should(gomega.BeNil())
	resp, err := uClient.Post(url, "application/json", bytes.NewReader(body))
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
}

// httpRegistryLogout removes registry credentials from Docker config JSON.
func httpRegistryLogout(registry string) {
	home, err := os.UserHomeDir()
	gomega.Expect(err).Should(gomega.BeNil())
	configPath := filepath.Join(home, ".docker", "config.json")

	data, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		// Config file doesn't exist, nothing to do
		return
	}

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return
	}

	auths, ok := config["auths"].(map[string]interface{})
	if !ok {
		return
	}

	delete(auths, registry)
	config["auths"] = auths

	updatedData, err := json.MarshalIndent(config, "", "  ")
	gomega.Expect(err).Should(gomega.BeNil())
	err = os.WriteFile(configPath, updatedData, 0600)
	gomega.Expect(err).Should(gomega.BeNil())
}

func containerShouldBeRunning(containerNames ...string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	containers := httpListContainers(uClient, version, false)
	for _, containerName := range containerNames {
		found := false
		for _, c := range containers {
			for _, name := range c.Names {
				if name == containerName || name == "/"+containerName {
					found = true
					break
				}
			}
		}
		gomega.Expect(found).To(gomega.BeTrue(), fmt.Sprintf("container %s should be running", containerName))
	}
}

func containerShouldNotBeRunning(containerNames ...string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	containers := httpListContainers(uClient, version, false)
	for _, containerName := range containerNames {
		found := false
		for _, c := range containers {
			for _, name := range c.Names {
				if name == containerName || name == "/"+containerName {
					found = true
					break
				}
			}
		}
		gomega.Expect(found).To(gomega.BeFalse(), fmt.Sprintf("container %s should not be running", containerName))
	}
}

func containerShouldExist(containerNames ...string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	containers := httpListContainers(uClient, version, true)
	for _, containerName := range containerNames {
		found := false
		for _, c := range containers {
			for _, name := range c.Names {
				if name == containerName || name == "/"+containerName {
					found = true
					break
				}
			}
		}
		gomega.Expect(found).To(gomega.BeTrue(), fmt.Sprintf("container %s should exist", containerName))
	}
}

func containerShouldNotExist(containerNames ...string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	containers := httpListContainers(uClient, version, true)
	for _, containerName := range containerNames {
		found := false
		for _, c := range containers {
			for _, name := range c.Names {
				if name == containerName || name == "/"+containerName {
					found = true
					break
				}
			}
		}
		gomega.Expect(found).To(gomega.BeFalse(), fmt.Sprintf("container %s should not exist", containerName))
	}
}

func imageTagMatches(tag, imageName string) bool {
	return tag == imageName || strings.HasSuffix(tag, "/"+imageName)
}

func imageShouldExist(imageName string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	waitForImageExist(uClient, version, imageName)
}

// waitForImageExist polls until the image appears in the list or times out.
func waitForImageExist(uClient *http.Client, version, imageName string) {
	filterJSON, _ := json.Marshal(map[string][]string{"reference": {imageName}})
	listURL := client.ConvertToFinchUrl(version, fmt.Sprintf("/images/json?filters=%s", url.QueryEscape(string(filterJSON))))
	for i := 0; i < 20; i++ {
		resp, err := uClient.Get(listURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			var images []types.ImageSummary
			if json.NewDecoder(resp.Body).Decode(&images) == nil && len(images) > 0 {
				resp.Body.Close()
				return
			}
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	gomega.Expect(false).To(gomega.BeTrue(), fmt.Sprintf("image %s should exist", imageName))
}

func imageShouldNotExist(imageName string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	images := httpListImages(uClient, version)
	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if imageTagMatches(tag, imageName) {
				found = true
				break
			}
		}
	}
	gomega.Expect(found).To(gomega.BeFalse(), fmt.Sprintf("image %s should not exist", imageName))
}

func volumeShouldExist(volumeName string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	vols := httpListVolumes(uClient, version)
	found := false
	for _, v := range vols.Volumes {
		if v.Name == volumeName {
			found = true
			break
		}
	}
	gomega.Expect(found).To(gomega.BeTrue(), fmt.Sprintf("volume %s should exist", volumeName))
}

func volumeShouldNotExist(volumeName string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	vols := httpListVolumes(uClient, version)
	found := false
	for _, v := range vols.Volumes {
		if v.Name == volumeName {
			found = true
			break
		}
	}
	gomega.Expect(found).To(gomega.BeFalse(), fmt.Sprintf("volume %s should not exist", volumeName))
}

func fileShouldExist(path, content string) {
	gomega.Expect(path).To(gomega.BeARegularFile())
	actualContent, err := os.ReadFile(filepath.Clean(path))
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(string(actualContent)).To(gomega.Equal(content))
}

func fileShouldNotExist(path string) {
	gomega.Expect(path).ToNot(gomega.BeAnExistingFile())
}

func fileShouldExistInContainer(containerName, path, content string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	output := httpExecContainer(uClient, version, containerName, []string{"cat", path})
	gomega.Expect(strings.TrimSpace(output)).To(gomega.Equal(content))
}

//nolint:unused // reserved for future use cases
func fileShouldNotExistInContainer(containerName, path string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	_, exitCode := httpExecContainerWithExitCode(uClient, version, containerName, []string{"cat", path})
	gomega.Expect(exitCode).NotTo(gomega.Equal(0))
}

//nolint:unused // reserved for future use cases
func buildImage(imageName string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	dockerfile := fmt.Sprintf(`FROM %s
		CMD ["echo", "finch-test-dummy-output"]
		`, defaultImage)
	buildContext := ffs.CreateBuildContext(dockerfile)
	ginkgo.DeferCleanup(os.RemoveAll, buildContext)
	httpBuildImage(uClient, version, imageName, buildContext)
}

// RemoveAll removes all containers, images, volumes, and non-default networks via HTTP API.
// Exported for use by the test suite setup/teardown in e2e_test.go.
func RemoveAll() {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	httpRemoveAll(uClient, version)
}

func GetDockerHostUrl() string {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		panic("DOCKER_HOST not set")
	}
	return dockerHost
}

func GetDockerApiVersion() string {
	version := os.Getenv("DOCKER_API_VERSION")
	if version == "" {
		panic("DOCKER_API_VERSION not set")
	}
	return version
}

// GetFinchExe gets the finch executable path from FINCH_ROOT environment variable, if set.
func GetFinchExe() string {
	finchdir := os.Getenv("FINCH_ROOT")

	// use default binary if env is not set
	if finchdir == "" {
		finchexe, err := exec.LookPath("finch")
		if err != nil {
			panic(err.Error())
		}
		return finchexe
	}

	finchexe := filepath.Join(finchdir, "bin/finch")
	if _, err := os.Stat(finchexe); errors.Is(err, os.ErrNotExist) {
		panic(fmt.Sprintf("%s not found. Is Finch installed?", finchexe))
	}
	return finchexe
}

func GetFinchDaemonExe() string {
	daemonPath := os.Getenv("DAEMON_ROOT")
	if daemonPath == "" {
		daemonPath = "./bin/finch-daemon" // fallback
	}
	return daemonPath
}
