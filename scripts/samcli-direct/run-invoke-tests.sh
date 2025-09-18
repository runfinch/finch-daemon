#!/bin/bash
set -e

cd aws-sam-cli

python -m pytest tests/integration/local/invoke -k 'not Terraform' -v --tb=short > invoke_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" invoke_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" invoke_output.txt || echo "No failures found"

cat > expected_invoke_failures.txt << 'EOF'
test_invoke_with_error_during_image_build
test_invoke_with_timeout_set_0_TimeoutFunction
test_invoke_with_timeout_set_1_TimeoutFunctionWithParameter
test_invoke_with_timeout_set_2_TimeoutFunctionWithStringParameter
test_building_new_rapid_image_removes_old_rapid_images
test_invoke_returns_expected_results_from_git_function
test_invoke_returns_expected_results_from_git_function_with_parameters
EOF

# Extract actual failures
grep "FAILED" invoke_output.txt | grep -o "test_[^[:space:]]*" > actual_invoke_failures.txt || true

# Find unexpected failures
UNEXPECTED=$(grep -v -f expected_invoke_failures.txt actual_invoke_failures.txt 2>/dev/null || true)

if [ -n "$UNEXPECTED" ]; then
  echo "❌ Unexpected failures found:"
  echo "$UNEXPECTED"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat invoke_output.txt || echo "No output file found"
  exit 1
else
  echo "✅ All failures were expected."
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" invoke_output.txt | tail -1 || echo "No pytest summary found"