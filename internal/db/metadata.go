package db

import (
	"context"
	"fmt"
)

// TableInfo holds metadata about a database table.
type TableInfo struct {
	Name        string
	RowCount    int
	ColumnCount int
}

// ListTables returns all tables in the database with metadata.
func (c *Client) ListTables(ctx context.Context) ([]TableInfo, error) {
	var query string
	switch c.dbType {
	case "postgres":
		query = `
			SELECT
				t.table_name,
				COALESCE((
					SELECT COUNT(*)
					FROM information_schema.columns
					WHERE table_schema = t.table_schema
					AND table_name = t.table_name
				), 0) as column_count
			FROM information_schema.tables t
			WHERE t.table_schema = 'public'
			ORDER BY t.table_name
		`
	case "mysql":
		query = `
			SELECT
				t.TABLE_NAME,
				COALESCE((
					SELECT COUNT(*)
					FROM information_schema.COLUMNS c
					WHERE c.TABLE_SCHEMA = t.TABLE_SCHEMA
					AND c.TABLE_NAME = t.TABLE_NAME
				), 0) as column_count
			FROM information_schema.TABLES t
			WHERE t.TABLE_SCHEMA = DATABASE()
			ORDER BY t.TABLE_NAME
		`
	default:
		return nil, fmt.Errorf("unsupported database type: %s", c.dbType)
	}

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.ColumnCount); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}

		rowCount, err := c.getTableRowCount(ctx, table.Name)
		if err != nil {
			table.RowCount = 0
		} else {
			table.RowCount = rowCount
		}

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tables: %w", err)
	}

	return tables, nil
}

func (c *Client) getTableRowCount(ctx context.Context, tableName string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var count int
	err := c.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
