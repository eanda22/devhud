package tui

import "github.com/eanda22/devhud/internal/db"

type OperationCompleteMsg struct {
	Success bool
	Message string
}

type LogsFetchedMsg struct {
	Logs  []string
	Error error
}

type TablesFetchedMsg struct {
	Tables []db.TableInfo
	Client *db.Client
	Error  error
}

type TableDataFetchedMsg struct {
	Columns []db.ColumnInfo
	Rows    db.RowData
	Error   error
}
