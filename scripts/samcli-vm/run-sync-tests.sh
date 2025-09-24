#!/bin/bash
set -e

echo "=== SYNC TESTS - Started at $(date) ==="
touch /tmp/sync_output.txt
su ec2-user -c "
  cd /Users/ec2-user/aws-sam-cli && \
  export PATH='/Users/ec2-user/Library/Python/$PYTHON_VERSION/bin:$PATH' && \
  export DOCKER_HOST='$DOCKER_HOST' && \
  AWS_DEFAULT_REGION='$AWS_DEFAULT_REGION' \
  BY_CANARY='$BY_CANARY' \
  SAM_CLI_DEV='$SAM_CLI_DEV' \
  SAM_CLI_TELEMETRY='$SAM_CLI_TELEMETRY' \
  '$PYTHON_BINARY' -m pytest tests/integration/sync -k 'image' -v --tb=short
" > /tmp/sync_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" /tmp/sync_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" /tmp/sync_output.txt || echo "No failures found"

# Should pass completely per test guide
if grep -q "FAILED" /tmp/sync_output.txt; then
  echo "❌ Sync tests failed (should pass completely)"
  grep "FAILED" /tmp/sync_output.txt
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat /tmp/sync_output.txt
  exit 1
else
  echo "✅ All sync tests passed as expected"
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" /tmp/sync_output.txt | tail -1 || echo "No pytest summary found"