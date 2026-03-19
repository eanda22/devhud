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

	normalModeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	commandModeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FFA500")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 1)

	searchModeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2ECC71")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	confirmDeleteStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#E74C3C")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1)

	commandBarBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FFA500")).
				Padding(0, 1)

	completionItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	completionSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500")).
				Bold(true)

	commandErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E74C3C"))
)
