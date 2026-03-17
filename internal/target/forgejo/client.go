package forgejo

import (
	"context"
	"fmt"
	"sync"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	"github.com/guilycst/gitmeout/internal/mirror"
)

type Client struct {
	client  *forgejo.Client
	url     string
	token   string
	user    string
	once    sync.Once
	userErr error
}

func NewClient(url, token string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("url is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	client, err := forgejo.NewClient(url, forgejo.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("failed to create forgejo client: %w", err)
	}

	return &Client{client: client, url: url, token: token}, nil
}

func (c *Client) getUsername(ctx context.Context) (string, error) {
	c.once.Do(func() {
		user, _, err := c.client.GetMyUserInfo()
		if err != nil {
			c.userErr = fmt.Errorf("failed to get authenticated user: %w", err)
			return
		}
		c.user = user.UserName
	})
	return c.user, c.userErr
}

func (c *Client) URL() string {
	return c.url
}

func (c *Client) GetAuthToken() string {
	return c.token
}

func (c *Client) CreateRepo(ctx context.Context, repo mirror.Repository) error {
	_, err := c.getUsername(ctx)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	opts := forgejo.CreateRepoOption{
		Name:          repo.Name,
		Private:       repo.Private,
		AutoInit:      false,
		DefaultBranch: repo.DefaultBranch,
	}

	_, _, err = c.client.CreateRepo(opts)
	if err != nil {
		return fmt.Errorf("failed to create repo %s: %w", repo.Name, err)
	}

	return nil
}

func (c *Client) Exists(ctx context.Context, repo mirror.Repository) (bool, error) {
	username, err := c.getUsername(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get username: %w", err)
	}

	_, resp, err := c.client.GetRepo(username, repo.Name)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if repo exists: %w", err)
	}
	return true, nil
}

func (c *Client) GetCloneURL(ctx context.Context, repo mirror.Repository) (string, error) {
	username, err := c.getUsername(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}

	r, _, err := c.client.GetRepo(username, repo.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get repo: %w", err)
	}

	return r.CloneURL, nil
}

func (c *Client) MigrateRepo(ctx context.Context, repo mirror.Repository) error {
	username, err := c.getUsername(ctx)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	opts := forgejo.MigrateRepoOption{
		RepoName:     repo.Name,
		RepoOwner:    username,
		CloneAddr:    repo.CloneURL,
		AuthToken:    repo.AuthToken,
		Service:      forgejo.GitServiceGithub,
		Mirror:       true,
		Private:      repo.Private,
		Wiki:         true,
		Issues:       true,
		Milestones:   true,
		Labels:       true,
		PullRequests: true,
		Releases:     true,
	}

	_, _, err = c.client.MigrateRepo(opts)
	if err != nil {
		return fmt.Errorf("failed to migrate repo %s: %w", repo.Name, err)
	}

	return nil
}
