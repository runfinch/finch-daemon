#!/bin/bash
set -e

cd aws-sam-cli

python -m pytest tests/integration/deploy -k 'image' -v --tb=short > deploy_output.txt 2>&1 || true

echo ""
echo "=== FAILURES ==="
grep "FAILED" deploy_output.txt || echo "No failures found"

echo ""
echo "=== PASSES ==="
grep "PASSED" deploy_output.txt || echo "No passes found"

# Expected passes - this test passes despite having an error in the output
cat > expected_deploy_passes.txt << 'EOF'
test_deploy_guided_image_auto_0_aws_serverless_function_image_yaml
EOF

# Extract actual passes - test names appear on the line after PASSED
grep -A1 "PASSED" deploy_output.txt | grep -o "test_[^[:space:]]*" > actual_deploy_passes.txt || true

# Find unexpected passes (passes that aren't in our expected list)
UNEXPECTED_PASSES=$(grep -v -f expected_deploy_passes.txt actual_deploy_passes.txt 2>/dev/null || true)

if [ -n "$UNEXPECTED_PASSES" ]; then
  echo "❌ Unexpected passes found:"
  echo "$UNEXPECTED_PASSES"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat deploy_output.txt || echo "No output file found"
  exit 1
else
  echo "✅ All failures and passes were expected (1 known pass with error, rest fail due to multi-arch)."
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" deploy_output.txt | tail -1 || echo "No pytest summary found"