package mirror

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/guilycst/gitmeout/internal/config"
	"github.com/guilycst/gitmeout/internal/git"
	"github.com/guilycst/gitmeout/internal/source/github"
)

type Repository struct {
	Owner         string
	Name          string
	FullName      string
	CloneURL      string
	DefaultBranch string
	Private       bool
	AuthToken     string
}

type Source interface {
	ListUserRepos(ctx context.Context) ([]github.Repository, error)
	ListOrgRepos(ctx context.Context, org string) ([]github.Repository, error)
	ListOwnerRepos(ctx context.Context, owner string) ([]github.Repository, error)
	GetRepo(ctx context.Context, owner, name string) (*github.Repository, error)
	GetAuthenticatedUser(ctx context.Context) (string, error)
}

type Target interface {
	Name() string
	MirrorType() string // "push" or "pull"
	CreateRepo(ctx context.Context, repo Repository) error
	MigrateRepo(ctx context.Context, repo Repository) error
	Exists(ctx context.Context, repo Repository) (bool, error)
	GetCloneURL(ctx context.Context, repo Repository) (string, error)
	GetAuthToken() string
}

type Service struct {
	source     Source
	targets    []Target
	sourceAuth string
}

func NewService(source Source, targets []Target, sourceAuth string) *Service {
	return &Service{
		source:     source,
		targets:    targets,
		sourceAuth: sourceAuth,
	}
}

func (s *Service) Run(ctx context.Context, filters config.Filters) error {
	repos, err := s.resolveRepositories(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to resolve repositories: %w", err)
	}

	slog.Info("resolved repositories", "count", len(repos))

	for _, repo := range repos {
		select {
		case <-ctx.Done():
			slog.Info("mirroring interrupted", "reason", ctx.Err())
			return ctx.Err()
		default:
		}

		if err := s.mirrorToTargets(ctx, repo); err != nil {
			slog.Error("failed to mirror repository", "repo", repo.FullName, "error", err)
			continue
		}
	}

	return nil
}

func (s *Service) resolveRepositories(ctx context.Context, filters config.Filters) ([]Repository, error) {
	var repos []Repository
	seen := make(map[string]bool)

	addRepo := func(r github.Repository) {
		if !seen[r.FullName] {
			seen[r.FullName] = true
			repos = append(repos, Repository{
				Owner:         r.Owner,
				Name:          r.Name,
				FullName:      r.FullName,
				CloneURL:      r.CloneURL,
				DefaultBranch: r.DefaultBranch,
				Private:       r.Private,
				AuthToken:     s.sourceAuth,
			})
		}
	}

	if filters.HasRepoFilter() {
		for _, spec := range filters.Repos {
			owner, repo, isWildcard := config.ParseRepoSpec(spec)

			if isWildcard {
				ownerRepos, err := s.source.ListOwnerRepos(ctx, owner)
				if err != nil {
					return nil, fmt.Errorf("failed to list repos for %s: %w", owner, err)
				}
				for _, r := range ownerRepos {
					addRepo(r)
				}
			} else {
				r, err := s.source.GetRepo(ctx, owner, repo)
				if err != nil {
					return nil, fmt.Errorf("failed to get repo %s: %w", spec, err)
				}
				addRepo(*r)
			}
		}
		return repos, nil
	}

	if filters.Personal {
		userRepos, err := s.source.ListUserRepos(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list personal repos: %w", err)
		}
		for _, r := range userRepos {
			addRepo(r)
		}
	}

	for _, org := range filters.Orgs {
		orgRepos, err := s.source.ListOrgRepos(ctx, org)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos for org %s: %w", org, err)
		}
		for _, r := range orgRepos {
			addRepo(r)
		}
	}

	return repos, nil
}

func (s *Service) mirrorToTargets(ctx context.Context, repo Repository) error {
	for _, target := range s.targets {
		exists, err := target.Exists(ctx, repo)
		if err != nil {
			return fmt.Errorf("failed to check if repo exists on %s: %w", target.Name(), err)
		}

		slog.Info("syncing mirror",
			"repo", repo.FullName,
			"target", target.Name(),
			"mirror_type", target.MirrorType(),
			"exists", exists)

		if target.MirrorType() == "pull" {
			if err := s.pullMirror(ctx, target, repo, exists); err != nil {
				return fmt.Errorf("failed to sync mirror on %s: %w", target.Name(), err)
			}
		} else {
			if err := s.pushMirror(ctx, target, repo, exists); err != nil {
				return fmt.Errorf("failed to sync mirror on %s: %w", target.Name(), err)
			}
		}

		slog.Info("mirror synced successfully",
			"repo", repo.FullName,
			"target", target.Name())
	}

	return nil
}

func (s *Service) pullMirror(ctx context.Context, target Target, repo Repository, exists bool) error {
	if exists {
		slog.Info("pull mirror already exists, skipping",
			"repo", repo.FullName,
			"target", target.Name())
		return nil
	}

	if err := target.MigrateRepo(ctx, repo); err != nil {
		return fmt.Errorf("failed to migrate repo: %w", err)
	}

	return nil
}

func (s *Service) pushMirror(ctx context.Context, target Target, repo Repository, exists bool) error {
	if !exists {
		if err := target.CreateRepo(ctx, repo); err != nil {
			return fmt.Errorf("failed to create repo: %w", err)
		}
	}

	destURL, err := target.GetCloneURL(ctx, repo)
	if err != nil {
		return fmt.Errorf("failed to get clone URL: %w", err)
	}

	gitClient, err := git.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create git client: %w", err)
	}
	defer gitClient.Close()

	destToken := target.GetAuthToken()
	if err := gitClient.CloneAndPush(ctx, repo.CloneURL, repo.AuthToken, destURL, destToken, repo.Name); err != nil {
		return fmt.Errorf("failed to push mirror: %w", err)
	}

	return nil
}
