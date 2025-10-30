/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// MySQLClient wraps a MySQL database connection
type MySQLClient struct {
	db *sql.DB
}

// TenantRow represents a row from the tenant table
type TenantRow struct {
	UID       string
	HostOrURL string
	Activate  string
	Extra     map[string]string
}

// MySQLConfig holds MySQL connection configuration
type MySQLConfig struct {
	Host     string
	Port     int32
	Username string
	Password string
	Database string
}

// NewMySQLClient creates a new MySQL client
func NewMySQLClient(config MySQLConfig) (*MySQLClient, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close() // Best effort close on error
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	return &MySQLClient{db: db}, nil
}

// Close closes the database connection
func (c *MySQLClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// QueryTenants queries active tenants from the database
func (c *MySQLClient) QueryTenants(ctx context.Context, table string, valueMappings ValueMappings, extraMappings map[string]string) ([]TenantRow, error) {
	// Build column list
	columns := []string{
		valueMappings.UID,
		valueMappings.HostOrURL,
		valueMappings.Activate,
	}

	// Add extra columns
	extraColumns := make([]string, 0, len(extraMappings))
	for _, col := range extraMappings {
		columns = append(columns, col)
		extraColumns = append(extraColumns, col)
	}

	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s", joinColumns(columns), table)

	// Execute query
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer func() {
		_ = rows.Close() // Best effort close
	}()

	// Scan results
	var tenants []TenantRow
	for rows.Next() {
		row := TenantRow{
			Extra: make(map[string]string),
		}

		// Prepare scan destinations
		scanDest := []interface{}{&row.UID, &row.HostOrURL, &row.Activate}

		// Add extra column destinations
		extraValues := make([]sql.NullString, len(extraColumns))
		for i := range extraValues {
			scanDest = append(scanDest, &extraValues[i])
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Map extra values - Build column index map first for stable mapping
		colIndex := make(map[string]int)
		for i, col := range extraColumns {
			colIndex[col] = i
		}

		// Map extra values using stable indices
		for key, col := range extraMappings {
			idx, ok := colIndex[col]
			if !ok {
				// Column not in result set (shouldn't happen)
				row.Extra[key] = ""
				continue
			}
			if extraValues[idx].Valid {
				row.Extra[key] = extraValues[idx].String
			} else {
				row.Extra[key] = "" // Null values become empty strings
			}
		}

		// Filter: only include active tenants
		if isActive(row.Activate) && row.HostOrURL != "" {
			tenants = append(tenants, row)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tenants, nil
}

// ValueMappings defines required column mappings
type ValueMappings struct {
	UID       string
	HostOrURL string
	Activate  string
}

// Helper functions

func joinColumns(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += "`" + col + "`" // Escape column names
	}
	return result
}

func isActive(value string) bool {
	// Truthy values: "1", "true", "TRUE", "yes", etc.
	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "Yes":
		return true
	default:
		return false
	}
}
