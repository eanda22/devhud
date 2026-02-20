package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/docker"
	"github.com/eanda22/devhud/internal/service"
)

type ActionMenuView struct {
	service       *service.Service
	dockerClient  *docker.Client
	actions       []string
	selectedIndex int
	shouldExit    bool
	executeAction string
	width         int
	height        int
}

func NewActionMenuView(svc *service.Service, dockerClient *docker.Client, w, h int) *ActionMenuView {
	return &ActionMenuView{
		service:       svc,
		dockerClient:  dockerClient,
		actions:       getActionMenuItems(svc),
		selectedIndex: 0,
		shouldExit:    false,
		executeAction: "",
		width:         w,
		height:        h,
	}
}

func (a *ActionMenuView) Init() tea.Cmd {
	return nil
}

func (a *ActionMenuView) Update(msg tea.Msg) (*ActionMenuView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "esc":
			a.shouldExit = true
			return a, nil
		case "up", "k":
			if a.selectedIndex > 0 {
				a.selectedIndex--
			}
		case "down", "j":
			if a.selectedIndex < len(a.actions)-1 {
				a.selectedIndex++
			}
		case "enter":
			if a.selectedIndex < len(a.actions) {
				a.executeAction = a.actions[a.selectedIndex]
				a.shouldExit = true
			}
			return a, nil
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	return a, nil
}

func (a *ActionMenuView) View() string {
	menuItems := make([]string, len(a.actions))
	for i, action := range a.actions {
		if i == a.selectedIndex {
			menuItems[i] = selectedActionStyle.Render("> " + action)
		} else {
			menuItems[i] = unselectedActionStyle.Render("  " + action)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("WHAT DO YOU WANT TO DO?")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[↑/↓] Select   [Enter] Execute   [Esc] Cancel")

	box := actionMenuBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, "", content, "", footer),
	)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, box)
}

func getActionMenuItems(svc *service.Service) []string {
	var items []string

	if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
		if svc.Status == "running" {
			items = append(items, "View Logs")
			if svc.DBType != "" {
				items = append(items, "Browse Database")
			}
			items = append(items, "Restart Container")
			items = append(items, "Stop Container")
			items = append(items, "Inspect JSON")
			items = append(items, "Open Shell (/bin/sh)")
			items = append(items, "Delete Container")
		} else {
			items = append(items, "Start Container")
			items = append(items, "Inspect JSON")
			items = append(items, "Delete Container")
		}
	} else if svc.Type == service.ServiceTypeProcess {
		items = append(items, "View Logs")
		items = append(items, "Kill Process")
	}

	return items
}
