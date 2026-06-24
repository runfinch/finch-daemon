// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/labels"
	"github.com/containerd/nerdctl/v2/pkg/ocihook"
	"github.com/containerd/typeurl/v2"
	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// ocihookMu serializes all ocihook.Run calls within finch-daemon.
// ocihook.Run temporarily redirects a package-level global logger to a
// multi-writer (restored in a defer), which races if called concurrently.
var ocihookMu sync.Mutex

func (s *service) Start(ctx context.Context, cid string, options types.ContainerStartOptions) error {
	cont, err := s.getContainer(ctx, cid)
	if err != nil {
		return err
	}
	if err := s.assertStartContainer(ctx, cont); err != nil {
		return err
	}
	s.logger.Debugf("starting container: %s", cid)

	if err := s.customStart(ctx, cid, cont, options); err != nil {
		s.logger.Errorf("Failed to start container: %s. Error: %v", cid, err)
		return err
	}

	s.logger.Debugf("successfully started: %s", cid)
	return nil
}

// customStart creates a container task with FIFO-based logging, sets up
// CNI networking, and starts the container process.
func (s *service) customStart(ctx context.Context, cid string, cont containerd.Container, options types.ContainerStartOptions) error {
	startTime := time.Now()
	s.cleanupOldTask(ctx, cont)

	containerLabels, err := cont.Labels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container labels: %w", err)
	}

	// Strip OCI hooks before creating the task. Networking is handled inline
	// by setupNetworking, not by runc forking a binary at lifecycle points.
	// This is needed for containers created via nerdctl CLI (which sets hooks)
	// as well as API-created containers (hooks already stripped in create.go).
	spec, err := cont.Spec(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container spec: %w", err)
	}

	specModified := false
	if spec.Hooks != nil {
		spec.Hooks = nil
		specModified = true
	}

	// If networking was pre-configured at create time, set the netns path
	// in the spec so the container joins the existing namespace.
	if netnsPath := containerLabels["nerdctl/network-namespace"]; netnsPath != "" {
		if _, statErr := os.Stat(netnsPath); statErr == nil {
			if spec.Linux == nil {
				spec.Linux = &specs.Linux{}
			}
			// Replace or add the network namespace entry.
			found := false
			for i, ns := range spec.Linux.Namespaces {
				if ns.Type == specs.NetworkNamespace {
					spec.Linux.Namespaces[i].Path = netnsPath
					found = true
					break
				}
			}
			if !found {
				spec.Linux.Namespaces = append(spec.Linux.Namespaces, specs.LinuxNamespace{
					Type: specs.NetworkNamespace,
					Path: netnsPath,
				})
			}
			specModified = true
		}
	}

	if specModified {
		if err := cont.Update(ctx, func(ctx context.Context, client *containerd.Client, c *containers.Container) error {
			newSpec, err := typeurl.MarshalAny(spec)
			if err != nil {
				return err
			}
			c.Spec = newSpec
			return nil
		}); err != nil {
			return fmt.Errorf("failed to update OCI spec: %w", err)
		}
	}
	s.logger.Debugf("customStart(%s): hooks stripped at +%dms", cid, time.Since(startTime).Milliseconds())

	task, directIO, err := s.createTaskWithFIFO(ctx, cont, spec)
	if err != nil {
		return err
	}
	s.logger.Debugf("customStart(%s): task created at +%dms", cid, time.Since(startTime).Milliseconds())

	ns := containerLabels[labels.Namespace]

	dataStore, err := s.nctlContainerSvc.GetDataStore()
	if err != nil {
		task.Delete(ctx)
		return fmt.Errorf("failed to get data store: %w", err)
	}

	// CNI setup — must happen after NewTask (needs task.Pid()) and before task.Start.
	// If setupNetworking fails (e.g., stale CNI state from a previous run),
	// tear down and retry once.
	globalOpts := s.nctlContainerSvc.GetGlobalOptions()
	networkingConfigured := false
	if err := s.setupNetworking(ctx, cont, task, containerLabels, ns, dataStore, globalOpts); err != nil {
		// Retry after teardown — handles restart of stopped containers
		// where CNI state persists from the previous run.
		runPostStop(cont.ID(), containerLabels, ns, dataStore, globalOpts)
		if retryErr := s.setupNetworking(ctx, cont, task, containerLabels, ns, dataStore, globalOpts); retryErr != nil {
			task.Delete(ctx)
			return retryErr
		}
		networkingConfigured = true
	} else if globalOpts.CNIPath != "" && globalOpts.CNINetConfPath != "" {
		// setupNetworking returned nil — check if it actually did work
		// (vs early-returning for non-CNI network modes).
		networksJSON := containerLabels[labels.Networks]
		if networksJSON != "" && networksJSON != "[]" {
			networkingConfigured = true
		}
	}

	if err := task.Start(ctx); err != nil {
		return err
	}
	s.logger.Debugf("customStart(%s): task.Start complete at +%dms", cid, time.Since(startTime).Milliseconds())

	// Start log copiers after task.Start to avoid goroutine leaks if Start fails.
	logPath := containerLogPath(dataStore, ns, cont.ID())
	startLogCopiers(logPath, directIO)

	// Watch for container exit and run CNI teardown (only if we set up networking).
	if networkingConfigured {
		watchPostStop(task, cont.ID(), containerLabels, ns, dataStore, globalOpts)
	}
	s.logger.Debugf("customStart(%s): complete at +%dms", cid, time.Since(startTime).Milliseconds())

	return nil
}

