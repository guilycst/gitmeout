# Agent Guidelines for gitmeout

## Overview

`gitmeout` is a Go CLI tool that mirrors repositories from a source Git provider (GitHub) to target Forgejo/Codeberg instances. It runs periodically via external schedulers (cron, Kubernetes CronJobs).

---

## Build, Lint, and Test Commands

### Using Just (recommended)

```bash
just build          # Build binary to bin/gitmeout
just run            # Run the application
just test           # Run all tests
just test-verbose   # Run tests with verbose output
just test-coverage  # Run tests with coverage report
just lint           # Run golangci-lint
just fmt            # Format code with gofmt
just vet            # Run go vet
just tidy           # Tidy go.mod
just docker-build   # Build Docker image
just docker-run     # Run Docker container locally
```

### Direct Go Commands

```bash
go build -o bin/gitmeout ./cmd/gitmeout
go run ./cmd/gitmeout
go test ./...
go test -v ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
golangci-lint run
gofmt -w .
go vet ./...
go mod tidy
```

### Running a Single Test

```bash
# Run a single test file
go test -v ./internal/config/...

# Run a specific test by name
go test -v -run TestConfig_Parse ./internal/config/...

# Run tests in a specific package
go test -v ./internal/mirror/...
```

---

## Project Structure

```
gitmeout/
├── cmd/
│   └── gitmeout/
│       └── main.go          # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go        # Config parsing and validation
│   │   └── config_test.go
│   ├── source/
│   │   └── github/
│   │       └── client.go    # GitHub API client
│   ├── target/
│   │   └── forgejo/
│   │       └── client.go    # Forgejo API client
│   └── mirror/
│       ├── service.go       # Core mirroring logic
│       └── service_test.go
├── deploy/
│   └── k8s/
│       ├── job.yaml         # K8s Job CR
│       └── cronjob.yaml     # K8s CronJob example
├── go.mod
├── go.sum
├── Justfile
├── Dockerfile
└── README.md
```

---

## Code Style Guidelines

### General Principles

- Follow standard Go conventions (Effective Go)
- Use `gofmt` for formatting
- Keep functions small and focused
- Prefer composition over inheritance
- Handle errors explicitly

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `config`, `mirror`)
- **Types**: PascalCase for exported, camelCase for unexported
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase or UPPER_SNAKE_CASE for exported constants
- **Interfaces**: typically end with `-er` (e.g., `RepositoryLister`, `Mirrorer`)

### Error Handling

- Always handle errors explicitly
- Wrap errors with context using `fmt.Errorf("operation failed: %w", err)`
- Use custom error types for domain-specific errors
- Never panic in library code

```go
func (s *Service) Mirror(ctx context.Context, repo *Repository) error {
    if err := s.target.CreateMirror(ctx, repo); err != nil {
        return fmt.Errorf("failed to create mirror for %s: %w", repo.Name, err)
    }
    return nil
}
```

### Logging

- Use structured logging (slog package from Go 1.21+)
- Log at appropriate levels: Debug, Info, Warn, Error
- Include relevant context in log entries

```go
logger.Info("mirroring repository", "repo", repo.Name, "target", target.Name)
```

### Testing

- Place tests in the same package with `_test.go` suffix
- Use table-driven tests for multiple test cases
- Use `t.Parallel()` where appropriate
- Aim for high coverage on core logic

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {name: "valid config", input: validYAML, want: validConfig, wantErr: false},
        {name: "missing token", input: missingTokenYAML, want: nil, wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConfig(strings.NewReader(tt.input))
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !cmp.Equal(got, tt.want) {
                t.Errorf("ParseConfig() = %v, want %v", got, cmp.Diff(got, tt.want))
            }
        })
    }
}
```

### Imports

Group imports in this order:
1. Standard library
2. External packages
3. Internal packages

```go
import (
    "context"
    "fmt"

    "github.com/google/go-github/v69/github"
    "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"

    "github.com/guilycst/gitmeout/internal/config"
)
```

---

## Configuration

- Config file: `config.yaml` (mounted as secret in K8s)
- Environment variable interpolation: `${VAR_NAME}` syntax
- Required: source token, all target tokens

---

## Git Workflow

### Commit Messages

Follow conventional commits:

```
<type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, test, chore

Examples:
feat(source): add support for filtering by org membership
fix(mirror): skip existing mirrors instead of failing
docs(readme): update installation instructions
```

### Branch Naming

- Use kebab-case
- Prefix: `feature/`, `fix/`, `chore/`
- Example: `feature/add-gitlab-source`

---

## Additional Notes

- Run `just lint` and `just test` before submitting changes
- Ensure `gofmt` has been run on all Go files
- Update go.mod when adding dependencies (`go mod tidy`)
- Keep the README.md up to date with usage examples
