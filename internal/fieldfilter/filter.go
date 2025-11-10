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

package fieldfilter

import (
	"fmt"

	"github.com/ohler55/ojg/jp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Filter handles JSONPath-based field filtering for Kubernetes resources
// Uses the ojg/jp library for complete JSONPath standard support
type Filter struct {
	ignoreFields []string
	parsedPaths  []jp.Expr
}

// NewFilter creates a new field filter with JSONPath expressions
// JSONPath expressions follow the standard JSONPath format:
// - Simple paths: $.spec.replicas
// - Nested paths: $.spec.template.spec.containers[0].image
// - Map keys: $.metadata.annotations['app.kubernetes.io/name']
// - Wildcards: $.spec.containers[*].image
// - Filters: $.items[?(@.status == 'active')]
func NewFilter(ignoreFields []string) (*Filter, error) {
	if len(ignoreFields) == 0 {
		return &Filter{
			ignoreFields: []string{},
			parsedPaths:  []jp.Expr{},
		}, nil
	}

	f := &Filter{
		ignoreFields: ignoreFields,
		parsedPaths:  make([]jp.Expr, 0, len(ignoreFields)),
	}

	// Parse and validate all paths using ojg/jp
	for _, pathStr := range ignoreFields {
		path, err := jp.ParseString(pathStr)
		if err != nil {
			return nil, fmt.Errorf("invalid JSONPath %q: %w", pathStr, err)
		}
		f.parsedPaths = append(f.parsedPaths, path)
	}

	return f, nil
}

// RemoveIgnoredFields removes fields matching ignoreFields from obj
// This modifies the object in-place using JSONPath deletion
func (f *Filter) RemoveIgnoredFields(obj *unstructured.Unstructured) error {
	if obj == nil {
		return fmt.Errorf("object cannot be nil")
	}

	// No-op if no ignore fields
	if len(f.parsedPaths) == 0 {
		return nil
	}

	// Get the underlying map from unstructured object
	data := obj.Object

	// Remove each ignored field using JSONPath
	for _, path := range f.parsedPaths {
		// Use Remove() which handles non-existent paths gracefully
		// Returns modified data and error
		_, err := path.Remove(data)
		if err != nil {
			// Silently continue - some paths may not exist and that's OK
			// In production, this could be logged at debug level
			continue
		}
	}

	return nil
}

// ValidateJSONPath validates a single JSONPath expression
// This is a convenience function for validation without creating a Filter
func ValidateJSONPath(pathStr string) error {
	_, err := jp.ParseString(pathStr)
	if err != nil {
		return fmt.Errorf("invalid JSONPath %q: %w", pathStr, err)
	}
	return nil
}

// GetMatchingFields returns the values at the specified JSONPath
// This is useful for debugging or inspecting what fields would be removed
func (f *Filter) GetMatchingFields(obj *unstructured.Unstructured) (map[string][]interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("object cannot be nil")
	}

	results := make(map[string][]interface{})
	data := obj.Object

	for i, path := range f.parsedPaths {
		matches := path.Get(data)
		if len(matches) > 0 {
			results[f.ignoreFields[i]] = matches
		}
	}

	return results, nil
}
