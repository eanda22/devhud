package db

import (
	"testing"
)

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		dbType  string
		want    string
		wantErr bool
	}{
		{
			name:   "postgres simple",
			input:  "users",
			dbType: "postgres",
			want:   `"users"`,
		},
		{
			name:   "postgres with double quote",
			input:  `my"table`,
			dbType: "postgres",
			want:   `"my""table"`,
		},
		{
			name:   "mysql simple",
			input:  "users",
			dbType: "mysql",
			want:   "`users`",
		},
		{
			name:   "mysql with backtick",
			input:  "my`table",
			dbType: "mysql",
			want:   "`my``table`",
		},
		{
			name:    "unsupported db type",
			input:   "users",
			dbType:  "sqlite",
			wantErr: true,
		},
		{
			name:   "postgres with spaces",
			input:  "my table",
			dbType: "postgres",
			want:   `"my table"`,
		},
		{
			name:   "mysql with spaces",
			input:  "my table",
			dbType: "mysql",
			want:   "`my table`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := quoteIdentifier(tt.input, tt.dbType)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("quoteIdentifier(%q, %q) = %q, want %q", tt.input, tt.dbType, got, tt.want)
			}
		})
	}
}
