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

# Check if ~/.finch/finch.yaml exists and if rosetta is set to true
echo "Checking Rosetta configuration..."
FINCH_CONFIG_DIR="/Users/ec2-user/.finch"
FINCH_CONFIG_FILE="${FINCH_CONFIG_DIR}/finch.yaml"

# Create config directory if it doesn't exist
su ec2-user -c "mkdir -p ${FINCH_CONFIG_DIR}"

# Check if finch.yaml exists and contains rosetta: true and memory: 16GiB
if su ec2-user -c "[ -f ${FINCH_CONFIG_FILE} ]"; then
  # Check if rosetta is already set to true
  if ! su ec2-user -c "grep -q 'rosetta: true' ${FINCH_CONFIG_FILE}"; then
    echo "Setting rosetta: true in ${FINCH_CONFIG_FILE}"
    # If the file exists but doesn't have rosetta: true, add it
    su ec2-user -c "sed -i '' '/^rosetta:/d' ${FINCH_CONFIG_FILE}" # Remove any existing rosetta line
    su ec2-user -c "echo 'rosetta: true' >> ${FINCH_CONFIG_FILE}"
  else
    echo "Rosetta is already set to true in ${FINCH_CONFIG_FILE}"
  fi
  
  # Check if memory is already set to 16GiB
  if ! su ec2-user -c "grep -q 'memory: 16GiB' ${FINCH_CONFIG_FILE}"; then
    echo "Setting memory: 16GiB in ${FINCH_CONFIG_FILE}"
    # If the file exists but doesn't have memory: 16GiB, add it
    su ec2-user -c "sed -i '' '/^memory:/d' ${FINCH_CONFIG_FILE}" # Remove any existing memory line
    su ec2-user -c "echo 'memory: 16GiB' >> ${FINCH_CONFIG_FILE}"
  else
    echo "Memory is already set to 16GiB in ${FINCH_CONFIG_FILE}"
  fi
  
  # Check if cpu is already set to 6
  if ! su ec2-user -c "grep -q 'cpus: 6' ${FINCH_CONFIG_FILE}"; then
    echo "Setting cpus: 6 in ${FINCH_CONFIG_FILE}"
    # If the file exists but doesn't have cpu: 6, add it
    su ec2-user -c "sed -i '' '/^cpus:/d' ${FINCH_CONFIG_FILE}" # Remove any existing cpu line
    su ec2-user -c "echo 'cpus: 6' >> ${FINCH_CONFIG_FILE}"
    echo "CPU configuration changed to 6 cores"
  else
    echo "CPU is already set to 6 in ${FINCH_CONFIG_FILE}"
  fi
else
  # If the file doesn't exist, create it with rosetta: true, memory: 16GiB, and cpu: 6
  echo "Creating ${FINCH_CONFIG_FILE} with rosetta: true, memory: 16GiB, and cpu: 6"
  su ec2-user -c "echo 'rosetta: true' > ${FINCH_CONFIG_FILE}"
  su ec2-user -c "echo 'memory: 16GiB' >> ${FINCH_CONFIG_FILE}"
  su ec2-user -c "echo 'cpu: 6' >> ${FINCH_CONFIG_FILE}"
  echo "CPU configuration set to 6 cores"
fi

su ec2-user -c 'finch vm init'
sleep 10  # Wait for services to be ready

echo "Checking Finch version..."
su ec2-user -c 'LIMA_HOME=/Applications/Finch/lima/data /Applications/Finch/lima/bin/limactl shell finch curl --unix-socket /var/run/finch.sock -X GET http:/v1.43/version'

echo "Checking binfmt ..."
su ec2-user -c 'LIMA_HOME=/Applications/Finch/lima/data /Applications/Finch/lima/bin/limactl shell finch ls -lh /proc/sys/fs/binfmt_misc/'

echo "Verifying Docker daemon is accessible..."
su ec2-user -c 'finch info' || echo "Finch info failed"
su ec2-user -c 'finch version' || echo "Finch version failed"

echo "✅ Finch VM build and setup complete"
