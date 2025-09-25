#!/bin/bash
set -e

echo "=== Setting up runner and dependencies for EC2 user ==="

# Configure Git for ec2-user
git config --global --add safe.directory "*"

# Configure Go for ec2-user
chown -R ec2-user:staff $GOPATH || true
chown -R ec2-user:staff $RUNNER_TOOL_CACHE/go || true

# Install Rosetta 2
su ec2-user -c 'echo "A" | /usr/sbin/softwareupdate --install-rosetta --agree-to-license || true'

# Configure Python for ec2-user
chown -R ec2-user:staff $($PYTHON_BINARY -c "import site; print(site.USER_BASE)") || true
ln -sf $(which $PYTHON_BINARY) /usr/local/bin/$PYTHON_BINARY || true

# Configure Homebrew for ec2-user
echo "Creating .brewrc file for ec2-user..."
cat > /Users/ec2-user/.brewrc << 'EOF'
# Homebrew environment setup
export PATH="/opt/homebrew/bin:/opt/homebrew/sbin:$PATH"
export HOMEBREW_PREFIX="/opt/homebrew"
export HOMEBREW_CELLAR="/opt/homebrew/Cellar"
export HOMEBREW_REPOSITORY="/opt/homebrew"
export HOMEBREW_NO_AUTO_UPDATE=1
EOF
chown ec2-user:staff /Users/ec2-user/.brewrc

# Fix Homebrew permissions
echo "Setting permissions for Homebrew directories..."
mkdir -p /opt/homebrew/Cellar
chown -R ec2-user:staff /opt/homebrew

# Install dependencies
echo "Installing dependencies as ec2-user..."
su ec2-user -c 'source /Users/ec2-user/.brewrc && brew install lz4 automake autoconf libtool yq'

echo "✅ Runner setup completed"