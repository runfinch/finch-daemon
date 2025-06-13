# Finch Daemon API Reference

This document provides a detailed reference for the Docker API endpoints implemented by Finch Daemon, including supported options, request formats, and response formats. Finch Daemon implements a subset of the [Docker API Spec v1.43](https://docs.docker.com/reference/api/engine/version/v1.43/).

## API Version

Finch Daemon implements Docker API v1.43 as the default version, with minimum supported version v1.35. All API endpoints should be prefixed with `/v1.43`.

## Authentication

Finch Daemon uses Unix socket-based authentication. Access to the socket file (`/run/finch.sock` by default) determines access to the API.

## API Compatibility

Finch Daemon implements a subset of the Docker API v1.43. The following sections detail which endpoints are supported and any limitations or differences from the official Docker API.

## Container Operations

### Create Container

**Endpoint**: `POST /containers/create`

**Description**: Creates a new container.

**Query Parameters**:
- `name` - Assign a name to the container

**Request Body**: Container configuration object with the following structure:

```json
{
  "Hostname": "string",
  "User": "string",
  "AttachStdin": false,
  "ExposedPorts": {
    "80/tcp": {}
  },
  "Tty": false,
  "Env": ["string"],
  "Cmd": ["string"],
  "Image": "string",
  "Volumes": {
    "volumeName": {}
  },
  "WorkingDir": "string",
  "Entrypoint": ["string"],
  "NetworkDisabled": false,
  "MacAddress": "string",
  "Labels": {
    "key": "value"
  },
  "StopSignal": "string",
  "StopTimeout": 10,
  "HostConfig": {
    "Binds": ["string"],
    "ContainerIDFile": "string",
    "LogConfig": {
      "Type": "json-file",
      "Config": {}
    },
    "NetworkMode": "bridge",
    "PortBindings": {
      "80/tcp": [
        {
          "HostIp": "0.0.0.0",
          "HostPort": "8080"
        }
      ]
    },
    "RestartPolicy": {
      "Name": "no",
      "MaximumRetryCount": 0
    },
    "AutoRemove": false,
    "VolumesFrom": ["string"],
    "CapAdd": ["string"],
    "CapDrop": ["string"],
    "CgroupnsMode": "private",
    "Dns": ["string"],
    "DnsOptions": ["string"],
    "DnsSearch": ["string"],
    "ExtraHosts": ["string"],
    "GroupAdd": ["string"],
    "IpcMode": "string",
    "OomKillDisable": false,
    "PidMode": "string",
    "Privileged": false,
    "ReadonlyRootfs": false,
    "SecurityOpt": ["string"],
    "Tmpfs": {
      "/tmp": "size=64m,exec"
    },
    "UTSMode": "string",
    "ShmSize": 67108864,
    "Sysctls": {
      "net.ipv4.ip_forward": "1"
    },
    "Runtime": "runc",
    "CpuShares": 512,
    "CpuPeriod": 100000,
    "CpuQuota": 50000,
    "CpusetCpus": "0-3",
    "CpusetMems": "0-3",
    "Memory": 536870912,
    "MemoryReservation": 268435456,
    "MemorySwap": 1073741824,
    "MemorySwappiness": 60,
    "BlkioWeight": 500,
    "BlkioWeightDevice": [
      {
        "Path": "/dev/sda",
        "Weight": 500
      }
    ],
    "BlkioDeviceReadBps": [
      {
        "Path": "/dev/sda",
        "Rate": 10485760
      }
    ],
    "BlkioDeviceWriteBps": [
      {
        "Path": "/dev/sda",
        "Rate": 10485760
      }
    ],
    "BlkioDeviceReadIOps": [
      {
        "Path": "/dev/sda",
        "Rate": 1000
      }
    ],
    "BlkioDeviceWriteIOps": [
      {
        "Path": "/dev/sda",
        "Rate": 1000
      }
    ],
    "Devices": [
      {
        "PathOnHost": "/dev/deviceName",
        "PathInContainer": "/dev/deviceName",
        "CgroupPermissions": "rwm"
      }
    ],
    "PidsLimit": 500
  }
}
```

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"Image":"nginx","ExposedPorts":{"80/tcp":{}},"HostConfig":{"PortBindings":{"80/tcp":[{"HostPort":"8080"}]}}}' \
  "http://localhost/v1.43/containers/create?name=my-nginx"
