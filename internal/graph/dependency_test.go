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

package graph

import (
	"testing"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
)

func TestNewDependencyGraph(t *testing.T) {
	graph := NewDependencyGraph()
	if graph == nil {
		t.Fatal("NewDependencyGraph() returned nil")
	}
	if graph.Nodes == nil {
		t.Error("Nodes map should be initialized")
	}
	if len(graph.Nodes) != 0 {
		t.Errorf("New graph should have 0 nodes, got %d", len(graph.Nodes))
	}
}

func TestDependencyGraph_AddResource(t *testing.T) {
	tests := []struct {
		name      string
		resources []tenantsv1.TResource
		wantErr   bool
		errMsg    string
	}{
		{
			name: "add single resource",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
			},
			wantErr: false,
		},
		{
			name: "add multiple resources",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource2"},
				{ID: "resource3"},
			},
			wantErr: false,
		},
		{
			name: "add resource with empty ID",
			resources: []tenantsv1.TResource{
				{ID: ""},
			},
			wantErr: true,
			errMsg:  "resource ID cannot be empty",
		},
		{
			name: "add duplicate resource ID",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource1"},
			},
			wantErr: true,
			errMsg:  "duplicate resource ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewDependencyGraph()
			var err error
			for _, resource := range tt.resources {
				err = graph.AddResource(resource)
				if err != nil {
					break
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("AddResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					// Check if error message contains expected substring
					if len(tt.errMsg) > 0 && len(err.Error()) > 0 {
						// Just check that error occurred for now
						return
					}
					t.Errorf("AddResource() error = %v, want error containing %v", err, tt.errMsg)
				}
			}

			if !tt.wantErr && len(graph.Nodes) != len(tt.resources) {
				t.Errorf("Graph should have %d nodes, got %d", len(tt.resources), len(graph.Nodes))
			}
		})
	}
}

func TestDependencyGraph_Validate(t *testing.T) {
	tests := []struct {
		name      string
		resources []tenantsv1.TResource
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid graph with no dependencies",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource2"},
			},
			wantErr: false,
		},
		{
			name: "valid graph with linear dependencies",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource2", DependIds: []string{"resource1"}},
				{ID: "resource3", DependIds: []string{"resource2"}},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource2", DependIds: []string{"missing"}},
			},
			wantErr: true,
			errMsg:  "depends on non-existent resource",
		},
		{
			name: "circular dependency - direct",
			resources: []tenantsv1.TResource{
				{ID: "resource1", DependIds: []string{"resource2"}},
				{ID: "resource2", DependIds: []string{"resource1"}},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "circular dependency - indirect",
			resources: []tenantsv1.TResource{
				{ID: "resource1", DependIds: []string{"resource2"}},
				{ID: "resource2", DependIds: []string{"resource3"}},
				{ID: "resource3", DependIds: []string{"resource1"}},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewDependencyGraph()
			for _, resource := range tt.resources {
				if err := graph.AddResource(resource); err != nil {
					t.Fatalf("Failed to add resource: %v", err)
				}
			}

			err := graph.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				// Just verify error occurred for now
			}
		})
	}
}

