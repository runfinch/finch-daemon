#!/bin/bash
# Set versions
RUNC_VERSION=1.3.0
NERDCTL_VERSION=2.1.2
BUILDKIT_VERSION=0.23.2
CNI_VERSION=1.6.2

apt update && apt install -y make gcc linux-libc-dev libseccomp-dev pkg-config git

# Download and install containerd
curl -sSL --output /tmp/containerd.tgz https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-${TARGETARCH:-amd64}.tar.gz
tar zxvf /tmp/containerd.tgz -C /usr/local/
rm -f /tmp/containerd.tgz

# Download and install runc
curl -sSL --output /tmp/runc https://github.com/opencontainers/runc/releases/download/v${RUNC_VERSION}/runc.${TARGETARCH:-amd64}
cp /tmp/runc /usr/local/bin/
chmod +x /usr/local/bin/runc
rm -f /tmp/runc
# Download and install nerdctl
curl -sSL --output /tmp/nerdctl.tgz https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${TARGETARCH:-amd64}.tar.gz
tar zxvf /tmp/nerdctl.tgz -C /usr/local/bin/
rm -f /tmp/nerdctl.tgz

#Download and install buildkit
curl -sSL --output /tmp/buildkit.tgz https://github.com/moby/buildkit/releases/download/v${BUILDKIT_VERSION}/buildkit-v${BUILDKIT_VERSION}.linux-amd64.tar.gz
tar zxvf /tmp/buildkit.tgz -C /usr/local/bin/
sudo mv /usr/local/bin/bin/* /usr/local/bin/
rm /tmp/buildkit.tgz

#Download and install cni-plugins
sudo rm -rf /opt/cni/bin/*
cd && wget https://github.com/containernetworking/plugins/releases/download/v${CNI_VERSION}/cni-plugins-linux-amd64-v${CNI_VERSION}.tgz
sudo mkdir -p /opt/cni/bin
sudo tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v${CNI_VERSION}.tgz

export PATH=$PATH:/usr/local/bin

echo "Starting containerd..."
sudo containerd &
CONTAINERD_PID=$!
echo "Containerd started with PID: $CONTAINERD_PID"

# Wait for containerd to be ready (up to 60 seconds)
echo "Waiting for containerd to be ready..."
for i in {1..60}; do
  if sudo ctr version >/dev/null 2>&1; then
    echo "containerd is ready after ${i} seconds"
    break
  fi
  if [ $i -eq 60 ]; then
    echo "ERROR: containerd failed to start after 60 seconds"
    exit 1
  fi
  sleep 1
done

# Extra conservative wait
sleep 3

echo "Starting buildkitd..."
sudo buildkitd &
BUILDKIT_PID=$!
echo "Buildkitd started with PID: $BUILDKIT_PID"

# Wait for buildkitd to be ready (up to 60 seconds)
echo "Waiting for buildkitd to be ready..."
for i in {1..60}; do
  if sudo buildctl debug workers >/dev/null 2>&1; then
    echo "buildkitd is ready after ${i} seconds"
    break
  fi
  if [ $i -eq 60 ]; then
    echo "ERROR: buildkitd failed to start after 60 seconds"
    exit 1
  fi
  sleep 1
done

# Find and fix buildkit socket permissions
echo "Finding buildkit socket..."
for i in {1..10}; do
  SOCKET=$(find /run /var/run /tmp -name "buildkitd.sock" 2>/dev/null | head -1)
  if [ -n "$SOCKET" ]; then
    echo "Found buildkit socket: $SOCKET"
    sudo chmod 666 "$SOCKET"
    echo "Fixed permissions on $SOCKET"
    break
  fi
  echo "Socket not found, waiting... (attempt $i/10)"
  sleep 2
done

# Extra conservative wait for full initialization
sleep 5

echo "All daemons are ready. PIDs: containerd=$CONTAINERD_PID, buildkitd=$BUILDKIT_PID"
