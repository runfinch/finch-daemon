#!/bin/bash
set -e

echo "=== Setting up Finch and Docker ==="

# Install Finch
echo "Installing Finch as ec2-user..."
su ec2-user -c 'source /Users/ec2-user/.brewrc && brew install finch --cask'
su ec2-user -c 'source /Users/ec2-user/.brewrc && brew list | grep finch || echo "finch not installed"'
mkdir -p /private/var/run/finch-lima
cat /etc/passwd
chown ec2-user:daemon /private/var/run/finch-lima

# Build binaries
echo "Building cross architecture binaries..."
su ec2-user -c "cd $GITHUB_WORKSPACE && STATIC=1 GOPROXY=direct GOOS=linux GOARCH=arm64 make"
su ec2-user -c 'finch vm remove -f' || true
cp -f $GITHUB_WORKSPACE/bin/finch-daemon /Applications/Finch/finch-daemon/finch-daemon

# Restart finch-daemon with new binary
su ec2-user -c 'finch vm stop' || true
su ec2-user -c 'finch vm start' || true

# Check Finch version
echo "Initializing VM and checking version..."
sudo pkill -f socket_vmnet || true
sudo rm -f /private/var/run/finch-lima/*.sock || true
su ec2-user -c 'finch vm init'
sleep 5  # Wait for services to be ready
echo "Checking Finch version..."
su ec2-user -c 'LIMA_HOME=/Applications/Finch/lima/data /Applications/Finch/lima/bin/limactl shell finch curl --unix-socket /var/run/finch.sock -X GET http:/v1.43/version'

# Install Docker CLI for SAM CLI compatibility
echo "Checking Docker CLI installation..."
if ! su ec2-user -c 'which docker' > /dev/null 2>&1; then
  echo "Installing Docker CLI..."
  su ec2-user -c 'source /Users/ec2-user/.brewrc && brew install --formula docker'
else
  echo "Docker CLI already installed"
fi

echo "✅ Finch setup completed"