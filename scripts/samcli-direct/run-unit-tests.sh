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

# Check for pytest summary
if ! grep -E "=+ .*(failed|passed|skipped|deselected|error).* =+$" unit_test_output.txt | tail -1; then
  echo "❌ No pytest summary found - tests may not have run"
  exit 1
fi

# Check for positive number of passes
PASS_COUNT=$(grep -c "PASSED" unit_test_output.txt || echo "0")
if [ "$PASS_COUNT" -eq 0 ]; then
  echo "❌ No tests passed - got $PASS_COUNT passes"
  exit 1
fi

# Check for errors in pytest summary
SUMMARY_LINE=$(grep -E "=+ .*(failed|passed|skipped|deselected|error).* =+$" unit_test_output.txt | tail -1)
if echo "$SUMMARY_LINE" | grep -q "error"; then
  echo "❌ Test errors found in summary: $SUMMARY_LINE"
  exit 1
fi

# Check coverage requirement
if grep -q "Required test coverage of.*reached" unit_test_output.txt; then
  echo "✅ Unit tests completed: $PASS_COUNT passes, required coverage reached"
  grep "Required test coverage of.*reached" unit_test_output.txt
else
  echo "❌ Required test coverage not reached"
  exit 1
fi