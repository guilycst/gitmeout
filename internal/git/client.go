package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Client struct {
	tempDir string
}

func NewClient() (*Client, error) {
	tempDir, err := os.MkdirTemp("", "gitmeout-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	return &Client{tempDir: tempDir}, nil
}

func (c *Client) Close() error {
	if c.tempDir != "" {
		return os.RemoveAll(c.tempDir)
	}
	return nil
}

func (c *Client) Clone(ctx context.Context, url, token, dest string) error {
	authURL := injectToken(url, token)

	cmd := exec.CommandContext(ctx, "git", "clone", "--bare", authURL, dest)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("clone cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("failed to clone %s: %w\n%s", url, err, string(output))
	}
	return nil
}

func (c *Client) AddRemote(ctx context.Context, repoPath, name, url, token string) error {
	authURL := injectToken(url, token)

	cmd := exec.CommandContext(ctx, "git", "remote", "add", name, authURL)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("add remote cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("failed to add remote: %w\n%s", err, string(output))
	}
	return nil
}

func (c *Client) PushMirror(ctx context.Context, repoPath, remote string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "--mirror", remote)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("push cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("failed to push mirror: %w\n%s", err, string(output))
	}
	return nil
}

func (c *Client) CloneAndPush(ctx context.Context, srcURL, srcToken, destURL, destToken, repoName string) error {
	repoPath := filepath.Join(c.tempDir, repoName)

	if err := c.Clone(ctx, srcURL, srcToken, repoPath); err != nil {
		return err
	}

	if err := c.AddRemote(ctx, repoPath, "target", destURL, destToken); err != nil {
		return err
	}

	if err := c.PushMirror(ctx, repoPath, "target"); err != nil {
		return err
	}

	return nil
}

func injectToken(url, token string) string {
	if token == "" {
		return url
	}

	if strings.HasPrefix(url, "https://") {
		return strings.Replace(url, "https://", "https://"+token+"@", 1)
	}
	if strings.HasPrefix(url, "http://") {
		return strings.Replace(url, "http://", "http://"+token+"@", 1)
	}
	return url
}
