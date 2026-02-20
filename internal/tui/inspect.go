package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/docker"
	"github.com/eanda22/devhud/internal/service"
)

type InspectView struct {
	service      *service.Service
	viewport     viewport.Model
	dockerClient *docker.Client
	error        error
	ready        bool
	shouldExit   bool
}

type InspectDataMsg struct {
	Data  string
	Error error
}

func NewInspectView(svc *service.Service, client *docker.Client, width, height int) *InspectView {
	vp := viewport.New(width-4, height-6)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	return &InspectView{
		service:      svc,
		viewport:     vp,
		dockerClient: client,
		ready:        false,
	}
}

func (i *InspectView) Init() tea.Cmd {
	return i.fetchInspectDataCmd()
}

func (i *InspectView) Update(msg tea.Msg) (*InspectView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return i, tea.Quit
		case "esc":
			i.shouldExit = true
			return i, nil
		}

	case InspectDataMsg:
		if msg.Error != nil {
			i.error = msg.Error
			i.viewport.SetContent("Error fetching inspect data: " + msg.Error.Error())
		} else {
			i.viewport.SetContent(msg.Data)
		}
		i.ready = true
		return i, nil

	case tea.WindowSizeMsg:
		i.viewport.Width = msg.Width - 4
		i.viewport.Height = msg.Height - 6
	}

	i.viewport, cmd = i.viewport.Update(msg)
	return i, cmd
}

func (i *InspectView) View() string {
	if !i.ready {
		return "Loading inspect data..."
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render(fmt.Sprintf("Inspect: %s", i.service.Name))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[esc] back  [↑/↓] scroll  [g/G] top/bottom")

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, i.viewport.View(), footer)
}

func (i *InspectView) fetchInspectDataCmd() tea.Cmd {
	return func() tea.Msg {
		if i.dockerClient == nil {
			return InspectDataMsg{
				Data:  "",
				Error: fmt.Errorf("Docker unavailable"),
			}
		}

		data, err := i.dockerClient.GetInspect(i.service.ContainerID)
		return InspectDataMsg{
			Data:  data,
			Error: err,
		}
	}
}
