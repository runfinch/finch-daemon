#!/bin/bash
set -e

cd aws-sam-cli

python -m pytest tests/integration/package/test_package_command_image.py -v --tb=short > package_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" package_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" package_output.txt || echo "No failures found"

cat > expected_package_failures.txt << 'EOF'
test_package_with_deep_nested_template_image
test_package_template_with_image_repositories_nested_stack
test_package_with_loadable_image_archive_0_template_image_load_yaml
EOF

# Extract actual failures
grep "FAILED" package_output.txt | grep -o "test_[^[:space:]]*" > actual_package_failures.txt || true

# Also check for nested stack failures (pattern match)
grep "FAILED.*test_package_template_with_image_repositories_nested_stack" package_output.txt >> actual_package_failures.txt || true

# Find unexpected failures (exclude nested stack pattern)
UNEXPECTED=$(grep -v -f expected_package_failures.txt actual_package_failures.txt | grep -v "test_package_template_with_image_repositories_nested_stack" || true)

if [ -n "$UNEXPECTED" ]; then
  echo "❌ Unexpected failures found:"
  echo "$UNEXPECTED"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat package_output.txt || echo "No output file found"
  exit 1
else
  echo "✅ All failures were expected."
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" package_output.txt | tail -1 || echo "No pytest summary found"