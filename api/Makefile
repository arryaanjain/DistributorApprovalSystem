.PHONY: dev build test clean

# Run backend with Air live-reloading
dev:
	air -c .air.toml

# Build binary
build:
	go build -o ./tmp/main ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf ./tmp
