# Using finch-daemon with systemd


# Configuring finch-daemon to support socket activation

This guide provides instructions for setting up and using socket activation for the Finch Daemon with systemd.

### Configure Socket and Service Files

Add the following configuration files to systemd:

### Socket Configuration

Create the socket unit file at `/etc/systemd/system/finch.socket`. An example can be found in [finch-socket-activation.socket](./sample-service-files/finch-socket-activation.socket)

### Service file Configuration

Create the service unit file at /etc/systemd/system/finch.service. An example can be found in [finch-socket-activation.service](./sample-service-files/finch-socket-activation.service)


### Enable the service

```bash
sudo systemctl enable finch.socket finch.service
sudo systemctl start finch.socket
```