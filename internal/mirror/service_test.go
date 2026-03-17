package mirror

import (
	"context"
	"errors"
	"testing"

	"github.com/guilycst/gitmeout/internal/config"
	"github.com/guilycst/gitmeout/internal/source/github"
)

type mockSource struct {
	userRepos  []github.Repository
	orgRepos   map[string][]github.Repository
	ownerRepos map[string][]github.Repository
	repo       map[string]*github.Repository
	user       string
	err        error
}

func (m *mockSource) ListUserRepos(ctx context.Context) ([]github.Repository, error) {
	return m.userRepos, m.err
}

func (m *mockSource) ListOrgRepos(ctx context.Context, org string) ([]github.Repository, error) {
	return m.orgRepos[org], m.err
}

func (m *mockSource) ListOwnerRepos(ctx context.Context, owner string) ([]github.Repository, error) {
	return m.ownerRepos[owner], m.err
}

func (m *mockSource) GetRepo(ctx context.Context, owner, name string) (*github.Repository, error) {
	key := owner + "/" + name
	return m.repo[key], m.err
}

func (m *mockSource) GetAuthenticatedUser(ctx context.Context) (string, error) {
	return m.user, m.err
}

type mockTarget struct {
	name       string
	existing   map[string]bool
	created    []string
	cloneURL   string
	mirrorType string
	err        error
}

func (m *mockTarget) Name() string {
	return m.name
}

func (m *mockTarget) MirrorType() string {
	if m.mirrorType == "" {
		return "push"
	}
	return m.mirrorType
}

func (m *mockTarget) Exists(ctx context.Context, repo Repository) (bool, error) {
	return m.existing[repo.FullName], m.err
}

func (m *mockTarget) CreateRepo(ctx context.Context, repo Repository) error {
	m.created = append(m.created, repo.FullName)
	return m.err
}

func (m *mockTarget) MigrateRepo(ctx context.Context, repo Repository) error {
	m.created = append(m.created, repo.FullName)
	return m.err
}

func (m *mockTarget) GetCloneURL(ctx context.Context, repo Repository) (string, error) {
	if m.cloneURL != "" {
		return m.cloneURL, nil
	}
	return "https://example.com/" + repo.Name, nil
}

func (m *mockTarget) GetAuthToken() string {
	return "test-token"
}

func (m *mockTarget) CreateMirror(ctx context.Context, repo Repository) error {
	m.created = append(m.created, repo.FullName)
	return m.err
}

func TestService_Run(t *testing.T) {
	tests := []struct {
		name        string
		source      *mockSource
		target      *mockTarget
		filters     config.Filters
		wantCreated []string
		wantErr     bool
	}{
		{
			name: "mirror new repo",
			source: &mockSource{
				userRepos: []github.Repository{
					{Owner: "user", Name: "repo1", FullName: "user/repo1", CloneURL: "https://github.com/user/repo1.git"},
				},
			},
			target:      &mockTarget{name: "forgejo", existing: map[string]bool{}},
			filters:     config.Filters{Personal: true},
			wantCreated: []string{"user/repo1"},
			wantErr:     false,
		},
		{
			name: "skip existing mirror",
			source: &mockSource{
				userRepos: []github.Repository{
					{Owner: "user", Name: "repo1", FullName: "user/repo1", CloneURL: "https://github.com/user/repo1.git"},
				},
			},
			target:      &mockTarget{name: "forgejo", existing: map[string]bool{"user/repo1": true}},
			filters:     config.Filters{Personal: true},
			wantCreated: []string{},
			wantErr:     false,
		},
		{
			name: "explicit repo list",
			source: &mockSource{
				repo: map[string]*github.Repository{
					"owner/repo1": {Owner: "owner", Name: "repo1", FullName: "owner/repo1", CloneURL: "https://github.com/owner/repo1.git"},
				},
			},
			target:      &mockTarget{name: "forgejo", existing: map[string]bool{}},
			filters:     config.Filters{Repos: []string{"owner/repo1"}},
			wantCreated: []string{"owner/repo1"},
			wantErr:     false,
		},
		{
			name: "wildcard repo spec",
			source: &mockSource{
				ownerRepos: map[string][]github.Repository{
					"owner": {
						{Owner: "owner", Name: "repo1", FullName: "owner/repo1", CloneURL: "https://github.com/owner/repo1.git"},
						{Owner: "owner", Name: "repo2", FullName: "owner/repo2", CloneURL: "https://github.com/owner/repo2.git"},
					},
				},
			},
			target:      &mockTarget{name: "forgejo", existing: map[string]bool{}},
			filters:     config.Filters{Repos: []string{"owner/*"}},
			wantCreated: []string{"owner/repo1", "owner/repo2"},
			wantErr:     false,
		},
		{
			name: "org repos",
			source: &mockSource{
				orgRepos: map[string][]github.Repository{
					"my-org": {
						{Owner: "my-org", Name: "org-repo", FullName: "my-org/org-repo", CloneURL: "https://github.com/my-org/org-repo.git"},
					},
				},
			},
			target:      &mockTarget{name: "forgejo", existing: map[string]bool{}},
			filters:     config.Filters{Personal: false, Orgs: []string{"my-org"}},
			wantCreated: []string{"my-org/org-repo"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.source, []Target{tt.target}, "test-token")
			err := svc.Run(context.Background(), tt.filters)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tt.target.created) != len(tt.wantCreated) {
				t.Errorf("Run() created %d repos, want %d", len(tt.target.created), len(tt.wantCreated))
				return
			}

			for i, want := range tt.wantCreated {
				if tt.target.created[i] != want {
					t.Errorf("Run() created[%d] = %q, want %q", i, tt.target.created[i], want)
				}
			}
		})
	}
}

func TestService_Run_TargetError(t *testing.T) {
	source := &mockSource{
		userRepos: []github.Repository{
			{Owner: "user", Name: "repo1", FullName: "user/repo1", CloneURL: "https://github.com/user/repo1.git"},
		},
	}
	target := &mockTarget{
		name:     "forgejo",
		existing: map[string]bool{},
		err:      errors.New("connection failed"),
	}

	svc := NewService(source, []Target{target}, "test-token")
	err := svc.Run(context.Background(), config.Filters{Personal: true})

	if err != nil {
		t.Errorf("Run() expected no error with continue-on-error behavior, got: %v", err)
	}
	if len(target.created) != 0 {
		t.Errorf("expected no repos created, got %d", len(target.created))
	}
}
