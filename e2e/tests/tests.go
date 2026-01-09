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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/ffs"
	"github.com/runfinch/common-tests/fnet"
	"github.com/runfinch/common-tests/option"

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
func SetupLocalRegistry(opt *option.Option) {
	command.RemoveAll(opt)
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()

	hostPort := fnet.GetFreePort()
	containerID := httpRunContainerWithOptions(uClient, version, localRegistryName, types.ContainerCreateRequest{
		ContainerConfig: types.ContainerConfig{
			Image: registryImage,
		},
		HostConfig: types.ContainerHostConfig{
			PortBindings: nat.PortMap{
				"5000/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", hostPort)}},
			},
		},
	})
	imageID := command.StdoutStr(opt, "images", "-q")
	command.SetLocalRegistryContainerID(containerID)
	command.SetLocalRegistryImageID(imageID)
	command.SetLocalRegistryImageName(registryImage)

	httpPullImage(uClient, version, alpineImage)
	defaultImage = fmt.Sprintf("localhost:%d/alpine:latest", hostPort)
	httpTagImage(uClient, version, alpineImage, defaultImage)
	httpPushImage(uClient, version, defaultImage)
	httpRemoveImage(uClient, version, alpineImage)
}

// CleanupLocalRegistry removes the local registry container and image. It's used together with SetupLocalRegistry,
// and should be invoked after running all the tests.
func CleanupLocalRegistry(opt *option.Option) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()

	containerID := command.StdoutStr(opt, "inspect", localRegistryName, "--format", "{{.ID}}")
	httpRemoveContainerForce(uClient, version, containerID)
	imageIDsOutput := command.StdoutStr(opt, "images", "-q")
	// Split by newlines and remove each image separately
	for _, imageID := range strings.Split(imageIDsOutput, "\n") {
		imageID = strings.TrimSpace(imageID)
		if imageID != "" {
			httpRemoveImageForce(uClient, version, imageID)
		}
	}
}

func pullImage(opt *option.Option, imageName string) string {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	httpPullImage(uClient, version, imageName)
	imageID := command.Stdout(opt, "images", "--quiet", imageName)
	gomega.Expect(imageID).ShouldNot(gomega.BeEmpty())
	return strings.TrimSpace(string(imageID))
}

func removeImage(opt *option.Option, imageName string) {
	uClient := client.NewClient(GetDockerHostUrl())
	version := GetDockerApiVersion()
	httpRemoveImageForce(uClient, version, imageName)
	imageID := command.Stdout(opt, "images", "--quiet", imageName)
	gomega.Expect(string(imageID)).Should(gomega.BeEmpty())
}

// Helper functions for HTTP API calls to replace command.Run usage

// httpPullImage pulls an image using the HTTP API.
func httpPullImage(uClient *http.Client, version, imageName string) {
	relativeUrl := fmt.Sprintf("/images/create?fromImage=%s", imageName)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	// Read body to completion to ensure image is fully pulled
	_, _ = io.Copy(io.Discard, resp.Body)
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
}

// httpTagImage tags an image using the HTTP API.
func httpTagImage(uClient *http.Client, version, sourceImage, targetImage string) {
	// Parse the target image to get repo and tag
	parts := strings.Split(targetImage, ":")
	repo := parts[0]
	tag := "latest"
	if len(parts) > 1 {
		tag = parts[1]
	}
	relativeUrl := fmt.Sprintf("/images/%s/tag?repo=%s&tag=%s", sourceImage, repo, tag)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))
}

// httpPushImage pushes an image using the HTTP API.
func httpPushImage(uClient *http.Client, version, imageName string) {
	relativeUrl := fmt.Sprintf("/images/%s/push", imageName)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Post(url, "application/json", nil)
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
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusCreated))
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
func httpContainerLogs(uClient *http.Client, version, containerID string) string {
	relativeUrl := fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true", containerID)
	url := client.ConvertToFinchUrl(version, relativeUrl)
	resp, err := uClient.Get(url)
	gomega.Expect(err).Should(gomega.BeNil())
	defer resp.Body.Close()
	gomega.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))
	output, _ := io.ReadAll(resp.Body)
	return string(output)
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

	// Read output
	output, _ := io.ReadAll(startResp.Body)
	return string(output)
}

func containerShouldBeRunning(opt *option.Option, containerNames ...string) {
	for _, containerName := range containerNames {
		gomega.Expect(command.Stdout(opt, "ps", "-q", "--filter",
			fmt.Sprintf("name=%s", containerName))).NotTo(gomega.BeEmpty())
	}
}

func containerShouldNotBeRunning(opt *option.Option, containerNames ...string) {
	for _, containerName := range containerNames {
		gomega.Expect(command.Stdout(opt, "ps", "-q", "--filter",
			fmt.Sprintf("name=%s", containerName))).To(gomega.BeEmpty())
	}
}

func containerShouldExist(opt *option.Option, containerNames ...string) {
	for _, containerName := range containerNames {
		gomega.Expect(command.Stdout(opt, "ps", "-a", "-q", "--filter",
			fmt.Sprintf("name=%s", containerName))).NotTo(gomega.BeEmpty())
	}
}

func containerShouldNotExist(opt *option.Option, containerNames ...string) {
	for _, containerName := range containerNames {
		gomega.Expect(command.Stdout(opt, "ps", "-a", "-q", "--filter",
			fmt.Sprintf("name=%s", containerName))).To(gomega.BeEmpty())
	}
}

func imageShouldExist(opt *option.Option, imageName string) {
	gomega.Expect(command.Stdout(opt, "images", "-q", imageName)).NotTo(gomega.BeEmpty())
}

func imageShouldNotExist(opt *option.Option, imageName string) {
	gomega.Expect(command.Stdout(opt, "images", "-q", imageName)).To(gomega.BeEmpty())
}

func volumeShouldExist(opt *option.Option, volumeName string) {
	gomega.Expect(command.Stdout(opt, "volume", "ls", "-q", "--filter",
		fmt.Sprintf("name=%s", volumeName))).NotTo(gomega.BeEmpty())
}

func volumeShouldNotExist(opt *option.Option, volumeName string) {
	gomega.Expect(command.Stdout(opt, "volume", "ls", "-q", "--filter",
		fmt.Sprintf("name=%s", volumeName))).To(gomega.BeEmpty())
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

func fileShouldExistInContainer(opt *option.Option, containerName, path, content string) {
	gomega.Expect(command.StdoutStr(opt, "exec", containerName, "cat", path)).To(gomega.Equal(content))
}

//nolint:unused // reserved for future use cases
func fileShouldNotExistInContainer(opt *option.Option, containerName, path string) {
	cmdOut := command.RunWithoutSuccessfulExit(opt, "exec", containerName, "cat", path)
	gomega.Expect(cmdOut.Err.Contents()).To(gomega.ContainSubstring("No such file or directory"))
}

// Note: buildImage uses command.Run for the build command since Docker build API
// requires sending a tar archive which is complex to implement via HTTP.
// This is intentionally kept as CLI command for simplicity.
//
//nolint:unused // reserved for future use cases
func buildImage(opt *option.Option, imageName string) {
	dockerfile := fmt.Sprintf(`FROM %s
		CMD ["echo", "finch-test-dummy-output"]
		`, defaultImage)
	buildContext := ffs.CreateBuildContext(dockerfile)
	ginkgo.DeferCleanup(os.RemoveAll, buildContext)
	command.Run(opt, "build", "-q", "-t", imageName, buildContext)
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
