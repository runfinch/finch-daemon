# Finch Daemon API Reference

This document provides a reference for the Docker API endpoints implemented by Finch Daemon, including status and unsupported options. Finch Daemon implements a subset of the [Docker API Spec v1.43](https://docs.docker.com/reference/api/engine/version/v1.43/).

For detailed API documentation of supported features, please refer to the generated API docs in the `docs/api/Apis` and `docs/api/Models` directories.

## API Version

Finch Daemon implements Docker API v1.43 as the default version, with minimum supported version v1.35. All API endpoints should be prefixed with `/v1.43`.

## Authentication

Finch Daemon uses Unix socket-based authentication. Access to the socket file (`/run/finch.sock` by default) determines access to the API.

## API Compatibility

Finch Daemon implements a subset of the Docker API v1.43. The following sections detail which endpoints are supported and any limitations or differences from the official Docker API.

## Container Operations

### Create Container

**Endpoint**: `POST /containers/create`

**Status**: ⚠️ Partially Implemented

**Status Codes**:
- `201 Created`: Container created successfully
- `400 Bad Request`: Invalid parameters or configuration
- `404 Not Found`: Image not found
- `409 Conflict`: Container name already in use
- `500 Internal Server Error`: Server error

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

**Status**: ⚠️ Partially Implemented

**Status Codes**:
- `200 OK`: List of containers returned successfully
- `400 Bad Request`: Invalid parameters (e.g., invalid filter format)
- `500 Internal Server Error`: Server error

**Unsupported Filters**:
- `ancestor` - Filter by image used to create the container
- `publish/expose` - Filter by published or exposed port
- `health` - Filter by health status
- `isolation` - Filter by container isolation technology
- `is-task` - Filter by containers that are tasks (part of a service)

### Inspect Container

**Endpoint**: `GET /containers/{id}/json`

**Status**: ✅ Implemented

### Get Container Logs

**Endpoint**: `GET /containers/{id}/logs`

**Status**: ✅ Implemented

### Start Container

**Endpoint**: `POST /containers/{id}/start`

**Status**: ✅ Implemented

**Status Codes**:
- `204 No Content`: Container started successfully
- `304 Not Modified`: Container already started
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

### Stop Container

**Endpoint**: `POST /containers/{id}/stop`

**Status**: ✅ Implemented

**Status Codes**:
- `204 No Content`: Container stopped successfully
- `304 Not Modified`: Container already stopped
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

### Restart Container

**Endpoint**: `POST /containers/{id}/restart`

**Status**: ✅ Implemented

**Status Codes**:
- `204 No Content`: Container restarted successfully
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

### Kill Container

**Endpoint**: `POST /containers/{id}/kill`

**Status**: ✅ Implemented

**Status Codes**:
- `204 No Content`: Container killed successfully
- `404 Not Found`: No such container
- `409 Conflict`: Container is not running
- `500 Internal Server Error`: Server error

### Pause Container

**Endpoint**: `POST /containers/{id}/pause`

**Status**: ✅ Implemented

**Status Codes**:
- `204 No Content`: Container paused successfully
- `404 Not Found`: No such container
- `409 Conflict`: Container is not running or is already paused
- `500 Internal Server Error`: Server error

### Unpause Container

**Endpoint**: `POST /containers/{id}/unpause`

**Status**: ✅ Implemented

**Status Codes**:
- `204 No Content`: Container unpaused successfully
- `404 Not Found`: No such container
- `409 Conflict`: Container is not paused
- `500 Internal Server Error`: Server error

### Remove Container

**Endpoint**: `DELETE /containers/{id}`

**Status**: ⚠️ Partially Implemented

**Status Codes**:
- `204 No Content`: Container removed successfully
- `404 Not Found`: No such container
- `409 Conflict`: Conflict error (e.g., container is running and force is not true)
- `500 Internal Server Error`: Server error

**Unsupported Options**:
- `link` - Remove the specified link (not implemented in nerdctl)

### Attach to Container

**Endpoint**: `POST /containers/{id}/attach`

**Status**: ⚠️ Partially Implemented

**Status Codes**:
- `200 OK` or `101 UPGRADED`: Connection successfully hijacked
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

**Unsupported Options**:
- `detachKeys` - Override the key sequence for detaching a container (not implemented in nerdctl)

**Limitations**:
- Currently only one attach session is allowed. When the second session tries to attach, no error will be returned from nerdctl. However, since behind the scenes, there's only one FIFO for stdin, stdout, and stderr respectively, if there are multiple sessions, all the sessions will be reading from and writing to the same 3 FIFOs, which will result in mixed input and partial output.

### Get Archive from Container

**Endpoint**: `GET /containers/{id}/archive`

**Status**: ✅ Implemented

### Extract Archive to Container

**Endpoint**: `PUT /containers/{id}/archive`

**Status**: ✅ Implemented

### Container Stats

**Endpoint**: `GET /containers/{id}/stats`

**Status**: ⚠️ Partially Implemented

**Status Codes**:
- `200 OK`: Statistics returned successfully
- `404 Not Found`: No such container
- `500 Internal Server Error`: Server error

**Unsupported Options**:
- `one-shot` - Only get a single stat instead of waiting for 2 cycles (not implemented in nerdctl)

### Rename Container

**Endpoint**: `POST /containers/{id}/rename`

**Status**: ✅ Implemented

### Wait Container

**Endpoint**: `POST /containers/{id}/wait`

**Status**: ✅ Implemented

## Image Operations

### List Images

**Endpoint**: `GET /images/json`

**Status**: ✅ Implemented

**Status Codes**:
- `200 OK`: Images listed successfully
- `500 Internal Server Error`: Server error

### Pull Image

**Endpoint**: `POST /images/create`

**Status**: ⚠️ Partially Implemented

**Status Codes**:
- `200 OK`: Pull operation started successfully
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Image not found
- `500 Internal Server Error`: Server error

**Unsupported Options**:
- `fromSrc` - Source to import from (not supported as importing images is not implemented)
- `repo` - Repository name (not supported as importing images is not implemented)
- `message` - Commit message (not supported as importing images is not implemented)
- `changes` - Apply dockerfile instructions to the created image (not supported)

**Limitations**:
- Importing images is not supported, only pulling from registries is implemented

### Push Image

**Endpoint**: `POST /images/{name}/push`

**Status**: ✅ Implemented

### Tag Image

**Endpoint**: `POST /images/{name}/tag`

**Status**: ✅ Implemented

### Remove Image

**Endpoint**: `DELETE /images/{name}`

**Status**: ✅ Implemented

### Inspect Image

**Endpoint**: `GET /images/{name}/json`

**Status**: ✅ Implemented

### Load Image

**Endpoint**: `POST /images/load`

**Status**: ✅ Implemented

## Build

### Build Image

**Endpoint**: `POST /build`

**Status**: ⚠️ Partially Implemented

**Status Codes**:
- `200 OK`: Build operation started successfully
- `400 Bad Request`: Invalid parameters
- `500 Internal Server Error`: Server error

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

## Networks

### List Networks

**Endpoint**: `GET /networks`

**Status**: ✅ Implemented

### Inspect Network

**Endpoint**: `GET /networks/{id}`

**Status**: ✅ Implemented

### Create Network

**Endpoint**: `POST /networks/create`

**Status**: ✅ Implemented

### Connect Container to Network

**Endpoint**: `POST /networks/{id}/connect`

**Status**: ✅ Implemented

### Remove Network

**Endpoint**: `DELETE /networks/{id}`

**Status**: ✅ Implemented

## Volumes

### List Volumes

**Endpoint**: `GET /volumes`

**Status**: ✅ Implemented

### Create Volume

**Endpoint**: `POST /volumes/create`

**Status**: ✅ Implemented

### Inspect Volume

**Endpoint**: `GET /volumes/{name}`

**Status**: ✅ Implemented

### Remove Volume

**Endpoint**: `DELETE /volumes/{name}`

**Status**: ✅ Implemented

## Exec

### Create Exec Instance

**Endpoint**: `POST /containers/{id}/exec`

**Status**: ✅ Implemented

### Start Exec Instance

**Endpoint**: `POST /exec/{id}/start`

**Status**: ✅ Implemented

### Resize Exec Instance

**Endpoint**: `POST /exec/{id}/resize`

**Status**: ✅ Implemented

### Inspect Exec Instance

**Endpoint**: `GET /exec/{id}/json`

**Status**: ✅ Implemented

## Distribution

### Inspect Distribution

**Endpoint**: `GET /distribution/{name}/json`

**Status**: ✅ Implemented

## System

### System Information

**Endpoint**: `GET /info`

**Status**: ✅ Implemented

### Version Information

**Endpoint**: `GET /version`

**Status**: ✅ Implemented

### Ping

**Endpoint**: `GET /_ping`

**Status**: ✅ Implemented

### Authentication

**Endpoint**: `POST /auth`

**Status**: ✅ Implemented

### Events

**Endpoint**: `GET /events`

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
