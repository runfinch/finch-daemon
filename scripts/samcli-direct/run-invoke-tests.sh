#!/bin/bash
set -e

cd aws-sam-cli

# Wrap samdev to capture subprocess debug output. The invoke tests run
# `sam local invoke` as a subprocess via Popen(stdout=PIPE), so its debug
# output is normally invisible. This wrapper tees stderr to a log file.
REAL_SAMDEV=$(which samdev)
mkdir -p /tmp/sam-debug-logs
mv "$REAL_SAMDEV" "${REAL_SAMDEV}.real"
cat > "$REAL_SAMDEV" << WRAPPER
#!/bin/bash
exec ${REAL_SAMDEV}.real --debug "\$@" 2>> /tmp/sam-debug-logs/invoke-subprocess.log
WRAPPER
chmod +x "$REAL_SAMDEV"

SAM_DEBUG=1 python -m pytest tests/integration/local/invoke -k 'timeout_set' -v --tb=short --log-cli-level=DEBUG 2>&1 | tee invoke_output.txt || true

# Restore original samdev
mv "${REAL_SAMDEV}.real" "$REAL_SAMDEV"

echo "=== SAM SUBPROCESS DEBUG LOGS (last 200 lines) ==="
tail -200 /tmp/sam-debug-logs/invoke-subprocess.log 2>/dev/null || echo "No subprocess logs captured"

# test_invoke_with_error_during_image_build: Build error message differs from expected.
# test_invoke_with_timeout_set_X_TimeoutFunction: Returns timeout message instead of empty string,
#         but matches actual Lambda service behavior.
# test_building_new_rapid_image_removes_old_rapid_images: Cannot remove images with same digest,
#         Docker creates different IDs for each.
# test_invoke_returns_expected_results_from_git_function: Layer download progress leaks into
#         stdout. SAM CLI test issue, not finch-daemon.
cat > expected_invoke_failures.txt << 'EOF'
test_invoke_with_error_during_image_build
test_invoke_with_timeout_set_0_TimeoutFunction
test_invoke_with_timeout_set_1_TimeoutFunctionWithParameter
test_invoke_with_timeout_set_2_TimeoutFunctionWithStringParameter
test_building_new_rapid_image_removes_old_rapid_images
test_invoke_returns_expected_results_from_git_function
EOF

# Validate test results
$GITHUB_WORKSPACE/scripts/validate-test-results.sh invoke_output.txt expected_invoke_failures.txt "Invoke tests"