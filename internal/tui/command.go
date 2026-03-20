package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eanda22/devhud/internal/service"
)

// Parsed represents a parsed command with category, action, and target.
type Parsed struct {
	Category string
	Action   string
	Target   string
	Raw      string
}

var categoryAliases = map[string]string{
	"c":          "containers",
	"containers": "containers",
	"p":          "processes",
	"processes":  "processes",
	"db":         "containers",
	"databases":  "containers",
}

var shortForms = map[string]string{
	"stop":    "containers",
	"start":   "containers",
	"restart": "containers",
	"logs":    "containers",
	"inspect": "containers",
	"kill":    "processes",
	"browse":  "containers",
}

func parseCommand(input string) Parsed {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return Parsed{Raw: input}
	}

	p := Parsed{Raw: input}

	if cat, ok := categoryAliases[strings.ToLower(parts[0])]; ok {
		p.Category = cat
		if len(parts) > 1 {
			p.Action = strings.ToLower(parts[1])
		}
		if len(parts) > 2 {
			p.Target = strings.Join(parts[2:], " ")
		}
		return p
	}

	verb := strings.ToLower(parts[0])
	if cat, ok := shortForms[verb]; ok {
		p.Category = cat
		p.Action = verb
		if len(parts) > 1 {
			p.Target = strings.Join(parts[1:], " ")
		}
		return p
	}

	p.Action = parts[0]
	if len(parts) > 1 {
		p.Target = strings.Join(parts[1:], " ")
	}
	return p
}

func (a *App) resolveService(name string) (*service.Service, string) {
	all := a.services.GetAll()
	lower := strings.ToLower(name)

	for _, svc := range all {
		if strings.ToLower(svc.Name) == lower {
			return svc, ""
		}
	}

	var matches []*service.Service
	for _, svc := range all {
		if strings.Contains(strings.ToLower(svc.Name), lower) {
			matches = append(matches, svc)
		}
	}

	if len(matches) == 1 {
		return matches[0], ""
	}
	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return nil, "ambiguous: " + strings.Join(names, ", ")
	}
	return nil, "not found: " + name
}

func (a *App) executeCommand(p Parsed) tea.Cmd {
	switch p.Category {
	case "containers":
		return a.executeContainerCommand(p)
	case "processes":
		return a.executeProcessCommand(p)
	default:
		switch strings.ToLower(p.Action) {
		case "help":
			a.helpView = NewHelpView(a.width, a.height)
			a.mode = "help"
			return a.helpView.Init()
		case "quit", "q":
			return tea.Quit
		default:
			a.statusMessage = "unknown command: " + p.Raw
		}
		return nil
	}
}

func (a *App) executeContainerCommand(p Parsed) tea.Cmd {
	switch p.Action {
	case "list":
		a.activeCatIndex = 0
		a.selectedIndex = 0
		a.focus = FocusMainList
		return a.fetchDiskUsageCmd()
	case "stop", "start", "restart", "logs", "inspect", "shell", "delete", "browse":
		if p.Target == "" {
			a.statusMessage = "usage: containers " + p.Action + " <name>"
			return nil
		}
		svc, errMsg := a.resolveService(p.Target)
		if svc == nil {
			a.statusMessage = errMsg
			return nil
		}
		actionMap := map[string]string{
			"stop":    "Stop Container",
			"start":   "Start Container",
			"restart": "Restart Container",
			"logs":    "View Logs",
			"inspect": "Inspect JSON",
			"shell":   "Open Shell (/bin/sh)",
			"delete":  "Delete Container",
			"browse":  "Browse Database",
		}
		return a.executeActionFromMenu(actionMap[p.Action], svc)
	default:
		a.statusMessage = "unknown containers action: " + p.Action
		return nil
	}
}

func (a *App) executeProcessCommand(p Parsed) tea.Cmd {
	switch p.Action {
	case "list":
		a.activeCatIndex = 1
		a.selectedIndex = 0
		a.focus = FocusMainList
		return nil
	case "kill":
		if p.Target == "" {
			a.statusMessage = "usage: processes kill <name>"
			return nil
		}
		svc, errMsg := a.resolveService(p.Target)
		if svc == nil {
			a.statusMessage = errMsg
			return nil
		}
		return a.executeActionFromMenu("Kill Process", svc)
	default:
		a.statusMessage = "unknown processes action: " + p.Action
		return nil
	}
}

func (a *App) completions(input string) []string {
	parts := strings.Fields(input)

	categories := []string{"containers", "processes", "help", "quit"}
	containerActions := []string{"list", "stop", "start", "restart", "logs", "inspect", "shell", "delete", "browse"}
	processActions := []string{"list", "kill"}

	if len(parts) == 0 {
		return categories
	}

	if len(parts) == 1 && !strings.HasSuffix(input, " ") {
		return filterPrefix(categories, parts[0])
	}

	cat := categoryAliases[strings.ToLower(parts[0])]
	if len(parts) == 2 && !strings.HasSuffix(input, " ") {
		switch cat {
		case "containers":
			return filterPrefix(containerActions, parts[1])
		case "processes":
			return filterPrefix(processActions, parts[1])
		}
		return nil
	}

	if len(parts) >= 2 && strings.HasSuffix(input, " ") || len(parts) >= 3 {
		var prefix string
		if len(parts) >= 3 {
			prefix = strings.Join(parts[2:], " ")
		}
		var names []string
		for _, svc := range a.services.GetAll() {
			names = append(names, svc.Name)
		}
		return filterPrefix(names, prefix)
	}

	return nil
}

func filterPrefix(candidates []string, prefix string) []string {
	prefix = strings.ToLower(prefix)
	var result []string
	for _, c := range candidates {
		if strings.HasPrefix(strings.ToLower(c), prefix) {
			result = append(result, c)
		}
	}
	return result
}