```

**Response Format**:
```json
{
  "Id": "3d5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b",
  "Warnings": []
}
```

**Status Codes**:
- `201 Created`: Container created successfully
- `400 Bad Request`: Invalid parameters or configuration
- `404 Not Found`: Image not found
- `409 Conflict`: Container name already in use
- `500 Internal Server Error`: Server error

**Status**: ⚠️ Partially Implemented

**Supported Options**:

**ContainerConfig**:
- **Hostname**: Container host name
- **User**: Username or UID to run commands inside the container
- **ExposedPorts**: Ports to expose from the container without publishing to the host
- **Tty**: Attach standard streams to a TTY, including stdin if it is not closed
- **Env**: List of environment variables in the form ["VAR=value", ...]
- **Cmd**: Command to run when starting the container
- **Image**: Name of the image as it was passed by the operator
- **Volumes**: List of volumes (mounts) used for the container
- **WorkingDir**: Current directory (PWD) in the command will be launched
- **Entrypoint**: Entrypoint to run when starting the container
- **NetworkDisabled**: Disable networking for the container
- **MacAddress**: Container MAC address (e.g., "12:34:56:78:9a:bc")
- **Labels**: Key-value map of container metadata
- **StopSignal**: Signal to stop a container (e.g., "SIGTERM")
- **StopTimeout**: Timeout (in seconds) to stop a container

**HostConfig**:
- **Binds**: List of volume bindings for this container (e.g., ["/host:/container:ro"])
- **ContainerIDFile**: File path where the container ID is written
- **LogConfig**: Log configuration for the container
- **NetworkMode**: Network mode to use for the container (e.g., "bridge", "host", "none")
- **PortBindings**: Port mapping between the exposed port (container) and the host
- **RestartPolicy**: Restart policy to be used for the container
- **AutoRemove**: Automatically remove the container when it exits
- **VolumesFrom**: List of volumes to take from other containers
- **CapAdd**: List of kernel capabilities to add to the container
- **CapDrop**: List of kernel capabilities to remove from the container
- **CgroupnsMode**: Cgroup namespace mode to use for the container ("host" or "private")
- **Dns**: List of DNS servers for the container
- **DnsOptions**: List of DNS options for the container
- **DnsSearch**: List of DNS search domains for the container
- **ExtraHosts**: List of hostnames/IP mappings to add to /etc/hosts
- **GroupAdd**: List of additional groups that the container process will run as
- **IpcMode**: IPC namespace to use for the container
- **OomKillDisable**: Whether to disable OOM Killer for the container
- **PidMode**: PID namespace to use for the container
- **Privileged**: Give extended privileges to this container
- **ReadonlyRootfs**: Mount the container's root filesystem as read only
- **SecurityOpt**: List of security options for the container
- **Tmpfs**: Temporary filesystems to mount (e.g., {"/tmp": "size=64m,exec"})
- **UTSMode**: UTS namespace to use for the container
- **ShmSize**: Size of /dev/shm in bytes
- **Sysctls**: Kernel parameters to set in the container
- **Runtime**: Runtime to use for this container
- **CpuShares**: CPU shares (relative weight vs. other containers)
- **CpuPeriod**: CPU CFS (Completely Fair Scheduler) period
- **CpuQuota**: CPU CFS (Completely Fair Scheduler) quota
- **CpusetCpus**: CPUs in which to allow execution (e.g., "0-3", "0,1")
- **CpusetMems**: Memory nodes (MEMs) in which to allow execution (e.g., "0-3", "0,1")
- **Memory**: Memory limit in bytes
- **MemoryReservation**: Memory soft limit in bytes
- **MemorySwap**: Total memory usage (memory + swap); set `-1` to enable unlimited swap
- **MemorySwappiness**: Tune container memory swappiness (0 to 100)
- **Ulimits**: List of ulimits to set in the container
- **BlkioWeight**: Block IO weight (relative weight vs. other containers)
- **BlkioWeightDevice**: Block IO weight for specific devices
- **BlkioDeviceReadBps**: Limit read rate from a device (bytes per second)
- **BlkioDeviceWriteBps**: Limit write rate to a device (bytes per second)
- **BlkioDeviceReadIOps**: Limit read rate from a device (IO per second)
- **BlkioDeviceWriteIOps**: Limit write rate to a device (IO per second)
- **Devices**: Expose host devices to the container
- **PidsLimit**: Tune container pids limit (set -1 for unlimited)

**Unsupported Options**:

**ContainerConfig**:
- **Domainname**: Container NIS domain name
- **AttachStdout**: Attach the standard output
- **AttachStderr**: Attach the standard error
- **AttachStdin**: Attach the standard input (explicitly rejected)
- **OpenStdin**: Open stdin
- **StdinOnce**: Close stdin after the 1 attached client disconnects
- **Healthcheck**: Healthcheck configuration for the container
- **ArgsEscaped**: True if command is already escaped (Windows specific)
- **OnBuild**: ONBUILD metadata that were defined on the image Dockerfile
- **Shell**: Shell for shell-form of RUN, CMD, ENTRYPOINT

**HostConfig**:
- **VolumeDriver**: Name of the volume driver used to mount volumes
- **ConsoleSize**: Initial console size [height,width]
- **Cgroup**: Cgroup to use for the container
- **Links**: List of links (in the name:alias form)
- **OomScoreAdj**: Tune container's OOM preferences (-1000 to 1000)
- **PublishAllPorts**: Publish all exposed ports to random ports on the host
- **StorageOpt**: Storage driver options per container
- **UsernsMode**: User namespace to use for the container
- **Isolation**: Isolation technology of the container (e.g., default, hyperv)
- **Mounts**: Specification for mounts to be added to the container
- **MaskedPaths**: Paths to mask inside the container
- **ReadonlyPaths**: Paths to set as read-only inside the container
- **Init**: Run an init inside the container

### List Containers

**Endpoint**: `GET /containers/json`

**Description**: Returns a list of containers.

**Query Parameters**:
- `all` - Show all containers (default shows just running). When set to `false` (default), only running containers are returned. When set to `true`, all containers are returned, including stopped containers.
- `limit` - Show `limit` last created containers. Limits the number of containers returned based on creation time, showing the most recently created containers first.
- `size` - Show the containers sizes. When set to `true`, includes disk size information for each container. Default is `false`.
- `filters` - Filter output based on conditions provided as a JSON encoded value. Format: `{"filter_name":{"filter_value":true}}` or legacy format: `{"filter_name":["filter_value"]}`.
- `format` - Format the output using the given Go template. Supported formats include:
  - `table` (default): Table format
  - `{{json .}}`: JSON format
  - `wide`: Wide table format
  - `json`: Alias of `{{json .}}`
- `no-trunc` - Don't truncate output
- `quiet` - Only display container IDs

**Response Format**:
```json
[
  {
    "Id": "8dfafdbc3a40",
    "Names": ["/boring_feynman"],
    "Image": "ubuntu:latest",
    "ImageID": "d74508fb6632491cea586a1fd7d748dfc5274cd6fdfedee309ecdcbc2bf5cb82",
    "Command": "echo 1",
    "Created": 1367854155,
    "State": "running",
    "Status": "",
    "Ports": [
      {
        "PrivatePort": 2222,
        "PublicPort": 3333,
        "Type": "tcp"
      }
    ],
    "Labels": {
      "com.example.vendor": "Acme",
      "com.example.license": "GPL"
    },
    "NetworkSettings": {
      "Networks": {
        "bridge": {
          "IPAddress": "172.17.0.3",
          "Gateway": "172.17.0.1"
        }
      }
    },
    "Mounts": [
      {
        "Type": "bind",
        "Source": "/host/path",
        "Destination": "/container/path",
        "Mode": "ro,Z",
        "RW": false,
        "Propagation": "rprivate"
      }
    ]
  }
]
```

**Status Codes**:
- `200 OK`: List of containers returned successfully
- `400 Bad Request`: Invalid parameters (e.g., invalid filter format)
- `500 Internal Server Error`: Server error

**Supported Filters**:
- `id` - Filter by container ID (both full ID and truncated ID are supported)
- `name` - Filter by container name
- `label` - Filter by container labels (e.g., `{"label":{"key=value":true}}` or `{"label":["key=value"]}`)
- `status` - Filter by container status (created, running, paused, stopped, exited, pausing, unknown). Note that restarting, removing, dead are not supported and will be ignored. When specifying this condition, it filters all containers.
- `exited` - Filter by container's exit code (only works with `--all`)
- `before` - Filter by container created before a given container ID or name
- `since` - Filter by container created since a given container ID or name
- `volume` - Filter by a given mounted volume or bind mount
- `network` - Filter by a given network

**Unsupported Filters**:
- `ancestor` - Filter by image used to create the container
- `publish/expose` - Filter by published or exposed port
- `health` - Filter by health status
- `isolation` - Filter by container isolation technology
- `is-task` - Filter by containers that are tasks (part of a service)

**Example**:
```bash
# List all containers
curl --unix-socket /run/finch.sock "http://localhost/v1.43/containers/json?all=true"

