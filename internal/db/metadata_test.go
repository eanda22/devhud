package db

import "testing"

func TestTableInfo(t *testing.T) {
	table := TableInfo{
		Name:        "users",
		RowCount:    100,
		ColumnCount: 5,
	}

	if table.Name != "users" {
		t.Errorf("expected Name to be 'users', got %s", table.Name)
	}
	if table.RowCount != 100 {
		t.Errorf("expected RowCount to be 100, got %d", table.RowCount)
	}
	if table.ColumnCount != 5 {
		t.Errorf("expected ColumnCount to be 5, got %d", table.ColumnCount)
	}
}

func TestColumnInfo(t *testing.T) {
	col := ColumnInfo{
		Name: "id",
		Type: "integer",
	}

	if col.Name != "id" {
		t.Errorf("expected Name to be 'id', got %s", col.Name)
	}
	if col.Type != "integer" {
		t.Errorf("expected Type to be 'integer', got %s", col.Type)
	}
}

func TestRowData(t *testing.T) {
	data := RowData{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	if len(data) != 2 {
		t.Errorf("expected 2 rows, got %d", len(data))
	}
	if data[0]["name"] != "Alice" {
		t.Errorf("expected first row name to be 'Alice', got %v", data[0]["name"])
	}
	if data[1]["id"] != 2 {
		t.Errorf("expected second row id to be 2, got %v", data[1]["id"])
	}
}
