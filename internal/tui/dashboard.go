package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eanda22/devhud/internal/docker"
	"github.com/eanda22/devhud/internal/service"
)

const (
	Version      = "0.2.0"
	sidebarWidth = 25
	detailWidth  = 30
)

func renderHeader(category string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render(fmt.Sprintf("DEVHUD v%s | %s", Version, category))
}

func RenderDashboard(a *App) string {
	services := a.getFilteredServices()
	panelHeight := a.height - 2

	if len(services) == 0 {
		msg := "No services discovered. Scanning..."
		if a.lastError != nil {
			msg += fmt.Sprintf("\nLast error: %v", a.lastError)
		}
		style := dashboardStyle.Copy().Height(panelHeight).Width(a.width - sidebarWidth - 6)
		return lipgloss.JoinHorizontal(lipgloss.Top, renderSidebar(a, panelHeight), style.Render(msg))
	}

	sidebar := renderSidebar(a, panelHeight)

	selectedCategory := ""
	if a.searchFilter != "" {
		selectedCategory = fmt.Sprintf("Search: %s", a.searchFilter)
	} else if a.activeCatIndex < len(a.categories) {
		selectedCategory = a.categories[a.activeCatIndex]
	}

	mainWidth := a.width - sidebarWidth - 6
	if a.showDetailPanel {
		mainWidth = a.width - sidebarWidth - detailWidth - 10
	}

	mainContent := renderMainPanel(a, services, selectedCategory, mainWidth, panelHeight)

	if a.showDetailPanel && a.selectedIndex < len(services) {
		detail := renderDetailPanel(services[a.selectedIndex], panelHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent, detail)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
}

func renderMainPanel(a *App, services []*service.Service, category string, width, height int) string {
	header := renderHeader(category)
	rows := buildServiceRows(services, a.activeCatIndex, a.selectedIndex, a.focus, a.operatingOnID, a.dockerDiskUsage)

	var footer string
	var footerLines int
	if a.mode == "action_menu" && a.actionMenuView != nil {
		footer = renderInlineActionMenu(a.actionMenuView)
		footerLines = len(a.actionMenuView.actions) + 9
	} else {
		footer = buildStatusLine(a)
		footerLines = 1
	}

	contentHeight := height - 4
	maxRows := contentHeight - footerLines - 1

	start, end := visibleWindow(len(rows), a.selectedIndex, maxRows)
	visibleRows := buildVisibleRows(rows, start, end)

	usedLines := 1 + len(visibleRows) + footerLines
	padLines := contentHeight - usedLines
	if padLines < 0 {
		padLines = 0
	}
	padding := strings.Repeat("\n", padLines)

	content := header + "\n" + strings.Join(visibleRows, "") + padding + footer
	style := dashboardStyle.Copy().Width(width).Height(height)
	return style.Render(content)
}

func buildVisibleRows(rows []string, start, end int) []string {
	var visible []string
	if start > 0 {
		visible = append(visible, subtleStyle.Render(fmt.Sprintf("  ↑ %d more", start))+"\n")
	}
	visible = append(visible, rows[start:end]...)
	if end < len(rows) {
		visible = append(visible, subtleStyle.Render(fmt.Sprintf("  ↓ %d more", len(rows)-end))+"\n")
	}
	return visible
}

func renderInlineActionMenu(a *ActionMenuView) string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("Actions: " + a.service.Name)

	var items []string
	for i, action := range a.actions {
		if i == a.selectedIndex {
			items = append(items, selectedActionStyle.Render("> "+action))
		} else {
			items = append(items, unselectedActionStyle.Render("  "+action))
		}
	}

	hint := subtleStyle.Render("[↑/↓] Select   [Enter] Execute   [Esc] Cancel")

	box := actionMenuBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, "", strings.Join(items, "\n"), "", hint),
	)

	return "\n" + box
}

