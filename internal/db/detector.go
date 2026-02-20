package db

import "strings"

// dbPatterns maps common database image name substrings to database types.
var dbPatterns = map[string]string{
	"postgres": "postgres",
	"mysql":    "mysql",
	"mariadb":  "mariadb",
	"mongo":    "mongodb",
	"mongodb":  "mongodb",
	"redis":    "redis",
}

// DetectType checks if a Docker image name matches known database patterns.
// Returns the database type or empty string if not a database image.
func DetectType(imageName string) string {
	lower := strings.ToLower(imageName)
	for pattern, dbType := range dbPatterns {
		if strings.Contains(lower, pattern) {
			return dbType
		}
	}
	return ""
}
