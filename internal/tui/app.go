package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	// We'll add state fields later
}

func NewApp() *App {
	return &App{}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *App) View() string {
	return RenderWelcome()
}
