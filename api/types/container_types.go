// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"io"
	"os"
	"time"

	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/dockercompat"
	dockertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

// AttachOptions defines the available options for the container attach call.
type AttachOptions struct {
	GetStreams func() (io.Writer, io.Writer, chan os.Signal, func(), error)
	UseStdin   bool
	UseStdout  bool
	UseStderr  bool
	Logs       bool
	Stream     bool
	// TODO: DetachKeys string
	MuxStreams bool
}

// ContainerConfig is from https://github.com/moby/moby/blob/v24.0.2/api/types/container/config.go#L64-L96
type ContainerConfig struct {
	Hostname string `json:",omitempty"` // Hostname
	// TODO: Domainname   string      // Domainname
	User        string `json:",omitempty"` // User that will run the command(s) inside the container, also support user:group
	AttachStdin bool   // Attach the standard input, makes possible user interaction
	// TODO: AttachStdout bool        // Attach the standard output
	// TODO: AttachStderr bool        // Attach the standard error
	ExposedPorts nat.PortSet `json:",omitempty"` // List of exposed ports
	Tty          bool        // Attach standard streams to a tty, including stdin if it is not closed.
	// TODO: OpenStdin    bool        // Open stdin
	// TODO: StdinOnce    bool        // If true, close stdin after the 1 attached client disconnects.
	Env []string `json:",omitempty"` // List of environment variable to set in the container
	Cmd []string `json:",omitempty"` // Command to run when starting the container
	// TODO Healthcheck     *HealthConfig       `json:",omitempty"` // Healthcheck describes how to check the container is healthy
	// TODO: ArgsEscaped     bool                `json:",omitempty"` // True if command is already escaped (meaning treat as a command line) (Windows specific).
	Image      string              // Name of the image as it was passed by the operator (e.g. could be symbolic)
	Volumes    map[string]struct{} `json:",omitempty"` // List of volumes (mounts) used for the container
	WorkingDir string              `json:",omitempty"` // Current directory (PWD) in the command will be launched
	Entrypoint []string            `json:",omitempty"` // Entrypoint to run when starting the container
	// TODO: NetworkDisabled bool                `json:",omitempty"` // Is network disabled
	// TODO: MacAddress      string              `json:",omitempty"` // Mac Address of the container
	// TODO: OnBuild         []string            // ONBUILD metadata that were defined on the image Dockerfile
	Labels      map[string]string `json:",omitempty"` // List of labels set to this container
	StopSignal  string            `json:",omitempty"` // Signal to stop a container
	StopTimeout *int              `json:",omitempty"` // Timeout (in seconds) to stop a container
	// TODO: Shell           []string            `json:",omitempty"` // Shell for shell-form of RUN, CMD, ENTRYPOINT
}

// HostConfig is from https://github.com/moby/moby/blob/v24.0.2/api/types/container/hostconfig.go#L376-L436
type ContainerHostConfig struct {
	// Applicable to all platforms
	Binds []string // List of volume bindings for this container
	// TODO: ContainerIDFile string            // File (path) where the containerId is written
	LogConfig     LogConfig     // Configuration of the logs for this container
	NetworkMode   string        // Network mode to use for the container
	PortBindings  nat.PortMap   // Port mapping between the exposed port (container) and the host
	RestartPolicy RestartPolicy // Restart policy to be used for the container
	AutoRemove    bool          // Automatically remove container when it exits
	// TODO: VolumeDriver    string            // Name of the volume driver used to mount volumes
	// TODO: VolumesFrom     []string          // List of volumes to take from other container
	// TODO: ConsoleSize     [2]uint           // Initial console size (height,width)
	// TODO: Annotations     map[string]string `json:",omitempty"` // Arbitrary non-identifying metadata attached to container and provided to the runtime

	// Applicable to UNIX platforms
	CapAdd []string // List of kernel capabilities to add to the container
	// TODO: CapDrop         strslice.StrSlice // List of kernel capabilities to remove from the container
	// TODO: CgroupnsMode    CgroupnsMode      // Cgroup namespace mode to use for the container
	DNS        []string `json:"Dns"`        // List of DNS server to lookup
	DNSOptions []string `json:"DnsOptions"` // List of DNSOption to look for
	DNSSearch  []string `json:"DnsSearch"`  // List of DNSSearch to look for
	ExtraHosts []string // List of extra hosts
	// TODO: GroupAdd        []string          // List of additional groups that the container process will run as
	// TODO: IpcMode         IpcMode           // IPC namespace to use for the container
	// TODO: Cgroup          CgroupSpec        // Cgroup to use for the container
	// TODO: Links           []string          // List of links (in the name:alias form)
	// TODO: OomScoreAdj     int               // Container preference for OOM-killing
	// TODO: PidMode         PidMode           // PID namespace to use for the container
	// TODO: Privileged      bool              // Is the container in privileged mode
	// TODO: PublishAllPorts bool              // Should docker publish all exposed port for the container
	// TODO: ReadonlyRootfs  bool              // Is the container root filesystem in read-only
	// TODO: SecurityOpt     []string          // List of string values to customize labels for MLS systems, such as SELinux.
	// TODO: StorageOpt      map[string]string `json:",omitempty"` // Storage driver options per container.
	// TODO: Tmpfs           map[string]string `json:",omitempty"` // List of tmpfs (mounts) used for the container
	// TODO: UTSMode         UTSMode           // UTS namespace to use for the container
	// TODO: UsernsMode      UsernsMode        // The user namespace to use for the container
	// TODO: ShmSize         int64             // Total shm memory usage
	// TODO: Sysctls         map[string]string `json:",omitempty"` // List of Namespaced sysctls used for the container
	// TODO: Runtime         string            `json:",omitempty"` // Runtime to use with this container

	// Applicable to Windows
	// TODO: Isolation Isolation // Isolation technology of the container (e.g. default, hyperv)

	// Contains container's resources (cgroups, ulimits)
	CPUShares int64 `json:"CpuShares"` // CPU shares (relative weight vs. other containers)
	Memory    int64 // Memory limit (in bytes)
	// TODO: Resources

	// Mounts specs used by the container
	// TODO: Mounts []mount.Mount `json:",omitempty"`

	// MaskedPaths is the list of paths to be masked inside the container (this overrides the default set of paths)
	// TODO: MaskedPaths []string

	// ReadonlyPaths is the list of paths to be set as read-only inside the container (this overrides the default set of paths)
	// TODO: ReadonlyPaths []string

	// Run a custom init inside the container, if null, use the daemon's configured settings
	// TODO: Init *bool `json:",omitempty"`
}

