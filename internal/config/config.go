package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Source  Source   `yaml:"source"`
	Targets []Target `yaml:"targets"`
}

type Source struct {
	Type    string  `yaml:"type"`
	Token   string  `yaml:"token"`
	Filters Filters `yaml:"filters"`
}

type Filters struct {
	Personal bool     `yaml:"personal"`
	Orgs     []string `yaml:"orgs"`
	Repos    []string `yaml:"repos"`
}

type Target struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	URL        string `yaml:"url"`
	Token      string `yaml:"token"`
	MirrorType string `yaml:"mirror_type"` // "push" or "pull", defaults to "push"
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func expandEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1]
		return os.Getenv(varName)
	})
}

func (c *Config) validate() error {
	if c.Source.Type != "github" {
		return fmt.Errorf("source type must be 'github', got %q", c.Source.Type)
	}

	if c.Source.Token == "" {
		return fmt.Errorf("source token is required")
	}

	if len(c.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	for i, t := range c.Targets {
		if t.Name == "" {
			return fmt.Errorf("target[%d].name is required", i)
		}
		if t.Type != "forgejo" {
			return fmt.Errorf("target[%d].type must be 'forgejo', got %q", i, t.Type)
		}
		if t.URL == "" {
			return fmt.Errorf("target[%d].url is required", i)
		}
		if t.Token == "" {
			return fmt.Errorf("target[%d].token is required", i)
		}
		if t.MirrorType != "" && t.MirrorType != "push" && t.MirrorType != "pull" {
			return fmt.Errorf("target[%d].mirror_type must be 'push' or 'pull', got %q", i, t.MirrorType)
		}
		if t.MirrorType == "" {
			c.Targets[i].MirrorType = "push"
		}
	}

	return nil
}

func Parse(data string) (*Config, error) {
	expanded := expandEnvVars(data)

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (f *Filters) HasRepoFilter() bool {
	return len(f.Repos) > 0
}

func (f *Filters) HasOrgFilter() bool {
	return len(f.Orgs) > 0
}

func ParseRepoSpec(spec string) (owner, repo string, isWildcard bool) {
	parts := strings.SplitN(spec, "/", 2)
	if len(parts) != 2 {
		return parts[0], "", false
	}
	return parts[0], parts[1], parts[1] == "*"
}
