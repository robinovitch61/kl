.PHONY: all test testsum lint build bench fmt goimports

all: goimports lint test build

# Run goimports formatting
goimports:
	goimports -w .

# Run linting
lint:
	golangci-lint run --output.text.print-issued-lines=false

# Run all tests
test:
	go test ./...

# Run all tests, show summary
testsum:
	gotestsum --format testname 2>&1 | grep -E "^FAIL" | sort | uniq

# Build the project
build:
	go build ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem -run=^$$ ./...

# Format code
fmt:
	go fmt ./...
