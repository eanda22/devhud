package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpView displays a scrollable keyboard reference overlay.
type HelpView struct {
	viewport   viewport.Model
	shouldExit bool
}

// NewHelpView creates a help overlay sized to the terminal.
func NewHelpView(w, h int) *HelpView {
	vp := viewport.New(w-4, h-6)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)
	vp.SetContent(helpContent())

	return &HelpView{
		viewport: vp,
	}
}

func (hv *HelpView) Init() tea.Cmd {
	return nil
}

func (hv *HelpView) Update(msg tea.Msg) (*HelpView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return hv, tea.Quit
		case "esc", "?":
			hv.shouldExit = true
			return hv, nil
		}
	case tea.WindowSizeMsg:
		hv.viewport.Width = msg.Width - 4
		hv.viewport.Height = msg.Height - 6
	}

	var cmd tea.Cmd
	hv.viewport, cmd = hv.viewport.Update(msg)
	return hv, cmd
}

func (hv *HelpView) View() string {
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("devhud — Keyboard Reference")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[↑/↓] scroll  [?/esc] close")

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, hv.viewport.View(), footer)
}

func helpContent() string {
	sections := []struct {
		title string
		keys  [][2]string
	}{
		{
			title: "Navigation",
			keys: [][2]string{
				{"j / ↓", "Move down"},
				{"k / ↑", "Move up"},
				{"h / ←", "Focus sidebar"},
				{"l / → / Enter", "Focus main list"},
				{"1 / 2 / 3", "Jump to Containers / Processes / Databases"},
				{"G", "Jump to last item"},
				{"gg", "Jump to first item"},
				{"Tab", "Toggle detail panel"},
			},
		},
		{
			title: "Actions (main list)",
			keys: [][2]string{
				{"Enter", "Open action menu"},
				{"s", "Start / Stop toggle"},
				{"r", "Restart"},
				{"l", "View logs"},
				{"d", "Delete (with confirm)"},
				{"i", "Inspect JSON"},
			},
		},
		{
			title: "Modes",
			keys: [][2]string{
				{"/", "Enter SEARCH mode"},
				{"?", "Open this help overlay"},
				{"Esc", "Return to NORMAL / clear filter"},
			},
		},
		{
			title: "Global",
			keys: [][2]string{
				{"q / Ctrl+C", "Quit"},
			},
		},
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(30)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	var lines []string
	for _, section := range sections {
		lines = append(lines, sectionStyle.Render(section.title))
		for _, kv := range section.keys {
			lines = append(lines, "  "+keyStyle.Render(kv[0])+descStyle.Render(kv[1]))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}
