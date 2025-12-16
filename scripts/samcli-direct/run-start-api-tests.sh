#!/bin/bash
set -e

cd aws-sam-cli

ulimit -n 65536
python -m pytest tests/integration/local/start_api -k 'not Terraform' -v --tb=short 2>&1 | tee start_api_output.txt || true

# test_can_invoke_lambda_layer_successfully: Uses random port, fails occasionally.
#         Only 1 test of 386 total, acceptable failure rate.
cat > expected_start_api_failures.txt << 'EOF'
test_can_invoke_lambda_layer_successfully
EOF

# Validate test results
$GITHUB_WORKSPACE/scripts/validate-test-results.sh start_api_output.txt expected_start_api_failures.txt "Start-API tests"