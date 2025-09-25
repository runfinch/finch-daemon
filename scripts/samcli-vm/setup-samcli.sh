#!/bin/bash
set -e

echo "=== Setting up SAM CLI ==="

# Move to ec2-user home and change ownership
sudo rm -rf /Users/ec2-user/aws-sam-cli || true
sudo mv aws-sam-cli /Users/ec2-user/aws-sam-cli
sudo chown -R ec2-user:staff /Users/ec2-user/aws-sam-cli

# Install and setup (use full path)
su ec2-user -c "cd /Users/ec2-user/aws-sam-cli && $PYTHON_BINARY -m pip install --upgrade pip --user"
su ec2-user -c "cd /Users/ec2-user/aws-sam-cli && SAM_CLI_DEV=1 $PYTHON_BINARY -m pip install -e \".[dev]\" --user"
su ec2-user -c "cd /Users/ec2-user/aws-sam-cli && export PATH=\"/Users/ec2-user/Library/Python/$PYTHON_VERSION/bin:\$PATH\" && samdev --version"

echo "✅ SAM CLI setup completed"