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

package datasource

import (
	"context"
	"fmt"
	"io"
)

// Datasource defines the interface that all datasource adapters must implement
type Datasource interface {
	// QueryNodes retrieves active node rows from the datasource
	QueryNodes(ctx context.Context, config QueryConfig) ([]NodeRow, error)

	// Close closes the datasource connection
	io.Closer
}

// NodeRow represents a row from the node datasource
type NodeRow struct {
	UID       string
	HostOrURL string
	Activate  string
	Extra     map[string]string
}

// QueryConfig holds configuration for querying nodes
type QueryConfig struct {
	// Table/Collection name
	Table string

	// Required column mappings
	ValueMappings ValueMappings

	// Extra column mappings
	ExtraMappings map[string]string
}

// ValueMappings defines required column mappings
type ValueMappings struct {
	UID       string
	HostOrURL string
	Activate  string
}

// Config holds generic datasource configuration
// Specific adapters will extract their needed fields
type Config struct {
	// MySQL/PostgreSQL fields
	Host     string
	Port     int32
	Username string
	Password string
	Database string

	// Connection pool settings (optional, adapter-specific defaults will be used if not set)
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime string // Duration string (e.g., "5m")
}

// SourceType represents the type of datasource
type SourceType string

const (
	// SourceTypeMySQL represents a MySQL datasource
	SourceTypeMySQL SourceType = "mysql"
	// SourceTypePostgreSQL represents a PostgreSQL datasource (planned)
	SourceTypePostgreSQL SourceType = "postgresql"
)

// NewDatasource creates a new datasource adapter based on the source type
func NewDatasource(sourceType SourceType, config Config) (Datasource, error) {
	switch sourceType {
	case SourceTypeMySQL:
		return NewMySQLAdapter(config)
	case SourceTypePostgreSQL:
		return nil, fmt.Errorf("postgresql datasource not yet implemented (planned for v1.2)")
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", sourceType)
	}
}
