#!/bin/bash
set -e

echo "=== DEPLOY TESTS - Started at $(date) ==="
touch /tmp/deploy_output.txt
chown ec2-user:staff /tmp/deploy_output.txt
su ec2-user -c 'cd /Users/ec2-user/aws-sam-cli && export PATH="/Users/ec2-user/Library/Python/3.11/bin:$PATH" && AWS_DEFAULT_REGION="$AWS_DEFAULT_REGION" BY_CANARY=true SAM_CLI_DEV=1 SAM_CLI_TELEMETRY=0 python3.11 -m pytest tests/integration/deploy -k "image" -v --tb=short' > /tmp/deploy_output.txt 2>&1 || true

echo ""
echo "=== FAILURES ==="
grep "FAILED" /tmp/deploy_output.txt || echo "No failures found"

echo ""
echo "=== PASSES ==="
grep "PASSED" /tmp/deploy_output.txt || echo "No passes found"

# Expected passes - this test passes despite having an error in the output
cat > expected_deploy_passes.txt << 'EOF'
test_deploy_guided_image_auto_0_aws_serverless_function_image_yaml
EOF

# Extract actual passes - test names appear on the line after PASSED
grep -A1 "PASSED" /tmp/deploy_output.txt | grep -o "test_[^[:space:]]*" > actual_deploy_passes.txt || true

# Find unexpected passes (passes that aren't in our expected list)
UNEXPECTED_PASSES=$(grep -v -f expected_deploy_passes.txt actual_deploy_passes.txt 2>/dev/null || true)

if [ -n "$UNEXPECTED_PASSES" ]; then
  echo "❌ Unexpected passes found:"
  echo "$UNEXPECTED_PASSES"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat /tmp/deploy_output.txt
  exit 1
else
  echo "✅ All failures and passes were expected (1 known pass with error, rest fail due to multi-arch)."
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" /tmp/deploy_output.txt | tail -1 || echo "No pytest summary found"