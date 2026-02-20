package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginTop(1).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	sidebarStyle = lipgloss.NewStyle().
			Width(25).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("241")). // Default gray
			Padding(0, 1).
			MarginRight(1)

	focusedBorderStyle = lipgloss.NewStyle().
				BorderForeground(lipgloss.Color("#7D56F4"))

	activeMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				PaddingLeft(1).
				Border(lipgloss.NormalBorder(), false, false, false, true). // Left border only
				BorderForeground(lipgloss.Color("#7D56F4"))

	inactiveMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				PaddingLeft(2) // Indent to match the border of active item

	actionMenuBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(2, 4)

	selectedActionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	unselectedActionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	operatingRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FFA500")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true)
)

// renders the welcome screen.
func RenderWelcome() string {
	title := titleStyle.Render("devhud")
	content := "Welcome to devhud!\n\nA unified local dev environment manager"
	help := subtleStyle.Render("\nPress 'q' to quit")

	return boxStyle.Render(title + "\n\n" + content + help)
}
