package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/docker"
	"github.com/eanda22/devhud/internal/service"
)

type LogsView struct {
	service      *service.Service
	viewport     viewport.Model
	dockerClient *docker.Client
	error        error
	ready        bool
	shouldExit   bool
}

// creates a new logs view for a service.
func NewLogsView(svc *service.Service, client *docker.Client, width, height int) *LogsView {
	vp := viewport.New(width-4, height-6)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	return &LogsView{
		service:      svc,
		viewport:     vp,
		dockerClient: client,
		ready:        false,
	}
}

func (l *LogsView) Init() tea.Cmd {
	return l.fetchLogsCmd()
}

func (l *LogsView) Update(msg tea.Msg) (*LogsView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return l, tea.Quit
		case "esc":
			l.shouldExit = true
			return l, nil
		case "r":
			return l, l.fetchLogsCmd()
		}

	case LogsFetchedMsg:
		if msg.Error != nil {
			l.error = msg.Error
			l.viewport.SetContent("Error fetching logs: " + msg.Error.Error())
		} else if len(msg.Logs) == 0 {
			l.viewport.SetContent("No logs found")
		} else {
			l.viewport.SetContent(strings.Join(msg.Logs, "\n"))
		}
		l.ready = true
		return l, nil

	case tea.WindowSizeMsg:
		l.viewport.Width = msg.Width - 4
		l.viewport.Height = msg.Height - 6
	}

	l.viewport, cmd = l.viewport.Update(msg)
	return l, cmd
}

func (l *LogsView) View() string {
	if !l.ready {
		return "Loading logs..."
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render(fmt.Sprintf("Logs: %s", l.service.Name))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[esc] back  [r]efresh  [↑/↓] scroll  [g/G] top/bottom")

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, l.viewport.View(), footer)
}

// fetches logs from Docker container.
func (l *LogsView) fetchLogsCmd() tea.Cmd {
	return func() tea.Msg {
		if l.service.Type == service.ServiceTypeProcess {
			return LogsFetchedMsg{
				Logs:  []string{"Logs not available for process-based services"},
				Error: nil,
			}
		}

		if l.dockerClient == nil {
			return LogsFetchedMsg{
				Logs:  nil,
				Error: fmt.Errorf("Docker unavailable"),
			}
		}

		logs, err := l.dockerClient.GetLogs(l.service.ContainerID, 100)
		return LogsFetchedMsg{
			Logs:  logs,
			Error: err,
		}
	}
}
