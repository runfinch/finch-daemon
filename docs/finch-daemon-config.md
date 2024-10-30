
# Configuring finch-daemon specific config

Finch Daemon takes parameters to configure daemon specific parameters.

 **Flag**            | **Description**                                        | **Default Value**        |
|---------------------|--------------------------------------------------------|--------------------------|
| `--socket-addr`     | The Unix socket address where the server listens.      | `/run/finch.sock`        |
| `--debug`           | Enable debug-level logging.                            | `false`                  |
| `--socket-owner`    | Set the UID and GID of the server socket owner.        | `-1` (no owner)          |
| `--config-file`     | Path to the daemon's configuration file (TOML format). | `/etc/finch/finch.toml` |


Example usage:
```bash
finch-daemon --socket-addr /tmp/finch.sock --debug --socket-owner 1001 --config-file /path/to/config.toml
```

# Configuring nerdctl with `finch.toml`

Finch daemon toml config is used to configure nerdctl parameters. For more details refer to nerdctl github page. [nerdctl configuration guide](https://github.com/containerd/nerdctl/blob/main/docs/config.md).


## File path
- `/etc/finch/finch.toml`

## Example

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

## Properties

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
| `experimental`      | `--experimental`                         | Enable [experimental features].                                                                          |
| `host_gateway_ip`   | `--host-gateway-ip`                      | IP address for the special 'host-gateway' in `--add-host`. Defaults to the host IP. Has no effect without `--add-host`.     |
