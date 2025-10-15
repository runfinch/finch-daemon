#!/bin/bash
set -e

echo "=== Building and Setting up Finch VM ==="

# Configure workspace permissions
echo "Configuring workspace permissions..."
chown -R ec2-user:staff "$GITHUB_WORKSPACE"

# Build and Install Finch from upstream
echo "Building Finch from upstream..."
su ec2-user -c "cd $GITHUB_WORKSPACE && make clean && make FINCH_OS_IMAGE_LOCATION_ROOT=/Applications/Finch && make install PREFIX=Applications/Finch"

# Build finch-daemon from PR and inject into VM
echo "Building finch-daemon from PR..."
su ec2-user -c "cd $GITHUB_WORKSPACE/finch-daemon-pr && STATIC=1 GOPROXY=direct GOOS=linux GOARCH=\$(go env GOARCH) make"
su ec2-user -c 'finch vm remove -f'
su ec2-user -c "cp $GITHUB_WORKSPACE/finch-daemon-pr/bin/finch-daemon /Applications/Finch/finch-daemon/finch-daemon"

# Check Finch version and initialize VM
echo "Initializing VM and checking version..."
# Clean up any leftover network state
sudo pkill -f socket_vmnet || true
sudo rm -f /private/var/run/finch-lima/*.sock || true

su ec2-user -c 'finch vm init'
sleep 10  # Wait for services to be ready

echo "Checking Finch version..."
su ec2-user -c 'LIMA_HOME=/Applications/Finch/lima/data /Applications/Finch/lima/bin/limactl shell finch curl --unix-socket /var/run/finch.sock -X GET http:/v1.43/version'

echo "Verifying Docker daemon is accessible..."
su ec2-user -c 'finch info' || echo "Finch info failed"
su ec2-user -c 'finch version' || echo "Finch version failed"

echo "✅ Finch VM build and setup complete"