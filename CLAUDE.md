# CLAUDE.md

## Project

**devhud** is a unified, interactive TUI tool for managing local development environments. It discovers and manages Docker containers, dev server processes, databases, logs, ports, and environment variables in a single terminal session.

Built in Go 1.22+ with Bubble Tea, Lip Gloss, Bubbles, and Cobra.

**Key Principle:** Business logic lives in `internal/` packages (service, scanner, docker, db, etc.). The TUI layer is thin—it coordinates and displays, not implements logic.

## Principles

1. **Simple and Concrete** — Don't over-abstract. Start with the simplest solution. Extract interfaces only when there's a second implementation or for testing.
2. **Zero Configuration** — No config files unless absolutely necessary. Discover and infer system state automatically.
3. **Graceful Degradation** — If Docker fails, still show process-based services. If DB connection fails, show the error but keep other features working.
4. **Native Only** — No CGo, no AI/LLM features, no external wrappers. Pure Go libraries and system calls via os/exec.
5. **Minimal Dependencies** — Use only Charm's ecosystem (Bubble Tea, Lip Gloss, Bubbles) for TUI. Docker SDK for container operations.

## Rules

### Go Style
- Follow `gofmt`, `go vet` — no unused imports or variables
- Use `internal/` for all non-exported packages
- Error handling: return errors, wrap with `fmt.Errorf("context: %w", err)`, never panic
- Keep functions short (under 40 lines preferred)
- Define interfaces where consumed, not where implemented
- Use table-driven tests; name test files `*_test.go`

## Code Comments
- Do not add inline comments that restate what the code does. The code should be self-documenting.
- Do not add comments explaining what changed or what something "used to be." That's what git history is for.
- Do not add TODO, FIXME, or NOTE comments unless explicitly asked.
- Every function/method gets a brief header comment explaining *why* it exists and what it returns — not *how* it works.
- If a piece of logic is genuinely non-obvious, a short comment explaining *why* is acceptable.
- When modifying existing code, do not add or change comments unless the function's purpose changed.
- Zero is the right number of comments for most lines of code.

### Commit Messages
- Keep commit messages concise: 1-2 sentences max.
- Be clear about what changed and why.
- No signatures or co-authored lines needed.

### Pull Requests
Before opening a PR:
1. Fetch latest main: `git fetch origin main`
2. Rebase feature branch: `git rebase origin/main`
3. Resolve any conflicts locally
4. Push to remote: `git push -u origin <branch-name>` (or `--force-with-lease` if rebasing existing branch)
5. Open PR with clear description of changes

This ensures the branch is up-to-date and conflicts are resolved before review.

### Bubble Tea
- Each TUI mode is a `tea.Model` (Dashboard, Logs, DB, Env)
- Root `App` model delegates to the active mode
- Use `tea.Cmd` for async operations (never block in Update)
- Use `tea.Batch` to combine commands
- Define custom message types per module (e.g., `ScanCompleteMsg`)
- View functions are pure — no side effects

### Project Structure
- Business logic goes in `internal/` packages
- TUI models coordinate, they don't implement logic
- `internal/service/` is the single source of truth for system state
- Docker operations only through `internal/docker/` — never call Docker SDK from TUI

### Docker
- Use Docker SDK for container operations
- Exception: Shell out to `docker compose` (too complex to reimplement)
- Graceful fallback if Docker daemon is unavailable

### Error Handling
- Display errors gracefully (status bar or inline message)
- Don't crash—show error and let user continue
- Degrade feature-by-feature (Docker failing shouldn't break Logs or DB)

## Commands

```bash
make build   # Compile to ./devhud
make run     # Run directly
make test    # Run test suite
make lint    # Run go vet
make fmt     # Format with gofmt
make clean   # Remove ./devhud binary
```

## Definition of Done

A feature or fix is done when:
- Code is tested (unit tests for logic, integration tests for workflows)
- All linters pass (`make lint`, `make fmt`)
- Error cases are handled (no unhandled panics or nil derefs)
- Degradation is graceful (failures don't cascade)
- Changes are minimal (only code necessary for the task)

## Documentation

### README Maintenance
- Keep README.md synchronized with implemented features
- When adding a new feature, move it from Roadmap to Features section
- When removing a feature, update README immediately
- README should always reflect the current state of the codebase, not future plans
- If you notice README is outdated, update it before implementing new features
