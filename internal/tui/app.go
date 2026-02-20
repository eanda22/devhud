package tui

import (
	"context"
	"fmt"
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
	dbTablesView     *DBTablesView
	width            int
	height           int
}

func NewApp() (*App, error) {
	store := service.NewStore()
	scan, err := scanner.NewScanner(store)
	if err != nil {
		return nil, fmt.Errorf("scanner: %w", err)
	}

	dockerClient, _ := docker.NewClient()

	return &App{
		services:     store,
		scanner:      scan,
		dockerClient: dockerClient,
		mode:         "dashboard",
	}, nil
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.scanCmd(),
		a.tickCmd(),
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

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			a.statusMessage = "Table data view not yet implemented"
			a.dbTablesView.openTable = ""
		}

		return a, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
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

		switch msg.String() {
		case "q", "ctrl+c":
			if a.dockerClient != nil {
				a.dockerClient.Close()
			}
			a.scanner.Close()
			return a, tea.Quit
		case "up", "k":
			if a.selectedIndex > 0 {
				a.selectedIndex--
			}
		case "down", "j":
			if a.selectedIndex < len(a.services.GetAll())-1 {
				a.selectedIndex++
			}
		case "s":
			services := a.services.GetAll()
			if a.selectedIndex < len(services) {
				svc := services[a.selectedIndex]
				if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
					a.statusMessage = "Starting container..."
					a.operatingOnID = svc.ID
					return a, a.startServiceCmd(svc.ContainerID)
				}
			}
		case "x":
			services := a.services.GetAll()
			if a.selectedIndex < len(services) {
				svc := services[a.selectedIndex]
				if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
					a.statusMessage = "Stopping container..."
					a.operatingOnID = svc.ID
					return a, a.stopServiceCmd(svc.ContainerID)
				} else if svc.Type == service.ServiceTypeProcess {
					a.statusMessage = "Stopping process..."
					a.operatingOnID = svc.ID
					return a, a.stopProcessCmd(svc.PID)
				}
			}
		case "r":
			services := a.services.GetAll()
			if a.selectedIndex < len(services) {
				svc := services[a.selectedIndex]
				if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
					a.statusMessage = "Restarting container..."
					a.operatingOnID = svc.ID
					return a, a.restartServiceCmd(svc.ContainerID)
				}
			}
		case "d":
			services := a.services.GetAll()
			if a.selectedIndex < len(services) {
				svc := services[a.selectedIndex]
				if svc.DBType != "" {
					a.dbTablesView = NewDBTablesView(svc, a.dockerClient, a.width, a.height)
					a.mode = "db_tables"
					return a, a.dbTablesView.Init()
				}
				if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
					a.confirmOperation = svc.ContainerID
				}
			}
		case "l":
			services := a.services.GetAll()
			if a.selectedIndex < len(services) {
				svc := services[a.selectedIndex]
				a.logsView = NewLogsView(svc, a.dockerClient, a.width, a.height)
				a.mode = "logs"
				return a, a.logsView.Init()
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
	}

	return a, nil
}

func (a *App) View() string {
	if a.mode == "logs" && a.logsView != nil {
		return a.logsView.View()
	}
	if a.mode == "db_tables" && a.dbTablesView != nil {
		return a.dbTablesView.View()
	}
	return RenderDashboard(a.services.GetAll(), a.selectedIndex, a.lastError, a.statusMessage, a.confirmOperation, a.operatingOnID)
}
