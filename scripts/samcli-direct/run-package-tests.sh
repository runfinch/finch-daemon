#!/bin/bash
set -e

cd aws-sam-cli

python -m pytest tests/integration/package/test_package_command_image.py -v --tb=short > package_output.txt 2>&1 || true

# test_package_with_deep_nested_template_image: Expects Docker-specific push stream pattern.
# test_package_template_with_image_repositories_nested_stack_x: Push API stream differs from Docker.
# test_package_with_loadable_image_archive_0_template_image_load_yaml: Docker imports by digest,
#         Finch imports as "overlayfs:" tag causing image info lookup to fail.
cat > expected_package_failures.txt << 'EOF'
test_package_with_deep_nested_template_image
test_package_template_with_image_repositories_nested_stack
test_package_with_loadable_image_archive_0_template_image_load_yaml
EOF

# Validate test results
$GITHUB_WORKSPACE/scripts/validate-test-results.sh package_output.txt expected_package_failures.txt "Package tests"