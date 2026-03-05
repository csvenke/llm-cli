package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	askcmd "llm/internal/cmd/ask"
	commitcmd "llm/internal/cmd/commit"
	prcmd "llm/internal/cmd/gh/pr"
	"llm/internal/gh"
	"llm/internal/git"
	"llm/internal/providers"
)

type stubProvider struct{}

func (s *stubProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	return "", nil
}

type stubGitClient struct{}

func (s *stubGitClient) GetStagedDiff() (string, error) {
	return "", nil
}

func (s *stubGitClient) GetDiffFromRevision(revision string) (string, error) {
	return "", nil
}

func (s *stubGitClient) GetCurrentBranch() (string, error) {
	return "", nil
}

func (s *stubGitClient) HasParentCommit() (bool, error) {
	return false, nil
}

func (s *stubGitClient) Commit(msg string, amend bool) error {
	return nil
}

type stubGHClient struct{}

func (s *stubGHClient) GetCurrentBranch() (string, error) {
	return "", nil
}

func (s *stubGHClient) GetDefaultBranch() (string, error) {
	return "", nil
}

func (s *stubGHClient) GetMergeBase(base, head string) (string, error) {
	return "", nil
}

func (s *stubGHClient) GetDiffRange(from, to string) (string, error) {
	return "", nil
}

func TestUsageIncludesNestedGHSubcommands(t *testing.T) {
	var out bytes.Buffer
	UsageTo(&out)

	if !strings.Contains(out.String(), "gh <subcommand>") {
		t.Fatalf("usage output = %q, want to include %q", out.String(), "gh <subcommand>")
	}

	if !strings.Contains(out.String(), "pr                Create a GitHub pull request") {
		t.Fatalf("usage output = %q, want to include gh pr subcommand help", out.String())
	}
}

func TestRunRoutesAskAndCommit(t *testing.T) {
	originalAskRun := askcmd.RunFunc
	originalCommitRun := commitcmd.RunFunc
	t.Cleanup(func() {
		askcmd.RunFunc = originalAskRun
		commitcmd.RunFunc = originalCommitRun
	})

	provider := &stubProvider{}
	deps := Dependencies{
		Provider: provider,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
		Git:      &stubGitClient{},
		GH:       &stubGHClient{},
	}

	t.Run("routes ask", func(t *testing.T) {
		var gotArgs []string
		askcmd.RunFunc = func(ctx context.Context, gotProvider providers.Provider, output io.Writer, stderr io.Writer, args []string) error {
			if gotProvider != provider {
				t.Fatalf("ask provider = %#v, want %#v", gotProvider, provider)
			}
			gotArgs = args
			return nil
		}

		err := defaultRegistry.Run(context.Background(), deps, []string{"ask", "what", "is", "go"})
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}

		if !reflect.DeepEqual(gotArgs, []string{"what", "is", "go"}) {
			t.Fatalf("ask args = %v, want %v", gotArgs, []string{"what", "is", "go"})
		}
	})

	t.Run("routes commit", func(t *testing.T) {
		var gotArgs []string
		commitcmd.RunFunc = func(ctx context.Context, gotProvider providers.Provider, gitClient git.Client, stderr io.Writer, args []string) error {
			if gotProvider != provider {
				t.Fatalf("commit provider = %#v, want %#v", gotProvider, provider)
			}
			gotArgs = args
			return nil
		}

		err := defaultRegistry.Run(context.Background(), deps, []string{"commit", "--amend"})
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}

		if !reflect.DeepEqual(gotArgs, []string{"--amend"}) {
			t.Fatalf("commit args = %v, want %v", gotArgs, []string{"--amend"})
		}
	})
}

