package tui

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eanda22/devhud/internal/docker"
	"github.com/eanda22/devhud/internal/process"
	"github.com/eanda22/devhud/internal/scanner"
	"github.com/eanda22/devhud/internal/service"
)

type ScanCompleteMsg struct {
	Services []*service.Service
	Error    error
}

type TickMsg struct{}

type DiskUsageMsg struct {
	Usage *docker.DiskUsage
	Error error
}

type App struct {
	services         *service.Store
	scanner          *scanner.Scanner
	dockerClient     *docker.Client
	ticker           *time.Ticker
	selectedIndex    int
	lastError        error
	statusMessage    string
	confirmOperation string
	operatingOnID    string
	mode             string
	logsView         *LogsView
	actionMenuView   *ActionMenuView
	inspectView      *InspectView
	dbTablesView     *DBTablesView
	dbDataView       *DBDataView
	width            int
	height           int
	focus            Focus
	categories       []string
	activeCatIndex   int
	dockerDiskUsage  *docker.DiskUsage
	showDetailPanel  bool
}

type Focus int

const (
	FocusSidebar Focus = iota
	FocusMainList
)

func NewApp() (*App, error) {
	store := service.NewStore()
	scan, err := scanner.NewScanner(store)
	if err != nil {
		return nil, fmt.Errorf("scanner: %w", err)
	}

	dockerClient, _ := docker.NewClient()

	return &App{
		services:       store,
		scanner:        scan,
		dockerClient:   dockerClient,
		mode:           "dashboard",
		categories:     []string{"Containers", "Local Procs", "Databases"},
		activeCatIndex: 0,
		focus:          FocusSidebar,
	}, nil
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.scanCmd(),
		a.tickCmd(),
		a.fetchDiskUsageCmd(),
	)
}

// performs service discovery.
func (a *App) scanCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := a.scanner.Scan(ctx); err != nil {
			return ScanCompleteMsg{Error: err}
		}

		return ScanCompleteMsg{Services: a.services.GetAll()}
	}
}

// triggers periodic refresh.
func (a *App) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// starts a Docker container.
func (a *App) startServiceCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		if a.dockerClient == nil {
			return OperationCompleteMsg{Success: false, Message: "Docker unavailable"}
		}
		if err := a.dockerClient.Start(containerID); err != nil {
			return OperationCompleteMsg{Success: false, Message: fmt.Sprintf("Start failed: %v", err)}
		}
		return OperationCompleteMsg{Success: true, Message: "Container started"}
	}
}

// stops a Docker container.
func (a *App) stopServiceCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		if a.dockerClient == nil {
			return OperationCompleteMsg{Success: false, Message: "Docker unavailable"}
		}
		if err := a.dockerClient.Stop(containerID); err != nil {
			return OperationCompleteMsg{Success: false, Message: fmt.Sprintf("Stop failed: %v", err)}
		}
		return OperationCompleteMsg{Success: true, Message: "Container stopped"}
	}
}

// restarts a Docker container.
func (a *App) restartServiceCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		if a.dockerClient == nil {
			return OperationCompleteMsg{Success: false, Message: "Docker unavailable"}
		}
		if err := a.dockerClient.Restart(containerID); err != nil {
			return OperationCompleteMsg{Success: false, Message: fmt.Sprintf("Restart failed: %v", err)}
		}
		return OperationCompleteMsg{Success: true, Message: "Container restarted"}
	}
}

// deletes a Docker container.
func (a *App) deleteServiceCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		if a.dockerClient == nil {
			return OperationCompleteMsg{Success: false, Message: "Docker unavailable"}
		}
		if err := a.dockerClient.Remove(containerID); err != nil {
			return OperationCompleteMsg{Success: false, Message: fmt.Sprintf("Delete failed: %v", err)}
		}
		return OperationCompleteMsg{Success: true, Message: "Container deleted"}
	}
}

