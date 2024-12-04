// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	gocni "github.com/containerd/go-cni"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/defaults"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/runfinch/finch-daemon/pkg/utility/maputility"
)

type containerCreateResponse struct {
	ID string `json:"Id"`
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	platform := r.URL.Query().Get("platform")

	var req types.ContainerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}

	// AttachStdin is currently not supported
	// TODO: Remove this check when attach supports stdin
	if req.AttachStdin {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg("AttachStdin is currently not supported during create"))
		return
	}

	// defaults
	rp := req.HostConfig.RestartPolicy
	restart := "no" // Docker API default.
	if rp.Name != "" {
		restart = rp.Name
		if rp.MaximumRetryCount > 0 {
			restart = fmt.Sprintf("%s:%d", restart, rp.MaximumRetryCount)
		}
	}
	stopSignal := "SIGTERM" // nerdctl default.
	if req.StopSignal != "" {
		stopSignal = req.StopSignal
	}
	stopTimeout := 10 // Docker API default.
	if req.StopTimeout != nil {
		stopTimeout = *req.StopTimeout
	}
	memory := ""
	if req.HostConfig.Memory > 0 {
		memory = fmt.Sprint(req.HostConfig.Memory)
	}
	lc := req.HostConfig.LogConfig
	logDriver := "json-file" // Docker API default
	if lc.Type != "" {
		logDriver = lc.Type
	}
	logOpt := []string{}
	if len(lc.Config) > 0 {
		logOpt = maputility.Flatten(lc.Config, maputility.KeyEqualsValueFormat)
	}

	// Volumes:
	// nerdctl expects volumes to be a list of bind mounts or individual user created volumes.
	// Each element is formatted as "HOST_PATH:CONTAINER_PATH:BIND_OPTIONS". Example: "/tmp/workdir:/workdir:ro".
	// Or simply "VOLUME", where VOLUME is a user created volume.
	volumes := req.HostConfig.Binds
	if req.Volumes != nil {
		for newVolume := range req.Volumes {
			// If a volume points to one of the directories already mapped to a host path in bind mounts, it should not be added as a separate volume.
			contained := false
			for _, volume := range volumes {
				bindOpts := strings.Split(volume, ":")
				if len(bindOpts) > 1 && newVolume == bindOpts[1] || newVolume == volume {
					contained = true
					break
				}
			}
			if !contained {
				volumes = append(volumes, newVolume)
			}
		}
	}

	// Labels:
	// labels are passed in as a map of strings,
	// but nerdctl expects an array of strings with format [LABEL1=VALUE1, LABEL2=VALUE2, ...].
	labels := []string{}
	if req.Labels != nil {
		for key, val := range req.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", key, val))
		}
	}

	// Annotations: TODO - available in nerdctl 2.0
	// Annotations are passed in as a map of strings,
	// but nerdctl expects an array of strings with format [annotations1=VALUE1, annotations2=VALUE2, ...].
	// annotations := []string{}
	// if req.HostConfig.Annotations != nil {
	// 	for key, val := range req.HostConfig.Annotations {
	// 		annotations = append(annotations, fmt.Sprintf("%s=%s", key, val))
	// 	}
	// }

	ulimits := []string{}
	if req.HostConfig.Ulimits != nil {
		for _, ulimit := range req.HostConfig.Ulimits {
			ulimits = append(ulimits, ulimit.String())
		}
	}
	// Environment vars:
	env := []string{}
	if req.Env != nil {
		env = append(env, req.Env...)
	}

	// Linux Capabilities
	capAdd := []string{}
	if req.HostConfig.CapAdd != nil {
		capAdd = req.HostConfig.CapAdd
	}

	capDrop := []string{}
	if req.HostConfig.CapDrop != nil {
		capDrop = req.HostConfig.CapDrop
	}

	memoryReservation := ""
	if req.HostConfig.MemoryReservation != 0 {
		memoryReservation = fmt.Sprint(req.HostConfig.MemoryReservation)
	}

	memorySwap := ""
	if req.HostConfig.MemorySwap != 0 {
		memorySwap = fmt.Sprint(req.HostConfig.MemorySwap)
	}

	memorySwappiness := int64(-1)
	if req.HostConfig.MemorySwappiness > 0 {
		memorySwappiness = req.HostConfig.MemorySwappiness
	}

	pidLimit := int64(-1)
	if req.HostConfig.PidsLimit > 0 {
		pidLimit = req.HostConfig.PidsLimit
	}

	CpuQuota := int64(-1)
	if req.HostConfig.CPUQuota != 0 {
		CpuQuota = req.HostConfig.CPUQuota
	}

	globalOpt := ncTypes.GlobalCommandOptions(*h.Config)
	createOpt := ncTypes.ContainerCreateOptions{
		Stdout:   nil,
		Stderr:   nil,
		GOptions: globalOpt,

		// #region for basic flags
		Interactive: false,                     // TODO: update this after attach supports STDIN
		TTY:         false,                     // TODO: update this after attach supports STDIN
		Detach:      true,                      // TODO: current implementation of create does not support AttachStdin, AttachStdout, and AttachStderr flags
		Restart:     restart,                   // Restart policy to apply when a container exits.
		Rm:          req.HostConfig.AutoRemove, // Automatically remove container upon exit.
		Pull:        "missing",                 // nerdctl default.
		StopSignal:  stopSignal,
		StopTimeout: stopTimeout,
		CidFile:     req.HostConfig.ContainerIDFile, // CidFile write the container ID to the file
		// #endregion

		// #region for platform flags
		Platform: platform, // target platform
		// #endregion

		// #region for isolation flags
		Isolation: "default", // nerdctl default.
		// #endregion

		// #region for resource flags
		CPUShares:          uint64(req.HostConfig.CPUShares), // CPU shares (relative weight)
		Memory:             memory,                           // memory limit (in bytes)
		CPUQuota:           CpuQuota,                         // CPUQuota limits the CPU CFS (Completely Fair Scheduler) quota
		MemorySwappiness64: memorySwappiness,                 // Tuning container memory swappiness behaviour
		PidsLimit:          pidLimit,                         // PidsLimit specifies the tune container pids limit
		Cgroupns:           defaults.CgroupnsMode(),          // nerdctl default.
		MemoryReservation:  memoryReservation,                // Memory soft limit (in bytes)
		MemorySwap:         memorySwap,                       // Total memory usage (memory + swap); set `-1` to enable unlimited swap
		Ulimit:             ulimits,                          // List of ulimits to be set in the container
		CPUPeriod:          uint64(req.HostConfig.CPUPeriod),
		// #endregion

		// #region for user flags
		User: req.User,
		// #endregion

		// #region for security flags
		SecurityOpt: []string{}, // nerdctl default.
		CapAdd:      capAdd,
		CapDrop:     capDrop,
		Privileged:  req.HostConfig.Privileged,
		// #endregion
		// #region for runtime flags
		Runtime: defaults.Runtime, // nerdctl default.
		// #endregion

		// #region for volume flags
		Volume: volumes,
		// #endregion

		// #region for env flags
		Env:               env,
		Workdir:           req.WorkingDir,
		Entrypoint:        req.Entrypoint,
		EntrypointChanged: len(req.Entrypoint) > 0,
		// #endregion

		// #region for metadata flags
		Name:  name,   // container name
		Label: labels, // container labels
		// #endregion

		// #region for logging flags
		LogDriver: logDriver, // logging driver for the container
		LogOpt:    logOpt,    // logging driver specific options
		// #endregion

		// #region for image pull and verify options
		ImagePullOpt: ncTypes.ImagePullOptions{
			GOptions:      globalOpt,
			VerifyOptions: ncTypes.ImageVerifyOptions{Provider: "none"},
			IPFSAddress:   "",
			Stdout:        nil,
			Stderr:        nil,
		},
		// #endregion
	}

	portMappings, err := translatePortMappings(req.HostConfig.PortBindings)
	if err != nil {
		logrus.Debugf("failed to parse port mappings: %s", err)
		response.JSON(w, http.StatusBadRequest, response.NewError(err))
		return
	}
	networkMode := req.HostConfig.NetworkMode
	if networkMode == "" || networkMode == "default" {
		networkMode = "bridge"
	}
	if req.NetworkDisabled {
		networkMode = "none"
	}
	dnsOpt := []string{}
	if req.HostConfig.DNSOptions != nil {
		dnsOpt = req.HostConfig.DNSOptions
	}
	netOpt := ncTypes.NetworkOptions{
		Hostname:             req.Hostname,
		NetworkSlice:         []string{networkMode},
		DNSServers:           req.HostConfig.DNS,       // Custom DNS lookup servers.
		DNSResolvConfOptions: dnsOpt,                   // DNS options.
		DNSSearchDomains:     req.HostConfig.DNSSearch, // Custom DNS search domains.
		PortMappings:         portMappings,
		AddHost:              req.HostConfig.ExtraHosts, // Extra hosts.
	}

	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)
	cid, err := h.service.Create(ctx, req.Image, req.Cmd, createOpt, netOpt)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsInvalidFormat(err):
			code = http.StatusBadRequest
		case errdefs.IsConflict(err):
			code = http.StatusConflict
		default:
			code = http.StatusInternalServerError
		}
		logrus.Debugf("Create Container API failed. Status code %d, Message: %s", code, err)
		response.SendErrorResponse(w, code, err)
		return
	}
	response.JSON(w, http.StatusCreated, containerCreateResponse{cid})
}

// translate docker port mappings to go-cni port mappings.
func translatePortMappings(portMappings nat.PortMap) ([]gocni.PortMapping, error) {
	ports := []gocni.PortMapping{}
	if portMappings == nil {
		return ports, nil
	}
	for portName, portBindings := range portMappings {
		for _, portBinding := range portBindings {
			hostPort, err := strconv.ParseInt(portBinding.HostPort, 10, 32)
			if err != nil {
				return []gocni.PortMapping{}, fmt.Errorf("failed to parse host port (%s) to integer: %w", portBinding.HostPort, err)
			}
			// Cannot use portName.Int() because it assumes nat.NewPort() was used
			// for error handling.
			containerPort, err := strconv.ParseInt(portName.Port(), 10, 32)
			if err != nil {
				return []gocni.PortMapping{}, fmt.Errorf("failed to parse container port (%s) to integer: %w", portName, err)
			}
			portMap := gocni.PortMapping{
				HostPort:      int32(hostPort),
				ContainerPort: int32(containerPort),
				Protocol:      portName.Proto(),
				HostIP:        portBinding.HostIP,
			}
			ports = append(ports, portMap)
		}
	}
	return ports, nil
}
