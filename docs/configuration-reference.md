# Finch Daemon Configuration Reference

This document provides a comprehensive reference for all configuration options available in Finch Daemon. Configuration can be specified through command-line flags and a TOML configuration file.

## Command-line Options

Finch Daemon accepts the following command-line options:

| **Flag**            | **Description**                                        | **Default Value**        | **Example**                |
|---------------------|--------------------------------------------------------|--------------------------|----------------------------|
| `--socket-addr`     | The Unix socket address where the server listens.      | `/run/finch.sock`        | `--socket-addr=/tmp/finch.sock` |
| `--debug`           | Enable debug-level logging.                            | `false`                  | `--debug`                  |
| `--socket-owner`    | Set the UID and GID of the server socket owner.        | `-1` (no owner)          | `--socket-owner=1000`      |
| `--debug-addr`      | Address for the debugging HTTP server (pprof).         | (empty, disabled)        | `--debug-addr=localhost:6060` |
| `--config-file`     | Path to the daemon's configuration file (TOML format). | `/etc/finch/finch.toml` | `--config-file=/path/to/config.toml` |
| `--pidfile`         | Path to the PID file.                                  | `/run/finch.pid`        | `--pidfile=/tmp/finch.pid` |

### Socket Address (`--socket-addr`)

The Unix socket address where Finch Daemon listens for API requests. This is the endpoint that Docker clients will connect to.

- **Default**: `/run/finch.sock`
- **Special value**: `fd://` - Use this value when using socket activation with systemd.

Example:
```bash
finch-daemon --socket-addr=/tmp/custom-finch.sock
```

### Debug Mode (`--debug`)

Enables debug-level logging, which provides more detailed information about the daemon's operations. This is useful for troubleshooting issues.

- **Default**: `false` (disabled)

Example:
```bash
finch-daemon --debug
```

### Socket Owner (`--socket-owner`)

Sets the UID and GID of the server socket owner. This determines who can access the socket. By default, the socket is owned by the user who starts the daemon (typically root).

- **Default**: `-1` (no explicit owner)
- **Note**: This option is not supported when using socket activation (`--socket-addr=fd://`).

Example:
```bash
finch-daemon --socket-owner=1000
```

### Debug Address (`--debug-addr`)

Specifies the address for the debugging HTTP server, which provides access to Go's pprof profiling tools. When set, Finch Daemon will start an HTTP server at the specified address that serves debugging endpoints.

- **Default**: (empty, disabled)

Example:
```bash
finch-daemon --debug-addr=localhost:6060
```

### Config File (`--config-file`)

Path to the daemon's configuration file in TOML format. This file contains settings for nerdctl, which is used by Finch Daemon.

- **Default**: `/etc/finch/finch.toml`

Example:
```bash
finch-daemon --config-file=/path/to/custom-config.toml
```

### PID File (`--pidfile`)

Path to the PID file, which stores the process ID of the running daemon. This is used to ensure only one instance of the daemon is running.

- **Default**: `/run/finch.pid`

Example:
```bash
finch-daemon --pidfile=/tmp/finch.pid
```

## Configuration File (finch.toml)

Finch Daemon uses a TOML configuration file to configure nerdctl parameters. The default location is `/etc/finch/finch.toml`.

### Example Configuration

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

### Configuration Properties

| **TOML Property**   | **CLI Flag(s)**                         | **Description**                                                                                                            | **Default Value**                |
|---------------------|------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|----------------------------------|
| `debug`             | `--debug`                                | Enable debug mode.                                                                                                         | `false`                          |
| `debug_full`        | `--debug-full`                           | Enable debug mode with full output.                                                                                        | `false`                          |
| `address`           | `--address`, `--host`, `-a`, `-H`        | Address of the containerd daemon.                                                                                          | `unix:///run/containerd/containerd.sock` |
| `namespace`         | `--namespace`, `-n`                      | containerd namespace.                                                                                                      | `finch` (overrides default)      |
| `snapshotter`       | `--snapshotter`, `--storage-driver`      | containerd snapshotter or storage driver.                                                                                  | (containerd default)             |
| `cni_path`          | `--cni-path`                             | Directory containing CNI binaries.                                                                                         | (nerdctl default)                |
| `cni_netconfpath`   | `--cni-netconfpath`                      | Directory containing CNI network configurations.                                                                           | (nerdctl default)                |
| `data_root`         | `--data-root`                            | Directory to store persistent state.                                                                                       | (nerdctl default)                |
| `cgroup_manager`    | `--cgroup-manager`                       | cgroup manager to use.                                                                                                     | (nerdctl default)                |
| `insecure_registry` | `--insecure-registry`                    | Allow communication with insecure registries.                                                                              | `false`                          |
| `hosts_dir`         | `--hosts-dir`                            | Directory for `certs.d` files.                                                                                             | (nerdctl default)                |
| `experimental`      | `--experimental`                         | Enable experimental features.                                                                                              | `false`                          |
| `host_gateway_ip`   | `--host-gateway-ip`                      | IP address for the special 'host-gateway' in `--add-host`. Defaults to the host IP. Has no effect without `--add-host`.     | (auto-detected)                  |