# List containers with a specific label
curl --unix-socket /run/finch.sock "http://localhost/v1.43/containers/json?filters={\"label\":{\"com.example.key=value\":true}}"

# List only exited containers
curl --unix-socket /run/finch.sock "http://localhost/v1.43/containers/json?all=true&filters={\"status\":[\"exited\"]}"
```

**Status**: ⚠️ Partially Implemented

### Inspect Container

**Endpoint**: `GET /containers/{id}/json`

**Description**: Returns detailed information about a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `size` - Include the size of the container

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/json"
```

**Status**: ✅ Implemented

### Get Container Logs

**Endpoint**: `GET /containers/{id}/logs`

**Description**: Get stdout and stderr logs from a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `follow` - Follow log output
- `stdout` - Show stdout log
- `stderr` - Show stderr log
- `since` - Show logs since timestamp
- `until` - Show logs before timestamp
- `timestamps` - Show timestamps
- `tail` - Number of lines to show from the end of the logs

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/logs?stdout=true&stderr=true"
```

**Status**: ✅ Implemented

### Start Container

**Endpoint**: `POST /containers/{id}/start`

**Description**: Start a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `detachKeys` - Override the key sequence for detaching a container

**Response Format**:
On success, this endpoint returns a `204 No Content` response with no response body.

**Status Codes**:
- `204 No Content`: Container started successfully
- `304 Not Modified`: Container already started
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/start"
```

