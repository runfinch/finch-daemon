#!/bin/bash
set -e

echo "=== START-LAMBDA TESTS - Started at $(date) ==="
touch /tmp/start_lambda_output.txt
chown ec2-user:staff /tmp/start_lambda_output.txt
su ec2-user -c 'cd /Users/ec2-user/aws-sam-cli && export PATH="/Users/ec2-user/Library/Python/3.11/bin:$PATH" && AWS_DEFAULT_REGION="$AWS_DEFAULT_REGION" BY_CANARY=true SAM_CLI_DEV=1 SAM_CLI_TELEMETRY=0 python3.11 -m pytest tests/integration/local/start_lambda -k "not Terraform" -v --tb=short' 2>&1 | tee /tmp/start_lambda_output.txt || true

echo ""
echo "=== PASSES ==="
grep "PASSED" /tmp/start_lambda_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" /tmp/start_lambda_output.txt || echo "No failures found"

# Should pass completely per test guide
if grep -q "FAILED" /tmp/start_lambda_output.txt; then
  echo "❌ Start-lambda tests failed (should pass completely)"
  grep "FAILED" /tmp/start_lambda_output.txt
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat /tmp/start_lambda_output.txt
  exit 1
else
  echo "✅ All start-lambda tests passed as expected"
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" /tmp/start_lambda_output.txt | tail -1 || echo "No pytest summary found"