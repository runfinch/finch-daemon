#!/bin/bash
set -e

cd aws-sam-cli

python -m pytest tests/integration/local/start_lambda -k 'not Terraform' -v --tb=short > start_lambda_output.txt 2>&1 || true

# Create empty expected failures file (should pass completely)
touch expected_start_lambda_failures.txt

# Validate test results
$GITHUB_WORKSPACE/scripts/validate-test-results.sh start_lambda_output.txt expected_start_lambda_failures.txt "Start-Lambda tests"