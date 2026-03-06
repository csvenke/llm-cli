package gh

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
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
	branch           string
	defaultBranch    string
	mergeBase        string
	diff             string
	mergeBaseFn      func(base, head string) (string, error)
	diffFn           func(from, to string) (string, error)
	branchErr        error
	defaultBranchErr error
	mergeErr         error
	diffErr          error
}

func (s *stubGitClient) GetCurrentBranch() (string, error) {
	return s.branch, s.branchErr
}

func (s *stubGitClient) GetDefaultBranch() (string, error) {
	return s.defaultBranch, s.defaultBranchErr
}

func (s *stubGitClient) GetMergeBase(base, head string) (string, error) {
	if s.mergeBaseFn != nil {
		return s.mergeBaseFn(base, head)
	}
	return s.mergeBase, s.mergeErr
}

func (s *stubGitClient) GetDiffRange(from, to string) (string, error) {
	if s.diffFn != nil {
		return s.diffFn(from, to)
	}
	return s.diff, s.diffErr
}

func TestBuildPrompt(t *testing.T) {
	got := BuildPrompt("diff content", "feature/test", "origin/main")
	want := "Branch: feature/test\nBase: origin/main\n\ndiff content"

	if got != want {
		t.Errorf("BuildPrompt() = %q, want %q", got, want)
	}
}

func TestGeneratePullRequest(t *testing.T) {
	tests := []struct {
		name    string
		resp    string
		err     error
		wantErr bool
		want    *PullRequest
	}{
		{
			name: "successfully parses title and body",
			resp: "<title>Add GH PR generation</title>\n<body>## Summary\n- Add PR generation workflow</body>",
			want: &PullRequest{
				Title: "Add GH PR generation",
				Body:  "## Summary\n- Add PR generation workflow",
			},
		},
		{
			name:    "provider error",
			err:     errors.New("provider failed"),
			wantErr: true,
		},
		{
			name:    "invalid response format",
			resp:    "just text",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			provider := &stubProvider{resp: tt.resp, err: tt.err}

			got, err := GeneratePullRequest(context.Background(), provider, "prompt", &stderr)

			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePullRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Title != tt.want.Title {
				t.Errorf("GeneratePullRequest() title = %q, want %q", got.Title, tt.want.Title)
			}

			if got.Body != tt.want.Body {
				t.Errorf("GeneratePullRequest() body = %q, want %q", got.Body, tt.want.Body)
			}
		})
	}
}

