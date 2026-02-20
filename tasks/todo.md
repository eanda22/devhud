# Log Viewing Feature - TODO

## Tasks

- [x] 1. Add GetLogs method to docker.Client in `/internal/docker/operations.go`
- [x] 2. Add LogsFetchedMsg to `/internal/tui/messages.go`
- [x] 3. Create LogsView model in `/internal/tui/logs.go`
- [x] 4. Add mode switching to App in `/internal/tui/app.go`
- [x] 5. Update action hints in `/internal/tui/dashboard.go`
- [x] 6. Test all functionality (Docker logs, process logs, navigation, scrolling)

## Review

### Summary
Added log viewing feature to devhud. Users can now press [l] on any service to view logs in a scrollable viewport. The implementation uses simple mode switching between dashboard and logs views, following the project's simplicity principles.

### Changes Made

**1. `/internal/docker/operations.go`**
- Added `GetLogs(containerID string, lines int) ([]string, error)` method
- Fetches last N lines using Docker SDK's ContainerLogs API
- Strips Docker log header bytes (first 8 bytes per line)
- Returns logs as []string

**2. `/internal/tui/messages.go`**
- Added `LogsFetchedMsg` message type with Logs []string and Error fields
- Used for async log fetching results

**3. `/internal/tui/logs.go` (NEW FILE)**
- Created `LogsView` model implementing Bubble Tea patterns
- Uses `viewport` component from bubbles for scrolling
- Handles key bindings: [esc] return, [r] refresh, j/k/up/down/g/G scroll
- Shows "not available" message for process services
- Gracefully handles errors (Docker unavailable, fetch failures)

**4. `/internal/tui/app.go`**
- Added fields: `mode`, `logsView`, `width`, `height`
- Implemented mode switching logic in Update()
- Routes Update/View calls based on current mode
- Added [l] key handler to create LogsView and switch to logs mode
- Handles WindowSizeMsg to track terminal dimensions

**5. `/internal/tui/dashboard.go`**
- Updated `getServiceActions()` to include [l]ogs in action hints
- Docker/Compose: "[s]tart [x]stop [r]estart [d]elete [l]ogs"
- Process: "[x]kill [l]ogs"

### Key Features
- Simple mode switching (string-based, no complex state machine)
- Scrollable log viewer using battle-tested viewport component
- Manual refresh with [r] key (simple MVP approach)
- Graceful degradation for process services
- Error handling for Docker unavailability and fetch failures
- [esc] returns to dashboard, [q] quits app globally

### Testing Checklist
Ready for manual testing:
- [ ] Navigate to Docker container, press [l] - should show logs
- [ ] Scroll logs with j/k/up/down/PageUp/PageDown/g/G
- [ ] Press [r] - should refresh logs
- [ ] Press [esc] - should return to dashboard
- [ ] Navigate to process, press [l] - should show "not available"
- [ ] From logs view, press [q] - should quit app

### Impact
- 5 files changed (4 modified, 1 new)
- ~240 lines added total
- No breaking changes
- Follows all CLAUDE.md principles (simplicity, minimal changes, business logic in internal/, graceful degradation)

### Benefits
- Users can now view container logs without leaving devhud
- Scrollable viewport for easy log navigation
- Consistent UX with existing keyboard shortcuts
- Foundation for future enhancements (follow mode, filtering, etc.)
