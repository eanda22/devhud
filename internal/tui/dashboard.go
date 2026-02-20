package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/service"
)

// renders the main service list and detail panel.
func RenderDashboard(services []*service.Service, serviceIndex int, lastError error, statusMessage string, confirmOperation string, operatingOnID string) string {
	if len(services) == 0 {
		msg := "No services discovered. Scanning..."
		if lastError != nil {
			msg += fmt.Sprintf("\nLast error: %v", lastError)
		}
		return dashboardStyle.Render(msg)
	}

	servicesList, detailPanel := renderServiceList(services, serviceIndex, statusMessage, confirmOperation, operatingOnID)

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
func renderServiceList(services []*service.Service, serviceIndex int, statusMessage string, confirmOperation string, operatingOnID string) (string, string) {
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

		var operatingStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FFA500")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)

		if svc.ID == operatingOnID && operatingOnID != "" {
			row = operatingStyle.Render(row)
			if i == serviceIndex {
				detailPanel = renderDetailPanel(svc)
			}
		} else if i == serviceIndex {
			row = selectedStyle.Render(row)
			detailPanel = renderDetailPanel(svc)
		}

		rows = append(rows, row+"\n")
	}

	footer := "\n"
	if confirmOperation != "" {
		footer += "⚠ Confirm delete? [y/N]\n"
	} else {
		footer += "[↑/↓] navigate [q]uit\n"
		if statusMessage != "" {
			footer += fmt.Sprintf("%s\n", statusMessage)
		}
	}

	servicesList := header + strings.Join(rows, "") + footer

	return servicesList, detailPanel
}

// renders detailed information for the selected service.
func renderDetailPanel(svc *service.Service) string {
	var serviceInfo string

	switch svc.Type {
	case service.ServiceTypeDocker, service.ServiceTypeCompose:
		serviceInfo = fmt.Sprintf(
			"Name: %-20s\nType: %-20s\nStatus: %-20s\nContainer ID: %-20s\nImage: %-20s\nUptime: %-20s\n",
			svc.Name,
			svc.Type,
			svc.Status,
			svc.ContainerID,
			svc.Image,
			formatUptime(svc.Uptime),
		)
	case service.ServiceTypeProcess:
		pid := fmt.Sprintf("%d", svc.PID)
		if svc.PID == 0 {
			pid = "-"
		}
		port := fmt.Sprintf("%d", svc.Port)
		if svc.Port == 0 {
			port = "-"
		}
		serviceInfo = fmt.Sprintf(
			"Name: %-20s\nType: %-20s\nStatus: %-20s\nPID: %-20s\nPort: %-20s\nUptime: %-20s\n",
			svc.Name,
			svc.Type,
			svc.Status,
			pid,
			port,
			formatUptime(svc.Uptime),
		)
	default:
		serviceInfo = fmt.Sprintf(
			"Name: %-20s\nType: %-20s\nStatus: %-20s\nUptime: %-20s\n",
			svc.Name,
			svc.Type,
			svc.Status,
			formatUptime(svc.Uptime),
		)
	}

	actions := getServiceActions(svc)
	if actions != "" {
		serviceInfo += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Actions: ") + actions + "\n"
	}

	return dashboardStyle.Render(serviceInfo)
}

// returns available action keys for a service type.
func getServiceActions(svc *service.Service) string {
	switch svc.Type {
	case service.ServiceTypeDocker, service.ServiceTypeCompose:
		return "[s]tart [x]stop [r]estart [d]elete"
	case service.ServiceTypeProcess:
		return "[x]kill"
	default:
		return ""
	}
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
