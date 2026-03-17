package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v69/github"
)

type Client struct {
	client *gh.Client
}

func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	client := gh.NewClient(nil).WithAuthToken(token)
	return &Client{client: client}, nil
}

type Repository struct {
	Owner         string
	Name          string
	FullName      string
	CloneURL      string
	SSHURL        string
	Private       bool
	DefaultBranch string
}

func (c *Client) ListUserRepos(ctx context.Context) ([]Repository, error) {
	opt := &gh.RepositoryListByAuthenticatedUserOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	var repos []Repository
	for {
		ghRepos, resp, err := c.client.Repositories.ListByAuthenticatedUser(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list user repos: %w", err)
		}

		for _, r := range ghRepos {
			repos = append(repos, toRepository(r))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return repos, nil
}

func (c *Client) ListOrgRepos(ctx context.Context, org string) ([]Repository, error) {
	opt := &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	var repos []Repository
	for {
		ghRepos, resp, err := c.client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list org repos for %s: %w", org, err)
		}

		for _, r := range ghRepos {
			repos = append(repos, toRepository(r))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return repos, nil
}

func (c *Client) ListOwnerRepos(ctx context.Context, owner string) ([]Repository, error) {
	opt := &gh.RepositoryListByUserOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	var repos []Repository
	for {
		ghRepos, resp, err := c.client.Repositories.ListByUser(ctx, owner, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos for owner %s: %w", owner, err)
		}

		for _, r := range ghRepos {
			repos = append(repos, toRepository(r))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return repos, nil
}

func (c *Client) GetRepo(ctx context.Context, owner, name string) (*Repository, error) {
	repo, _, err := c.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo %s/%s: %w", owner, name, err)
	}

	r := toRepository(repo)
	return &r, nil
}

func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user, _, err := c.client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	return user.GetLogin(), nil
}

func toRepository(r *gh.Repository) Repository {
	return Repository{
		Owner:         r.GetOwner().GetLogin(),
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		CloneURL:      r.GetCloneURL(),
		SSHURL:        r.GetSSHURL(),
		Private:       r.GetPrivate(),
		DefaultBranch: r.GetDefaultBranch(),
	}
}
