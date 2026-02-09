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
)

func RenderWelcome() string {
	title := titleStyle.Render("devhud")
	content := "Welcome to devhud!\n\nA unified local dev environment manager"
	help := subtleStyle.Render("\nPress 'q' to quit")

	return boxStyle.Render(title + "\n\n" + content + help)
}
