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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatasource(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceType
		config     Config
		wantErr    bool
		errMessage string
	}{
		{
			name:       "mysql datasource",
			sourceType: SourceTypeMySQL,
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "nodes",
			},
			wantErr: true, // Will fail without real MySQL, but validates factory logic
		},
		{
			name:       "postgresql datasource (not yet implemented)",
			sourceType: SourceTypePostgreSQL,
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "nodes",
			},
			wantErr:    true,
			errMessage: "postgresql datasource not yet implemented",
		},
		{
			name:       "unsupported datasource type",
			sourceType: SourceType("mongodb"),
			config:     Config{},
			wantErr:    true,
			errMessage: "unsupported datasource type: mongodb",
		},
		{
			name:       "empty datasource type",
			sourceType: SourceType(""),
			config:     Config{},
			wantErr:    true,
			errMessage: "unsupported datasource type:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDatasource(tt.sourceType, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestSourceType(t *testing.T) {
	// Test that source type constants are defined correctly
	assert.Equal(t, SourceType("mysql"), SourceTypeMySQL)
	assert.Equal(t, SourceType("postgresql"), SourceTypePostgreSQL)
}

func TestNodeRow(t *testing.T) {
	// Test NodeRow structure
	row := NodeRow{
		UID:       "node1",
		HostOrURL: "https://node1.example.com",
		Activate:  "1",
		Extra: map[string]string{
			"planId": "premium",
			"region": "us-east-1",
		},
	}

	assert.Equal(t, "node1", row.UID)
	assert.Equal(t, "https://node1.example.com", row.HostOrURL)
	assert.Equal(t, "1", row.Activate)
	assert.Equal(t, "premium", row.Extra["planId"])
	assert.Equal(t, "us-east-1", row.Extra["region"])
}

func TestQueryConfig(t *testing.T) {
	// Test QueryConfig structure
	config := QueryConfig{
		Table: "nodes",
		ValueMappings: ValueMappings{
			UID:       "id",
			HostOrURL: "url",
			Activate:  "active",
		},
		ExtraMappings: map[string]string{
			"planId": "subscription_plan",
			"region": "deployment_region",
		},
	}

	assert.Equal(t, "nodes", config.Table)
	assert.Equal(t, "id", config.ValueMappings.UID)
	assert.Equal(t, "url", config.ValueMappings.HostOrURL)
	assert.Equal(t, "active", config.ValueMappings.Activate)
	assert.Equal(t, "subscription_plan", config.ExtraMappings["planId"])
	assert.Equal(t, "deployment_region", config.ExtraMappings["region"])
}

func TestConfig(t *testing.T) {
	// Test Config structure with all fields
	config := Config{
		Host:            "db.example.com",
		Port:            3307,
		Username:        "app_user",
		Password:        "secret",
		Database:        "production",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: "10m",
	}

	assert.Equal(t, "db.example.com", config.Host)
	assert.Equal(t, int32(3307), config.Port)
	assert.Equal(t, "app_user", config.Username)
	assert.Equal(t, "secret", config.Password)
	assert.Equal(t, "production", config.Database)
	assert.Equal(t, 50, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, "10m", config.ConnMaxLifetime)
}
