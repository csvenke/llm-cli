package gh

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"llm/internal/loading"
	"llm/internal/providers"
)

//go:embed prompt.md
var systemPrompt string

var (
	lookPath     = lookPathExec
	runGHCommand = func(args ...string) (string, error) {
		cmd := execCommand("gh", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("gh %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
		}

		return strings.TrimSpace(stdout.String()), nil
	}
	runGitCommand = func(args ...string) (string, error) {
		cmd := execCommand("git", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
		}

		return strings.TrimSpace(stdout.String()), nil
	}
)

type PullRequest struct {
	Title string
	Body  string
}

func BuildPrompt(diff, branch, base string) string {
	if diff == "" {
		return ""
	}

	return fmt.Sprintf("Branch: %s\nBase: %s\n\n%s", branch, base, diff)
}

func GeneratePullRequest(ctx context.Context, provider providers.Provider, prompt string, stderr io.Writer) (*PullRequest, error) {
	ind := loading.Start(stderr)
	raw, err := provider.Complete(ctx, systemPrompt, prompt)
	ind.Stop()

	if err != nil {
		return nil, err
	}

	pr, err := parsePullRequest(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing generated pull request content: %w", err)
	}

	return pr, nil
}

func Run(ctx context.Context, provider providers.Provider, git Client, stderr io.Writer, args []string) (*PullRequest, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("usage: llm gh pr")
	}

	if _, err := lookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI is required. Install GitHub CLI from https://cli.github.com/")
	}

	diff, branch, base, err := getBaseDiffContext(git)
	if err != nil {
		return nil, err
	}

	prompt := BuildPrompt(diff, branch, base)
	return GeneratePullRequest(ctx, provider, prompt, stderr)
}

func CreatePullRequest(pr *PullRequest) (string, error) {
	if pr == nil {
		return "", fmt.Errorf("pull request content is required")
	}

	if strings.TrimSpace(pr.Title) == "" {
		return "", fmt.Errorf("pull request title is required")
	}

	if strings.TrimSpace(pr.Body) == "" {
		return "", fmt.Errorf("pull request body is required")
	}

	output, err := runGHCommand("pr", "create", "--title", pr.Title, "--body", pr.Body)
	if err != nil {
		return "", err
	}

	return output, nil
}

func EnsureBranchPushed(git Client) error {
	if _, err := runGitCommand("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); err == nil {
		if _, err := runGitCommand("push"); err != nil {
			return fmt.Errorf("pushing current branch: %w", err)
		}
		return nil
	}

	branch, err := git.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("getting current branch for push: %w", err)
	}

	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("unable to determine current branch for push")
	}

	if _, err := runGitCommand("push", "-u", "origin", branch); err != nil {
		return fmt.Errorf("pushing current branch to origin: %w", err)
	}

	return nil
}

func getBaseDiffContext(git Client) (string, string, string, error) {
	base, mergeBase, err := resolveBaseBranch(git)
	if err != nil {
		return "", "", "", err
	}

	diff, err := git.GetDiffRange(mergeBase, "HEAD")
	if err != nil {
		return "", "", "", fmt.Errorf("getting diff context from merge-base: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		return "", "", "", fmt.Errorf("no branch changes found against base branch %q", base)
	}

	branch, _ := git.GetCurrentBranch()
	if branch == "" {
		branch = "HEAD"
	}

	return diff, branch, base, nil
}

func resolveBaseBranch(git Client) (string, string, error) {
	var candidates []string

	defaultBranch, _ := git.GetDefaultBranch()
	if defaultBranch != "" {
		candidates = append(candidates, defaultBranch)
	}

	fallbacks := []string{"origin/main", "origin/master", "main", "master"}
	for _, fallback := range fallbacks {
		if slices.Contains(candidates, fallback) {
			continue
		}
		candidates = append(candidates, fallback)
	}

	for _, candidate := range candidates {
		mergeBase, err := git.GetMergeBase(candidate, "HEAD")
		if err != nil || mergeBase == "" {
			continue
		}

		return candidate, mergeBase, nil
	}

	if len(candidates) == 0 {
		return "", "", fmt.Errorf("unable to resolve pull request base branch")
	}

	return "", "", fmt.Errorf("unable to resolve pull request base branch from candidates: %s", strings.Join(candidates, ", "))
}

func parsePullRequest(raw string) (*PullRequest, error) {
	re := regexp.MustCompile(`(?s)<title>\s*(.*?)\s*</title>\s*<body>\s*(.*?)\s*</body>`)
	m := re.FindStringSubmatch(raw)
	if len(m) != 3 {
		return nil, fmt.Errorf("expected <title> and <body> tags in provider response")
	}

	title := strings.TrimSpace(m[1])
	body := strings.TrimSpace(m[2])

	if title == "" {
		return nil, fmt.Errorf("generated pull request title is empty")
	}

	if body == "" {
		return nil, fmt.Errorf("generated pull request body is empty")
	}

	return &PullRequest{Title: title, Body: body}, nil
}
