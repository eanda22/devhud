package tui

type OperationCompleteMsg struct {
	Success bool
	Message string
}

type LogsFetchedMsg struct {
	Logs  []string
	Error error
}
