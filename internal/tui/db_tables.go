package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/db"
	"github.com/eanda22/devhud/internal/docker"
	"github.com/eanda22/devhud/internal/service"
)

type DBTablesView struct {
	service       *service.Service
	dockerClient  *docker.Client
	dbClient      *db.Client
	tables        []db.TableInfo
	selectedIndex int
	viewport      viewport.Model
	error         error
	ready         bool
	shouldExit    bool
	openTable     string
}

// creates a new database tables view for a service.
func NewDBTablesView(svc *service.Service, dockerClient *docker.Client, width, height int) *DBTablesView {
	vp := viewport.New(width-4, height-6)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	return &DBTablesView{
		service:      svc,
		dockerClient: dockerClient,
		viewport:     vp,
		ready:        false,
	}
}

func (v *DBTablesView) Init() tea.Cmd {
	return v.fetchTablesCmd()
}

func (v *DBTablesView) Update(msg tea.Msg) (*DBTablesView, tea.Cmd) {
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
			return v, v.fetchTablesCmd()
		case "up", "k":
			if v.selectedIndex > 0 {
				v.selectedIndex--
				v.updateViewportContent()
			}
		case "down", "j":
			if v.selectedIndex < len(v.tables)-1 {
				v.selectedIndex++
				v.updateViewportContent()
			}
		case "enter":
			if v.selectedIndex < len(v.tables) {
				v.openTable = v.tables[v.selectedIndex].Name
				return v, nil
			}
		}

	case TablesFetchedMsg:
		if msg.Error != nil {
			v.error = msg.Error
			v.viewport.SetContent("Error connecting to database: " + msg.Error.Error())
		} else if len(msg.Tables) == 0 {
			v.viewport.SetContent("No tables found")
		} else {
			v.tables = msg.Tables
			v.dbClient = msg.Client
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

func (v *DBTablesView) View() string {
	if !v.ready {
		return "Connecting to database..."
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render(fmt.Sprintf("Database Tables: %s", v.service.Name))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[esc] back  [r]efresh  [↑/↓] navigate  [enter] view table")

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, v.viewport.View(), footer)
}

func (v *DBTablesView) updateViewportContent() {
	var lines []string
	for i, table := range v.tables {
		prefix := "  "
		if i == v.selectedIndex {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%-40s  Rows: %-10d  Columns: %d",
			prefix,
			table.Name,
			table.RowCount,
			table.ColumnCount,
		)
		lines = append(lines, line)
	}
	v.viewport.SetContent(strings.Join(lines, "\n"))
}

// fetches tables from the database.
func (v *DBTablesView) fetchTablesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		config, err := db.DiscoverConfig(ctx, v.dockerClient.GetRawClient(), v.service.ContainerID, v.service.DBType)
		if err != nil {
			return TablesFetchedMsg{
				Tables: nil,
				Client: nil,
				Error:  fmt.Errorf("discover config: %w", err),
			}
		}

		client, err := db.NewClient(ctx, config, v.service.DBType)
		if err != nil {
			return TablesFetchedMsg{
				Tables: nil,
				Client: nil,
				Error:  fmt.Errorf("connect to database: %w", err),
			}
		}

		tables, err := client.ListTables(ctx)
		if err != nil {
			client.Close()
			return TablesFetchedMsg{
				Tables: nil,
				Client: nil,
				Error:  fmt.Errorf("list tables: %w", err),
			}
		}

		return TablesFetchedMsg{
			Tables: tables,
			Client: client,
			Error:  nil,
		}
	}
}
