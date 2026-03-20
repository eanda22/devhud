package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommandBar provides a `:` command input with history and tab completion.
const maxVisibleCompletions = 5

type CommandBar struct {
	input       textinput.Model
	history     []string
	histIndex   int
	active      bool
	candidates  []string
	selectedIdx int
	errMsg      string
	cycling     bool
}

func newCommandBar() *CommandBar {
	ti := textinput.New()
	ti.Prompt = ":"
	ti.CharLimit = 128
	return &CommandBar{
		input:       ti,
		histIndex:   -1,
		selectedIdx: -1,
	}
}

func (cb *CommandBar) Focus() tea.Cmd {
	cb.active = true
	cb.histIndex = -1
	cb.selectedIdx = -1
	cb.errMsg = ""
	cb.input.SetValue("")
	cb.input.Focus()
	return cb.input.Cursor.BlinkCmd()
}

func (cb *CommandBar) Blur() {
	cb.active = false
	cb.input.Blur()
}

func (cb *CommandBar) Value() string {
	return cb.input.Value()
}

func (cb *CommandBar) View() string {
	return cb.input.View()
}

// BoxView renders the completion popup above the bar and error/hint below.
func (cb *CommandBar) BoxView(width int) string {
	var popup string
	if len(cb.candidates) > 0 {
		visible := cb.candidates
		var extra int
		if len(visible) > maxVisibleCompletions {
			extra = len(visible) - maxVisibleCompletions
			visible = visible[:maxVisibleCompletions]
		}
		var lines []string
		for i, c := range visible {
			if i == cb.selectedIdx {
				lines = append(lines, completionSelectedStyle.Render("  > "+c))
			} else {
				lines = append(lines, completionItemStyle.Render("    "+c))
			}
		}
		if extra > 0 {
			lines = append(lines, completionItemStyle.Render(fmt.Sprintf("    +%d more", extra)))
		}
		popup = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	bar := commandBarBoxStyle.Copy().Width(width).Render(cb.input.View())

	var top string
	if cb.errMsg != "" {
		top = commandErrorStyle.Render("  error: "+cb.errMsg) + "\n"
	}

	var bottom string
	if len(cb.candidates) > 0 {
		bottom = "\n" + popup
	} else {
		bottom = "\n" + subtleStyle.Render("  [Tab] Complete  [↑/↓] History  [Enter] Run  [Esc] Cancel")
	}

	return "\n" + top + bar + bottom
}

func (cb *CommandBar) applyCompletion(addSpace bool) {
	if cb.selectedIdx < 0 || cb.selectedIdx >= len(cb.candidates) {
		return
	}
	parts := strings.Fields(cb.input.Value())
	if len(parts) > 0 {
		parts = parts[:len(parts)-1]
	}
	parts = append(parts, cb.candidates[cb.selectedIdx])
	val := strings.Join(parts, " ")
	if addSpace {
		val += " "
	}
	cb.input.SetValue(val)
	cb.input.CursorEnd()
}

// SetCompletions provides pre-computed tab completion candidates.
func (cb *CommandBar) SetCompletions(candidates []string) {
	if cb.cycling {
		return
	}
	cb.candidates = candidates
	cb.selectedIdx = -1
}

// SetError stores an error message for inline display.
func (cb *CommandBar) SetError(msg string) {
	cb.errMsg = msg
}

// ClearError removes any displayed error message.
func (cb *CommandBar) ClearError() {
	cb.errMsg = ""
}

// CompletionLines returns the number of lines the completion popup occupies.
func (cb *CommandBar) CompletionLines() int {
	n := len(cb.candidates)
	if n == 0 {
		return 0
	}
	lines := n
	if lines > maxVisibleCompletions {
		lines = maxVisibleCompletions + 1
	}
	return lines
}

// Update handles key input. Returns (cmd, shouldExecute, shouldCancel).
func (cb *CommandBar) Update(msg tea.Msg) (tea.Cmd, bool, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		cb.input, cmd = cb.input.Update(msg)
		return cmd, false, false
	}

	switch keyMsg.String() {
	case "enter":
		val := strings.TrimSpace(cb.input.Value())
		if val != "" && (len(cb.history) == 0 || cb.history[len(cb.history)-1] != val) {
			cb.history = append(cb.history, val)
		}
		return nil, true, false

	case "esc":
		return nil, false, true

	case "up":
		if len(cb.candidates) > 0 {
			cb.cycling = true
			cb.selectedIdx--
			if cb.selectedIdx < 0 {
				cb.selectedIdx = len(cb.candidates) - 1
			}
			return nil, false, false
		}
		if len(cb.history) == 0 {
			return nil, false, false
		}
		if cb.histIndex == -1 {
			cb.histIndex = len(cb.history) - 1
		} else if cb.histIndex > 0 {
			cb.histIndex--
		}
		cb.input.SetValue(cb.history[cb.histIndex])
		cb.input.CursorEnd()
		return nil, false, false

	case "down":
		if len(cb.candidates) > 0 {
			cb.cycling = true
			cb.selectedIdx++
			if cb.selectedIdx >= len(cb.candidates) {
				cb.selectedIdx = 0
			}
			return nil, false, false
		}
		if cb.histIndex == -1 {
			return nil, false, false
		}
		if cb.histIndex < len(cb.history)-1 {
			cb.histIndex++
			cb.input.SetValue(cb.history[cb.histIndex])
		} else {
			cb.histIndex = -1
			cb.input.SetValue("")
		}
		cb.input.CursorEnd()
		return nil, false, false

	case "tab":
		if len(cb.candidates) == 0 {
			return nil, false, false
		}
		if len(cb.candidates) == 1 {
			cb.selectedIdx = 0
			cb.cycling = false
			cb.applyCompletion(true)
			return nil, false, false
		}
		cb.cycling = true
		cb.selectedIdx++
		if cb.selectedIdx >= len(cb.candidates) {
			cb.selectedIdx = 0
		}
		cb.applyCompletion(false)
		return nil, false, false

	case "shift+tab":
		if len(cb.candidates) == 0 {
			return nil, false, false
		}
		if len(cb.candidates) == 1 {
			cb.selectedIdx = 0
			cb.cycling = false
			cb.applyCompletion(true)
			return nil, false, false
		}
		cb.cycling = true
		cb.selectedIdx--
		if cb.selectedIdx < 0 {
			cb.selectedIdx = len(cb.candidates) - 1
		}
		cb.applyCompletion(false)
		return nil, false, false
	}

	cb.cycling = false
	var cmd tea.Cmd
	cb.input, cmd = cb.input.Update(msg)
	return cmd, false, false
}
