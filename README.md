# gitmeout

> Mirror GitHub repositories to Forgejo/Codeberg instances

`gitmeout` is a Go CLI tool that reads repositories from a source Git provider (GitHub) and mirrors them to one or more target Forgejo/Codeberg instances. Designed to run periodically via cron or Kubernetes CronJobs.

## Features

- **GitHub as source** - Mirror personal and organization repositories
- **Forgejo/Codeberg as targets** - Create mirror repos on any Forgejo instance
- **Flexible filtering** - Filter by personal repos, orgs, or explicit repo list
- **Wildcard support** - Use `owner/*` to mirror all repos from an owner
- **Idempotent** - Skip existing mirrors, only create new ones
- **Token-only auth** - Secure authentication via Personal Access Tokens
- **Container-ready** - Docker image and Kubernetes manifests included

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
