package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/service"
)

// renders the main service list and detail panel.
func RenderDashboard(services []*service.Service, serviceIndex int, lastError error) string {
	if len(services) == 0 {
		msg := "No services discovered. Scanning..."
		if lastError != nil {
			msg += fmt.Sprintf("\nLast error: %v", lastError)
		}
		return dashboardStyle.Render(msg)
	}

	servicesList, detailPanel := renderServiceList(services, serviceIndex)

	if detailPanel != "" {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			dashboardStyle.Render(servicesList),
			detailPanel,
		)
	}

	return dashboardStyle.Render(servicesList)
}

// renders the service table with header and footer.
func renderServiceList(services []*service.Service, serviceIndex int) (string, string) {
	header := fmt.Sprintf("%-6s %-40s %-10s %-6s %-10s\n",
		"STATUS", "NAME", "TYPE", "PORT", "UPTIME")

	var detailPanel string
	rows := []string{}

	for i, svc := range services {
		status := statusIcon(svc.Status)
		port := fmt.Sprintf("%d", svc.Port)
		if svc.Port == 0 {
			port = "-"
		}
		uptime := formatUptime(svc.Uptime)

		row := fmt.Sprintf("%-6s %-40s %-10s %-6s %-10s",
			status,
			truncate(svc.Name, 18),
			string(svc.Type),
			port,
			uptime,
		)

		var selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

		if i == serviceIndex {
			row = selectedStyle.Render(row)
			detailPanel = renderDetailPanel(svc)
		}

		rows = append(rows, row+"\n")
	}

	footer := "\n[↑/↓] navigate  [q]uit\n"
	servicesList := header + strings.Join(rows, "") + footer

	return servicesList, detailPanel
}

// renders detailed information for the selected service.
func renderDetailPanel(svc *service.Service) string {
	serviceInfo := fmt.Sprintf(
		"Name: %-20s\nType: %-20s\nStatus: %-20s\nContainer ID: %-20s\nImage: %-20s\nProject: %-20s\n",
		svc.Name,
		svc.Type,
		svc.Status,
		svc.ContainerID,
		svc.Image,
		svc.Project,
	)

	return dashboardStyle.Render(serviceInfo)
}

// returns the appropriate icon for a service status.
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

// converts duration to human-readable uptime string.
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

// shortens strings exceeding max length with ellipsis.
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
