# fr0g-ai-bridge

A gRPC and REST API bridge service that forwards chat completion requests to OpenWebUI, with support for persona prompts.

## Features

- **Dual Protocol Support**: Both gRPC and REST API endpoints
- **Persona Integration**: Inject persona prompts into chat completions
- **OpenWebUI Compatible**: Forwards requests to OpenWebUI's chat completion API
- **Health Monitoring**: Built-in health check endpoints
- **Configurable**: YAML configuration with environment variable overrides
- **Docker Ready**: Containerized deployment support

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Protocol Buffers compiler (`protoc`)
- OpenWebUI instance running and accessible

### Installation

1. **Install protobuf tools:**
   ```bash
   make install-proto-tools
   export PATH="$(go env GOPATH)/bin:$PATH"
   ```

2. **Build the application:**
   ```bash
   make build-with-grpc
   ```

3. **Configure the service:**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your OpenWebUI settings
   ```

4. **Run the service:**
   ```bash
   # Run both HTTP and gRPC servers
   ./bin/fr0g-ai-bridge
   
   # Or run only HTTP REST API
   ./bin/fr0g-ai-bridge -http-only
   
   # Or run only gRPC server
   ./bin/fr0g-ai-bridge -grpc-only
   ```

## Configuration

### Configuration File

Create a `config.yaml` file based on `config.example.yaml`:

```yaml
server:
  http_port: 8080
  grpc_port: 9090
  host: "0.0.0.0"

openwebui:
  base_url: "http://localhost:3000"
  api_key: "your-openwebui-api-key"
  timeout: 30

logging:
  level: "info"
  format: "json"
```

### Environment Variables

You can override configuration with environment variables:

- `HTTP_PORT`: HTTP server port
- `GRPC_PORT`: gRPC server port  
- `HOST`: Server host
- `OPENWEBUI_BASE_URL`: OpenWebUI base URL
- `OPENWEBUI_API_KEY`: OpenWebUI API key
- `OPENWEBUI_TIMEOUT`: Request timeout in seconds
- `LOG_LEVEL`: Logging level

## API Usage

### REST API

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Chat Completion
```bash
curl -X POST http://localhost:8080/api/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "persona_prompt": "You are a helpful AI assistant with expertise in software development."
  }'
```

### gRPC API

The gRPC service runs on port 9090 by default. Use your preferred gRPC client or generate client code from the protobuf definition in `proto/fr0g_ai_bridge.proto`.

## Persona Prompts

The bridge service supports persona prompts that are automatically injected as system messages:

```json
{
  "model": "llama3.1",
  "messages": [
    {"role": "user", "content": "Explain microservices"}
  ],
  "persona_prompt": "You are a senior software architect with 15 years of experience in distributed systems."
}
```

The persona prompt will be:
- Added as a system message if none exists
- Prepended to existing system messages

## Development

### Available Make Targets

```bash
make help                 # Show all available targets
make build               # Build the application
make build-with-grpc     # Build with full gRPC support
make proto               # Generate protobuf code
make test                # Run tests
make test-coverage       # Run tests with coverage
make run-http            # Run HTTP server only
make run-grpc            # Run gRPC server only
make run-both            # Run both servers
make clean               # Clean build artifacts
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Docker Deployment

### Build Docker Image

```bash
make docker-build
```

### Run with Docker

```bash
# Using docker run
docker run -p 8080:8080 -p 9090:9090 \
  -e OPENWEBUI_BASE_URL=http://your-openwebui:3000 \
  -e OPENWEBUI_API_KEY=your-api-key \
  fr0g-ai-bridge

# Using docker-compose
version: '3.8'
services:
  fr0g-ai-bridge:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - OPENWEBUI_BASE_URL=http://openwebui:3000
      - OPENWEBUI_API_KEY=your-api-key
    depends_on:
      - openwebui
```

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client App    │    │  fr0g-ai-bridge  │    │   OpenWebUI     │
│                 │    │                  │    │   Instance      │
│  ┌───────────┐  │    │                  │    │                 │
│  │ REST API  │──┼────┼─► HTTP Server    │    │                 │
│  └───────────┘  │    │                  │    │                 │
│                 │    │  ┌─────────────┐ │    │  ┌────────────┐ │
│  ┌───────────┐  │    │  │   Persona   │ │    │  │    Chat    │ │
│  │ gRPC API  │──┼────┼─►│ Integration │─┼────┼─►│Completions │ │
│  └───────────┘  │    │  └─────────────┘ │    │  │    API     │ │
└─────────────────┘    │                  │    │  └────────────┘ │
                       │  ┌─────────────┐ │    │                 │
                       │  │ gRPC Server │ │    │                 │
                       │  └─────────────┘ │    │                 │
                       └──────────────────┘    └─────────────────┘
```

## License

This project is licensed under the GPL-3.0 License - see the LICENSE file for details.
