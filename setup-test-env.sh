#!/bin/bash

# Set versions
RUNC_VERSION=1.3.0
NERDCTL_VERSION=2.1.2
BUILDKIT_VERSION=0.23.2
CNI_VERSION=1.6.2
DEFAULT_RUNC_VERSION="1.3.0"
DEFAULT_CONTAINERD_VERSION="1.7.27"
DEFAULT_NERDCTL_VERSION="2.1.5"
DEFAULT_BUILDKIT_VERSION="0.23.2"
DEFAULT_CNI_VERSION="1.6.2"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --runc-version)
      RUNC_VERSION="$2"
      shift 2
      ;;
    --containerd-version)
      CONTAINERD_VERSION="$2"
      shift 2
      ;;
    --nerdctl-version)
      NERDCTL_VERSION="$2"
      shift 2
      ;;
    --buildkit-version)
      BUILDKIT_VERSION="$2"
      shift 2
      ;;
    --cni-version)
      CNI_VERSION="$2"
      shift 2
      ;;
    --help)
      echo "Usage: $0 [OPTIONS]"
      echo "Options:"
      echo "  --runc-version VERSION       Set runc version (default: $DEFAULT_RUNC_VERSION)"
      echo "  --containerd-version VERSION Set containerd version (default: $DEFAULT_CONTAINERD_VERSION)"
      echo "  --nerdctl-version VERSION    Set nerdctl version (default: $DEFAULT_NERDCTL_VERSION)"
      echo "  --buildkit-version VERSION   Set buildkit version (default: $DEFAULT_BUILDKIT_VERSION)"
      echo "  --cni-version VERSION        Set CNI plugins version (default: $DEFAULT_CNI_VERSION)"
      echo "  --help                     Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Set default versions if not provided
CONTAINERD_VERSION=${CONTAINERD_VERSION:-$DEFAULT_CONTAINERD_VERSION}
RUNC_VERSION=${RUNC_VERSION:-$DEFAULT_RUNC_VERSION}
NERDCTL_VERSION=${NERDCTL_VERSION:-$DEFAULT_NERDCTL_VERSION}
BUILDKIT_VERSION=${BUILDKIT_VERSION:-$DEFAULT_BUILDKIT_VERSION}
CNI_VERSION=${CNI_VERSION:-$DEFAULT_CNI_VERSION}

echo "Using dependency versions:"
echo "  containerd: $CONTAINERD_VERSION"
echo "  runc: $RUNC_VERSION"
echo "  nerdctl: $NERDCTL_VERSION"
echo "  buildkit: $BUILDKIT_VERSION"
echo "  cni: $CNI_VERSION"

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

# Create BuildKit config directory and file to ensure finch namespace
sudo mkdir -p /etc/buildkit
sudo tee /etc/buildkit/buildkitd.toml > /dev/null << 'EOF'
root = "/var/lib/buildkit"

[worker.oci]
  enabled = false

[worker.containerd]
  enabled = true
  namespace = "finch"
EOF

sudo containerd &
sudo buildkitd --config /etc/buildkit/buildkitd.toml &

sleep 2