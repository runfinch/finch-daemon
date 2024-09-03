#!/bin/bash
# Set versions
CONTAINERD_VERSION=1.6.34
RUNC_VERSION=1.1.12
NERDCTL_VERSION=1.7.1
BUILDKIT_VERSION=0.15.2
CNI_VERSION=1.4.1

apt update && apt install -y make gcc linux-libc-dev libseccomp-dev pkg-config git
apk add --no-cache \
    btrfs-progs-libs \
    curl \
    fuse \
    gcc \
    libc6-compat \
    libseccomp-dev \
    pigz \
    zlib-dev

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

sudo containerd &
sudo buildkitd &

sleep 2