**Status**: ✅ Implemented

### Stop Container

**Endpoint**: `POST /containers/{id}/stop`

**Description**: Stop a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `t` - Seconds to wait before killing the container (default: 10)

**Response Format**:
On success, this endpoint returns a `204 No Content` response with no response body.

**Status Codes**:
- `204 No Content`: Container stopped successfully
- `304 Not Modified`: Container already stopped
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/stop?t=5"
```

**Status**: ✅ Implemented

### Restart Container

**Endpoint**: `POST /containers/{id}/restart`

**Description**: Restart a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `t` - Seconds to wait before killing the container (default: 10)

**Response Format**:
On success, this endpoint returns a `204 No Content` response with no response body.

**Status Codes**:
- `204 No Content`: Container restarted successfully
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/restart?t=5"
```

**Status**: ✅ Implemented

### Kill Container

**Endpoint**: `POST /containers/{id}/kill`

**Description**: Kill a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `signal` - Signal to send to the container (default: "KILL")

**Response Format**:
On success, this endpoint returns a `204 No Content` response with no response body.

**Status Codes**:
- `204 No Content`: Container killed successfully
- `404 Not Found`: No such container
- `409 Conflict`: Container is not running
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/kill?signal=SIGTERM"
```

**Status**: ✅ Implemented

### Pause Container

**Endpoint**: `POST /containers/{id}/pause`

**Description**: Pause all processes within a container.

**Path Parameters**:
- `id` - Container ID or name

**Response Format**:
On success, this endpoint returns a `204 No Content` response with no response body.

**Status Codes**:
- `204 No Content`: Container paused successfully
- `404 Not Found`: No such container
- `409 Conflict`: Container is not running or is already paused
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/pause"
```

**Status**: ✅ Implemented

### Unpause Container

**Endpoint**: `POST /containers/{id}/unpause`

**Description**: Unpause a paused container.

**Path Parameters**:
- `id` - Container ID or name

**Response Format**:
On success, this endpoint returns a `204 No Content` response with no response body.

**Status Codes**:
- `204 No Content`: Container unpaused successfully
- `404 Not Found`: No such container
- `409 Conflict`: Container is not paused
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/unpause"
```

**Status**: ✅ Implemented

### Remove Container

**Endpoint**: `DELETE /containers/{id}`

**Description**: Remove a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `v` - Remove anonymous volumes associated with the container (implemented as `--volumes` in nerdctl)
- `force` - Force the removal of a running|paused|unknown container (uses SIGKILL) (implemented as `--force` in nerdctl)

**Response Format**:
On success, this endpoint returns a `204 No Content` response with no response body.

**Status Codes**:
- `204 No Content`: Container removed successfully
- `404 Not Found`: No such container
- `409 Conflict`: Conflict error (e.g., container is running and force is not true)
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X DELETE --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx?force=true&v=true"
```

**Unsupported Options**:
- `link` - Remove the specified link (not implemented in nerdctl)

**Status**: ⚠️ Partially Implemented

### Attach to Container

**Endpoint**: `POST /containers/{id}/attach`

**Description**: Attach to a container to read its output or send input.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `logs` - Return logs
- `stream` - Return stream
- `stdin` - Attach to stdin
- `stdout` - Attach to stdout
- `stderr` - Attach to stderr

**Response Format**:
On success, this endpoint returns a hijacked connection for streaming container output and/or input.

