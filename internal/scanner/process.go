package scanner

import (
	"os/exec"
	"strings"
)

type ProcessScanner struct {
	targets map[string]bool
}

func NewProcessScanner() *ProcessScanner {
	return &ProcessScanner{
		targets: map[string]bool{
			"node":   true,
			"python": true,
			"ruby":   true,
			"java":   true,
			"go":     true,
			"next":   true,
			"vite":   true,
		},
	}
}

type ProcessInfo struct {
	PID     string
	Command string
	User    string
}

func (ps *ProcessScanner) FindProcesses() ([]ProcessInfo, error) {
	var found []ProcessInfo

	cmd := exec.Command("ps", "-eo", "pid,comm,args")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid := fields[0]
		processName := fields[1]
		fullCommand := strings.Join(fields[1:], " ")

		for target := range ps.targets {
			if strings.Contains(processName, target) {
				found = append(found, ProcessInfo{
					PID:     pid,
					Command: fullCommand,
					User:    "",
				})
				break
			}
		}
	}

	return found, nil
}
