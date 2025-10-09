#!/bin/bash
set -e

echo "=== INVOKE TESTS - Started at $(date) ==="
touch /tmp/invoke_test_output.txt
chown ec2-user:staff /tmp/invoke_test_output.txt

echo "=== System Information ===" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "uname -a" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "arch" | tee -a /tmp/invoke_test_output.txt
echo "=== Finch VM Status ===" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "finch vm status" | tee -a /tmp/invoke_test_output.txt

echo "=== Checking Rosetta Configuration ===" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "cat /Users/ec2-user/.finch/finch.yaml" | tee -a /tmp/invoke_test_output.txt

echo "=== Checking Docker Info Before Test ===" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "finch system prune -f" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "finch info" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "finch image ls" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "finch container ls -a" | tee -a /tmp/invoke_test_output.txt

echo "=== Running Specific Failing Test with Extra Verbosity ===" | tee -a /tmp/invoke_test_output.txt
su ec2-user -c "
  cd /Users/ec2-user/aws-sam-cli && \
  export PATH='/Users/ec2-user/Library/Python/$PYTHON_VERSION/bin:$PATH' && \
  export DOCKER_HOST='$DOCKER_HOST' && \
  export SAM_CLI_CONTAINER_CONNECTION_TIMEOUT=40 && \
  AWS_DEFAULT_REGION='$AWS_DEFAULT_REGION' \
  BY_CANARY='$BY_CANARY' \
  SAM_CLI_DEV='$SAM_CLI_DEV' \
  SAM_CLI_TELEMETRY='$SAM_CLI_TELEMETRY' \
  '$PYTHON_BINARY' -m pytest tests/integration/local/invoke/runtimes/test_with_runtime_zips.py::TestWithDifferentLambdaRuntimeZips::test_custom_provided_runtime -vvs --no-header --showlocals
" 2>&1 | tee /tmp/invoke_test_output.txt || true

su ec2-user -c 'LIMA_HOME=/Applications/Finch/lima/data /Applications/Finch/lima/bin/limactl shell finch sudo journalctl -xeu finch@$UID'
su ec2-user -c 'LIMA_HOME=/Applications/Finch/lima/data /Applications/Finch/lima/bin/limactl shell finch sudo journalctl -xeu buildkit'

# test_invoke_with_error_during_image_build: Build error message differs from expected.
# test_invoke_with_timeout_set_X_TimeoutFunction: Returns timeout message instead of empty string,
#         but matches actual Lambda service behavior.
# test_building_new_rapid_image_removes_old_rapid_images: Cannot remove images with same digest,
#         Docker creates different IDs for each.
# test_caching_two_layers and test_caching_two_layers_with_layer_cache_env_set: error due to sequential
#         test runs within invoke. Work when run in isolation and locally.
# test_successful_invoke: Related to symlink mount errors due to permissions. Works locally.
cat > expected_invoke_failures.txt << 'EOF'
test_invoke_with_error_during_image_build
test_invoke_with_timeout_set_0_TimeoutFunction
test_invoke_with_timeout_set_1_TimeoutFunctionWithParameter
test_invoke_with_timeout_set_2_TimeoutFunctionWithStringParameter
test_building_new_rapid_image_removes_old_rapid_images
test_caching_two_layers
test_caching_two_layers_with_layer_cache_env_set
test_successful_invoke
EOF

# Validate test results
$GITHUB_WORKSPACE/finch-daemon-pr/scripts/validate-test-results.sh /tmp/invoke_test_output.txt expected_invoke_failures.txt "Invoke tests"