func buildServiceRows(
	services []*service.Service,
	activeCatIndex int,
	selectedIndex int,
	focus Focus,
	operatingOnID string,
	dockerDiskUsage *docker.DiskUsage,
) []string {
	diskOrPortHeader := "DISK"
	if activeCatIndex == 1 {
		diskOrPortHeader = "PORT"
	}

	headerLine := fmt.Sprintf("%-6s %-40s %-10s %-10s %-10s\n",
		"STATUS", "NAME", "TYPE", diskOrPortHeader, "UPTIME")

	rows := []string{headerLine}

	for i, svc := range services {
		row := formatServiceRow(svc, activeCatIndex, dockerDiskUsage)

		if svc.ID == operatingOnID && operatingOnID != "" {
			row = operatingRowStyle.Render(row)
		} else if i == selectedIndex && focus == FocusMainList {
			row = selectedRowStyle.Render(row)
		}

		rows = append(rows, row+"\n")
	}

	return rows
}

func formatServiceRow(svc *service.Service, activeCatIndex int, dockerDiskUsage *docker.DiskUsage) string {
	status := statusIcon(svc.Status)
	uptime := formatUptime(svc.Uptime)

	serviceName := svc.Name
	if svc.DBType != "" {
		serviceName += " [DB]"
	}

	diskColumn := getDiskColumn(svc, dockerDiskUsage)

	return fmt.Sprintf("%-6s %-40s %-10s %-10s %-10s",
		status,
		truncate(serviceName, 38),
		string(svc.Type),
		diskColumn,
		uptime,
	)
}

func getDiskColumn(svc *service.Service, dockerDiskUsage *docker.DiskUsage) string {
	if svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose {
		if dockerDiskUsage != nil && dockerDiskUsage.ContainerSizes != nil {
			if size, ok := dockerDiskUsage.ContainerSizes[svc.ContainerID]; ok {
				return formatBytes(size)
			}
		}
		return "-"
	}
	if svc.Port == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", svc.Port)
}

func buildStatusLine(a *App) string {
	if a.confirmOperation != "" {
		return "\n" + confirmDeleteStyle.Render(" DELETE ") + "  Confirm delete? [y/N]"
	}

	var mode, hints string

	switch a.inputMode {
	case ModeCommand:
		mode = commandModeStyle.Render(" COMMAND ")
		hints = "Type command  [Esc] Cancel"
	case ModeSearch:
		mode = searchModeStyle.Render(" SEARCH ")
		hints = a.searchInput.View() + "  [Enter] Lock  [Esc] Cancel"
	default:
		mode = normalModeStyle.Render(" NORMAL ")
		if a.searchFilter != "" {
			mode += "  " + subtleStyle.Render("filter: "+a.searchFilter+" [/ edit, Esc clear]")
		}
		if a.focus == FocusSidebar {
			hints = "[j/k] Nav  [l] Select  [Tab] Details  [/] Search  [:] Cmd"
		} else {
			hints = "[j/k] Nav  [s]top  [r]estart  [l]ogs  [d]el  [i]nspect  [1-3] Cat  [:] Cmd"
		}
	}

	line := "\n" + mode + "  " + hints
	if a.statusMessage != "" {
		line += "  " + subtleStyle.Render(a.statusMessage)
	}

	return line
}

func visibleWindow(totalItems, selectedIndex, maxVisible int) (int, int) {
	if totalItems <= maxVisible {
		return 0, totalItems
	}

	half := maxVisible / 2
	start := selectedIndex - half
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > totalItems {
		end = totalItems
		start = end - maxVisible
	}
	return start, end
}

func renderDetailPanel(svc *service.Service, height int) string {
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

	style := dashboardStyle.Copy().Width(detailWidth).Height(height)
	return style.Render(serviceInfo)
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

func renderSidebar(a *App, panelHeight int) string {
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

	var diskInfo string
	selectedCategory := ""
	if a.activeCatIndex < len(a.categories) {
		selectedCategory = a.categories[a.activeCatIndex]
	}
	if selectedCategory == "Containers" && a.dockerDiskUsage != nil {
		diskInfo = subtleStyle.Render(fmt.Sprintf("Disk: %s", formatBytes(a.dockerDiskUsage.Total)))
	}

	if diskInfo != "" {
		contentLines := strings.Count(content, "\n") + 1
		diskLines := 1
		innerHeight := panelHeight - 2
		padCount := innerHeight - contentLines - diskLines
		if padCount < 2 {
			padCount = 2
		}
		content += strings.Repeat("\n", padCount) + diskInfo
	}

	style := sidebarStyle.Copy().Height(panelHeight)
	if a.focus == FocusSidebar {
		style = style.BorderForeground(lipgloss.Color("#7D56F4"))
	}

	return style.Render(content)
}

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