// LogConfig represents the logging configuration of the container.
// From https://github.com/moby/moby/blob/v24.0.2/api/types/container/hostconfig.go#L319-L323
type LogConfig struct {
	Type   string
	Config map[string]string
}

// RestartPolicy represents the restart policies of the container.
// From https://github.com/moby/moby/blob/v24.0.2/api/types/container/hostconfig.go#L272-L276
type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

type ContainerCreateRequest struct {
	ContainerConfig
	HostConfig ContainerHostConfig
	// TODO: NetworkingConfig ContainerNetworkingConfig
}

// Container mimics a `docker container inspect` object.
// From https://github.com/moby/moby/blob/v24.0.2/api/types/types.go#L445-L486
type Container struct {
	ID             string `json:"Id"`
	Created        string
	Path           string
	Args           []string
	State          *dockercompat.ContainerState
	Image          string
	ResolvConfPath string
	HostnamePath   string
	// TODO: HostsPath      string
	LogPath string
	// Unimplemented: Node            *ContainerNode `json:",omitempty"` // Node is only propagated by Docker Swarm standalone API
	Name         string
	RestartCount int
	Driver       string
	Platform     string
	// TODO: MountLabel      string
	// TODO: ProcessLabel    string
	AppArmorProfile string
	// TODO: ExecIDs         []string
	// TODO: HostConfig      *container.HostConfig
	// TODO: GraphDriver     GraphDriverData
	// TODO: SizeRw     *int64 `json:",omitempty"`
	// TODO: SizeRootFs *int64 `json:",omitempty"`

	Mounts          []dockercompat.MountPoint
	Config          *ContainerConfig
	NetworkSettings *dockercompat.NetworkSettings
}

type ContainerListItem struct {
	Id              string   `json:"Id"`
	Names           []string `json:"Names"`
	Image           string
	CreatedAt       int64  `json:"Created"`
	State           string `json:"State"`
	Labels          map[string]string
	NetworkSettings *dockercompat.NetworkSettings
	Mounts          []dockercompat.MountPoint
	// TODO: Other fields
}

// LogsOptions defines the available options for the container logs call.
type LogsOptions struct {
	GetStreams func() (io.Writer, io.Writer, chan os.Signal, func(), error)
	Stdout     bool
	Stderr     bool
	Follow     bool
	Since      int64
	Until      int64
	Timestamps bool
	Tail       string
	MuxStreams bool
}

// PutArchiveOptions defines the parameters for [PutContainerArchive API](https://docs.docker.com/engine/api/v1.41/#tag/Container/operation/PutContainerArchive)
type PutArchiveOptions struct {
	ContainerId string
	Path        string
	Overwrite   bool
	CopyUIDGID  bool
}

// CPUStats aggregates and wraps all CPU related info of container
// From https://github.com/moby/moby/blob/v24.0.2/api/types/stats.go#L42-L55
type CPUStats struct {
	// CPU Usage. Linux and Windows.
	CPUUsage dockertypes.CPUUsage `json:"cpu_usage"`

	// System Usage. Linux only.
	SystemUsage uint64 `json:"system_cpu_usage,omitempty"`

	// Online CPUs. Linux only.
	OnlineCPUs uint32 `json:"online_cpus,omitempty"`

	// Throttling Data. Linux only.
	// TODO: ThrottlingData ThrottlingData `json:"throttling_data,omitempty"`
}

// Stats is Ultimate struct aggregating all types of stats of one container
// From https://github.com/moby/moby/blob/v24.0.2/api/types/stats.go#L152-L170
type Stats struct {
	// Common stats
	Read    time.Time `json:"read"`
	PreRead time.Time `json:"preread"`

	// Linux specific stats, not populated on Windows.
	PidsStats  dockertypes.PidsStats  `json:"pids_stats,omitempty"`
	BlkioStats dockertypes.BlkioStats `json:"blkio_stats,omitempty"`

	// Windows specific stats, not populated on Linux.
	// NumProcs     uint32       `json:"num_procs"`
	// StorageStats StorageStats `json:"storage_stats,omitempty"`

	// Shared stats
	CPUStats    CPUStats                `json:"cpu_stats,omitempty"`
	PreCPUStats CPUStats                `json:"precpu_stats,omitempty"` // "Pre"="Previous"
	MemoryStats dockertypes.MemoryStats `json:"memory_stats,omitempty"`
}

// StatsJSON is the JSON response for container stats api
// From https://github.com/moby/moby/blob/v24.0.2/api/types/stats.go#L172-L181
type StatsJSON struct {
	Stats

	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`

	// Networks request version >=1.21
	Networks map[string]dockertypes.NetworkStats `json:"networks,omitempty"`
}
