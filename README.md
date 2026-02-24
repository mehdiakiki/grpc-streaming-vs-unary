# gRPC Unary vs Streaming

A demonstration of the difference between gRPC unary and streaming modes, focusing on memory allocation.

## The Difference

- **Unary**: Loads entire file into memory, then sends → Grows to ~file size
- **Streaming**: Sends file in chunks incrementally → Stays at ~chunk size

## Quick Start

```bash
# Run the test (creates 100MB test file automatically)
./test.sh
```

Or manually:

```bash
# Start server
go run ./cmd/server &

# Test unary
go run ./cmd/client -mode unary -file test.bin

# Test streaming
go run ./cmd/client -mode stream -file test.bin -chunk 1048576
```

## Expected Output

![Example output](https://github.com/user-attachments/assets/8d9dcec2-acc9-4b9f-a798-7ef39d5be3c6)

## When to Use Each

- **Unary** - Simple, good for small requests
- **Streaming** - Large files, real-time data, memory-constrained environments

## Files

- `proto/file.proto` - Protocol buffer definitions
- `cmd/server/main.go` - gRPC server
- `cmd/client/main.go` - Client with memory tracking
- `cmd/bench/main.go` - Peak memory benchmark tool
- `test.sh` - Automated test script
