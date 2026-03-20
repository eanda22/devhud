package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eanda22/devhud/internal/service"
)

// Parsed represents a parsed command with category, action, and target.
type Parsed struct {
	Category  string
	Action    string
	Target    string
	Raw       string
	VerbFirst bool
}

type verbDef struct {
	filter   func(*service.Service) bool
	action   string
	errMsg   string
	category string
}

const errFiltered = "filtered"

var categoryAliases = map[string]string{
	"c":          "containers",
	"containers": "containers",
	"p":          "processes",
	"processes":  "processes",
	"db":         "containers",
	"databases":  "containers",
}

var verbRegistry map[string]*verbDef

func isDockerOrCompose(svc *service.Service) bool {
	return svc.Type == service.ServiceTypeDocker || svc.Type == service.ServiceTypeCompose
}

func init() {
	stop := &verbDef{
		filter:   func(svc *service.Service) bool { return isDockerOrCompose(svc) && svc.Status == service.StatusRunning },
		action:   "Stop Container",
		errMsg:   "already stopped",
		category: "containers",
	}
	start := &verbDef{
		filter:   func(svc *service.Service) bool { return isDockerOrCompose(svc) && svc.Status == service.StatusStopped },
		action:   "Start Container",
		errMsg:   "already running",
		category: "containers",
	}
	restart := &verbDef{
		filter:   func(svc *service.Service) bool { return isDockerOrCompose(svc) && svc.Status == service.StatusRunning },
		action:   "Restart Container",
		errMsg:   "restart not available",
		category: "containers",
	}
	kill := &verbDef{
		filter: func(svc *service.Service) bool {
			return svc.Type == service.ServiceTypeProcess && svc.Status == service.StatusRunning
		},
		action:   "Kill Process",
		errMsg:   "not a running process",
		category: "processes",
	}
	logs := &verbDef{
		action:   "View Logs",
		category: "containers",
	}
	inspect := &verbDef{
		filter:   func(svc *service.Service) bool { return isDockerOrCompose(svc) },
		action:   "Inspect JSON",
		errMsg:   "inspect not available",
		category: "containers",
	}
	shell := &verbDef{
		filter:   func(svc *service.Service) bool { return isDockerOrCompose(svc) && svc.Status == service.StatusRunning },
		action:   "Open Shell (/bin/sh)",
		errMsg:   "shell not available",
		category: "containers",
	}
	del := &verbDef{
		filter:   func(svc *service.Service) bool { return isDockerOrCompose(svc) },
		action:   "Delete Container",
		errMsg:   "delete not available",
		category: "containers",
	}
	browse := &verbDef{
		filter:   func(svc *service.Service) bool { return svc.DBType != "" },
		action:   "Browse Database",
		errMsg:   "not a database container",
		category: "containers",
	}
	toggle := &verbDef{
		filter: func(svc *service.Service) bool {
			return isDockerOrCompose(svc) || svc.Type == service.ServiceTypeProcess
		},
		errMsg: "no stop/start action for this service",
	}

	verbRegistry = map[string]*verbDef{
		"stop": stop, "start": start, "restart": restart,
		"kill": kill, "logs": logs, "inspect": inspect,
		"shell": shell, "delete": del, "browse": browse,
		"r": restart, "l": logs, "i": inspect,
		"d": del, "b": browse, "s": toggle,
	}
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
	if vd, ok := verbRegistry[verb]; ok {
		p.VerbFirst = true
		p.Action = verb
		p.Category = vd.category
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

// resolveService finds a service by name, optionally filtered by a predicate.
func (a *App) resolveService(name string, filter func(*service.Service) bool) (*service.Service, string) {
	all := a.services.GetAll()
	lower := strings.ToLower(name)

	var nameMatched bool
	for _, svc := range all {
		if strings.ToLower(svc.Name) == lower {
			if filter == nil || filter(svc) {
				return svc, ""
			}
			nameMatched = true
		}
	}

	if !nameMatched {
		var matches []*service.Service
		for _, svc := range all {
			if strings.Contains(strings.ToLower(svc.Name), lower) {
				if filter == nil || filter(svc) {
					matches = append(matches, svc)
				} else {
					nameMatched = true
				}
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
	}

	if nameMatched {
		return nil, errFiltered
	}
	return nil, "not found: " + name
}

func (a *App) executeCommand(p Parsed) tea.Cmd {
	if p.VerbFirst {
		return a.executeVerbCommand(p)
	}
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

func (a *App) executeVerbCommand(p Parsed) tea.Cmd {
	vd := verbRegistry[p.Action]
	if vd == nil {
		a.statusMessage = "unknown command: " + p.Raw
		return nil
	}
	if p.Target == "" {
		a.statusMessage = "usage: " + p.Action + " <name>"
		return nil
	}
	svc, errMsg := a.resolveService(p.Target, vd.filter)
	if svc == nil {
		a.statusMessage = resolveErrorMessage(errMsg, p.Target, vd.errMsg)
		return nil
	}
	if vd.action == "" {
		return a.executeToggle(svc)
	}
	return a.executeActionFromMenu(vd.action, svc)
}

// executeToggle dispatches stop/start/kill based on service state.
func (a *App) executeToggle(svc *service.Service) tea.Cmd {
	if isDockerOrCompose(svc) {
		if svc.Status == service.StatusRunning {
			return a.executeActionFromMenu("Stop Container", svc)
		}
		return a.executeActionFromMenu("Start Container", svc)
	}
	if svc.Type == service.ServiceTypeProcess {
		return a.executeActionFromMenu("Kill Process", svc)
	}
	a.statusMessage = "no stop/start action for this service"
	return nil
}

func resolveErrorMessage(errMsg, target, filterErrMsg string) string {
	if errMsg == errFiltered {
		if filterErrMsg != "" {
			return target + ": " + filterErrMsg
		}
		return "not found: " + target
	}
	return errMsg
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
		var filter func(*service.Service) bool
		var filterErr string
		if vd, ok := verbRegistry[p.Action]; ok {
			filter = vd.filter
			filterErr = vd.errMsg
		}
		svc, errMsg := a.resolveService(p.Target, filter)
		if svc == nil {
			a.statusMessage = resolveErrorMessage(errMsg, p.Target, filterErr)
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
		var filter func(*service.Service) bool
		var filterErr string
		if vd, ok := verbRegistry["kill"]; ok {
			filter = vd.filter
			filterErr = vd.errMsg
		}
		svc, errMsg := a.resolveService(p.Target, filter)
		if svc == nil {
			a.statusMessage = resolveErrorMessage(errMsg, p.Target, filterErr)
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
	trailingSpace := strings.HasSuffix(input, " ")

	verbNames := []string{"stop", "start", "restart", "kill", "logs", "inspect", "shell", "delete", "browse"}
	categories := []string{"containers", "processes"}
	builtins := []string{"help", "quit"}
	topLevel := make([]string, 0, len(verbNames)+len(categories)+len(builtins))
	topLevel = append(topLevel, verbNames...)
	topLevel = append(topLevel, categories...)
	topLevel = append(topLevel, builtins...)

	containerActions := []string{"list", "stop", "start", "restart", "logs", "inspect", "shell", "delete", "browse"}
	processActions := []string{"list", "kill"}

	if len(parts) == 0 {
		return topLevel
	}

	first := strings.ToLower(parts[0])

	if len(parts) == 1 && !trailingSpace {
		return filterPrefix(topLevel, first)
	}

	if cat, ok := categoryAliases[first]; ok {
		if len(parts) == 1 && trailingSpace {
			switch cat {
			case "containers":
				return containerActions
			case "processes":
				return processActions
			}
			return nil
		}
		if len(parts) == 2 && !trailingSpace {
			switch cat {
			case "containers":
				return filterPrefix(containerActions, parts[1])
			case "processes":
				return filterPrefix(processActions, parts[1])
			}
			return nil
		}
		if (len(parts) == 2 && trailingSpace) || len(parts) >= 3 {
			action := strings.ToLower(parts[1])
			var prefix string
			if len(parts) >= 3 {
				prefix = strings.Join(parts[2:], " ")
			}
			if vd, ok := verbRegistry[action]; ok {
				return a.filteredServiceNames(vd.filter, prefix)
			}
			return a.filteredServiceNames(nil, prefix)
		}
		return nil
	}

	if vd, ok := verbRegistry[first]; ok {
		var prefix string
		if len(parts) >= 2 {
			prefix = strings.Join(parts[1:], " ")
		}
		return a.filteredServiceNames(vd.filter, prefix)
	}

	return nil
}

func (a *App) filteredServiceNames(filter func(*service.Service) bool, prefix string) []string {
	var names []string
	for _, svc := range a.services.GetAll() {
		if filter == nil || filter(svc) {
			names = append(names, svc.Name)
		}
	}
	return filterPrefix(names, prefix)
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
