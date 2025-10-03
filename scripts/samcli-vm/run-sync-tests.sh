#!/bin/bash
set -e

echo "=== SYNC TESTS - Started at $(date) ==="
touch /tmp/sync_test_output.txt
chown ec2-user:staff /tmp/sync_test_output.txt
su ec2-user -c "
  cd /Users/ec2-user/aws-sam-cli && \
  export PATH='/Users/ec2-user/Library/Python/$PYTHON_VERSION/bin:$PATH' && \
  export DOCKER_HOST='$DOCKER_HOST' && \
  AWS_DEFAULT_REGION='$AWS_DEFAULT_REGION' \
  BY_CANARY='$BY_CANARY' \
  SAM_CLI_DEV='$SAM_CLI_DEV' \
  SAM_CLI_TELEMETRY='$SAM_CLI_TELEMETRY' \
  '$PYTHON_BINARY' -m pytest tests/integration/sync -k 'image' -v --tb=short
" > /tmp/sync_test_output.txt 2>&1 || true

# Create empty expected failures file (should pass completely)
touch expected_sync_failures.txt

# Validate test results
$(dirname "$0")/../validate-test-results.sh /tmp/sync_test_output.txt expected_sync_failures.txt "Sync tests"