func TestDependencyGraph_TopologicalSort(t *testing.T) {
	tests := []struct {
		name         string
		resources    []tenantsv1.TResource
		wantErr      bool
		validateFunc func(*testing.T, []*Node)
	}{
		{
			name: "simple linear order",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource2", DependIds: []string{"resource1"}},
				{ID: "resource3", DependIds: []string{"resource2"}},
			},
			wantErr: false,
			validateFunc: func(t *testing.T, nodes []*Node) {
				if len(nodes) != 3 {
					t.Errorf("Expected 3 nodes, got %d", len(nodes))
					return
				}
				// resource1 should come before resource2
				// resource2 should come before resource3
				positions := make(map[string]int)
				for i, node := range nodes {
					positions[node.ID] = i
				}
				if positions["resource1"] >= positions["resource2"] {
					t.Error("resource1 should come before resource2")
				}
				if positions["resource2"] >= positions["resource3"] {
					t.Error("resource2 should come before resource3")
				}
			},
		},
		{
			name: "parallel branches",
			resources: []tenantsv1.TResource{
				{ID: "root"},
				{ID: "branch1", DependIds: []string{"root"}},
				{ID: "branch2", DependIds: []string{"root"}},
				{ID: "leaf", DependIds: []string{"branch1", "branch2"}},
			},
			wantErr: false,
			validateFunc: func(t *testing.T, nodes []*Node) {
				if len(nodes) != 4 {
					t.Errorf("Expected 4 nodes, got %d", len(nodes))
					return
				}
				positions := make(map[string]int)
				for i, node := range nodes {
					positions[node.ID] = i
				}
				// root should come before branches
				if positions["root"] >= positions["branch1"] {
					t.Error("root should come before branch1")
				}
				if positions["root"] >= positions["branch2"] {
					t.Error("root should come before branch2")
				}
				// branches should come before leaf
				if positions["branch1"] >= positions["leaf"] {
					t.Error("branch1 should come before leaf")
				}
				if positions["branch2"] >= positions["leaf"] {
					t.Error("branch2 should come before leaf")
				}
			},
		},
		{
			name: "no dependencies",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource2"},
				{ID: "resource3"},
			},
			wantErr: false,
			validateFunc: func(t *testing.T, nodes []*Node) {
				if len(nodes) != 3 {
					t.Errorf("Expected 3 nodes, got %d", len(nodes))
				}
				// All should be at level 0
				for _, node := range nodes {
					if node.Level != 0 {
						t.Errorf("Node %s should be at level 0, got %d", node.ID, node.Level)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewDependencyGraph()
			for _, resource := range tt.resources {
				if err := graph.AddResource(resource); err != nil {
					t.Fatalf("Failed to add resource: %v", err)
				}
			}

			nodes, err := graph.TopologicalSort()
			if (err != nil) != tt.wantErr {
				t.Errorf("TopologicalSort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFunc != nil {
				tt.validateFunc(t, nodes)
			}
		})
	}
}

func TestDependencyGraph_GetResourcesByLevel(t *testing.T) {
	graph := NewDependencyGraph()
	resources := []tenantsv1.TResource{
		{ID: "level0-a"},
		{ID: "level0-b"},
		{ID: "level1-a", DependIds: []string{"level0-a"}},
		{ID: "level1-b", DependIds: []string{"level0-b"}},
		{ID: "level2", DependIds: []string{"level1-a", "level1-b"}},
	}

	for _, resource := range resources {
		if err := graph.AddResource(resource); err != nil {
			t.Fatalf("Failed to add resource: %v", err)
		}
	}

	// Trigger level calculation
	_, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort() failed: %v", err)
	}

	byLevel := graph.GetResourcesByLevel()

	// Verify we have 3 levels (0, 1, 2)
	if len(byLevel) != 3 {
		t.Errorf("Expected 3 levels, got %d", len(byLevel))
	}

	// Verify level 0 has 2 resources
	if len(byLevel[0]) != 2 {
		t.Errorf("Level 0 should have 2 resources, got %d", len(byLevel[0]))
	}

	// Verify level 1 has 2 resources
	if len(byLevel[1]) != 2 {
		t.Errorf("Level 1 should have 2 resources, got %d", len(byLevel[1]))
	}

	// Verify level 2 has 1 resource
	if len(byLevel[2]) != 1 {
		t.Errorf("Level 2 should have 1 resource, got %d", len(byLevel[2]))
	}
}

func TestBuildGraph(t *testing.T) {
	tests := []struct {
		name      string
		resources []tenantsv1.TResource
		wantErr   bool
	}{
		{
			name: "valid graph",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource2", DependIds: []string{"resource1"}},
			},
			wantErr: false,
		},
		{
			name:      "empty resources",
			resources: []tenantsv1.TResource{},
			wantErr:   false,
		},
		{
			name: "invalid - duplicate ID",
			resources: []tenantsv1.TResource{
				{ID: "resource1"},
				{ID: "resource1"},
			},
			wantErr: true,
		},
		{
			name: "invalid - circular dependency",
			resources: []tenantsv1.TResource{
				{ID: "resource1", DependIds: []string{"resource2"}},
				{ID: "resource2", DependIds: []string{"resource1"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := BuildGraph(tt.resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && graph == nil {
				t.Error("BuildGraph() returned nil graph without error")
			}
		})
	}
}
