#!/bin/bash
set -e

echo "=== UNIT TESTS - Started at $(date) ==="
touch /tmp/unit_test_output.txt
chown ec2-user:staff /tmp/unit_test_output.txt
su ec2-user -c "
  cd /Users/ec2-user/aws-sam-cli && \
  export PATH='/Users/ec2-user/Library/Python/$PYTHON_VERSION/bin:$PATH' && \
  ulimit -n 65536 && \
  AWS_DEFAULT_REGION='$AWS_DEFAULT_REGION' \
  BY_CANARY='$BY_CANARY' \
  SAM_CLI_DEV='$SAM_CLI_DEV' \
  SAM_CLI_TELEMETRY='$SAM_CLI_TELEMETRY' \
  make test
" > /tmp/unit_test_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" /tmp/unit_test_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" /tmp/unit_test_output.txt || echo "No failures found"

echo ""
if grep -q "Required test coverage of.*reached" /tmp/unit_test_output.txt; then
  echo "✅ Unit tests completed with required coverage"
  grep "Required test coverage of.*reached" /tmp/unit_test_output.txt
else
  echo "❌ Required test coverage not reached"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat /tmp/unit_test_output.txt
  exit 1
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" /tmp/unit_test_output.txt | tail -1 || echo "No pytest summary found"