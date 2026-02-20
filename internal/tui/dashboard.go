package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/docker"
	"github.com/eanda22/devhud/internal/service"
)

const Version = "0.2.0"

func renderHeader(category string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render(fmt.Sprintf("DEVHUD v%s | %s", Version, category))
}

// renders the main service list and detail panel.
func RenderDashboard(a *App) string {
	services := a.getFilteredServices()

	if len(services) == 0 {
		msg := "No services discovered. Scanning..."
		if a.lastError != nil {
			msg += fmt.Sprintf("\nLast error: %v", a.lastError)
		}
		return dashboardStyle.Render(msg)
	}

	sidebar := renderSidebar(a)

	selectedCategory := ""
	if a.activeCatIndex < len(a.categories) {
		selectedCategory = a.categories[a.activeCatIndex]
	}
	header := renderHeader(selectedCategory)

	servicesList, detailPanel := renderServiceList(
		services,
		a.activeCatIndex,
		a.selectedIndex,
		a.focus,
		a.confirmOperation,
		a.operatingOnID,
		a.dockerDiskUsage,
		a.statusMessage,
	)

	mainContent := dashboardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, servicesList))

	if detailPanel != "" {
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, mainContent, detailPanel)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
}

// renders the service table with header and footer.
func renderServiceList(
	services []*service.Service,
	activeCatIndex int,
	selectedIndex int,
	focus Focus,
	confirmOperation string,
	operatingOnID string,
	dockerDiskUsage *docker.DiskUsage,
	statusMessage string,
) (string, string) {
	// Column header: DISK for Containers(0) and Databases(2), PORT for Local Procs(1)
	diskOrPortHeader := "DISK"
	if activeCatIndex == 1 {
		diskOrPortHeader = "PORT"
	}

	header := fmt.Sprintf("%-6s %-40s %-10s %-10s %-10s\n",
		"STATUS", "NAME", "TYPE", diskOrPortHeader, "UPTIME")

	var detailPanel string
	rows := []string{}

	for i, svc := range services {
		status := statusIcon(svc.Status)
		uptime := formatUptime(svc.Uptime)

		serviceName := svc.Name
		if svc.DBType != "" {
			serviceName += " [DB]"
		}

		var diskColumn string
		if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
			if dockerDiskUsage != nil && dockerDiskUsage.ContainerSizes != nil {
				if size, ok := dockerDiskUsage.ContainerSizes[svc.ContainerID]; ok {
					diskColumn = formatBytes(size)
				} else {
					diskColumn = "-"
				}
			} else {
				diskColumn = "-"
			}
		} else {
			port := fmt.Sprintf("%d", svc.Port)
			if svc.Port == 0 {
				port = "-"
			}
			diskColumn = port
		}

		row := fmt.Sprintf("%-6s %-40s %-10s %-10s %-10s",
			status,
			truncate(serviceName, 38),
			string(svc.Type),
			diskColumn,
			uptime,
		)

		if svc.ID == operatingOnID && operatingOnID != "" {
			row = operatingRowStyle.Render(row)
			if i == selectedIndex && focus == FocusMainList {
				detailPanel = renderDetailPanel(svc)
			}
		} else if i == selectedIndex && focus == FocusMainList {
			row = selectedRowStyle.Render(row)
			detailPanel = renderDetailPanel(svc)
		}

		rows = append(rows, row+"\n")
	}

	footer := "\n"
	if confirmOperation != "" {
		footer += "⚠ Confirm delete? [y/N]\n"
	} else {
		if focus == FocusSidebar {
			footer += "[↑/↓] Navigate   [→] Select   [q] Quit\n"
		} else {
			footer += "[↑/↓] Navigate   [←] Back   [Enter] Actions   [q] Quit\n"
		}
		if statusMessage != "" {
			footer += fmt.Sprintf("%s\n", statusMessage)
		}
	}

	servicesList := header + strings.Join(rows, "") + footer

	return servicesList, detailPanel
}

func renderDetailPanel(svc *service.Service) string {
	var serviceInfo string

	switch svc.Type {
	case service.ServiceTypeDocker, service.ServiceTypeCompose:
		serviceInfo = fmt.Sprintf(
			"Name: %-20s\nType: %-20s\nStatus: %-20s\nContainer ID: %-20s\nImage: %-20s\nUptime: %-20s\n",
			svc.Name,
			svc.Type,
			svc.Status,
			truncate(svc.ContainerID, 20),
			truncate(svc.Image, 20),
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

// renders the sidebar with categories.
func renderSidebar(a *App) string {
	var items []string

	for i, cat := range a.categories {
		var line string
		if i == a.activeCatIndex {
			line = activeMenuItemStyle.Render(cat)
		} else {
			line = inactiveMenuItemStyle.Render(cat)
		}
		items = append(items, line)
	}

	content := strings.Join(items, "\n")

	selectedCategory := ""
	if a.activeCatIndex < len(a.categories) {
		selectedCategory = a.categories[a.activeCatIndex]
	}

	if selectedCategory == "Containers" && a.dockerDiskUsage != nil {
		diskInfo := fmt.Sprintf("\n\nDisk: %s", formatBytes(a.dockerDiskUsage.Total))
		content += subtleStyle.Render(diskInfo)
	}

	style := sidebarStyle
	if a.focus == FocusSidebar {
		style = style.Copy().BorderForeground(lipgloss.Color("#7D56F4"))
	}

	return style.Render(content)
}

// converts bytes to human-readable format.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

var dashboardStyle = lipgloss.NewStyle().
	Padding(1, 2).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#7D56F4"))
