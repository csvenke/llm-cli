package commit

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

type stubProvider struct {
	resp string
	err  error
}

func (s *stubProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	return s.resp, s.err
}

type stubGitClient struct {
	diff      string
	branch    string
	hasParent bool
	commitErr error
	diffErr   error
	branchErr error
}

func (s *stubGitClient) GetStagedDiff() (string, error) {
	return s.diff, s.diffErr
}

func (s *stubGitClient) GetDiffFromRevision(revision string) (string, error) {
	return s.diff, s.diffErr
}

func (s *stubGitClient) GetCurrentBranch() (string, error) {
	return s.branch, s.branchErr
}

func (s *stubGitClient) HasParentCommit() (bool, error) {
	return s.hasParent, nil
}

func (s *stubGitClient) Commit(msg string, amend bool) error {
	return s.commitErr
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantAmend bool
	}{
		{
			name:      "no flags",
			args:      []string{},
			wantAmend: false,
		},
		{
			name:      "with -a flag",
			args:      []string{"-a"},
			wantAmend: true,
		},
		{
			name:      "with --amend flag",
			args:      []string{"--amend"},
			wantAmend: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseConfig(tt.args)
			if err != nil {
				t.Errorf("ParseConfig() unexpected error: %v", err)
				return
			}
			if cfg.Amend != tt.wantAmend {
				t.Errorf("ParseConfig() cfg.Amend = %v, want %v", cfg.Amend, tt.wantAmend)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name   string
		diff   string
		branch string
		want   string
	}{
		{
			name:   "diff with JIRA issue in branch",
			diff:   "some diff content",
			branch: "feature/PROJ-123-add-feature",
			want:   "Branch: feature/PROJ-123-add-feature (Issue: PROJ-123)\n\nsome diff content",
		},
		{
			name:   "diff without issue in branch",
			diff:   "some diff content",
			branch: "feature/add-feature",
			want:   "some diff content",
		},
		{
			name:   "empty diff returns empty",
			diff:   "",
			branch: "main",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPrompt(tt.diff, tt.branch)
			if got != tt.want {
				t.Errorf("BuildPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		resp    string
		err     error
		want    string
		wantErr bool
	}{
		{
			name:    "successful message generation",
			prompt:  "test prompt",
			resp:    "feat: add new feature",
			want:    "feat: add new feature",
			wantErr: false,
		},
		{
			name:    "provider returns error",
			prompt:  "test prompt",
			err:     errors.New("API error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			provider := &stubProvider{resp: tt.resp, err: tt.err}

			msg, err := GenerateCommitMessage(context.Background(), provider, tt.prompt, &stderr)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCommitMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && msg != tt.want {
				t.Errorf("GenerateCommitMessage() = %q, want %q", msg, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		git     *stubGitClient
		resp    string
		provErr error
		wantErr bool
	}{
		{
			name: "successful commit flow",
			args: []string{},
			git: &stubGitClient{
				diff:   "some changes",
				branch: "feature/test",
			},
			resp:    "feat: add feature",
			wantErr: false,
		},
		{
			name: "successful commit with issue in branch",
			args: []string{},
			git: &stubGitClient{
				diff:   "some changes",
				branch: "feature/PROJ-123-fix",
			},
			resp:    "fix: resolve bug",
			wantErr: false,
		},
		{
			name: "no staged changes",
			args: []string{},
			git: &stubGitClient{
				diff: "",
			},
			wantErr: true,
		},
		{
			name: "no changes to amend",
			args: []string{"-a"},
			git: &stubGitClient{
				diff:      "",
				hasParent: true, // HEAD~1 exists
			},
			wantErr: true,
		},
		{
			name: "provider fails",
			args: []string{},
			git: &stubGitClient{
				diff:   "some changes",
				branch: "main",
			},
			provErr: errors.New("API error"),
			wantErr: true,
		},
		{
			name: "git commit fails",
			args: []string{},
			git: &stubGitClient{
				diff:      "some changes",
				branch:    "main",
				commitErr: errors.New("commit failed"),
			},
			resp:    "test commit",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			provider := &stubProvider{resp: tt.resp, err: tt.provErr}
			err := Run(context.Background(), provider, tt.git, &stderr, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
