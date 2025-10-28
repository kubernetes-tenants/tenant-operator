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

package template

import (
	"testing"
)

func TestEngine_Render(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		vars     Variables
		want     string
		wantErr  bool
	}{
		{
			name:     "simple substitution",
			template: "tenant-{{ .uid }}",
			vars:     Variables{"uid": "42"},
			want:     "tenant-42",
			wantErr:  false,
		},
		{
			name:     "with host extraction",
			template: "{{ .host }}",
			vars:     Variables{"host": "example.com"},
			want:     "example.com",
			wantErr:  false,
		},
		{
			name:     "with default function",
			template: "{{ default \"nginx:stable\" .deployImage }}",
			vars:     Variables{},
			want:     "nginx:stable",
			wantErr:  false,
		},
		{
			name:     "with trunc63",
			template: "{{ .longName | trunc63 }}",
			vars:     Variables{"longName": "this-is-a-very-long-name-that-exceeds-kubernetes-label-limit-of-sixtythree-characters"},
			want:     "this-is-a-very-long-name-that-exceeds-kubernetes-label-limit-of",
			wantErr:  false,
		},
		{
			name:     "complex template",
			template: "{{ .uid }}-{{ default \"api\" .service }}-{{ .host | trunc63 }}",
			vars:     Variables{"uid": "123", "host": "very-long-hostname-example.com"},
			want:     "123-api-very-long-hostname-example.com",
			wantErr:  false,
		},
		{
			name:     "invalid template syntax",
			template: "{{ .uid",
			vars:     Variables{"uid": "42"},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Render(tt.template, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngine_RenderMap(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name    string
		input   map[string]string
		vars    Variables
		want    map[string]string
		wantErr bool
	}{
		{
			name: "render multiple keys",
			input: map[string]string{
				"app":    "{{ .uid }}",
				"tenant": "{{ .uid }}-{{ .host }}",
			},
			vars: Variables{"uid": "42", "host": "example.com"},
			want: map[string]string{
				"app":    "42",
				"tenant": "42-example.com",
			},
			wantErr: false,
		},
		{
			name:    "nil map",
			input:   nil,
			vars:    Variables{},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.RenderMap(tt.input, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil && tt.want == nil {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("RenderMap() got %d items, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("RenderMap()[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func Test_toHost(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "full URL with scheme",
			input: "https://example.com/path",
			want:  "example.com",
		},
		{
			name:  "URL with port",
			input: "https://example.com:8080/path",
			want:  "example.com",
		},
		{
			name:  "just hostname",
			input: "example.com",
			want:  "example.com",
		},
		{
			name:  "hostname with port",
			input: "example.com:8080",
			want:  "example.com",
		},
		{
			name:  "subdomain",
			input: "https://api.example.com",
			want:  "api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toHost(tt.input); got != tt.want {
				t.Errorf("toHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trunc63(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short string",
			input: "short",
			want:  "short",
		},
		{
			name:  "exact 63 chars",
			input: "123456789012345678901234567890123456789012345678901234567890123",
			want:  "123456789012345678901234567890123456789012345678901234567890123",
		},
		{
			name:  "long string",
			input: "this-is-a-very-long-string-that-exceeds-the-kubernetes-label-limit-of-sixtythree-characters",
			want:  "this-is-a-very-long-string-that-exceeds-the-kubernetes-label-li",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trunc63(tt.input)
			if got != tt.want {
				t.Errorf("trunc63() = %v, want %v", got, tt.want)
			}
			if len(got) > 63 {
				t.Errorf("trunc63() result length %d exceeds 63", len(got))
			}
		})
	}
}

func TestBuildVariables(t *testing.T) {
	tests := []struct {
		name          string
		uid           string
		hostOrURL     string
		activate      string
		extraMappings map[string]string
		wantKeys      []string
	}{
		{
			name:      "basic variables",
			uid:       "42",
			hostOrURL: "https://example.com",
			activate:  "true",
			extraMappings: map[string]string{
				"deployImage": "my-image:latest",
			},
			wantKeys: []string{"uid", "hostOrUrl", "activate", "host", "deployImage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildVariables(tt.uid, tt.hostOrURL, tt.activate, tt.extraMappings)

			// Check all expected keys exist
			for _, key := range tt.wantKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("BuildVariables() missing key %s", key)
				}
			}

			// Check values
			if result["uid"] != tt.uid {
				t.Errorf("BuildVariables() uid = %v, want %v", result["uid"], tt.uid)
			}
			if result["hostOrUrl"] != tt.hostOrURL {
				t.Errorf("BuildVariables() hostOrUrl = %v, want %v", result["hostOrUrl"], tt.hostOrURL)
			}
			// host should be auto-extracted
			if result["host"] != "example.com" {
				t.Errorf("BuildVariables() host = %v, want example.com", result["host"])
			}
		})
	}
}
