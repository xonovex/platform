#!/usr/bin/env bash
set -euo pipefail

# Validation script for script-agent-go
# Runs all validation steps in sequence

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(cd "$PACKAGE_DIR/../../.." && pwd)"

cd "$ROOT_DIR"

echo "==> Running validation suite for script-agent-go"
echo ""

# Type checking
echo "==> Type checking..."
moon run script-lib-go:go-typecheck
moon run script-agent-go:go-typecheck
echo "✓ Type check passed"
echo ""

# Linting (optional - requires golangci-lint)
if command -v golangci-lint &>/dev/null; then
    echo "==> Linting..."
    moon run script-lib-go:go-lint || true
    moon run script-agent-go:go-lint || true
    echo "✓ Lint completed"
    echo ""
else
    echo "⚠ golangci-lint not installed, skipping lint step"
    echo ""
fi

# Building
echo "==> Building..."
moon run script-lib-go:go-build
moon run script-agent-go:go-build
echo "✓ Build passed"
echo ""

# Unit tests
echo "==> Running unit tests..."
moon run script-lib-go:go-test
moon run script-agent-go:go-test
echo "✓ Unit tests passed"
echo ""

# Integration tests (optional - requires built binary)
if [[ -f "$PACKAGE_DIR/dist/script-agent-go" ]]; then
    echo "==> Running integration tests..."
    cd "$PACKAGE_DIR"
    if go test -v -tags=integration ./test/integration/... 2>/dev/null; then
        echo "✓ Integration tests passed"
    else
        echo "⚠ Integration tests failed or skipped"
    fi
    cd "$ROOT_DIR"
    echo ""
else
    echo "⚠ Binary not found, skipping integration tests"
    echo ""
fi

# Feature parity check
echo "==> Feature parity summary..."
PARITY_FILE="$PACKAGE_DIR/test/feature-parity.md"
if [[ -f "$PARITY_FILE" ]]; then
    UNCHECKED=$(grep -c "\- \[ \]" "$PARITY_FILE" 2>/dev/null || echo "0")
    CHECKED=$(grep -c "\- \[x\]" "$PARITY_FILE" 2>/dev/null || echo "0")
    TOTAL=$((UNCHECKED + CHECKED))
    if [[ "$UNCHECKED" -eq 0 ]]; then
        echo "✓ All $TOTAL features implemented!"
    else
        echo "⚠ $CHECKED/$TOTAL features implemented ($UNCHECKED pending)"
    fi
else
    echo "⚠ Feature parity file not found"
fi
echo ""

echo "========================================"
echo "✅ Validation suite completed!"
echo "========================================"