func TestRunGHSubcommands(t *testing.T) {
	originalGHRun := prcmd.RunFunc
	originalGHCreatePullRequest := prcmd.CreatePullRequestFunc
	originalEnsureBranchPushed := prcmd.EnsureBranchPushedFunc
	t.Cleanup(func() {
		prcmd.RunFunc = originalGHRun
		prcmd.CreatePullRequestFunc = originalGHCreatePullRequest
		prcmd.EnsureBranchPushedFunc = originalEnsureBranchPushed
	})

	deps := Dependencies{
		Provider: &stubProvider{},
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
		Git:      &stubGitClient{},
		GH:       &stubGHClient{},
	}

	tests := []struct {
		name          string
		args          []string
		runErr        error
		createErr     error
		pushErr       error
		createOutput  string
		wantErr       bool
		wantErrSubstr string
		wantOutput    string
		wantRunArgs   []string
	}{
		{
			name:         "runs gh pr and prints command output",
			args:         []string{"gh", "pr"},
			createOutput: "https://github.com/example/repo/pull/123",
			wantOutput:   "https://github.com/example/repo/pull/123\n",
			wantRunArgs:  []string{},
		},
		{
			name:          "fails when gh subcommand is missing",
			args:          []string{"gh"},
			wantErr:       true,
			wantErrSubstr: "usage: llm gh <subcommand>",
		},
		{
			name:          "fails when gh subcommand is unknown",
			args:          []string{"gh", "issue"},
			wantErr:       true,
			wantErrSubstr: "unknown gh subcommand",
		},
		{
			name:          "fails when pr receives trailing args",
			args:          []string{"gh", "pr", "--extra"},
			wantErr:       true,
			wantErrSubstr: "usage: llm gh pr",
		},
		{
			name:          "returns generation error",
			args:          []string{"gh", "pr"},
			runErr:        errors.New("generation failed"),
			wantErr:       true,
			wantErrSubstr: "generation failed",
		},
		{
			name:          "returns push error",
			args:          []string{"gh", "pr"},
			pushErr:       errors.New("push failed"),
			wantErr:       true,
			wantErrSubstr: "push failed",
		},
		{
			name:          "returns create error",
			args:          []string{"gh", "pr"},
			createErr:     errors.New("gh pr create failed"),
			wantErr:       true,
			wantErrSubstr: "gh pr create failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotRunArgs []string
			var gotCreatedPR *gh.PullRequest
			var stdout bytes.Buffer
			deps.Stdout = &stdout

			prcmd.RunFunc = func(ctx context.Context, provider providers.Provider, client gh.Client, stderr io.Writer, args []string) (*gh.PullRequest, error) {
				gotRunArgs = args
				if tt.runErr != nil {
					return nil, tt.runErr
				}
				return &gh.PullRequest{Title: "Add gh pr command", Body: "## Summary\n- Add command wiring"}, nil
			}

			prcmd.CreatePullRequestFunc = func(pr *gh.PullRequest) (string, error) {
				gotCreatedPR = pr
				if tt.createErr != nil {
					return "", tt.createErr
				}
				return tt.createOutput, nil
			}

			prcmd.EnsureBranchPushedFunc = func(client gh.Client) error {
				return tt.pushErr
			}

			err := defaultRegistry.Run(context.Background(), deps, tt.args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.wantErrSubstr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr)) {
					t.Errorf("Run() error = %v, want substring %q", err, tt.wantErrSubstr)
				}
				return
			}

			if !reflect.DeepEqual(gotRunArgs, tt.wantRunArgs) {
				t.Errorf("Run() args passed to gh.Run = %v, want %v", gotRunArgs, tt.wantRunArgs)
			}

			wantCreatedPR := &gh.PullRequest{Title: "Add gh pr command", Body: "## Summary\n- Add command wiring"}
			if !reflect.DeepEqual(gotCreatedPR, wantCreatedPR) {
				t.Errorf("Run() PR passed to CreatePullRequest = %#v, want %#v", gotCreatedPR, wantCreatedPR)
			}

			if stdout.String() != tt.wantOutput {
				t.Errorf("Run() stdout = %q, want %q", stdout.String(), tt.wantOutput)
			}
		})
	}
}

func TestRunReturnsUnknownCommandForUnknownTopLevel(t *testing.T) {
	err := defaultRegistry.Run(context.Background(), Dependencies{}, []string{"unknown"})
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("Run() error = %v, want ErrUnknownCommand", err)
	}
}