**Status Codes**:
- `200 OK` or `101 UPGRADED`: Connection successfully hijacked
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/attach?stdout=true&stderr=true"
```

**Unsupported Options**:
- `detachKeys` - Override the key sequence for detaching a container (not implemented in nerdctl)

**Limitations**:
- Currently only one attach session is allowed. When the second session tries to attach, no error will be returned from nerdctl. However, since behind the scenes, there's only one FIFO for stdin, stdout, and stderr respectively, if there are multiple sessions, all the sessions will be reading from and writing to the same 3 FIFOs, which will result in mixed input and partial output.

**Status**: ⚠️ Partially Implemented

### Get Archive from Container

**Endpoint**: `GET /containers/{id}/archive`

**Description**: Get a tar archive of a resource in the container filesystem.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `path` - Resource path in the container

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/archive?path=/etc/nginx/nginx.conf"
```

**Status**: ✅ Implemented

### Extract Archive to Container

**Endpoint**: `PUT /containers/{id}/archive`

**Description**: Extract a tar archive to a path in the container filesystem.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `path` - Path to extract the archive to
- `noOverwriteDirNonDir` - If "1", "true", or "True" then it will be an error if unpacking the given content would cause an existing directory to be replaced with a non-directory and vice versa

**Example**:
```bash
curl -X PUT --unix-socket /run/finch.sock -H "Content-Type: application/x-tar" \
  --data-binary '@files.tar' \
  "http://localhost/v1.43/containers/my-nginx/archive?path=/app"
```

**Status**: ✅ Implemented

### Container Stats

**Endpoint**: `GET /containers/{id}/stats`

**Description**: Get container resource usage statistics.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `stream` - Stream statistics (default: true)

**Response Format**:
On success, this endpoint returns a JSON object with container statistics. If streaming is enabled, it continuously sends updated statistics.

**Status Codes**:
- `200 OK`: Statistics returned successfully
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/stats"
```

**Unsupported Options**:
- `one-shot` - Only get a single stat instead of waiting for 2 cycles (not implemented in nerdctl)

**Status**: ⚠️ Partially Implemented

### Rename Container

**Endpoint**: `POST /containers/{id}/rename`

**Description**: Rename a container.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `name` - New name for the container

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/rename?name=new-nginx"
```

**Status**: ✅ Implemented

### Wait Container

**Endpoint**: `POST /containers/{id}/wait`

**Description**: Block until a container stops, then returns the exit code.

**Path Parameters**:
- `id` - Container ID or name

**Query Parameters**:
- `condition` - Wait until a container state reaches the given condition, either 'not-running' (default), 'next-exit', or 'removed'

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/containers/my-nginx/wait"
```

**Status**: ✅ Implemented

## Image Operations

### List Images

**Endpoint**: `GET /images/json`

**Description**: Returns a list of images.

**Query Parameters**:
- `all` - Show all images (default hides intermediate images) (unimplemented in nerdctl)
- `filters` - Filter output based on conditions provided as a JSON encoded value
- `digests` - Show digests (compatible with Docker, unlike ID)
- `quiet` - Only show numeric IDs
- `no-trunc` - Don't truncate output
- `format` - Format the output using the given Go template. Supported formats include:
  - `table` (default): Table format
  - `{{json .}}`: JSON format
  - `wide`: Wide table format
  - `json`: Alias of `{{json .}}`

**Response Format**:
On success, this endpoint returns a JSON array of image objects.

**Status Codes**:
- `200 OK`: Images listed successfully
- `500 Internal Server Error`: Server error

**Supported Filters**:
- `dangling` - Show dangling images (true/false)
- `label` - Filter by image label (e.g., `label=<key>` or `label=<key>=<value>`)
- `before` - Filter by image created before given image (exclusive)
- `since` - Filter by image created since given image (exclusive)
- `reference` - Filter by image reference (matches both Docker compatible wildcard pattern and regexp match)

**Example**:
```bash
# List all images
curl --unix-socket /run/finch.sock "http://localhost/v1.43/images/json"

# List only dangling images
curl --unix-socket /run/finch.sock "http://localhost/v1.43/images/json?filters={\"dangling\":[\"true\"]}"

# List images with a specific label
curl --unix-socket /run/finch.sock "http://localhost/v1.43/images/json?filters={\"label\":[\"maintainer=nginx\"]}"

# List images in JSON format
curl --unix-socket /run/finch.sock "http://localhost/v1.43/images/json?format=json"
```

**Status**: ✅ Implemented

### Pull Image

**Endpoint**: `POST /images/create`

**Description**: Pull an image from a registry.

**Query Parameters**:
- `fromImage` - Name of the image to pull
- `tag` - Tag of the image to pull
- `platform` - Platform in the format os[/arch[/variant]]

**Headers**:
- `X-Registry-Auth` - Base64-encoded AuthConfig object

**Response Format**:
On success, this endpoint returns a stream of JSON objects with status updates about the pull operation.

**Status Codes**:
- `200 OK`: Pull operation started successfully
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Image not found
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/images/create?fromImage=nginx&tag=latest"
```