// cleanupOldTask deletes any existing task for the container. This handles
// the case where a previous task exists from a prior run (e.g., container restart).
func (s *service) cleanupOldTask(ctx context.Context, cont containerd.Container) {
	if oldTask, err := cont.Task(ctx, nil); err == nil {
		if _, err := oldTask.Delete(ctx); err != nil {
			s.logger.Debugf("failed to delete old task for %s: %v", cont.ID(), err)
		}
	}
}

// createTaskWithFIFO creates a new container task using FIFO-based IO instead of
// the binary:// log driver. containerd's shim writes container stdout/stderr to
// the FIFO write ends; the returned DirectIO provides the read ends that our
// copier goroutines (see copier.go) consume.
func (s *service) createTaskWithFIFO(ctx context.Context, cont containerd.Container, spec *specs.Spec) (containerd.Task, *cio.DirectIO, error) {
	// Terminal mode multiplexes stdout+stderr into a single stream.
	isTerminal := spec.Process != nil && spec.Process.Terminal

	var directIO *cio.DirectIO
	ioCreator := func(id string) (cio.IO, error) {
		// Create named pipes under containerd's default FIFO directory.
		fifos, err := s.client.NewFIFOSetInDir(defaults.DefaultFIFODir, id, isTerminal)
		if err != nil {
			return nil, err
		}
		fifos.Stdin = ""

		// Open the FIFOs and return DirectIO with read ends for stdout/stderr.
		directIO, err = s.client.NewDirectCIO(ctx, fifos)
		if err != nil {
			return nil, err
		}
		return directIO, nil
	}

	task, err := cont.NewTask(ctx, ioCreator)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, directIO, nil
}

// setupNetworking calls ocihook.Run("createRuntime") to configure CNI networking
// for the container. This replaces the OCI createRuntime hook that runc would
// normally execute by forking the nerdctl binary.
func (s *service) setupNetworking(ctx context.Context, cont containerd.Container, task containerd.Task, containerLabels map[string]string, ns, dataStore string, globalOpts *types.GlobalCommandOptions) error {
	// Skip CNI setup if CNI paths are not configured (e.g., unit tests or
	// environments without CNI). ocihook.Run requires non-empty paths.
	if globalOpts.CNIPath == "" || globalOpts.CNINetConfPath == "" {
		s.logger.Debugf("skipping CNI setup for container %s: CNI paths not configured", cont.ID())
		return nil
	}

	// Skip CNI setup for non-CNI network modes. These don't need any
	// networking configuration and ocihook.Run would no-op anyway, but
	// calling it still has overhead (lock acquisition, state dir creation).
	networksJSON := containerLabels[labels.Networks]
	if networksJSON == "" || networksJSON == "[]" {
		return nil
	}
	s.logger.Debugf("setupNetworking(%s): networks=%s", cont.ID(), networksJSON)

	// If networking was already configured at create time, skip.
	if netnsPath := containerLabels["nerdctl/network-namespace"]; netnsPath != "" {
		if _, err := os.Stat(netnsPath); err == nil {
			s.logger.Debugf("setupNetworking(%s): skipping, already configured at create time", cont.ID())
			return nil
		}
		// Netns file doesn't exist — fall through to configure inline.
		s.logger.Debugf("setupNetworking(%s): netns annotation set but file missing, configuring inline", cont.ID())
	}
	// Check for host/none/container network modes
	var networks []string
	if err := json.Unmarshal([]byte(networksJSON), &networks); err == nil {
		if len(networks) == 1 {
			switch networks[0] {
			case "host", "none":
				return nil
			}
			if len(networks[0]) > 0 && networks[0][0] == '/' {
				// container:<id> network mode stored as absolute path
				return nil
			}
		}
	}

	state := specs.State{
		Version:     "1.0.0",
		ID:          cont.ID(),
		Status:      specs.StateCreated,
		Pid:         int(task.Pid()),
		Bundle:      fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/%s/%s", ns, cont.ID()),
		Annotations: containerLabels,
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal hook state: %w", err)
	}

	if err := func() error {
		lockStart := time.Now()
		ocihookMu.Lock()
		s.logger.Debugf("setupNetworking(%s): acquired mutex after %dms", cont.ID(), time.Since(lockStart).Milliseconds())
		defer ocihookMu.Unlock()
		hookStart := time.Now()
		err := ocihook.Run(
			bytes.NewReader(stateJSON),
			os.Stderr,
			"createRuntime",
			dataStore,
			globalOpts.CNIPath,
			globalOpts.CNINetConfPath,
			globalOpts.BridgeIP,
		)
		s.logger.Debugf("setupNetworking(%s): ocihook.Run took %dms", cont.ID(), time.Since(hookStart).Milliseconds())
		return err
	}(); err != nil {
		return fmt.Errorf("CNI setup failed: %w", err)
	}

	s.logger.Infof("hookless CNI setup complete for container %s", cont.ID())
	return nil
}

