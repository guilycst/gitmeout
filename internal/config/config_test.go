package config

import (
	"os"
	"strings"
	"testing"
)

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "no env vars",
			input:    "hello world",
			envVars:  nil,
			expected: "hello world",
		},
		{
			name:     "single env var",
			input:    "token: ${TOKEN}",
			envVars:  map[string]string{"TOKEN": "secret123"},
			expected: "token: secret123",
		},
		{
			name:     "multiple env vars",
			input:    "${VAR1} and ${VAR2}",
			envVars:  map[string]string{"VAR1": "first", "VAR2": "second"},
			expected: "first and second",
		},
		{
			name:     "missing env var",
			input:    "token: ${MISSING}",
			envVars:  nil,
			expected: "token: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got := expandEnvVars(tt.input)
			if got != tt.expected {
				t.Errorf("expandEnvVars() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			input: `
source:
  type: github
  token: test-token
  filters:
    personal: true
    orgs: []
    repos: []
targets:
  - name: codeberg
    type: forgejo
    url: https://codeberg.org
    token: test-token
`,
			wantErr: false,
		},
		{
			name: "missing source token",
			input: `
source:
  type: github
  token: ""
targets:
  - name: codeberg
    type: forgejo
    url: https://codeberg.org
    token: test-token
`,
			wantErr: true,
			errMsg:  "source token is required",
		},
		{
			name: "invalid source type",
			input: `
source:
  type: gitlab
  token: test-token
targets:
  - name: codeberg
    type: forgejo
    url: https://codeberg.org
    token: test-token
`,
			wantErr: true,
			errMsg:  "source type must be 'github'",
		},
		{
			name: "no targets",
			input: `
source:
  type: github
  token: test-token
targets: []
`,
			wantErr: true,
			errMsg:  "at least one target is required",
		},
		{
			name: "missing target token",
			input: `
source:
  type: github
  token: test-token
targets:
  - name: codeberg
    type: forgejo
    url: https://codeberg.org
    token: ""
`,
			wantErr: true,
			errMsg:  "token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Parse() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestParseRepoSpec(t *testing.T) {
	tests := []struct {
		name         string
		spec         string
		wantOwner    string
		wantRepo     string
		wantWildcard bool
	}{
		{
			name:         "specific repo",
			spec:         "owner/repo",
			wantOwner:    "owner",
			wantRepo:     "repo",
			wantWildcard: false,
		},
		{
			name:         "wildcard",
			spec:         "owner/*",
			wantOwner:    "owner",
			wantRepo:     "*",
			wantWildcard: true,
		},
		{
			name:         "no slash",
			spec:         "owner",
			wantOwner:    "owner",
			wantRepo:     "",
			wantWildcard: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepo, gotWildcard := ParseRepoSpec(tt.spec)
			if gotOwner != tt.wantOwner {
				t.Errorf("ParseRepoSpec() owner = %q, want %q", gotOwner, tt.wantOwner)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("ParseRepoSpec() repo = %q, want %q", gotRepo, tt.wantRepo)
			}
			if gotWildcard != tt.wantWildcard {
				t.Errorf("ParseRepoSpec() wildcard = %v, want %v", gotWildcard, tt.wantWildcard)
			}
		})
	}
}

func TestFilters(t *testing.T) {
	t.Run("HasRepoFilter", func(t *testing.T) {
		if (&Filters{Repos: []string{"a/b"}}).HasRepoFilter() != true {
			t.Error("expected HasRepoFilter to be true")
		}
		if (&Filters{}).HasRepoFilter() != false {
			t.Error("expected HasRepoFilter to be false")
		}
	})

	t.Run("HasOrgFilter", func(t *testing.T) {
		if (&Filters{Orgs: []string{"org1"}}).HasOrgFilter() != true {
			t.Error("expected HasOrgFilter to be true")
		}
		if (&Filters{}).HasOrgFilter() != false {
			t.Error("expected HasOrgFilter to be false")
		}
	})
}
