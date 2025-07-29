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

#Download and install cni-plugins with retry logic
echo "Installing CNI plugins..."
sudo rm -rf /opt/cni/bin/*
sudo mkdir -p /opt/cni/bin

# Retry logic for CNI download
CNI_SUCCESS=false
for attempt in 1 2 3; do
  echo "CNI download attempt $attempt/3..."
  cd /tmp
  rm -f cni-plugins-linux-amd64-v${CNI_VERSION}.tgz
  
  if wget --timeout=30 --tries=3 https://github.com/containernetworking/plugins/releases/download/v${CNI_VERSION}/cni-plugins-linux-amd64-v${CNI_VERSION}.tgz; then
    if [ -f "cni-plugins-linux-amd64-v${CNI_VERSION}.tgz" ]; then
      echo "CNI download successful, extracting..."
      if sudo tar -xzf cni-plugins-linux-amd64-v${CNI_VERSION}.tgz -C /opt/cni/bin/; then
        # Verify critical plugins exist
        if [ -f "/opt/cni/bin/bridge" ] && [ -f "/opt/cni/bin/loopback" ] && [ -f "/opt/cni/bin/host-local" ]; then
          echo "CNI plugins installed successfully"
          CNI_SUCCESS=true
          break
        else
          echo "CNI extraction failed - missing critical plugins"
        fi
      else
        echo "CNI extraction failed"
      fi
    else
      echo "CNI download failed - file not found"
    fi
  else
    echo "CNI download failed - network error"
  fi
  
  if [ $attempt -lt 3 ]; then
    echo "Retrying in 10 seconds..."
    sleep 10
  fi
done

if [ "$CNI_SUCCESS" = false ]; then
  echo "ERROR: Failed to install CNI plugins after 3 attempts"
  echo "Listing /opt/cni/bin contents:"
  ls -la /opt/cni/bin/ || echo "Directory does not exist"
  exit 1
fi

# Verify installation
echo "Verifying CNI installation:"
ls -la /opt/cni/bin/bridge /opt/cni/bin/loopback /opt/cni/bin/host-local

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

echo "Starting BuildKit daemon..."
sudo mkdir -p /run/buildkit
sudo buildkitd --addr unix:///run/buildkit/buildkitd.sock --group $(id -gn) &
BUILDKIT_PID=$!
echo "BuildKit daemon started with PID: $BUILDKIT_PID"

# Wait for BuildKit to be ready
echo "Waiting for BuildKit to be ready..."
for i in {1..30}; do
  if buildctl --addr unix:///run/buildkit/buildkitd.sock debug info >/dev/null 2>&1; then
    echo "BuildKit is ready after ${i} seconds"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "ERROR: BuildKit failed to start after 30 seconds"
    exit 1
  fi
  sleep 1
done

echo "Setup complete. Containerd PID: $CONTAINERD_PID, BuildKit PID: $BUILDKIT_PID"
