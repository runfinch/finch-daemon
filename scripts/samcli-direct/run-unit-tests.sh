#!/bin/bash
set -e

cd aws-sam-cli

ulimit -n 65536
make test > unit_test_output.txt 2>&1 || true

echo ""
echo "=== PASSES ==="
grep "PASSED" unit_test_output.txt || echo "No passes found"

echo ""
echo "=== FAILURES ==="
grep "FAILED" unit_test_output.txt || echo "No failures found"

if grep -q "Required test coverage of.*reached" unit_test_output.txt; then
  echo "✅ Unit tests completed with required coverage"
  grep "Required test coverage of.*reached" unit_test_output.txt
else
  echo "❌ Required test coverage not reached"
  echo ""
  echo "=== FULL OUTPUT FOR DEBUGGING ==="
  cat unit_test_output.txt || echo "No output file found"
  exit 1
fi

echo ""
echo "=== PYTEST SUMMARY ==="
grep -E "=+ .*(failed|passed|skipped|deselected).* =+$" unit_test_output.txt | tail -1 || echo "No pytest summary found"