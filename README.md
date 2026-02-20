# devhud

A unified, interactive TUI tool for managing local development environments.

Manage Docker containers, dev server processes, databases, logs, ports, and environment variables in a single terminal session.

## Installation

### Download Pre-Built Binary

Download the latest release for your platform from the [Releases](https://github.com/eanda22/devhud/releases) page.

```bash
# Extract the archive
tar -xzf devhud_*_*.tar.gz

# Move to PATH (optional)
mv devhud /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/eanda22/devhud
cd devhud
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