func TestRun(t *testing.T) {
	originalLookPath := lookPath
	t.Cleanup(func() {
		lookPath = originalLookPath
	})

	tests := []struct {
		name          string
		args          []string
		git           *stubGitClient
		providerResp  string
		providerErr   error
		lookPathErr   error
		wantErr       bool
		wantErrSubstr string
		want          *PullRequest
	}{
		{
			name: "successful PR generation",
			git: &stubGitClient{
				branch:        "feature/test",
				defaultBranch: "origin/main",
				mergeBase:     "abc123",
				diff:          "diff content",
			},
			providerResp: "<title>Add PR workflow</title>\n<body>## Summary\n- Add gh PR generation</body>",
			want: &PullRequest{
				Title: "Add PR workflow",
				Body:  "## Summary\n- Add gh PR generation",
			},
		},
		{
			name:          "fails when gh CLI is missing",
			lookPathErr:   errors.New("not found"),
			git:           &stubGitClient{},
			wantErr:       true,
			wantErrSubstr: "gh CLI is required",
		},
		{
			name: "fails when base branch cannot be resolved",
			git: &stubGitClient{
				defaultBranchErr: errors.New("no default branch"),
				mergeErr:         errors.New("no merge base"),
			},
			wantErr:       true,
			wantErrSubstr: "unable to resolve pull request base branch",
		},
		{
			name: "uses fallback base branch candidates",
			git: &stubGitClient{
				mergeBaseFn: func(base, head string) (string, error) {
					if base != "origin/main" || head != "HEAD" {
						return "", errors.New("unexpected merge-base range")
					}
					return "abc123", nil
				},
				diff: "diff content",
			},
			providerResp: "<title>Add PR workflow</title>\n<body>## Summary\n- Add gh PR generation</body>",
			want: &PullRequest{
				Title: "Add PR workflow",
				Body:  "## Summary\n- Add gh PR generation",
			},
		},
		{
			name: "fails when merge-base cannot be computed",
			git: &stubGitClient{
				defaultBranch: "origin/main",
				mergeErr:      errors.New("unrelated histories"),
			},
			wantErr:       true,
			wantErrSubstr: "unable to resolve pull request base branch",
		},
		{
			name: "fails when diff context is empty",
			git: &stubGitClient{
				defaultBranch: "origin/main",
				mergeBase:     "abc123",
				diff:          "",
			},
			wantErr:       true,
			wantErrSubstr: "no branch changes found against base branch",
		},
		{
			name: "uses base branch merge-base flow for diff context",
			git: func() *stubGitClient {
				base := "origin/main"
				mergeBase := "abc123"
				return &stubGitClient{
					branch:        "feature/test",
					defaultBranch: base,
					mergeBaseFn: func(base, head string) (string, error) {
						if base != "origin/main" || head != "HEAD" {
							return "", errors.New("unexpected merge-base range")
						}
						return mergeBase, nil
					},
					diffFn: func(from, to string) (string, error) {
						if from != mergeBase || to != "HEAD" {
							return "", errors.New("unexpected diff range")
						}
						return "diff content", nil
					},
				}
			}(),
			providerResp: "<title>Add PR workflow</title>\n<body>## Summary\n- Add gh PR generation</body>",
			want: &PullRequest{
				Title: "Add PR workflow",
				Body:  "## Summary\n- Add gh PR generation",
			},
		},
		{
			name: "succeeds when tracked branch diff is empty but base diff exists",
			git: func() *stubGitClient {
				return &stubGitClient{
					defaultBranch: "origin/main",
					mergeBaseFn: func(base, head string) (string, error) {
						if base != "origin/main" || head != "HEAD" {
							return "", errors.New("unexpected merge-base range")
						}
						return "abc123", nil
					},
					diffFn: func(from, to string) (string, error) {
						if from != "abc123" || to != "HEAD" {
							return "", errors.New("unexpected diff range")
						}
						return "diff content against base", nil
					},
				}
			}(),
			providerResp: "<title>Add PR workflow</title>\n<body>## Summary\n- Add gh PR generation</body>",
			want: &PullRequest{
				Title: "Add PR workflow",
				Body:  "## Summary\n- Add gh PR generation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookPath = func(file string) (string, error) {
				if tt.lookPathErr != nil {
					return "", tt.lookPathErr
				}
				return "/stub/bin/gh", nil
			}

			var stderr bytes.Buffer
			provider := &stubProvider{resp: tt.providerResp, err: tt.providerErr}
			got, err := Run(context.Background(), provider, tt.git, &stderr, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrSubstr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr)) {
					t.Errorf("Run() error = %v, want substring %q", err, tt.wantErrSubstr)
				}
				return
			}

			if got.Title != tt.want.Title {
				t.Errorf("Run() title = %q, want %q", got.Title, tt.want.Title)
			}

			if got.Body != tt.want.Body {
				t.Errorf("Run() body = %q, want %q", got.Body, tt.want.Body)
			}
		})
	}
}

func TestCreatePullRequest(t *testing.T) {
	originalRunGHCommand := runGHCommand
	t.Cleanup(func() {
		runGHCommand = originalRunGHCommand
	})

	tests := []struct {
		name          string
		pr            *PullRequest
		runErr        error
		runOutput     string
		wantErr       bool
		wantErrSubstr string
		wantArgs      []string
		wantOutput    string
	}{
		{
			name: "creates PR with generated title and body",
			pr: &PullRequest{
				Title: "Add gh pr command",
				Body:  "## Summary\n- Add routing and PR creation",
			},
			runOutput:  "https://github.com/example/repo/pull/123",
			wantArgs:   []string{"pr", "create", "--title", "Add gh pr command", "--body", "## Summary\n- Add routing and PR creation"},
			wantOutput: "https://github.com/example/repo/pull/123",
		},
		{
			name:          "fails when PR content is nil",
			pr:            nil,
			wantErr:       true,
			wantErrSubstr: "pull request content is required",
		},
		{
			name: "fails when title is empty",
			pr: &PullRequest{
				Body: "body",
			},
			wantErr:       true,
			wantErrSubstr: "pull request title is required",
		},
		{
			name: "fails when body is empty",
			pr: &PullRequest{
				Title: "title",
			},
			wantErr:       true,
			wantErrSubstr: "pull request body is required",
		},
		{
			name: "returns gh command error",
			pr: &PullRequest{
				Title: "title",
				Body:  "body",
			},
			runErr:        errors.New("gh pr create failed"),
			wantErr:       true,
			wantErrSubstr: "gh pr create failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotArgs []string
			runGHCommand = func(args ...string) (string, error) {
				gotArgs = args
				if tt.runErr != nil {
					return "", tt.runErr
				}
				return tt.runOutput, nil
			}

			output, err := CreatePullRequest(tt.pr)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CreatePullRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.wantErrSubstr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr)) {
					t.Errorf("CreatePullRequest() error = %v, want substring %q", err, tt.wantErrSubstr)
				}
				return
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("CreatePullRequest() args = %v, want %v", gotArgs, tt.wantArgs)
			}

			if output != tt.wantOutput {
				t.Errorf("CreatePullRequest() output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}
