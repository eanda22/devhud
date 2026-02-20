package db

import (
	"context"
	"fmt"
)

// ColumnInfo holds metadata about a table column.
type ColumnInfo struct {
	Name string
	Type string
}

// RowData represents a slice of rows, where each row is a map of column names to values.
type RowData []map[string]interface{}

// GetTableColumns returns column information for a table.
func (c *Client) GetTableColumns(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	var query string
	switch c.dbType {
	case "postgres":
		query = `
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = $1
			ORDER BY ordinal_position
		`
	case "mysql":
		query = `
			SELECT COLUMN_NAME, DATA_TYPE
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION
		`
	default:
		return nil, fmt.Errorf("unsupported database type: %s", c.dbType)
	}

	rows, err := c.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate columns: %w", err)
	}

	return columns, nil
}

// GetTableData returns paginated rows from a table.
func (c *Client) GetTableData(ctx context.Context, tableName string, limit, offset int) (RowData, error) {
	var query string
	switch c.dbType {
	case "postgres":
		query = fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", tableName, limit, offset)
	case "mysql":
		query = fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", tableName, limit, offset)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", c.dbType)
	}

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query table data: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	var data RowData
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return data, nil
}
