# gitmeout

[![CI](https://github.com/guilycst/gitmeout/actions/workflows/ci.yml/badge.svg)](https://github.com/guilycst/gitmeout/actions/workflows/ci.yml)
[![Release](https://github.com/guilycst/gitmeout/actions/workflows/release.yml/badge.svg)](https://github.com/guilycst/gitmeout/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/guilycst/gitmeout)](https://goreportcard.com/report/github.com/guilycst/gitmeout)
[![GoDoc](https://godoc.org/github.com/guilycst/gitmeout?status.svg)](https://godoc.org/github.com/guilycst/gitmeout)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)

**Container Images:**
[![GHCR](https://img.shields.io/badge/ghcr.io-guilycst/gitmeout-blue)](https://github.com/guilycst/gitmeout/pkgs/container/gitmeout)
[![Forgejo](https://img.shields.io/badge/code.decastro.me-guilherme/gitmeout-blue)](https://code.decastro.me/guilherme/-/packages/container/gitmeout)

> Mirror GitHub repositories to Forgejo/Codeberg instances

A fast, reliable CLI tool for mirroring repositories from GitHub to Forgejo/Codeberg. Supports both push and pull mirror modes, with flexible filtering and container-ready deployment.

**Topics:** `github` `forgejo` `codeberg` `mirror` `backup` `git` `golang` `cli` `kubernetes` `docker`

## Features

- **GitHub as source** - Mirror personal and organization repositories
- **Forgejo/Codeberg as targets** - Create mirror repos on any Forgejo instance
- **Flexible filtering** - Filter by personal repos, orgs, or explicit repo list
- **Wildcard support** - Use `owner/*` to mirror all repos from an owner
- **Dual mirror types** - Push mirrors (clone + push) or pull mirrors (Forgejo MigrateRepo API)
- **Token-only auth** - Secure authentication via Personal Access Tokens
- **Container-ready** - Docker image and Kubernetes manifests included
- **Multiarch support** - Linux amd64/arm64 binaries and Docker images

## ⚠️ Disclaimer

**Use this tool responsibly and only for repositories you own or have permission to mirror.**

- **Personal Use**: This tool is designed for mirroring your own repositories or repositories you have explicit permission to mirror.
- **GitHub ToS**: Ensure your usage complies with [GitHub's Terms of Service](https://docs.github.com/en/site-policy/github-terms/github-terms-of-service), including API rate limits and acceptable use policies.
- **Rate Limits**: GitHub's API has rate limits (5,000 authenticated requests/hour). Be mindful when mirroring large numbers of repositories.
- **Data Portability**: GitHub's ToS supports user data portability for your own content.
- **No Warranty**: This software is provided "as is" without warranty of any kind. The authors are not responsible for any issues arising from its use.

By using this tool, you agree to:
- Only mirror repositories you own or have permission to mirror
- Comply with GitHub's API Terms of Service
- Respect rate limits and not abuse the API
- Use it for legitimate data portability purposes

## Quick Start

### 1. Create a config file

```yaml
source:
  type: github
  token: ${GITHUB_TOKEN}
  filters:
    personal: true
    orgs:
      - my-org
    repos:
      - owner/repo1
      - owner/*

targets:
  - name: codeberg
    type: forgejo
    url: https://codeberg.org
    token: ${CODEBERG_TOKEN}
```

### 2. Set environment variables

```bash
export GITHUB_TOKEN="ghp_xxx"
export CODEBERG_TOKEN="xxx"
```

### 3. Run

```bash
# Using Just
just run

# Or directly
go run ./cmd/gitmeout

# Or with Docker
docker run --rm \
  -e GITHUB_TOKEN \
  -e CODEBERG_TOKEN \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  ghcr.io/guilycst/gitmeout:latest
```

## Configuration

### Source Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Provider type. Currently only `github` |
| `token` | string | Yes | GitHub Personal Access Token. Use `${VAR}` for env vars |
| `filters.personal` | bool | No | Mirror personal repos. Default: `true` |
| `filters.orgs` | []string | No | List of organization names to mirror |
| `filters.repos` | []string | No | Explicit repo list. Use `owner/*` for all repos from an owner |

### Target Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Friendly name for logging |
| `type` | string | Yes | Provider type. Currently only `forgejo` |
| `url` | string | Yes | Base URL of the Forgejo instance |
| `token` | string | Yes | Forgejo Personal Access Token. Use `${VAR}` for env vars |
| `mirror_type` | string | No | `push` (default) or `pull` - see Mirror Types below |

### Mirror Types

**Push Mirrors** (`mirror_type: push`):
- Clones the repository from GitHub, then pushes `--mirror` to the target
- Syncs every run - changes on source are pushed to target
- Use for: Codeberg (pull mirrors disabled by admin)

**Pull Mirrors** (`mirror_type: pull`):
- Uses Forgejo's MigrateRepo API to create a pull mirror
- Forgejo periodically fetches changes from the source automatically
- Use for: Self-hosted Forgejo instances where pull mirrors are enabled
- Skips if a mirror already exists

### Environment Variable Interpolation

Use `${VAR_NAME}` syntax in config values to reference environment variables:

```yaml
source:
  token: ${GITHUB_TOKEN}  # Reads from $GITHUB_TOKEN env var
```

## Behavior

### Repository Resolution

1. **Personal repos**: If `personal: true`, fetches all repos owned by the authenticated user
2. **Organization repos**: For each org in `orgs`, fetches all repos the user can access
3. **Explicit repos**: Resolves each entry in `repos`:
   - `owner/repo-name`: Specific repo
   - `owner/*`: All repos from that owner

### Mirror Creation

- **Existing mirrors**: If a mirror repo already exists on the target, it is skipped
- **New mirrors**: Creates a new mirror repo pointing to the source

### Required Token Permissions

**GitHub**:
- `repo` (for private repos)
- `read:org` (if filtering by orgs)

**Forgejo/Codeberg**:
- `read:user` - required to get authenticated user info
- `write:repository` - required to create mirror repositories

## Deployment

### Docker

```bash
docker build -t gitmeout:latest .
docker run --rm \
  -e GITHUB_TOKEN \
  -e CODEBERG_TOKEN \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  gitmeout:latest
```

### Kubernetes

Create a secret for your config:

```bash
kubectl create secret generic gitmeout-config --from-file=config.yaml=config.yaml
```

Apply the CronJob:

```bash
kubectl apply -f deploy/k8s/cronjob.yaml
```

See `deploy/k8s/` for Job and CronJob examples.

### Kubernetes Custom Resource

For GitOps workflows, see `deploy/k8s/job.yaml` for a Job CR that can be used with tools like ArgoCD or Flux.

## Development

```bash
# Install dependencies
go mod download

# Build
just build

# Run tests
just test

# Lint
just lint

# Format
just fmt

# Run all checks
just validate
```

## License

MIT License - see [LICENSE](LICENSE) for details.
