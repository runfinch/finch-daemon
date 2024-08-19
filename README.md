# finch-daemon

The Finch Daemon project is a container runtime engine that enables users to seamelssly integrate their software which has programmatic dependencies on Docker RESTful APIs. This project currently implements the [Docker API Spec v1.43](https://docs.docker.com/engine/api/v1.43/).

## Onboarding & Development

Please review [CONTRIBUTING.md](./CONTRIBUTING.md) for onboarding as well for an overview of the development cycle for this project. Additionally, check the [Makefile](./Makefile) for additional information on setup & configuration.

## Quickstart

### macOS
Note that with macOS, it is not possible to run unit tests with `make run-unit-tests` or do code-gen with `make code-gen` directly as this is a Linux project. It is possible, however, to build the project on macOS with `make` and run the linux-related commands in the finch vm.

1. Get finch, `brew install finch`
2. Add the unix socket forwarding to `/Applications/Finch/lima/data/finch/lima.yaml`:
   ```yaml
   portForwards:
   - guestSocket: "/run/fnich.sock"
     hostSocket: "{{.Dir}}/sock/finch.sock"
   ```
3. Init the vm to apply the changes (or restart if Finch was already running):
   ```bash
   # init
   finch vm init
   
   # restart
   finch vm stop
   finch vm start
   ```
4. Pull finch-daemon from the GitHub repo to somewhere under your home directory: `git clone https://github.com/runfinch/finch-daemon.git`
5. Check if you have go installed with `go version`, if not, install go with `brew install go`
6. Build finch-daemon with `make`
7. Spin up the finch-daemon server in finch's vm:
   ```bash
   LIMA_HOME=/Applications/Finch/lima/data /Applications/Finch/lima/bin/limactl shell finch
   cd <finch-daemon-dir>
   sudo bin/finch-daemon --debug --socket-owner $UID
   ```
8. Test the project and use finch to confirm changes
   - Using curl:
     ```bash
     curl -v -X GET --unix-socket /Applications/Finch/lima/data/finch/sock/finch.sock 'http://localhost/version'
     ```
   - Using Docker CLI:
     ```bash
     DOCKER_HOST=unix:///Applications/Finch/lima/data/finch/sock/finch.sock docker images
     ```
   - Using the Docker python SDK (after installing python3):
     ```bash
     # Install docker python SDK
     python3 -m pip install docker
     # Start python 
     DOCKER_HOST=unix:///Applications/Finch/lima/data/finch/sock/finch.sock python3
     ``` 
     ```python
     >>> import docker
     >>> client = docker.from_env()
     >>> con = client.containers.get("test-container")
     >>> con.attach(stream=True, logs=True, demux=True)
     ```