**Unsupported Options**:
- `fromSrc` - Source to import from (not supported as importing images is not implemented)
- `repo` - Repository name (not supported as importing images is not implemented)
- `message` - Commit message (not supported as importing images is not implemented)
- `changes` - Apply dockerfile instructions to the created image (not supported)

**Limitations**:
- Importing images is not supported, only pulling from registries is implemented

**Status**: ⚠️ Partially Implemented

### Push Image

**Endpoint**: `POST /images/{name}/push`

**Description**: Push an image to a registry.

**Path Parameters**:
- `name` - Image name

**Query Parameters**:
- `tag` - The tag to associate with the image on the registry

**Headers**:
- `X-Registry-Auth` - Base64-encoded AuthConfig object

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "X-Registry-Auth: {base64_auth_json}" \
  "http://localhost/v1.43/images/myregistry/myimage/push?tag=latest"
```

**Status**: ✅ Implemented

### Tag Image

**Endpoint**: `POST /images/{name}/tag`

**Description**: Tag an image.

**Path Parameters**:
- `name` - Image name or ID

**Query Parameters**:
- `repo` - The repository to tag in
- `tag` - The name of the new tag

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/images/nginx/tag?repo=myregistry/nginx&tag=v1"
```

**Status**: ✅ Implemented

### Remove Image

**Endpoint**: `DELETE /images/{name}`

**Description**: Remove an image.

**Path Parameters**:
- `name` - Image name or ID

**Query Parameters**:
- `force` - Force removal of the image
- `noprune` - Do not delete untagged parent images

**Example**:
```bash
curl -X DELETE --unix-socket /run/finch.sock "http://localhost/v1.43/images/nginx?force=true"
```

**Status**: ✅ Implemented

### Inspect Image

**Endpoint**: `GET /images/{name}/json`

**Description**: Return detailed information about an image.

**Path Parameters**:
- `name` - Image name or ID

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/images/nginx/json"
```

**Status**: ✅ Implemented

### Load Image

**Endpoint**: `POST /images/load`

**Description**: Load an image from a tar archive.

**Query Parameters**:
- `quiet` - Suppress progress details during load

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/x-tar" \
  --data-binary '@image.tar' \
  "http://localhost/v1.43/images/load"
```

**Status**: ✅ Implemented

## Build

### Build Image

**Endpoint**: `POST /build`

**Description**: Build an image from a Dockerfile.

**Query Parameters**:
- `t` - Name and optionally a tag in the 'name:tag' format
- `q` - Suppress verbose build output
- `nocache` - Do not use the cache when building the image
- `rm` - Remove intermediate containers after a successful build (default: true)
- `dockerfile` - Name of the Dockerfile (default: "Dockerfile")
- `target` - Set the target build stage to build
- `platform` - Set platform if server is multi-platform capable
- `buildargs` - Build-time variables as JSON
- `labels` - Set metadata for an image as JSON
- `cachefrom` - Images to consider as cache sources as JSON
- `networkmode` - Set the networking mode for the RUN instructions during build
- `output` - Output destination

**Response Format**:
On success, this endpoint returns a stream of JSON objects with status updates about the build operation.

**Status Codes**:
- `200 OK`: Build operation started successfully
- `400 Bad Request`: Invalid parameters
- `500 Internal Server Error`: Server error

**Request Body**: Tar archive containing Dockerfile and build context.

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/tar" \
  --data-binary '@build.tar' \
  "http://localhost/v1.43/build?t=myapp:latest"
```

**Unsupported Options**:
- `squash` - Squash the resulting images layers into a single layer
- `forcerm` - Always remove intermediate containers, even upon failure
- `memory` - Memory limit for the build container
- `memswap` - Total memory (memory + swap) limit for the build container
- `cpushares` - CPU shares for the build container
- `cpusetcpus` - CPUs in which to allow execution for the build container
- `cpuperiod` - CPU CFS (Completely Fair Scheduler) period for the build container
- `cpuquota` - CPU CFS (Completely Fair Scheduler) quota for the build container
- `buildid` - BuildID for the build
- `shmsize` - Size of /dev/shm for the build container
- `ulimits` - Ulimit options for the build container
- `compress` - Compress the build context using gzip
- `securityopt` - Security options for the build container
- `extrahosts` - Add hosts to the build container's /etc/hosts
- `isolation` - Container isolation technology

**Status**: ⚠️ Partially Implemented

## Networks

### List Networks

**Endpoint**: `GET /networks`

**Description**: Returns a list of networks.

**Query Parameters**:
- `filters` - Filter output based on conditions provided as a JSON encoded value

**Supported Filters**:
- `driver` - Filter by network driver
- `id` - Filter by network ID
- `label` - Filter by network label
- `name` - Filter by network name
- `scope` - Filter by network scope (swarm, global, local)
- `type` - Filter by network type (custom, builtin)

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/networks"
```

