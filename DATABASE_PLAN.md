# Database Management Feature - Implementation Plan

## Summary

Add read-only database management to devhud TUI. Users can browse database tables and view row data for Postgres/MySQL containers. Follows existing patterns (LogsView as reference), keeps business logic in `internal/`, graceful degradation on errors.

## Git Workflow

```bash
git checkout -b feature/database-management
```

All work happens in this branch. Open PR when complete (don't push to main).

## Todo List

### Phase 1: Database Detection
- [ ] Create `internal/db/detector.go` with image pattern matching
- [ ] Add `DBType string` field to `service.Service` struct
- [ ] Modify `internal/scanner/docker.go` to call detector during scan
- [ ] Write `internal/db/detector_test.go` unit tests
- [ ] Test: Verify DBType populates for postgres/mysql containers
- [ ] Commit: "Add database detection logic with image pattern matching"

### Phase 2: Connection Discovery
- [ ] Create `internal/db/connection.go` for env var discovery
- [ ] Implement `DiscoverConfig()` using Docker ContainerInspect
- [ ] Implement `BuildConnectionString()` for postgres and mysql
- [ ] Write `internal/db/connection_test.go` with mocked responses
- [ ] Test: Verify connection strings build correctly
- [ ] Commit: "Add connection discovery from Docker container env vars"

### Phase 3: Database Client & Queries
- [ ] Add dependencies: `go get github.com/lib/pq github.com/go-sql-driver/mysql`
- [ ] Create `internal/db/client.go` with NewClient, Close, Ping
- [ ] Create `internal/db/metadata.go` with ListTables (postgres/mysql)
- [ ] Create `internal/db/queries.go` with GetTableData, GetTableColumns
- [ ] Write `internal/db/metadata_test.go` with sql.DB mocks
- [ ] Test: Connect to local postgres container, list tables
- [ ] Commit: "Add database client and metadata query operations"

### Phase 4: Tables View UI
- [ ] Create `internal/tui/db_tables.go` (model LogsView structure)
- [ ] Add `dbTablesView *DBTablesView` field to App struct
- [ ] Add "db_tables" mode handling in `App.Update()`
- [ ] Modify [d] key handler to check DBType (not just delete confirm)
- [ ] Update `getServiceActions()` to show [d]atabase for DB services
- [ ] Add `TablesFetchedMsg` to `internal/tui/messages.go`
- [ ] Add view delegation in `App.View()`
- [ ] Test: Press [d] on postgres container, see tables list
- [ ] Commit: "Add tables view UI for browsing database tables"

### Phase 5: Table Data View UI
- [ ] Create `internal/tui/db_data.go` for row viewing
- [ ] Add `dbDataView *DBDataView` field to App struct
- [ ] Add "db_data" mode handling in `App.Update()`
- [ ] Modify DBTablesView to set `openTable` field on [enter]
- [ ] Add mode transition logic (tables → data → tables)
- [ ] Add `TableDataFetchedMsg` to messages.go
- [ ] Implement pagination ([n]/[p] keys)
- [ ] Test: Full flow - dashboard → tables → data → back
- [ ] Commit: "Add table data view UI with pagination support"

### Phase 6: Polish & Error Handling
- [ ] Add graceful error messages for connection failures
- [ ] Add loading states ("Connecting to database...")
- [ ] Add [r]efresh functionality to tables/data views
- [ ] Add [DB] indicator to dashboard service list
- [ ] Test: Stopped container shows error (doesn't crash)
- [ ] Test: Missing env vars shows error gracefully
- [ ] Commit: "Add error handling and polish for database features"

### Phase 7: Testing & Documentation
- [ ] Run `make lint` and fix issues
- [ ] Run `make fmt`
- [ ] Run `make test` - all tests pass
- [ ] Manual test: Postgres container full flow
- [ ] Manual test: MySQL container full flow
- [ ] Manual test: Error cases (stopped, no creds, timeout)
- [ ] Commit: "Add tests and documentation for database management"

### Phase 8: Open PR
- [ ] Push branch: `git push -u origin feature/database-management`
- [ ] Open PR: `gh pr create --title "Add database management features"`
- [ ] Verify PR description includes summary, features, technical details

## Implementation Details

### Database Detection
**File:** `internal/db/detector.go` (~50 lines)

```go
// Pattern matching for database images
var dbPatterns = map[string]string{
    "postgres":  "postgres",
    "mysql":     "mysql",
    "mariadb":   "mariadb",
    "mongo":     "mongodb",
    "mongodb":   "mongodb",
    "redis":     "redis",
}

func DetectType(imageName string) string {
    // Check if image contains any DB pattern
    // Return DB type string or empty string
}
```

**Changes:** Add `DBType string` to `service.Service` (line 24-37 in service/service.go)

### Connection Discovery
**File:** `internal/db/connection.go` (~120 lines)

```go
type ConnectionConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
}

// Inspects container to extract env vars
func DiscoverConfig(ctx, dockerClient, containerID, dbType) (*ConnectionConfig, error)

// Builds DSN string for database/sql
func BuildConnectionString(config, dbType) string
```

**Postgres env vars:** POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
**MySQL env vars:** MYSQL_USER, MYSQL_PASSWORD, MYSQL_DATABASE

### Database Client
**File:** `internal/db/client.go` (~80 lines)

```go
type Client struct {
    db     *sql.DB
    dbType string
    config *ConnectionConfig
}

func NewClient(ctx, config, dbType) (*Client, error)  // Opens connection
func (c *Client) Close() error
func (c *Client) Ping(ctx) error
```

**File:** `internal/db/metadata.go` (~100 lines)

```go
type TableInfo struct {
    Name       string
    RowCount   int
    ColumnCount int
}

func (c *Client) ListTables(ctx) ([]TableInfo, error)
```

**Postgres query:** `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
**MySQL query:** `SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()`

**File:** `internal/db/queries.go` (~80 lines)

```go
type ColumnInfo struct {
    Name string
    Type string
}

type RowData []map[string]interface{}

func (c *Client) GetTableColumns(ctx, tableName) ([]ColumnInfo, error)
func (c *Client) GetTableData(ctx, tableName, limit, offset) (RowData, error)
```

### Tables View
**File:** `internal/tui/db_tables.go` (~200 lines)

```go
type DBTablesView struct {
    service       *service.Service
    dbClient      *db.Client
    tables        []db.TableInfo
    selectedIndex int
    viewport      viewport.Model
    error         error
    ready         bool
    shouldExit    bool   // true = return to dashboard
    openTable     string // non-empty = open table data view
}

func NewDBTablesView(svc, dockerClient, width, height) *DBTablesView
func (v *DBTablesView) Init() tea.Cmd
func (v *DBTablesView) Update(msg) (*DBTablesView, tea.Cmd)
func (v *DBTablesView) View() string
```

**Navigation:**
- [esc] → sets shouldExit=true, App returns to dashboard
- [enter] → sets openTable=tableName, App switches to data view
- [r] → refetch tables
- [↑/↓] → navigate tables

### Table Data View
**File:** `internal/tui/db_data.go` (~220 lines)

```go
type DBDataView struct {
    service    *service.Service
    tableName  string
    dbClient   *db.Client
    columns    []db.ColumnInfo
    rows       db.RowData
    viewport   viewport.Model
    error      error
    ready      bool
    shouldExit bool  // true = return to tables view
    page       int
    pageSize   int
}

func NewDBDataView(svc, tableName, dbClient, width, height) *DBDataView
func (v *DBDataView) Init() tea.Cmd
func (v *DBDataView) Update(msg) (*DBDataView, tea.Cmd)
func (v *DBDataView) View() string
```

**Navigation:**
- [esc] → sets shouldExit=true, App returns to tables view
- [n] → next page (page++, refetch)
- [p] → previous page (page--, refetch)
- [r] → refresh current page
- [↑/↓] → scroll viewport

### App Mode Switching
**File:** `internal/tui/app.go` (modify existing)

Add fields (line 22-36):
```go
type App struct {
    // ... existing fields ...
    dbTablesView *DBTablesView
    dbDataView   *DBDataView
}
```

Add mode handling in `Update()` (before existing switch statement):
```go
// Handle db_tables mode
if a.mode == "db_tables" && a.dbTablesView != nil {
    updatedView, cmd := a.dbTablesView.Update(msg)
    a.dbTablesView = updatedView

    if a.dbTablesView.shouldExit {
        a.mode = "dashboard"
        if a.dbTablesView.dbClient != nil {
            a.dbTablesView.dbClient.Close()
        }
        a.dbTablesView = nil
        return a, a.scanCmd()
    }

    if a.dbTablesView.openTable != "" {
        tableName := a.dbTablesView.openTable
        a.dbDataView = NewDBDataView(a.dbTablesView.service, tableName,
            a.dbTablesView.dbClient, a.width, a.height)
        a.mode = "db_data"
        return a, a.dbDataView.Init()
    }

    return a, cmd
}

// Handle db_data mode
if a.mode == "db_data" && a.dbDataView != nil {
    updatedView, cmd := a.dbDataView.Update(msg)
    a.dbDataView = updatedView

    if a.dbDataView.shouldExit {
        a.mode = "db_tables"
        a.dbDataView = nil
        return a, nil
    }

    return a, cmd
}
```

Modify [d] key handler (line 233-240):
```go
case "d":
    services := a.services.GetAll()
    if a.selectedIndex < len(services) {
        svc := services[a.selectedIndex]
        if svc.DBType != "" {  // Is a database
            a.dbTablesView = NewDBTablesView(svc, a.dockerClient, a.width, a.height)
            a.mode = "db_tables"
            return a, a.dbTablesView.Init()
        }
        // Existing delete confirmation for non-DB containers
        if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
            a.confirmOperation = svc.ContainerID
        }
    }
```

Add view delegation in `View()` (line 270-275):
```go
func (a *App) View() string {
    if a.mode == "logs" && a.logsView != nil {
        return a.logsView.View()
    }
    if a.mode == "db_tables" && a.dbTablesView != nil {
        return a.dbTablesView.View()
    }
    if a.mode == "db_data" && a.dbDataView != nil {
        return a.dbDataView.View()
    }
    return RenderDashboard(...)
}
```

### Dashboard Changes
**File:** `internal/tui/dashboard.go`

Modify `getServiceActions()` (line 148-158):
```go
func getServiceActions(svc *service.Service) string {
    if svc.DBType != "" {
        return "[s]tart [x]stop [r]estart [d]atabase [l]ogs"
    }
    switch svc.Type {
    case service.ServiceTypeDocker, service.ServiceTypeCompose:
        return "[s]tart [x]stop [r]estart [d]elete [l]ogs"
    case service.ServiceTypeProcess:
        return "[x]kill [l]ogs"
    default:
        return ""
    }
}
```

Add [DB] indicator in service list rendering (line 51-54):
```go
serviceName := svc.Name
if svc.DBType != "" {
    serviceName += " [DB]"
}
row := fmt.Sprintf("%-6s %-40s %-10s %-6s %-10s",
    status,
    truncate(serviceName, 18),
    string(svc.Type),
    port,
    uptime,
)
```

### Messages
**File:** `internal/tui/messages.go` (add new types)

```go
type TablesFetchedMsg struct {
    Tables []db.TableInfo
    Error  error
}

type TableDataFetchedMsg struct {
    Columns []db.ColumnInfo
    Rows    db.RowData
    Error   error
}
```

## Critical Files

### Files to CREATE
1. `internal/db/detector.go` - Database image pattern matching
2. `internal/db/connection.go` - Env var discovery, connection strings
3. `internal/db/client.go` - Database connection wrapper
4. `internal/db/metadata.go` - List tables queries
5. `internal/db/queries.go` - Table data fetching, pagination
6. `internal/tui/db_tables.go` - Tables list view
7. `internal/tui/db_data.go` - Table data view
8. `internal/db/detector_test.go` - Detection unit tests
9. `internal/db/connection_test.go` - Connection unit tests

### Files to MODIFY
1. `internal/service/service.go` - Add DBType field to Service struct
2. `internal/scanner/docker.go` - Call detector during scan
3. `internal/tui/app.go` - Add db views, mode switching, [d] key handler
4. `internal/tui/dashboard.go` - Update getServiceActions(), add [DB] indicator
5. `internal/tui/messages.go` - Add TablesFetchedMsg, TableDataFetchedMsg
6. `go.mod` - Add lib/pq and go-sql-driver/mysql dependencies

## Verification

### Manual Testing Checklist
1. **Database Detection:**
   - Start postgres container: `docker run -d --name test-postgres -e POSTGRES_PASSWORD=pass -p 5432:5432 postgres`
   - Run devhud, verify postgres service shows `[DB]` indicator
   - Detail panel shows `[d]atabase` action

2. **Tables View:**
   - Select postgres service, press [d]
   - Verify tables view opens
   - Shows list of tables with row counts
   - [↑/↓] navigates tables
   - [esc] returns to dashboard

3. **Table Data View:**
   - In tables view, press [enter] on a table
   - Verify data view opens showing rows
   - Columns display correctly
   - [↑/↓] scrolls viewport
   - [n] loads next page (if >100 rows)
   - [p] loads previous page
   - [esc] returns to tables view

4. **Error Handling:**
   - Stop postgres container: `docker stop test-postgres`
   - Press [d] on stopped container
   - Verify error message displays (doesn't crash)
   - [esc] returns to dashboard gracefully

5. **MySQL Support:**
   - Start mysql container: `docker run -d --name test-mysql -e MYSQL_ROOT_PASSWORD=pass -p 3306:3306 mysql`
   - Repeat tables/data view tests
   - Verify works identically to postgres

6. **Refresh:**
   - In tables view, press [r]
   - Verify tables list refreshes
   - In data view, press [r]
   - Verify current page refreshes

### Tests Pass
```bash
make lint   # No errors
make fmt    # Code formatted
make test   # All unit tests pass
```

## Dependencies

**Add to go.mod:**
- `github.com/lib/pq` v1.10.9 (Postgres, CGo-free)
- `github.com/go-sql-driver/mysql` v1.8.1 (MySQL, CGo-free)

**Import in internal/db/client.go:**
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
    _ "github.com/go-sql-driver/mysql"
)
```

## Success Criteria

Feature complete when:
1. Database containers show `[DB]` indicator in dashboard
2. Pressing [d] on DB service opens tables view
3. Pressing [enter] on table opens data view with paginated rows
4. All navigation works: [esc] returns, [r] refreshes, [n/p] paginates
5. Connection errors display gracefully (no crashes)
6. Works with both postgres and mysql containers
7. All unit tests pass
8. No linter errors
9. PR opened with comprehensive description

## PR Description Template

```markdown
## Summary
Add read-only database management features to devhud TUI (Phase 1).

## Features
- Auto-detect database containers (postgres, mysql, mariadb) by image name
- Show [DB] indicator in dashboard for database services
- Press [d] on database service to view tables list
- Press [enter] on table to view data (paginated, 100 rows per page)
- Press [esc] to navigate back (data → tables → dashboard)
- Graceful error handling for connection failures

## Technical Details
- New `internal/db` package: detection, connection discovery, metadata queries
- Two new TUI views: DBTablesView, DBDataView (follow LogsView pattern)
- Added dependencies: lib/pq (postgres), go-sql-driver/mysql (both CGo-free)
- Connection strings built from container env vars (POSTGRES_USER, MYSQL_USER, etc.)
- Business logic in internal/, thin TUI layer (follows CLAUDE.md principles)

## Testing
- Unit tests for detector, connection builder, metadata queries
- Manual testing with postgres and mysql containers
- Error cases handled (stopped containers, missing credentials, timeouts)

## Future Work (Phase 2)
- Insert/update/delete operations
- MongoDB and Redis support
- Connection credential management UI
```

## Notes

- Keep functions under 40 lines
- Only add function header comments (explain why, not how)
- No inline comments unless logic is genuinely non-obvious
- Graceful degradation - DB failures don't affect Docker/Process features
- Simple and concrete - no over-abstraction
- Follow LogsView pattern exactly for new views
