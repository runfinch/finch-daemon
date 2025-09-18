#!/bin/bash
set -e

cd aws-sam-cli

python -m pytest tests/integration/local/start_lambda -k 'not Terraform' -v --tb=short > start_lambda_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" start_lambda_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" start_lambda_output.txt || echo "No failures found"

# Should pass completely per test guide
if grep -q "FAILED" start_lambda_output.txt; then
  echo "❌ Start-lambda tests failed (should pass completely)"
  grep "FAILED" start_lambda_output.txt
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat start_lambda_output.txt || echo "No output file found"
  exit 1
else
  echo "✅ All start-lambda tests passed as expected"
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" start_lambda_output.txt | tail -1 || echo "No pytest summary found"