#!/bin/bash
set -e

TEST_FILE="test.bin"
SIZE_MB=100
CHUNK_SIZE=1048576

echo "=== gRPC Unary vs Streaming Test ==="
echo ""

if [ ! -f "$TEST_FILE" ]; then
	echo "Creating test file ($SIZE_MB MB)..."
	dd if=/dev/zero of="$TEST_FILE" bs=1M count=$SIZE_MB 2>/dev/null
fi

FILE_SIZE=$(stat -f%z "$TEST_FILE" 2>/dev/null || stat -c%s "$TEST_FILE")
echo "Test file: $TEST_FILE ($((FILE_SIZE / 1024 / 1024)) MB)"
echo ""

pkill -f "cmd/server" 2>/dev/null || true
sleep 1

echo "Starting server..."
go run ./cmd/server &
SERVER_PID=$!
sleep 1

cleanup() {
	kill $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

echo ""
echo "--- Unary (loads entire file into memory) ---"
go run ./cmd/client -mode unary -file "$TEST_FILE"

echo ""
echo "--- Streaming (chunks sent incrementally) ---"
go run ./cmd/client -mode stream -file "$TEST_FILE" -chunk $CHUNK_SIZE

echo ""
echo "=== Key Difference ==="
echo "Unary: Alloc grows to ~100MB (entire file in memory)"
echo "Streaming: Stays at ~8MB (only chunk size in memory)"