// stops a process with SIGTERM.
func (a *App) stopProcessCmd(pid int) tea.Cmd {
	return func() tea.Msg {
		if err := process.Stop(pid); err != nil {
			return OperationCompleteMsg{Success: false, Message: fmt.Sprintf("Stop failed: %v", err)}
		}
		return OperationCompleteMsg{Success: true, Message: "Process stopped"}
	}
}

// fetches Docker disk usage asynchronously.
func (a *App) fetchDiskUsageCmd() tea.Cmd {
	return func() tea.Msg {
		if a.dockerClient == nil {
			return DiskUsageMsg{Error: fmt.Errorf("docker unavailable")}
		}
		usage, err := a.dockerClient.GetDiskUsage()
		return DiskUsageMsg{Usage: usage, Error: err}
	}
}

// executes action selected from action menu.
func (a *App) executeActionFromMenu(actionName string, svc *service.Service) tea.Cmd {
	switch actionName {
	case "View Logs":
		a.logsView = NewLogsView(svc, a.dockerClient, a.width, a.height)
		a.mode = "logs"
		return a.logsView.Init()

	case "Restart Container":
		a.mode = "dashboard"
		a.statusMessage = "Restarting container..."
		a.operatingOnID = svc.ID
		return a.restartServiceCmd(svc.ContainerID)

	case "Stop Container":
		a.mode = "dashboard"
		a.statusMessage = "Stopping container..."
		a.operatingOnID = svc.ID
		return a.stopServiceCmd(svc.ContainerID)

	case "Start Container":
		a.mode = "dashboard"
		a.statusMessage = "Starting container..."
		a.operatingOnID = svc.ID
		return a.startServiceCmd(svc.ContainerID)

	case "Kill Process":
		a.mode = "dashboard"
		a.statusMessage = "Stopping process..."
		a.operatingOnID = svc.ID
		return a.stopProcessCmd(svc.PID)

	case "Browse Database":
		a.dbTablesView = NewDBTablesView(svc, a.dockerClient, a.width, a.height)
		a.mode = "db_tables"
		return a.dbTablesView.Init()

	case "Delete Container":
		a.mode = "dashboard"
		a.confirmOperation = svc.ContainerID
		return nil

	case "Inspect JSON":
		a.inspectView = NewInspectView(svc, a.dockerClient, a.width, a.height)
		a.mode = "inspect"
		return a.inspectView.Init()

	case "Open Shell (/bin/sh)":
		a.mode = "dashboard"
		c := exec.Command("docker", "exec", "-it", svc.ContainerID, "/bin/sh")
		return tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return OperationCompleteMsg{Success: false, Message: fmt.Sprintf("Shell failed: %v", err)}
			}
			return TickMsg{}
		})

	default:
		a.mode = "dashboard"
		return nil
	}
}