**Status**: ✅ Implemented

### Inspect Network

**Endpoint**: `GET /networks/{id}`

**Description**: Return detailed information about a network.

**Path Parameters**:
- `id` - Network ID or name

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/networks/bridge"
```

**Status**: ✅ Implemented

### Create Network

**Endpoint**: `POST /networks/create`

**Description**: Create a network.

**Request Body**: Network configuration object:

```json
{
  "Name": "my-network",
  "Driver": "bridge",
  "IPAM": {
    "Driver": "default",
    "Config": [
      {
        "Subnet": "172.20.0.0/16",
        "Gateway": "172.20.0.1"
      }
    ]
  },
  "Options": {
    "com.docker.network.bridge.name": "my-bridge"
  },
  "Labels": {
    "key": "value"
  }
}
```

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"Name":"my-network","Driver":"bridge"}' \
  "http://localhost/v1.43/networks/create"
```

**Status**: ✅ Implemented

### Connect Container to Network

**Endpoint**: `POST /networks/{id}/connect`

**Description**: Connect a container to a network.

**Path Parameters**:
- `id` - Network ID or name

**Request Body**: Connection configuration:

```json
{
  "Container": "container_id_or_name",
  "EndpointConfig": {
    "IPAddress": "172.20.0.2",
    "IPAMConfig": {
      "IPv4Address": "172.20.0.2"
    }
  }
}
```

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"Container":"my-nginx"}' \
  "http://localhost/v1.43/networks/my-network/connect"
```

**Status**: ✅ Implemented

### Remove Network

**Endpoint**: `DELETE /networks/{id}`

**Description**: Remove a network.

**Path Parameters**:
- `id` - Network ID or name

**Example**:
```bash
curl -X DELETE --unix-socket /run/finch.sock "http://localhost/v1.43/networks/my-network"
```

**Status**: ✅ Implemented

## Volumes

### List Volumes

**Endpoint**: `GET /volumes`

**Description**: Returns a list of volumes.

**Query Parameters**:
- `filters` - Filter output based on conditions provided as a JSON encoded value

**Supported Filters**:
- `driver` - Filter by volume driver
- `label` - Filter by volume label
- `name` - Filter by volume name

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/volumes"
```

**Status**: ✅ Implemented

### Create Volume

**Endpoint**: `POST /volumes/create`

**Description**: Create a volume.

**Request Body**: Volume configuration object:

```json
{
  "Name": "my-volume",
  "Driver": "local",
  "DriverOpts": {
    "type": "tmpfs",
    "device": "tmpfs",
    "o": "size=100m,uid=1000"
  },
  "Labels": {
    "key": "value"
  }
}
```

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"Name":"my-volume"}' \
  "http://localhost/v1.43/volumes/create"
```

**Status**: ✅ Implemented

### Inspect Volume

**Endpoint**: `GET /volumes/{name}`

**Description**: Return detailed information about a volume.

**Path Parameters**:
- `name` - Volume name

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/volumes/my-volume"
```

**Status**: ✅ Implemented

### Remove Volume

**Endpoint**: `DELETE /volumes/{name}`

**Description**: Remove a volume.

**Path Parameters**:
- `name` - Volume name

**Query Parameters**:
- `force` - Force removal of the volume

**Example**:
```bash
curl -X DELETE --unix-socket /run/finch.sock "http://localhost/v1.43/volumes/my-volume?force=true"
```

**Status**: ✅ Implemented

## Exec

### Create Exec Instance

**Endpoint**: `POST /containers/{id}/exec`

**Description**: Create an exec instance in a container.

**Path Parameters**:
- `id` - Container ID or name

**Request Body**: Exec configuration object:

```json
{
  "AttachStdin": false,
  "AttachStdout": true,
  "AttachStderr": true,
  "DetachKeys": "ctrl-p,ctrl-q",
  "Tty": false,
  "Cmd": ["ls", "-la"],
  "Env": ["FOO=bar"],
  "WorkingDir": "/root",
  "Privileged": false,
  "User": "root"
}
```

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"AttachStdout":true,"AttachStderr":true,"Cmd":["ls","-la"]}' \
  "http://localhost/v1.43/containers/my-nginx/exec"
