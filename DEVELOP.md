# Development Guide

This document explains how to work with the Needy codebase.

## Prerequisites

- **Go 1.21+** (or latest stable)
- **Git**

## Quick Start

```bash
# Clone the repository
git clone https://github.com/akafred/needy.git
cd needy

# Build binaries
go build -o nd ./cmd/nd
go build -o ndadm ./cmd/ndadm

# Run tests
go test ./...
```

## Project Structure

```
.
├── cmd/                    # CLI entry points
│   ├── nd/main.go         # Agent CLI
│   └── ndadm/main.go      # Admin CLI
├── pkg/                    # Public packages
│   ├── client/            # NATS client wrapper
│   ├── config/            # Configuration handling
│   ├── messages/          # Message models
│   └── payload/           # Payload storage
├── internal/               # Private packages
│   └── server/            # Embedded NATS server
├── tests/                  # Go integration tests
│   └── integration/       # Scenario tests (invoke binaries)
├── docs/
│   ├── adr/               # Architecture Decision Records
│   └── guides/            # User guides
├── .github/
│   └── workflows/         # CI/CD workflows
├── .goreleaser.yaml        # Release configuration
└── go.mod
```

## Building

### Development Build

```bash
# Build both CLIs
go build -o nd ./cmd/nd
go build -o ndadm ./cmd/ndadm

# With version info
go build -ldflags "-X main.version=dev" -o nd ./cmd/nd
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o nd-linux ./cmd/nd

# Windows
GOOS=windows GOARCH=amd64 go build -o nd.exe ./cmd/nd

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o nd-arm64 ./cmd/nd
```

### Release Build

Use GoReleaser for official releases:

```bash
# Dry run (local)
goreleaser release --snapshot --clean

# Actual release (CI handles this)
goreleaser release --clean
```

## Testing

### Unit Tests

```bash
# Run package unit tests
go test ./pkg/... ./internal/... -v

# With coverage
go test -coverprofile=coverage.out ./pkg/... ./internal/...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests invoke the compiled `nd` and `ndadm` binaries:

```bash
# Build first (required)
go build -o nd ./cmd/nd
go build -o ndadm ./cmd/ndadm

# Run integration tests
go test ./tests/integration/... -v
```

### All Tests

```bash
# Run everything
go test ./... -v

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Code Quality

### Linting

```bash
go vet ./...
golangci-lint run
```

### Formatting

```bash
go fmt ./...
```

## Git Hooks

Git hooks are automatically installed when you run:

```bash
./scripts/setup-hooks.sh
```

The hooks include:
- **pre-commit**: Run `go fmt` and `go vet`
- **post-commit**: Report test coverage

## Adding a New Command

1. Add the command definition in `cmd/nd/main.go` or `cmd/ndadm/main.go`
2. Add any new packages to `pkg/` if needed
3. Add an integration test in `tests/integration/scenarios_test.go`
4. Update the README with CLI documentation

Example adding a new command:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func main() {
    rootCmd.AddCommand(myCmd)
    // ...
}
```

## Adding a New Package

1. Create directory under `pkg/` (public) or `internal/` (private)
2. Add Go files with package declaration
3. Import from other packages using full module path

```go
import "github.com/akafred/needy/pkg/mypackage"
```

## Debugging

### Verbose Output

```bash
# Enable NATS debug logging
NATS_DEBUG=true ./ndadm start
```

### Testing Against Local Server

```bash
# Terminal 1: Start server
./ndadm start --port 4222

# Terminal 2: Run commands
./nd init --name test-agent
./nd watch
```

## Release Process

Releases are automated via GitHub Actions:

1. Tag a version: `git tag v1.0.0`
2. Push the tag: `git push origin v1.0.0`
3. GitHub Actions runs GoReleaser
4. Binaries are published to GitHub Releases

Users can then update via:
```bash
nd update
ndadm update
```

## Troubleshooting

### "nd: command not found"

Ensure the binary is in your PATH:
```bash
export PATH=$PATH:$(pwd)
```

### NATS connection refused

Check if server is running:
```bash
./ndadm start --port 4222
```

### Tests failing with "nd binary not found"

Build the binaries first:
```bash
go build -o nd ./cmd/nd
go build -o ndadm ./cmd/ndadm
```

## Architecture Decisions

See [docs/adr/](docs/adr/) for design rationale:

- [ADR-001](docs/adr/ADR-001-messaging-technology.md) - Why NATS
- [ADR-002](docs/adr/ADR-002-message-protocol.md) - Message format
- [ADR-003](docs/adr/ADR-003-network-identity.md) - Agent identity (planned)
- [ADR-004](docs/adr/ADR-004-go-migration.md) - Python to Go migration
