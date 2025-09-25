#!/bin/bash
set -e

echo "=== Test Results Summary ==="
echo "Unit tests: ${UNIT_EXIT_CODE:-1}"
echo "Sync tests: ${SYNC_EXIT_CODE:-1}"
echo "Package tests: ${PACKAGE_EXIT_CODE:-1}"
echo "Start-API tests: ${START_API_EXIT_CODE:-1}"
echo "Start-Lambda tests: ${START_LAMBDA_EXIT_CODE:-1}"
echo "Patch: ${PATCH_EXIT_CODE:-1}"
echo "Invoke tests: ${INVOKE_EXIT_CODE:-1}"

# Check if any tests failed
if [ "${UNIT_EXIT_CODE:-1}" -ne 0 ] || [ "${SYNC_EXIT_CODE:-1}" -ne 0 ] || [ "${PACKAGE_EXIT_CODE:-1}" -ne 0 ] || [ "${START_API_EXIT_CODE:-1}" -ne 0 ] || [ "${START_LAMBDA_EXIT_CODE:-1}" -ne 0 ] || [ "${PATCH_EXIT_CODE:-1}" -ne 0 ] || [ "${INVOKE_EXIT_CODE:-1}" -ne 0 ]; then
  echo "❌ One or more tests failed"
  exit 1
else
  echo "✅ All tests passed"
fi