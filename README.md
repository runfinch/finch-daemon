[[⬇️ **Download]**](https://github.com/runfinch/finch-daemon/releases)
[[🚀 **All Releases]**](https://github.com/runfinch/finch-daemon/releases)
[[📚 **Installation]**](#quickstart)
[[✏️ **Contributing]**](CONTRIBUTING.md)

# Finch Daemon

The Finch Daemon project is an open source container runtime engine that enables users to integrate any software that uses Docker's RESTful APIs as a programmatic dependency. Some core features include:

 - Full implementation of the [Docker API Spec v1.43](https://docs.docker.com/engine/api/v1.43/)
 - Native support for Linux environments

## Onboarding & Development

Please review [CONTRIBUTING.md](./CONTRIBUTING.md) for onboarding, as well as for an overview of the development cycle for this project.
Additionally, check the [Makefile](./Makefile) for additional information on setup & configuration.

## Quickstart
Make sure [NerdCTL](https://github.com/containerd/nerdctl) and 
[Containerd](https://github.com/containerd/containerd) are installed and set up

### Linux
Getting started with Finch Daemon on Linux only requires a few steps:

1. Clone the repository - `git clone https://github.com/runfinch/finch-daemon.git`
2. `cd finch-daemon`
3. Install Go and make sure that `go version` is >= 1.22
4. Build and spin up the finch-daemon server with 
   ```bash 
   make
   sudo bin/finch-daemon --debug --socket-owner $UID
   ```
5. Test any changes with `make run-unit-tests` and `sudo make run-e2e-tests`


## Creating a systemd service
Sometimes, you may wish to have a script be controlled by systemd, 
or to have the script restart by itself when it gets killed. 
You can configure such a service using systemd on Linux by following these steps:

1. `cd /path/to/finch-daemon`
2. `sudo cp finch.service /etc/systemd/system/`
3. Refresh the service files to include the new service - `sudo systemctl daemon-reload`
4. Start the service - `sudo systemctl start finch.service`
5. To check the status of the service - `sudo systemctl status finch.service`
6. To enable the service on every reboot - `sudo systemctl enable finch.service`
7. To disable the service on every reboot - `sudo systemctl disable finch.service`

-----


Legal
=====

*Brought to you courtesy of our legal counsel. For more context,
please see the [NOTICE](https://github.com/runfinch/finch-daemon/blob/main/NOTICE) document in this repo.*

Use and transfer of Finch Daemon may be subject to certain restrictions by the
United States and other governments.

It is your responsibility to ensure that your use and/or transfer does not
violate applicable laws.

For more information, please see https://www.bis.doc.gov

Licensing
=========
Finch Daemon is licensed under the Apache License, Version 2.0. 
See [LICENSE](https://github.com/runfinch/finch-daemon/blob/main/LICENSE) for the full license text.