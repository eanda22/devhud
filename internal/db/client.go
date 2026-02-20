package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// Client wraps a database connection with type information.
type Client struct {
	db     *sql.DB
	dbType string
	config *ConnectionConfig
}

// NewClient creates a new database client and opens a connection.
func NewClient(ctx context.Context, config *ConnectionConfig, dbType string) (*Client, error) {
	dsn := BuildConnectionString(config, dbType)
	if dsn == "" {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	var driverName string
	switch dbType {
	case "postgres":
		driverName = "postgres"
	case "mysql":
		driverName = "mysql"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	client := &Client{
		db:     db,
		dbType: dbType,
		config: config,
	}

	if err := client.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return client, nil
}

// Close closes the database connection.
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping verifies the database connection is alive.
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