```

**Status**: ✅ Implemented

### Start Exec Instance

**Endpoint**: `POST /exec/{id}/start`

**Description**: Start an exec instance.

**Path Parameters**:
- `id` - Exec instance ID

**Request Body**: Exec start configuration:

```json
{
  "Detach": false,
  "Tty": false
}
```

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"Detach":false,"Tty":false}' \
  "http://localhost/v1.43/exec/{exec_id}/start"
```

**Status**: ✅ Implemented

### Resize Exec Instance

**Endpoint**: `POST /exec/{id}/resize`

**Description**: Resize the TTY session used by an exec instance.

**Path Parameters**:
- `id` - Exec instance ID

**Query Parameters**:
- `h` - Height of the TTY session in characters
- `w` - Width of the TTY session in characters

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock "http://localhost/v1.43/exec/{exec_id}/resize?h=40&w=80"
```

**Status**: ✅ Implemented

### Inspect Exec Instance

**Endpoint**: `GET /exec/{id}/json`

**Description**: Return detailed information about an exec instance.

**Path Parameters**:
- `id` - Exec instance ID

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/exec/{exec_id}/json"
```

**Status**: ✅ Implemented

## Distribution

### Inspect Distribution

**Endpoint**: `GET /distribution/{name}/json`

**Description**: Return image digest and platform information.

**Path Parameters**:
- `name` - Image name

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/distribution/nginx/json"
```

**Status**: ✅ Implemented

## System

### System Information

**Endpoint**: `GET /info`

**Description**: Get system information.

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/info"
```

**Status**: ✅ Implemented

### Version Information

**Endpoint**: `GET /version`

**Description**: Get version information.

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/version"
```

**Status**: ✅ Implemented

### Ping

**Endpoint**: `GET /_ping`

**Description**: Ping the Docker server.

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/_ping"
```

**Status**: ✅ Implemented

### Authentication

**Endpoint**: `POST /auth`

**Description**: Authenticate with a registry.

**Request Body**: Authentication configuration:

```json
{
  "username": "myusername",
  "password": "mypassword",
  "serveraddress": "https://index.docker.io/v1/"
}
```

**Example**:
```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"username":"myusername","password":"mypassword","serveraddress":"https://index.docker.io/v1/"}' \
  "http://localhost/v1.43/auth"
```

**Status**: ✅ Implemented

### Events

**Endpoint**: `GET /events`

**Description**: Monitor Docker events.

**Query Parameters**:
- `since` - Show events created since this timestamp
- `until` - Show events created until this timestamp
- `filters` - Filter output based on conditions provided

**Example**:
```bash
curl --unix-socket /run/finch.sock "http://localhost/v1.43/events"
```

**Status**: ✅ Implemented

## Unimplemented Endpoints

The following Docker API endpoints are **not** currently implemented in Finch Daemon:

- `POST /containers/{id}/update` - Update a container
- `POST /containers/{id}/prune` - Delete stopped containers
- `POST /containers/{id}/resize` - Resize a container TTY
- `POST /containers/{id}/export` - Export a container
- `POST /containers/{id}/changes` - Get changes on a container's filesystem
- `POST /images/{name}/prune` - Delete unused images
- `POST /images/prune` - Delete unused images
- `POST /images/{name}/commit` - Create a new image from a container
- `POST /images/{name}/get` - Export an image
- `POST /images/get` - Export several images
- `POST /images/search` - Search images
- `POST /networks/{id}/disconnect` - Disconnect a container from a network
- `POST /networks/prune` - Delete unused networks
- `POST /volumes/prune` - Delete unused volumes
- `POST /plugins` - Plugin related endpoints

## Docker Swarm API

All Docker Swarm related endpoints are **not implemented** and there is **no plan to implement** them:

- `POST /swarm` - Swarm related endpoints
- `POST /services` - Swarm service related endpoints
- `POST /tasks` - Swarm task related endpoints
- `POST /secrets` - Swarm secret related endpoints
- `POST /configs` - Swarm config related endpoints
- `POST /nodes` - Swarm node related endpoints

## API Response Formats

Most API endpoints return JSON-formatted responses. The exact format of each response follows the Docker API specification.

## Error Handling

Errors are returned as JSON objects with the following structure:

```json
{
  "message": "Error message"
}
```

HTTP status codes are used to indicate the success or failure of an API request:

- 2xx: Success
- 4xx: Client error
- 5xx: Server error

## Further Reading

For more detailed information about the Docker API, refer to the [official Docker API documentation](https://docs.docker.com/reference/api/engine/version/v1.43/).
