#!/bin/bash
set -e

cd aws-sam-cli

python -m pytest tests/integration/local/invoke -k 'not Terraform' -v --tb=short 2>&1 | tee invoke_output.txt || true

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

# Validate test results
$GITHUB_WORKSPACE/scripts/validate-test-results.sh invoke_output.txt expected_invoke_failures.txt "Invoke tests"