### Debug (`debug`)

Enables debug mode for nerdctl operations. This provides more detailed logging.

- **Default**: `false`

### Full Debug (`debug_full`)

Enables debug mode with full output for nerdctl operations. This provides even more detailed logging than the standard debug mode.

- **Default**: `false`

### Containerd Address (`address`)

The address of the containerd daemon that Finch Daemon will connect to. This is typically a Unix socket path.

- **Default**: `unix:///run/containerd/containerd.sock`

Example:
```toml
address = "unix:///path/to/containerd.sock"
```

### Namespace (`namespace`)

The containerd namespace to use. Namespaces provide isolation between different users of containerd.

- **Default**: `finch` (overrides the default namespace)

Example:
```toml
namespace = "myapp"
```

### Snapshotter (`snapshotter`)

The containerd snapshotter or storage driver to use. Snapshotters are responsible for managing container filesystem layers.

- **Default**: (containerd default, typically `overlayfs`)

Example:
```toml
snapshotter = "soci"
```

### CNI Path (`cni_path`)

Directory containing CNI (Container Network Interface) binaries. CNI is used for container networking.

- **Default**: (nerdctl default)

Example:
```toml
cni_path = "/opt/cni/bin"
```

### CNI Network Configuration Path (`cni_netconfpath`)

Directory containing CNI network configurations. These configurations define how container networks are set up.

- **Default**: (nerdctl default)

Example:
```toml
cni_netconfpath = "/etc/cni/net.d"
```

### Data Root (`data_root`)

Directory to store persistent state, such as downloaded images and container data.

- **Default**: (nerdctl default)

Example:
```toml
data_root = "/var/lib/finch"
```

### Cgroup Manager (`cgroup_manager`)

The cgroup manager to use for container resource control. Options include `cgroupfs` and `systemd`.

- **Default**: (nerdctl default)

Example:
```toml
cgroup_manager = "cgroupfs"
```

### Insecure Registry (`insecure_registry`)

Allow communication with insecure registries (those without TLS certificates). This is useful for development or internal registries.

- **Default**: `false`

Example:
```toml
insecure_registry = true
```

### Hosts Directory (`hosts_dir`)

Directory for `certs.d` files, which contain registry certificates and configuration. This can be a single path or an array of paths.

- **Default**: (nerdctl default)

Example:
```toml
hosts_dir = ["/etc/containerd/certs.d", "/etc/docker/certs.d"]
```

### Experimental Features (`experimental`)

Enable experimental features in nerdctl. This may enable features that are not yet stable.

- **Default**: `false`

Example:
```toml
experimental = true
```

### Host Gateway IP (`host_gateway_ip`)

IP address for the special 'host-gateway' in `--add-host`. Defaults to the host IP. Has no effect without `--add-host`.

- **Default**: (auto-detected)

Example:
```toml
host_gateway_ip = "192.168.1.1"
```

## Environment Variables

Finch Daemon respects the following environment variables:

### DOCKER_CONFIG

When building an image via finch-daemon, buildctl uses this variable to load auth configs. Finch Daemon automatically sets this to `$HOME/.finch` for the user specified by `--socket-owner`.

## Configuration Precedence

Configuration options are applied in the following order of precedence (highest to lowest):

1. Command-line flags
2. Configuration file (finch.toml)
3. Default values

This means that command-line flags override settings in the configuration file, and both override default values.

## Configuration Examples

### Basic Configuration

```bash
finch-daemon --debug --socket-owner=$UID
```

### Custom Socket and Config File

```bash
finch-daemon --socket-addr=/tmp/finch.sock --config-file=/path/to/config.toml
```

### With Debugging Enabled

```bash
finch-daemon --debug --debug-addr=localhost:6060
```

### Socket Activation with Systemd

```bash
finch-daemon --socket-addr=fd:// --debug
```

## Configuration File Examples

### Basic Configuration

```toml
debug = true
namespace = "finch"
```

### Custom Containerd Address

```toml
address = "unix:///path/to/containerd.sock"
namespace = "finch"
```

### With Experimental Features

```toml
experimental = true
debug = true
namespace = "finch"
```

### With Custom Network Configuration

```toml
cni_path = "/opt/cni/bin"
cni_netconfpath = "/etc/cni/net.d"
namespace = "finch"
```

## See Also

- [Main Documentation](./DOCUMENTATION.md)
- [API Reference](./api-reference.md)
- [Docker CLI Setup](./docker-cli-setup.md)
