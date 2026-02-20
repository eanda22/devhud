package db

import "testing"

func TestDetectType(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		want      string
	}{
		{
			name:      "postgres image",
			imageName: "postgres:15-alpine",
			want:      "postgres",
		},
		{
			name:      "mysql image",
			imageName: "mysql:8.0",
			want:      "mysql",
		},
		{
			name:      "mariadb image",
			imageName: "mariadb:latest",
			want:      "mariadb",
		},
		{
			name:      "mongodb image",
			imageName: "mongo:6.0",
			want:      "mongodb",
		},
		{
			name:      "mongodb explicit",
			imageName: "mongodb:latest",
			want:      "mongodb",
		},
		{
			name:      "redis image",
			imageName: "redis:7-alpine",
			want:      "redis",
		},
		{
			name:      "non-database image",
			imageName: "nginx:latest",
			want:      "",
		},
		{
			name:      "empty image",
			imageName: "",
			want:      "",
		},
		{
			name:      "uppercase postgres",
			imageName: "POSTGRES:15",
			want:      "postgres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectType(tt.imageName)
			if got != tt.want {
				t.Errorf("DetectType(%q) = %q, want %q", tt.imageName, got, tt.want)
			}
		})
	}
}