// setupNetworkingAtCreate runs CNI setup at container create time using a
// pre-created network namespace. This eliminates networking latency from the
// Start path. The netns path is stored as an annotation so ocihook.Run can
// use it without needing a container PID.
func (s *service) setupNetworkingAtCreate(ctx context.Context, cont containerd.Container, globalOpts types.GlobalCommandOptions) error {
	if globalOpts.CNIPath == "" || globalOpts.CNINetConfPath == "" {
		return nil
	}

	containerLabels, err := cont.Labels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container labels: %w", err)
	}

	networksJSON := containerLabels[labels.Networks]
	if networksJSON == "" || networksJSON == "[]" {
		return nil
	}

	// Skip for non-CNI network modes
	var networks []string
	if err := json.Unmarshal([]byte(networksJSON), &networks); err == nil {
		if len(networks) == 1 {
			switch networks[0] {
			case "host", "none":
				return nil
			}
			if len(networks[0]) > 0 && networks[0][0] == '/' {
				return nil
			}
		}
	}

	ns := containerLabels[labels.Namespace]
	dataStore, err := s.nctlContainerSvc.GetDataStore()
	if err != nil {
		return fmt.Errorf("failed to get data store: %w", err)
	}

	// Create a named network namespace for CNI to configure.
	netnsPath := fmt.Sprintf("/var/run/netns/%s", cont.ID())
	if err := createNetns(netnsPath); err != nil {
		return fmt.Errorf("failed to create netns: %w", err)
	}

	// Store the netns path as a container label so Start can find it
	// and ocihook.Run can use it instead of /proc/<pid>/ns/net.
	annotationLabels := make(map[string]string, len(containerLabels)+1)
	for k, v := range containerLabels {
		annotationLabels[k] = v
	}
	annotationLabels["nerdctl/network-namespace"] = netnsPath

	// Update container labels to include the netns annotation.
	if _, err := cont.SetLabels(ctx, map[string]string{
		"nerdctl/network-namespace": netnsPath,
	}); err != nil {
		removeNetns(netnsPath)
		return fmt.Errorf("failed to set netns label: %w", err)
	}

	state := specs.State{
		Version:     "1.0.0",
		ID:          cont.ID(),
		Status:      specs.StateCreated,
		Pid:         0, // No process yet — ocihook uses the netns annotation instead
		Bundle:      fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/%s/%s", ns, cont.ID()),
		Annotations: annotationLabels,
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		removeNetns(netnsPath)
		return fmt.Errorf("failed to marshal hook state: %w", err)
	}

	ocihookMu.Lock()
	err = ocihook.Run(
		bytes.NewReader(stateJSON),
		os.Stderr,
		"createRuntime",
		dataStore,
		globalOpts.CNIPath,
		globalOpts.CNINetConfPath,
		globalOpts.BridgeIP,
	)
	ocihookMu.Unlock()
	if err != nil {
		removeNetns(netnsPath)
		return fmt.Errorf("pre-create CNI setup failed: %w", err)
	}

	s.logger.Infof("pre-create CNI setup complete for container %s (netns=%s)", cont.ID(), netnsPath)
	return nil
}

// createNetns creates a persistent named network namespace.
func createNetns(path string) error {
	// Use the container ID as the netns name (basename of path).
	name := filepath.Base(path)
	cmd := exec.Command("ip", "netns", "add", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ip netns add: %s: %w", string(out), err)
	}
	return nil
}

// removeNetns removes a named network namespace.
func removeNetns(path string) {
	name := filepath.Base(path)
	exec.Command("ip", "netns", "delete", name).Run() //nolint:errcheck // best-effort cleanup
}

// watchPostStop spawns a goroutine that waits for the container to exit and
// then runs CNI teardown. This replaces the OCI postStop hook.
func watchPostStop(task containerd.Task, containerID string, containerLabels map[string]string, ns, dataStore string, globalOpts *types.GlobalCommandOptions) {
	go func() {
		exitCh, err := task.Wait(context.Background())
		if err != nil {
			return
		}
		<-exitCh
		// Kill port reserver immediately — before the potentially slow/contended
		// runPostStop call. This ensures clients connected to the container's
		// mapped ports get a connection reset without delay.
		killPortReserver(ns, containerID)
		runPostStop(containerID, containerLabels, ns, dataStore, globalOpts)
	}()
}

func (s *service) assertStartContainer(ctx context.Context, c containerd.Container) error {
	status := s.client.GetContainerStatus(ctx, c)
	switch status {
	case containerd.Running:
		return errdefs.NewNotModified(fmt.Errorf("container already running"))
	case containerd.Pausing:
	case containerd.Paused:
		return fmt.Errorf("cannot start a paused container, try unpause instead")
	}
	return nil
}
