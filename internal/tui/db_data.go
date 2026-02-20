package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/db"
	"github.com/eanda22/devhud/internal/service"
)

type DBDataView struct {
	service    *service.Service
	tableName  string
	dbClient   *db.Client
	columns    []db.ColumnInfo
	rows       db.RowData
	viewport   viewport.Model
	error      error
	ready      bool
	shouldExit bool
	page       int
	pageSize   int
}

// creates a new database data view for a table.
func NewDBDataView(svc *service.Service, tableName string, dbClient *db.Client, width, height int) *DBDataView {
	vp := viewport.New(width-4, height-6)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	return &DBDataView{
		service:   svc,
		tableName: tableName,
		dbClient:  dbClient,
		viewport:  vp,
		ready:     false,
		page:      0,
		pageSize:  100,
	}
}

func (v *DBDataView) Init() tea.Cmd {
	return v.fetchDataCmd()
}

func (v *DBDataView) Update(msg tea.Msg) (*DBDataView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return v, tea.Quit
		case "esc":
			v.shouldExit = true
			return v, nil
		case "r":
			return v, v.fetchDataCmd()
		case "n":
			v.page++
			return v, v.fetchDataCmd()
		case "p":
			if v.page > 0 {
				v.page--
				return v, v.fetchDataCmd()
			}
		}

	case TableDataFetchedMsg:
		if msg.Error != nil {
			v.error = msg.Error
			v.viewport.SetContent("Error fetching table data: " + msg.Error.Error())
		} else {
			v.columns = msg.Columns
			v.rows = msg.Rows
			v.updateViewportContent()
		}
		v.ready = true
		return v, nil

	case tea.WindowSizeMsg:
		v.viewport.Width = msg.Width - 4
		v.viewport.Height = msg.Height - 6
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *DBDataView) View() string {
	if !v.ready {
		return "Loading table data..."
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render(fmt.Sprintf("Table: %s (Page %d)", v.tableName, v.page+1))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[esc] back  [r]efresh  [n]ext page  [p]revious page  [↑/↓] scroll")

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, v.viewport.View(), footer)
}

func (v *DBDataView) updateViewportContent() {
	if len(v.rows) == 0 {
		v.viewport.SetContent("No data found")
		return
	}

	var lines []string

	colNames := make([]string, len(v.columns))
	for i, col := range v.columns {
		colNames[i] = col.Name
	}
	lines = append(lines, strings.Join(colNames, " | "))
	lines = append(lines, strings.Repeat("-", 80))

	for _, row := range v.rows {
		values := make([]string, len(v.columns))
		for i, col := range v.columns {
			val := row[col.Name]
			if val == nil {
				values[i] = "NULL"
			} else {
				values[i] = fmt.Sprintf("%v", val)
				if len(values[i]) > 30 {
					values[i] = values[i][:27] + "..."
				}
			}
		}
		lines = append(lines, strings.Join(values, " | "))
	}

	v.viewport.SetContent(strings.Join(lines, "\n"))
}

// fetches table data from the database.
func (v *DBDataView) fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		columns, err := v.dbClient.GetTableColumns(ctx, v.tableName)
		if err != nil {
			return TableDataFetchedMsg{
				Columns: nil,
				Rows:    nil,
				Error:   fmt.Errorf("get columns: %w", err),
			}
		}

		offset := v.page * v.pageSize
		rows, err := v.dbClient.GetTableData(ctx, v.tableName, v.pageSize, offset)
		if err != nil {
			return TableDataFetchedMsg{
				Columns: nil,
				Rows:    nil,
				Error:   fmt.Errorf("get data: %w", err),
			}
		}

		return TableDataFetchedMsg{
			Columns: columns,
			Rows:    rows,
			Error:   nil,
		}
	}
}
