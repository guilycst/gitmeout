package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/guilycst/gitmeout/internal/config"
	"github.com/guilycst/gitmeout/internal/mirror"
	"github.com/guilycst/gitmeout/internal/source/github"
	"github.com/guilycst/gitmeout/internal/target/forgejo"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting gitmeout", "version", version)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	sourceClient, err := github.NewClient(cfg.Source.Token)
	if err != nil {
		slog.Error("failed to create github client", "error", err)
		os.Exit(1)
	}

	var targets []mirror.Target
	for _, t := range cfg.Targets {
		client, err := forgejo.NewClient(t.URL, t.Token)
		if err != nil {
			slog.Error("failed to create forgejo client", "error", err, "target", t.Name)
			os.Exit(1)
		}
		targets = append(targets, &forgejoTarget{
			name:       t.Name,
			client:     client,
			mirrorType: t.MirrorType,
		})
	}

	svc := mirror.NewService(sourceClient, targets, cfg.Source.Token)

	err = svc.Run(ctx, cfg.Source.Filters)
	if err != nil {
		if ctx.Err() != nil {
			slog.Info("mirroring cancelled by user", "signal", "SIGINT/SIGTERM")
		} else {
			slog.Error("mirroring failed", "error", err)
			os.Exit(1)
		}
		return
	}

	slog.Info("mirroring completed successfully")
}

type forgejoTarget struct {
	name       string
	client     *forgejo.Client
	mirrorType string
}

func (t *forgejoTarget) Name() string {
	return t.name
}

func (t *forgejoTarget) MirrorType() string {
	return t.mirrorType
}

func (t *forgejoTarget) CreateRepo(ctx context.Context, repo mirror.Repository) error {
	return t.client.CreateRepo(ctx, repo)
}

func (t *forgejoTarget) MigrateRepo(ctx context.Context, repo mirror.Repository) error {
	return t.client.MigrateRepo(ctx, repo)
}

func (t *forgejoTarget) Exists(ctx context.Context, repo mirror.Repository) (bool, error) {
	return t.client.Exists(ctx, repo)
}

func (t *forgejoTarget) GetCloneURL(ctx context.Context, repo mirror.Repository) (string, error) {
	return t.client.GetCloneURL(ctx, repo)
}

func (t *forgejoTarget) GetAuthToken() string {
	return t.client.GetAuthToken()
}
