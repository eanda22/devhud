package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/service"
)

func RenderDashboard(services []*service.Service) string {
	if len(services) == 0 {
		return dashboardStyle.Render("No services discovered. Scanning...")
	}

	header := fmt.Sprintf("%-6s %-20s %-10s %-6s %-10s\n",
		"STATUS", "NAME", "TYPE", "PORT", "UPTIME")

	rows := []string{}
	for _, svc := range services {
		status := statusIcon(svc.Status)
		port := fmt.Sprintf("%d", svc.Port)
		if svc.Port == 0 {
			port = "-"
		}
		uptime := formatUptime(svc.Uptime)

		row := fmt.Sprintf("%-6s %-20s %-10s %-6s %-10s\n",
			status,
			truncate(svc.Name, 18),
			string(svc.Type),
			port,
			uptime,
		)
		rows = append(rows, row)
	}

	footer := "\n[q]uit\n"

	return dashboardStyle.Render(header + strings.Join(rows, "") + footer)
}

func statusIcon(s service.Status) string {
	switch s {
	case service.StatusRunning:
		return "●"
	case service.StatusStopped:
		return "○"
	case service.StatusUnhealthy:
		return "⚠"
	default:
		return "?"
	}
}

func formatUptime(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-1] + "…"
	}
	return s
}

var dashboardStyle = lipgloss.NewStyle().
	Padding(1, 2).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#7D56F4"))
