# Finch Daemon Documentation

For a quick overview and getting started guide, please refer to the [README.md](../README.md) in the root of the repository.

## Table of Contents
- [Overview](#overview)
- [Installation & Setup](#installation--setup)
- [Configuration](#configuration)
- [API Support](#api-support)
- [Docker CLI Integration](#docker-cli-integration)
- [System Integration](#system-integration)
- [Debugging & Troubleshooting](#debugging--troubleshooting)
- [Contributing](#contributing)
- [API Reference](#api-reference)
- [Experimental Features](#experimental-features)

## Overview

### What is Finch Daemon?

Finch Daemon is an open-source container runtime engine that enables users to integrate software that uses Docker's RESTful APIs as a programmatic dependency. It serves as a bridge between Docker-compatible client applications and the underlying container technologies like containerd and nerdctl.

As described in the [README.md](../README.md), the Finch Daemon project is actively taking contributions, especially to improve API compatibility.

### Key Features

- Partial implementation of the [Docker API Spec v1.43](https://docs.docker.com/engine/api/v1.43/)
- Native support for Linux environments
- Integration with containerd and nerdctl
- Systemd service support
- Socket activation

## Installation & Setup

### Prerequisites

Before installing Finch Daemon, ensure you have the following prerequisites:

1. [Containerd](https://github.com/containerd/containerd) - Container runtime
2. [NerdCTL](https://github.com/containerd/nerdctl) - Docker-compatible CLI for containerd

### Installation Steps

#### From Source

Follow these steps as outlined in the [README.md Quickstart section](../README.md#quickstart):

1. Clone the repository:
   ```bash
   git clone https://github.com/runfinch/finch-daemon.git
   ```

2. Change to the project directory:
   ```bash
   cd finch-daemon
   ```

3. Install Go (version >= 1.22) if not already installed.

4. Build the daemon:
   ```bash
   make
   ```

5. Run the daemon:
   ```bash
   sudo bin/finch-daemon --debug --socket-owner $UID
   ```

6. Test any changes with:
   ```bash
   make test-unit
   sudo make test-e2e
   ```

### Quick Start

Once the daemon is running, you can interact with it using any Docker API client by pointing it to the Finch Daemon socket (default: `/run/finch.sock`).

## Configuration

Finch Daemon can be configured through command-line flags and a TOML configuration file.

### Command-line Options

| **Flag**            | **Description**                                        | **Default Value**        |
|---------------------|--------------------------------------------------------|--------------------------|
| `--socket-addr`     | The Unix socket address where the server listens.      | `/run/finch.sock`        |
| `--debug`           | Enable debug-level logging.                            | `false`                  |
| `--socket-owner`    | Set the UID and GID of the server socket owner.        | `-1` (no owner)          |
| `--debug-addr`      | Address for the debugging HTTP server (pprof).         | (empty, disabled)        |
| `--config-file`     | Path to the daemon's configuration file (TOML format). | `/etc/finch/finch.toml` |
| `--pidfile`         | Path to the PID file.                                  | `/run/finch.pid`        |

Example usage:
```bash
finch-daemon --socket-addr /tmp/finch.sock --debug --socket-owner 1001 --config-file /path/to/config.toml
```

### Configuration File (finch.toml)

Finch Daemon uses a TOML configuration file to configure nerdctl parameters. The default location is `/etc/finch/finch.toml`.

Example configuration:

```toml
debug          = false
debug_full     = false
address        = "unix:///run/k3s/containerd/containerd.sock"
namespace      = "k8s.io"
snapshotter    = "soci"
cgroup_manager = "cgroupfs"
hosts_dir      = ["/etc/containerd/certs.d", "/etc/docker/certs.d"]
experimental   = true
```

#### Configuration Properties

| **TOML Property**   | **CLI Flag(s)**                         | **Description**                                                                                                            |
|---------------------|------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| `debug`             | `--debug`                                | Enable debug mode.                                                                                                         |
| `debug_full`        | `--debug-full`                           | Enable debug mode with full output.                                                                                        |
| `address`           | `--address`, `--host`, `-a`, `-H`        | Address of the containerd daemon.                                                                                          |
| `namespace`         | `--namespace`, `-n`                      | containerd namespace.                                                                                                      |
| `snapshotter`       | `--snapshotter`, `--storage-driver`      | containerd snapshotter or storage driver.                                                                                  |
| `cni_path`          | `--cni-path`                             | Directory containing CNI binaries.                                                                                         |
| `cni_netconfpath`   | `--cni-netconfpath`                      | Directory containing CNI network configurations.                                                                           |
| `data_root`         | `--data-root`                            | Directory to store persistent state.                                                                                       |
| `cgroup_manager`    | `--cgroup-manager`                       | cgroup manager to use.                                                                                                     |
| `insecure_registry` | `--insecure-registry`                    | Allow communication with insecure registries.                                                                              |
| `hosts_dir`         | `--hosts-dir`                            | Directory for `certs.d` files.                                                                                             |
| `experimental`      | `--experimental`                         | Enable experimental features.                                                                                              |
| `host_gateway_ip`   | `--host-gateway-ip`                      | IP address for the special 'host-gateway' in `--add-host`. Defaults to the host IP. Has no effect without `--add-host`.     |

## API Support

Finch Daemon implements a subset of the [Docker API Spec v1.43](https://docs.docker.com/engine/api/v1.43/). The implementation focuses on the most commonly used endpoints.

For a detailed reference of the Docker API endpoints implemented by Finch Daemon, including supported options, request formats, and response formats, please refer to the [API Reference](./api/README.md) document.

### Supported API Endpoints

Based on the repository structure, the following API endpoints are supported:

#### Container Operations
- Create containers
- List containers
- Inspect containers
- Start/stop/restart containers
- Pause/unpause containers
- Kill containers
- Remove containers
- Get container logs
- Attach to containers
- Execute commands in containers
- Get container stats
- Get container filesystem archives

#### Image Operations
- List images
- Pull images
- Push images
- Remove images
- Tag images
- Build images

#### Network Operations
- Create networks
- List networks
- Inspect networks
- Remove networks

#### Volume Operations
- Create volumes
- List volumes
- Inspect volumes
- Remove volumes

#### System Operations
- Get system information
- Get version information
- Get events

### API Usage Examples

#### List Containers

```bash
curl --unix-socket /run/finch.sock http://localhost/v1.43/containers/json
```

#### Create a Container

```bash
curl -X POST --unix-socket /run/finch.sock -H "Content-Type: application/json" \
  -d '{"Image":"nginx","ExposedPorts":{"80/tcp":{}},"HostConfig":{"PortBindings":{"80/tcp":[{"HostPort":"8080"}]}}}' \
  http://localhost/v1.43/containers/create
```

#### Start a Container

```bash
curl -X POST --unix-socket /run/finch.sock http://localhost/v1.43/containers/{container_id}/start
```

## Docker CLI Integration

You can use the Docker CLI with Finch Daemon by setting the `DOCKER_HOST` environment variable to point to the Finch Daemon socket.

### Setting up Docker CLI

```bash
export DOCKER_HOST=unix:///run/finch.sock
```

With this configuration, Docker CLI commands will be sent to Finch Daemon instead of the Docker daemon.

### Example Commands

```bash
# List containers
docker ps

# Run a container
docker run -d -p 8080:80 nginx

# Build an image
docker build -t myapp .

# Push an image
docker push myregistry/myapp
```

## System Integration

### Running as a Systemd Service

Finch Daemon can be configured to run as a systemd service for automatic startup and management, as described in the [README.md](../README.md#creating-a-systemd-service).

#### Standard Service Setup

1. Copy the service file to the systemd directory:
   ```bash
   sudo cp docs/sample-service-files/finch.service /etc/systemd/system/
   ```

2. Reload systemd to recognize the new service:
   ```bash
   sudo systemctl daemon-reload
   ```

3. Start the service:
   ```bash
   sudo systemctl start finch.service
   ```

4. Enable the service to start on boot:
   ```bash
   sudo systemctl enable finch.service
   ```

5. Check the service status:
   ```bash
   sudo systemctl status finch.service
   ```

6. To disable the service on every reboot:
   ```bash
   sudo systemctl disable finch.service
   ```

#### Socket Activation

Socket activation allows systemd to start Finch Daemon on demand when a client connects to its socket.

1. Copy the socket and service files:
   ```bash
   sudo cp docs/sample-service-files/finch-socket-activation.socket /etc/systemd/system/finch.socket
   sudo cp docs/sample-service-files/finch-socket-activation.service /etc/systemd/system/finch.service
   ```

2. Reload systemd:
   ```bash
   sudo systemctl daemon-reload
   ```

3. Enable and start the socket:
   ```bash
   sudo systemctl enable finch.socket
   sudo systemctl start finch.socket
   ```

The service will now start automatically when a client connects to the socket.

## Debugging & Troubleshooting

### Logs

If Finch Daemon is running as a systemd service, you can view its logs using journalctl:

```bash
sudo journalctl -u finch
```

If you started Finch Daemon manually, logs are output to stderr/stdout.

### Development Setup

For detailed information on onboarding and the development cycle, please refer to the [CONTRIBUTING.md](../CONTRIBUTING.md) file as mentioned in the [README.md](../README.md#onboarding--development).

1. Fork and clone the repository
2. Install dependencies
3. Make your changes
4. Run tests: `make test-unit` and `sudo make test-e2e`
5. Submit a pull request

### Testing

Finch Daemon includes unit tests and end-to-end tests:

- Unit tests: `make test-unit`
- End-to-end tests: `sudo make test-e2e`

### Pull Request Process

1. Ensure your code passes all tests
2. Update documentation as needed
3. Submit a pull request with a clear description of the changes
4. Address any feedback from reviewers

## Experimental Features

As described in the [README.md](../README.md#experimental-features), Finch Daemon includes experimental features that can be enabled using the `--experimental` flag.

### Using Experimental Features

To enable experimental features, use the `--experimental` flag when starting the daemon:

```bash
sudo bin/finch-daemon --debug --socket-owner $UID --experimental
```

### Current Experimental Features

#### OPA Authorization Middleware

The OPA (Open Policy Agent) middleware allows you to define authorization policies for API requests using Rego policy language. This feature requires both the `--experimental` flag and the `--rego-file` flag to be set.

Example usage:
```bash
sudo bin/finch-daemon --debug --socket-owner $UID --experimental --rego-file /path/to/policy.rego
```

For detailed documentation on the OPA middleware, see [opa-middleware.md](opa-middleware.md).
