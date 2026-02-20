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

## Features

- **Dashboard** - Interactive view of all running services with sidebar navigation and filtering
- **Docker Management** - Start, stop, restart, delete containers; inspect and shell access
- **Process Control** - Discover and manage local dev server processes
- **Log Viewer** - Tail logs from Docker containers with scrolling support
- **Database Explorer** - Browse tables and query data from containerized databases

## Roadmap

- Environment variable management
- Port conflict detection and resolution

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
