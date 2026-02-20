package db

import "testing"

func TestBuildConnectionString(t *testing.T) {
	tests := []struct {
		name   string
		config *ConnectionConfig
		dbType string
		want   string
	}{
		{
			name: "postgres with password",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
			},
			dbType: "postgres",
			want:   "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "postgres without password",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "",
				Database: "postgres",
			},
			dbType: "postgres",
			want:   "postgres://postgres:@localhost:5432/postgres?sslmode=disable",
		},
		{
			name: "mysql with password",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     "3306",
				User:     "root",
				Password: "rootpass",
				Database: "mydb",
			},
			dbType: "mysql",
			want:   "root:rootpass@tcp(localhost:3306)/mydb",
		},
		{
			name: "mysql without password",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     "3306",
				User:     "root",
				Password: "",
				Database: "mysql",
			},
			dbType: "mysql",
			want:   "root:@tcp(localhost:3306)/mysql",
		},
		{
			name: "postgres custom port",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     "54321",
				User:     "admin",
				Password: "secret",
				Database: "appdb",
			},
			dbType: "postgres",
			want:   "postgres://admin:secret@localhost:54321/appdb?sslmode=disable",
		},
		{
			name: "mysql custom port",
			config: &ConnectionConfig{
				Host:     "127.0.0.1",
				Port:     "33060",
				User:     "app",
				Password: "pass123",
				Database: "application",
			},
			dbType: "mysql",
			want:   "app:pass123@tcp(127.0.0.1:33060)/application",
		},
		{
			name: "unsupported database type",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     "6379",
				User:     "user",
				Password: "pass",
				Database: "db",
			},
			dbType: "redis",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildConnectionString(tt.config, tt.dbType)
			if got != tt.want {
				t.Errorf("BuildConnectionString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		want   string
	}{
		{
			name:   "postgres default port",
			dbType: "postgres",
			want:   "5432",
		},
		{
			name:   "mysql default port",
			dbType: "mysql",
			want:   "3306",
		},
		{
			name:   "unknown database type",
			dbType: "mongodb",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultPort(tt.dbType)
			if got != tt.want {
				t.Errorf("getDefaultPort(%q) = %q, want %q", tt.dbType, got, tt.want)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envMap       map[string]string
		key          string
		defaultValue string
		want         string
	}{
		{
			name: "key exists",
			envMap: map[string]string{
				"USER": "testuser",
			},
			key:          "USER",
			defaultValue: "default",
			want:         "testuser",
		},
		{
			name:         "key does not exist",
			envMap:       map[string]string{},
			key:          "USER",
			defaultValue: "default",
			want:         "default",
		},
		{
			name: "key exists but empty",
			envMap: map[string]string{
				"USER": "",
			},
			key:          "USER",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEnv(tt.envMap, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv() = %q, want %q", got, tt.want)
			}
		})
	}
}
