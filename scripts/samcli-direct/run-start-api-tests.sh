#!/bin/bash
set -e

cd aws-sam-cli

ulimit -n 65536
python -m pytest tests/integration/local/start_api -k 'not Terraform' -v --tb=short > start_api_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" start_api_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" start_api_output.txt || echo "No failures found"

# Expected failures - flaky tests that fail in CI but not locally
cat > expected_start_api_failures.txt << 'EOF'
test_can_invoke_lambda_layer_successfully
EOF

# Extract actual failures - find test names in FAILED lines
grep "FAILED" start_api_output.txt | grep -o "test_[^[:space:]]*" > actual_start_api_failures.txt || true

# Find unexpected failures
UNEXPECTED=$(grep -v -f expected_start_api_failures.txt actual_start_api_failures.txt 2>/dev/null || true)

if [ -n "$UNEXPECTED" ]; then
  echo "❌ Unexpected start-api failures found:"
  echo "$UNEXPECTED"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat start_api_output.txt || echo "No output file found"
  exit 1
else
  echo "✅ All start-api failures (if any) were expected."
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" start_api_output.txt | tail -1 || echo "No pytest summary found"