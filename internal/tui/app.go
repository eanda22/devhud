package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eanda22/devhud/internal/scanner"
	"github.com/eanda22/devhud/internal/service"
)

type ScanCompleteMsg struct {
	Services []*service.Service
	Error    error
}

type TickMsg struct{}

type App struct {
	services *service.Store
	scanner  *scanner.Scanner
	ticker   *time.Ticker
}

func NewApp() (*App, error) {
	store := service.NewStore()
	scan, err := scanner.NewScanner(store)
	if err != nil {
		return nil, fmt.Errorf("scanner: %w", err)
	}

	return &App{
		services: store,
		scanner:  scan,
	}, nil
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.scanCmd(),
		a.tickCmd(),
	)
}

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

func (a *App) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			a.scanner.Close()
			return a, tea.Quit
		}

	case ScanCompleteMsg:
		if msg.Error != nil {
			fmt.Printf("Scan error: %v\n", msg.Error)
		}
		return a, nil

	case TickMsg:
		return a, a.scanCmd()
	}

	return a, nil
}

func (a *App) View() string {
	return RenderDashboard(a.services.GetAll())
}
