.PHONY: help build clean proto run-http run-grpc run-both test deps fmt install-proto-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  build-with-grpc    - Build with full gRPC support (generates protobuf)"
	@echo "  clean              - Clean build artifacts"
	@echo "  proto              - Generate protobuf code"
	@echo "  proto-if-needed    - Generate protobuf code only if missing"
	@echo "  run-http           - Run HTTP REST server only"
	@echo "  run-grpc           - Run gRPC server only"
	@echo "  run-both           - Run both HTTP and gRPC servers"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  deps               - Install/update dependencies"
	@echo "  fmt                - Format code"
	@echo "  install-proto-tools - Install protobuf generation tools"

# Build targets
build: proto-if-needed
	@echo "Building fr0g-ai-bridge..."
	go build -o bin/fr0g-ai-bridge ./cmd/fr0g-ai-bridge

build-with-grpc: proto
	@echo "Building fr0g-ai-bridge with gRPC support..."
	go build -o bin/fr0g-ai-bridge ./cmd/fr0g-ai-bridge

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf internal/pb/*.pb.go

# Protocol Buffers
proto:
	@echo "Generating protobuf code..."
	@mkdir -p internal/pb
	protoc --go_out=. --go_opt=module=github.com/fr0g-vibe/fr0g-ai-bridge \
		--go-grpc_out=. --go-grpc_opt=module=github.com/fr0g-vibe/fr0g-ai-bridge \
		proto/fr0g_ai_bridge.proto

proto-if-needed:
	@if [ ! -f internal/pb/fr0g_ai_bridge.pb.go ]; then \
		echo "Protobuf files missing, generating..."; \
		$(MAKE) proto; \
	fi

# Run targets
run-http: build
	@echo "Starting HTTP REST server..."
	./bin/fr0g-ai-bridge -http-only

run-grpc: build
	@echo "Starting gRPC server..."
	./bin/fr0g-ai-bridge -grpc-only

run-both: build
	@echo "Starting both HTTP and gRPC servers..."
	./bin/fr0g-ai-bridge

# Development targets
test:
	@echo "Running tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

deps:
	@echo "Installing/updating dependencies..."
	go mod tidy
	go mod download

fmt:
	@echo "Formatting code..."
	go fmt ./...

install-proto-tools:
	@echo "Installing protobuf generation tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Make sure protoc is installed and in your PATH"
	@echo "On Ubuntu/Debian: sudo apt install protobuf-compiler"
	@echo "On macOS: brew install protobuf"
	@echo "On Arch Linux: sudo pacman -S protobuf"

# Docker targets (optional)
docker-build:
	@echo "Building Docker image..."
	docker build -t fr0g-ai-bridge .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9090:9090 fr0g-ai-bridge
