#!/bin/bash
#
# Coverage report generation script for restic
#

set -e

OUTPUT_DIR="coverage_reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "Running tests with coverage..."
go test -coverprofile="$OUTPUT_DIR/coverage_${TIMESTAMP}.out" ./... 2>&1 | tee "$OUTPUT_DIR/test_output_${TIMESTAMP}.txt"

if [ -f "$OUTPUT_DIR/coverage_${TIMESTAMP}.out" ]; then
    echo ""
    echo "Generating coverage reports..."

    # Generate HTML report
    go tool cover -html="$OUTPUT_DIR/coverage_${TIMESTAMP}.out" -o "$OUTPUT_DIR/coverage_${TIMESTAMP}.html"
    echo "HTML report: $OUTPUT_DIR/coverage_${TIMESTAMP}.html"

    # Generate function-level coverage
    go tool cover -func="$OUTPUT_DIR/coverage_${TIMESTAMP}.out" > "$OUTPUT_DIR/coverage_${TIMESTAMP}_func.txt"
    echo "Function coverage: $OUTPUT_DIR/coverage_${TIMESTAMP}_func.txt"

    # Calculate total coverage
    TOTAL_COVERAGE=$(go tool cover -func="$OUTPUT_DIR/coverage_${TIMESTAMP}.out" | grep total | awk '{print $3}')
    echo ""
    echo "Total coverage: $TOTAL_COVERAGE"

    # Extract per-package coverage
    echo ""
    echo "Coverage by package:"
    go tool cover -func="$OUTPUT_DIR/coverage_${TIMESTAMP}.out" | \
        awk '{if ($1 != "total:") {pkg=$1; sub(/\/[^\/]+$/, "", pkg); if (pkg != last_pkg) {print pkg; last_pkg=pkg}}}' | \
        sort -u

    # Create summary
    cat > "$OUTPUT_DIR/summary_${TIMESTAMP}.txt" << EOF
Coverage Report Summary
Generated: $(date)
Total Coverage: $TOTAL_COVERAGE

Reports generated:
- HTML: coverage_${TIMESTAMP}.html
- Function level: coverage_${TIMESTAMP}_func.txt
- Test output: test_output_${TIMESTAMP}.txt
EOF

    cat "$OUTPUT_DIR/summary_${TIMESTAMP}.txt"
else
    echo "ERROR: Coverage file was not generated"
    exit 1
fi

echo ""
echo "Coverage reports saved to: $OUTPUT_DIR/"
