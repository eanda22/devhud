package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
)

// ConnectionConfig holds database connection parameters.
type ConnectionConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// DiscoverConfig inspects a Docker container to extract database connection parameters from env vars.
func DiscoverConfig(ctx context.Context, dockerClient *client.Client, containerID, dbType string) (*ConnectionConfig, error) {
	inspect, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("inspect container: %w", err)
	}

	config := &ConnectionConfig{
		Host: "localhost",
		Port: getDefaultPort(dbType),
	}

	envMap := make(map[string]string)
	for _, env := range inspect.Config.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	switch dbType {
	case "postgres":
		config.User = getEnv(envMap, "POSTGRES_USER", "postgres")
		config.Password = getEnv(envMap, "POSTGRES_PASSWORD", "")
		config.Database = getEnv(envMap, "POSTGRES_DB", "postgres")
	case "mysql":
		config.User = getEnv(envMap, "MYSQL_USER", "root")
		config.Password = getEnv(envMap, "MYSQL_PASSWORD", getEnv(envMap, "MYSQL_ROOT_PASSWORD", ""))
		config.Database = getEnv(envMap, "MYSQL_DATABASE", "mysql")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if inspect.NetworkSettings != nil && len(inspect.NetworkSettings.Ports) > 0 {
		for portBinding := range inspect.NetworkSettings.Ports {
			portStr := portBinding.Port()
			if portStr == getDefaultPort(dbType) {
				bindings := inspect.NetworkSettings.Ports[portBinding]
				if len(bindings) > 0 && bindings[0].HostPort != "" {
					config.Port = bindings[0].HostPort
					break
				}
			}
		}
	}

	return config, nil
}

// BuildConnectionString formats a DSN string for database/sql drivers.
func BuildConnectionString(config *ConnectionConfig, dbType string) string {
	switch dbType {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)
	default:
		return ""
	}
}

func getEnv(envMap map[string]string, key, defaultValue string) string {
	if val, ok := envMap[key]; ok && val != "" {
		return val
	}
	return defaultValue
}

func getDefaultPort(dbType string) string {
	switch dbType {
	case "postgres":
		return "5432"
	case "mysql":
		return "3306"
	default:
		return ""
	}
}
