#!/bin/bash
set -e

echo "=== INVOKE TESTS - Started at $(date) ==="
touch /tmp/invoke_output.txt
chown ec2-user:staff /tmp/invoke_output.txt

su ec2-user -c 'finch system prune -f' || true
su ec2-user -c 'cd /Users/ec2-user/aws-sam-cli && export PATH="/Users/ec2-user/Library/Python/3.11/bin:$PATH" && AWS_DEFAULT_REGION="$AWS_DEFAULT_REGION" BY_CANARY=true SAM_CLI_DEV=1 SAM_CLI_TELEMETRY=0 python3.11 -m pytest tests/integration/local/invoke -k "not Terraform" -v --tb=short' > /tmp/invoke_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" /tmp/invoke_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" /tmp/invoke_output.txt || echo "No failures found"

# test_invoke_with_error_during_image_build: Build error message differs from expected.
# test_invoke_with_timeout_set_X_TimeoutFunction: Returns timeout message instead of empty string,
#         but matches actual Lambda service behavior.
# test_building_new_rapid_image_removes_old_rapid_images: Cannot remove images with same digest,
#         Docker creates different IDs for each.
cat > expected_invoke_failures.txt << 'EOF'
test_invoke_with_error_during_image_build
test_invoke_with_timeout_set_0_TimeoutFunction
test_invoke_with_timeout_set_1_TimeoutFunctionWithParameter
test_invoke_with_timeout_set_2_TimeoutFunctionWithStringParameter
test_building_new_rapid_image_removes_old_rapid_images
EOF

# Extract actual failures
grep "FAILED" /tmp/invoke_output.txt | grep -o "test_[^[:space:]]*" > actual_invoke_failures.txt || true

# Find unexpected failures
UNEXPECTED=$(grep -v -f expected_invoke_failures.txt actual_invoke_failures.txt 2>/dev/null || true)

if [ -n "$UNEXPECTED" ]; then
  echo "❌ Unexpected failures found:"
  echo "$UNEXPECTED"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat /tmp/invoke_output.txt
  exit 1
else
  echo "✅ All failures were expected"
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" /tmp/invoke_output.txt | tail -1 || echo "No pytest summary found"