// returns services filtered by active category.
func (a *App) getFilteredServices() []*service.Service {
	if a.activeCatIndex == 0 {
		containers := append(
			a.services.GetByType(service.ServiceTypeDocker),
			a.services.GetByType(service.ServiceTypeCompose)...,
		)
		return containers
	} else if a.activeCatIndex == 1 {
		return a.services.GetByType(service.ServiceTypeProcess)
	} else if a.activeCatIndex == 2 {
		var databases []*service.Service
		for _, svc := range a.services.GetAll() {
			if svc.DBType != "" {
				databases = append(databases, svc)
			}
		}
		return databases
	}
	return a.services.GetAll()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.mode == "action_menu" && a.actionMenuView != nil {
		updatedView, cmd := a.actionMenuView.Update(msg)
		a.actionMenuView = updatedView

		if a.actionMenuView.shouldExit {
			if a.actionMenuView.executeAction != "" {
				actionCmd := a.executeActionFromMenu(a.actionMenuView.executeAction, a.actionMenuView.service)
				a.actionMenuView = nil
				return a, actionCmd
			}
			a.mode = "dashboard"
			a.actionMenuView = nil
			return a, nil
		}

		return a, cmd
	}

	if a.mode == "logs" && a.logsView != nil {
		updatedLogsView, cmd := a.logsView.Update(msg)
		a.logsView = updatedLogsView
		if a.logsView.shouldExit {
			a.mode = "dashboard"
			a.logsView = nil
			return a, a.scanCmd()
		}
		return a, cmd
	}

	if a.mode == "inspect" && a.inspectView != nil {
		updatedView, cmd := a.inspectView.Update(msg)
		a.inspectView = updatedView
		if a.inspectView.shouldExit {
			a.mode = "dashboard"
			a.inspectView = nil
			return a, a.scanCmd()
		}
		return a, cmd
	}

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
			a.dbDataView = NewDBDataView(a.dbTablesView.service, tableName, a.dbTablesView.dbClient, a.width, a.height)
			a.mode = "db_data"
			a.dbTablesView.openTable = ""
			return a, a.dbDataView.Init()
		}

		return a, cmd
	}

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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if a.dockerClient != nil {
				a.dockerClient.Close()
			}
			a.scanner.Close()
			return a, tea.Quit
		}

		if a.confirmOperation != "" {
			switch msg.String() {
			case "y", "Y":
				containerID := a.confirmOperation
				services := a.services.GetAll()
				if a.selectedIndex < len(services) {
					a.operatingOnID = services[a.selectedIndex].ID
				}
				a.confirmOperation = ""
				a.statusMessage = "Deleting container..."
				return a, a.deleteServiceCmd(containerID)
			case "n", "N":
				a.confirmOperation = ""
				a.statusMessage = "Delete cancelled"
			default:
				a.confirmOperation = ""
			}
			return a, nil
		}

		if msg.String() == "tab" {
			a.showDetailPanel = !a.showDetailPanel
			return a, nil
		}

		if a.focus == FocusSidebar {
			switch msg.String() {
			case "up", "k":
				if a.activeCatIndex > 0 {
					a.activeCatIndex--
					a.selectedIndex = 0
					if a.activeCatIndex == 0 || a.activeCatIndex == 2 {
						return a, a.fetchDiskUsageCmd()
					}
				}
			case "down", "j":
				if a.activeCatIndex < len(a.categories)-1 {
					a.activeCatIndex++
					a.selectedIndex = 0
					if a.activeCatIndex == 0 || a.activeCatIndex == 2 {
						return a, a.fetchDiskUsageCmd()
					}
				}
			case "enter", "right", "l":
				a.focus = FocusMainList
			}
		} else {
			switch msg.String() {
			case "left", "h":
				a.focus = FocusSidebar
			case "up", "k":
				if a.selectedIndex > 0 {
					a.selectedIndex--
				}
			case "down", "j":
				if a.selectedIndex < len(a.getFilteredServices())-1 {
					a.selectedIndex++
				}
			case "enter":
				services := a.getFilteredServices()
				if a.selectedIndex < len(services) {
					svc := services[a.selectedIndex]
					a.actionMenuView = NewActionMenuView(svc, a.dockerClient, a.width, a.height)
					a.mode = "action_menu"
					return a, a.actionMenuView.Init()
				}
			}
		}

	case ScanCompleteMsg:
		if msg.Error != nil {
			a.lastError = msg.Error
		}
		return a, nil

	case OperationCompleteMsg:
		a.statusMessage = msg.Message
		a.operatingOnID = ""
		return a, a.scanCmd()

	case TickMsg:
		return a, a.scanCmd()

	case DiskUsageMsg:
		if msg.Error == nil {
			a.dockerDiskUsage = msg.Usage
		}
		return a, nil
	}

	return a, nil
}

func (a *App) View() string {
	if a.mode == "logs" && a.logsView != nil {
		return a.logsView.View()
	}
	if a.mode == "inspect" && a.inspectView != nil {
		return a.inspectView.View()
	}
	if a.mode == "db_tables" && a.dbTablesView != nil {
		return a.dbTablesView.View()
	}
	if a.mode == "db_data" && a.dbDataView != nil {
		return a.dbDataView.View()
	}
	return RenderDashboard(a)
}
