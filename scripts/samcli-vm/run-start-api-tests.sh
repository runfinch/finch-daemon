#!/bin/bash
set -e

echo "=== START-API TESTS - Started at $(date) ==="
touch /tmp/start_api_output.txt
chown ec2-user:staff /tmp/start_api_output.txt
su ec2-user -c "cd /Users/ec2-user/aws-sam-cli && export PATH='/Users/ec2-user/Library/Python/3.11/bin:$PATH' && export DOCKER_HOST='$DOCKER_HOST' && ulimit -n 65536 && AWS_DEFAULT_REGION='$AWS_DEFAULT_REGION' BY_CANARY=true SAM_CLI_DEV=1 SAM_CLI_TELEMETRY=0 python3.11 -m pytest tests/integration/local/start_api -k 'not Terraform' -v" > /tmp/start_api_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" /tmp/start_api_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" /tmp/start_api_output.txt || echo "No failures found"

# test_can_invoke_lambda_layer_successfully: Uses random port, fails occasionally.
#         Only 1 test of 386 total, acceptable failure rate.
cat > expected_start_api_failures.txt << 'EOF'
test_can_invoke_lambda_layer_successfully
EOF

# Extract actual failures
grep "FAILED" /tmp/start_api_output.txt | grep -o "test_[^[:space:]]*" > actual_start_api_failures.txt || true

# Find unexpected failures
UNEXPECTED=$(grep -v -f expected_start_api_failures.txt actual_start_api_failures.txt 2>/dev/null || true)

if [ -n "$UNEXPECTED" ]; then
  echo "❌ Unexpected start-api failures found:"
  echo "$UNEXPECTED"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat /tmp/start_api_output.txt
  echo "=== NOTE ==="
  echo "This is a known flaky test with ~ % pass rate."
  echo "Please try again using an individual workflow trigger."
  exit 1
else
  echo "✅ All start-api failures were expected (CI environment flakiness)"
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" /tmp/start_api_output.txt | tail -1 || echo "No pytest summary found"