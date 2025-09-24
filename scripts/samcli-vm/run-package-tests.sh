#!/bin/bash
set -e

echo "=== PACKAGE TESTS - Started at $(date) ==="
touch /tmp/package_output.txt
chown ec2-user:staff /tmp/package_output.txt
su ec2-user -c "cd /Users/ec2-user/aws-sam-cli && export PATH='/Users/ec2-user/Library/Python/3.11/bin:$PATH' && export DOCKER_HOST='$DOCKER_HOST' && AWS_DEFAULT_REGION='$AWS_DEFAULT_REGION' BY_CANARY=true SAM_CLI_DEV=1 SAM_CLI_TELEMETRY=0 python3.11 -m pytest tests/integration/package/test_package_command_image.py -v" > /tmp/package_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" /tmp/package_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" /tmp/package_output.txt || echo "No failures found"

# test_package_with_deep_nested_template_image: Expects Docker-specific push stream pattern.
# test_package_template_with_image_repositories_nested_stack_x: Push API stream differs from Docker.
# test_package_with_loadable_image_archive_0_template_image_load_yaml: Docker imports by digest,
#         Finch imports as "overlayfs:" tag causing image info lookup to fail.
cat > expected_package_failures.txt << 'EOF'
test_package_with_deep_nested_template_image
test_package_template_with_image_repositories_nested_stack
test_package_with_loadable_image_archive_0_template_image_load_yaml
EOF

# Extract actual failures
grep "FAILED" /tmp/package_output.txt | grep -o "test_[^[:space:]]*" > actual_package_failures.txt || true

# Also check for nested stack failures (pattern match)
grep "FAILED.*test_package_template_with_image_repositories_nested_stack" /tmp/package_output.txt >> actual_package_failures.txt || true

# Find unexpected failures (exclude nested stack pattern)
UNEXPECTED=$(grep -v -f expected_package_failures.txt actual_package_failures.txt | grep -v "test_package_template_with_image_repositories_nested_stack" || true)

if [ -n "$UNEXPECTED" ]; then
  echo "❌ Unexpected failures found:"
  echo "$UNEXPECTED"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat /tmp/package_output.txt
  exit 1
else
  echo "✅ All failures were expected"
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" /tmp/package_output.txt | tail -1 || echo "No pytest summary found"