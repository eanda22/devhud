# devhud

A unified, interactive TUI tool for managing local development environments.

Manage Docker containers, dev server processes, databases, logs, ports, and environment variables in a single terminal session.

## Status

**Phase 0: Project Setup** - Initial scaffolding complete

## Quick Start

```bash
# Install dependencies
go mod download

# Run
go run .

# Build
make build
./devhud
```

## Features (Planned)

- Dashboard view of all running services
- Docker container management
- Process discovery and control
- Multi-service log tailing with filtering
- Database explorer and query runner
- Environment variable management
- Port conflict detection

## Tech Stack

- Go 1.22+
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## Development

```bash
make build    # Compile binary
make run      # Run directly
make test     # Run tests
make lint     # Run linter
make fmt      # Format code
```

## License

MIT - See LICENSE file
