# Contributing to Needy

## Development Setup

```bash
# Clone and setup
git clone <repo-url>
cd needy

# Build binaries
go build -o nd ./cmd/nd
go build -o ndadm ./cmd/ndadm

# Setup git hooks
./scripts/setup-hooks.sh
```

## Running Tests

```bash
# Unit tests
go test ./pkg/... ./internal/... -v

# Integration tests (invoke compiled binaries)
go test ./tests/integration/... -v

# All tests
go test ./... -v
```

## Code Style

- Use `go fmt` and `go vet` (enforced by pre-commit hook)
- Cobra for CLI commands
- Standard Go idioms for error handling
- Table-driven tests

## Adding Features

1. **New CLI Command**: Add to `cmd/nd/main.go`, test in `tests/integration/`
2. **New Message Type**: Update `pkg/messages/`, test in `pkg/messages/`
3. **Architecture Decision**: Document in `docs/adr/ADR-NNN-*.md`

## Architecture Decision Records

We document significant decisions using [Nygard-style ADRs](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions):

- **Status**: Proposed → Accepted/Rejected → Deprecated/Superseded
- **Context**: What problem are we solving?
- **Decision**: What did we decide?
- **Consequences**: What are the trade-offs?

See existing ADRs in `docs/adr/` for examples.

## Pull Request Process

1. Ensure all tests pass: `go test ./...`
2. Add tests for new functionality
3. Update documentation (README, AGENTS.md) if needed
4. Create ADR for architectural changes
