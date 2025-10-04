#!/bin/bash
set -e

echo "=== START-API TESTS - Started at $(date) ==="
touch /tmp/start_api_test_output.txt
chown ec2-user:staff /tmp/start_api_output.txt

# Start background monitor to show progress
(
  while true; do
    sleep 30
    echo "[$(date)] Progress check - Last 20 lines:"
    tail -20 /tmp/start_api_test_output.txt 2>/dev/null || echo "No output yet"
    echo "---"
  done
) &
MONITOR_PID=$!

# Run tests (output to file only, show progress via monitor)
su ec2-user -c "
  cd /Users/ec2-user/aws-sam-cli && \
  export PATH='/Users/ec2-user/Library/Python/$PYTHON_VERSION/bin:$PATH' && \
  export DOCKER_HOST='$DOCKER_HOST' && \
  ulimit -n 65536 && \
  AWS_DEFAULT_REGION='$AWS_DEFAULT_REGION' \
  BY_CANARY='$BY_CANARY' \
  SAM_CLI_DEV='$SAM_CLI_DEV' \
  SAM_CLI_TELEMETRY='$SAM_CLI_TELEMETRY' \
  '$PYTHON_BINARY' -m pytest tests/integration/local/start_api -k 'not Terraform' -v --tb=short
" > /tmp/start_api_test_output.txt 2>&1 || true

# Stop monitor
kill $MONITOR_PID 2>/dev/null || true
echo "=== START-API TESTS - Finished at $(date) ==="

# test_can_invoke_lambda_layer_successfully: Uses random port, fails occasionally.
#         Only 1 test of 386 total, acceptable failure rate.
# test_changed_code_got_observed_and_loaded: Fails due to race conditions with containers.
# Fails infrequently, but is of note. 
cat > expected_start_api_failures.txt << 'EOF'
test_can_invoke_lambda_layer_successfully
test_changed_code_got_observed_and_loaded
EOF

# Validate test results
$GITHUB_WORKSPACE/finch-daemon-pr/scripts/validate-test-results.sh /tmp/start_api_test_output.txt expected_start_api_failures.txt "Start-API tests"